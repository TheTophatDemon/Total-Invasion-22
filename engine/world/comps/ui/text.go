package ui

import (
	"fmt"
	"path"
	"strings"
	"text/scanner"
	"unicode"
	"unicode/utf8"

	"github.com/fzipp/bmfont"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/fonts"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type TextAlign int

const (
	TEXT_ALIGN_LEFT TextAlign = iota
	TEXT_ALIGN_CENTER
	TEXT_ALIGN_RIGHT
)

type Text struct {
	alignment      TextAlign
	color          color.Color
	text           string
	textDirty      bool
	mesh           *geom.Mesh
	texture        *textures.Texture
	font           *fonts.Font
	dest           math2.Rect
	scale          float32
	transform      mgl32.Mat4
	transformDirty bool
}

func NewText(fontPath, text string) (Text, error) {
	txt := Text{
		alignment:      TEXT_ALIGN_LEFT,
		color:          color.White,
		text:           text,
		textDirty:      true,
		dest:           math2.Rect{},
		scale:          1.0,
		transform:      mgl32.Ident4(),
		transformDirty: true,
	}
	if err := txt.SetFont(fontPath); err != nil {
		return Text{}, err
	}
	return txt, nil
}

func (txt *Text) SetAlignment(align TextAlign) *Text {
	if txt.alignment != align {
		txt.alignment = align
		txt.textDirty = true
	}
	return txt
}

func (txt *Text) GetAlignment() TextAlign {
	return txt.alignment
}

func (txt *Text) SetColor(col color.Color) *Text {
	txt.color = col
	return txt
}

func (txt *Text) Color() color.Color {
	return txt.color
}

func (txt *Text) SetText(newText string) *Text {
	if newText != txt.text {
		txt.text = newText
		txt.textDirty = true
	}
	return txt
}

func (txt *Text) Text() string {
	return txt.text
}

// Calculates positions for each character's rectangle
func (txt *Text) generateBoxes() ([]math2.Rect, []bmfont.Char) {
	var originX, originY float32 = 0.0, 0.0
	cursorX, cursorY := originX, originY
	var prevRune rune = scanner.EOF

	boxes := make([]math2.Rect, 0, len(txt.text))
	chars := make([]bmfont.Char, 0, len(txt.text))

	var scan scanner.Scanner
	scan.Init(strings.NewReader(txt.text))
	// Don't skip spaces or newlines
	scan.Whitespace ^= (1 << '\n') | (1 << ' ')
	scan.Mode = scanner.ScanIdents

	numCharsInLine := 0
	// Applies text alignment to all character boxes in the current line, then starts a new one
	newLine := func() {
		// Set cursor to next line position
		cursorX = originX
		cursorY += float32(txt.font.Common.LineHeight)

		if txt.alignment != TEXT_ALIGN_LEFT {
			lastBox := boxes[len(boxes)-1]

			shiftAmount := (txt.dest.Width - (lastBox.X + lastBox.Width - originX)) // Amount of remaining space within the text's bounds
			if txt.alignment == TEXT_ALIGN_CENTER {
				shiftAmount *= 0.5
			}

			// Shift all characters in the line to the right depending on text alignment
			for i := 0; i < numCharsInLine; i += 1 {
				boxes[len(boxes)-1-i].X += shiftAmount
			}
		}

		numCharsInLine = 0
	}

	for token := scan.Scan(); token != scanner.EOF; token = scan.Scan() {
		if token == '\n' {
			newLine()
			continue
		} else if unicode.IsSpace(token) {
			cursorX += 16
			continue
		}

		word := scan.TokenText()

		var firstBox math2.Rect
	restartPoint:
		runeIndex := 0
		for i, w := 0, 0; i < len(word); i += w {
			var r rune
			r, w = utf8.DecodeRuneInString(word[i:])
			char, ok := txt.font.Chars[r]
			if !ok {
				// Add blank space for unknown character
				cursorX += 16
				continue
			}

			// Find character position
			charRect := math2.Rect{
				X:      originX + cursorX + float32(char.XOffset),
				Y:      originY + cursorY - float32(txt.font.Common.Base+char.YOffset),
				Width:  float32(char.Size().X),
				Height: float32(char.Size().Y),
			}
			if i == 0 {
				firstBox = charRect
			}

			// Stop drawing when out of bounds
			if charRect.Y+charRect.Height >= txt.dest.Height {
				break
			}

			if i == len(word)-w {
				// On the last character in the word, determine if the word should go on a new line
				overflowsBounds := (charRect.X+charRect.Width >= txt.dest.Width)
				firstWordOnLine := (firstBox.X == originX)

				if overflowsBounds && !firstWordOnLine {
					// Remove the previous letters in this word
					boxes = boxes[:len(boxes)-runeIndex]
					chars = chars[:len(chars)-runeIndex]

					numCharsInLine -= runeIndex
					newLine()

					// Restart building the word
					goto restartPoint
				}
			}

			boxes = append(boxes, charRect)
			chars = append(chars, char)

			cursorX += float32(char.XAdvance)

			// Add kerning
			if prevRune != scanner.EOF {
				pair := bmfont.CharPair{First: prevRune, Second: r}
				kerning, ok := txt.font.Kerning[pair]
				if ok {
					cursorX += float32(kerning.Amount)
				}
			}

			prevRune = r
			runeIndex += 1
			numCharsInLine += 1
		}
	}
	newLine()

	return boxes, chars
}

// Retrieves the mesh corresponding to the text, regenerating if there have been any changes.
func (txt *Text) Mesh() (*geom.Mesh, error) {
	if txt.font == nil {
		return nil, fmt.Errorf("font not assigned")
	}

	if txt.textDirty {
		txt.textDirty = false
		if txt.mesh != nil {
			txt.mesh.Free()
		}

		boxes, chars := txt.generateBoxes()

		// Regenerate text mesh
		numVertsGuess := len(txt.text) * 4
		verts := geom.Vertices{
			Pos:      make([]mgl32.Vec3, 0, numVertsGuess),
			TexCoord: make([]mgl32.Vec2, 0, numVertsGuess),
			Color:    make([]mgl32.Vec4, 0, numVertsGuess),
		}

		inds := make([]uint32, 0, len(txt.text)*2)

		// Generate mesh data for the boxes
		for b, charRect := range boxes {
			indexBase := uint32(len(verts.Pos))

			// Generate mesh data
			verts.Pos = append(verts.Pos,
				mgl32.Vec3{charRect.X, charRect.Y, 0.0},
				mgl32.Vec3{charRect.X + charRect.Width, charRect.Y, 0.0},
				mgl32.Vec3{charRect.X + charRect.Width, charRect.Y + charRect.Height, 0.0},
				mgl32.Vec3{charRect.X, charRect.Y + charRect.Height, 0.0},
			)

			pageW, pageH := float32(txt.font.Common.ScaleW), float32(txt.font.Common.ScaleH)

			srcRect := math2.Rect{
				X:      1.0 - (float32(chars[b].X+chars[b].Width) / pageW),
				Y:      float32(chars[b].Y) / pageH,
				Width:  float32(chars[b].Width) / pageW,
				Height: float32(chars[b].Height) / pageH,
			}

			verts.TexCoord = append(verts.TexCoord,
				mgl32.Vec2{srcRect.X + srcRect.Width, srcRect.Y},
				mgl32.Vec2{srcRect.X, srcRect.Y},
				mgl32.Vec2{srcRect.X, srcRect.Y + srcRect.Height},
				mgl32.Vec2{srcRect.X + srcRect.Width, srcRect.Y + srcRect.Height},
			)

			verts.Color = append(verts.Color,
				mgl32.Vec4{1.0, 1.0, 1.0, 1.0},
				mgl32.Vec4{1.0, 1.0, 1.0, 1.0},
				mgl32.Vec4{1.0, 1.0, 1.0, 1.0},
				mgl32.Vec4{1.0, 1.0, 1.0, 1.0},
			)

			inds = append(inds, indexBase+0, indexBase+1, indexBase+2, indexBase+0, indexBase+2, indexBase+3)
		}

		txt.mesh = geom.CreateMesh(verts, inds)
	}
	return txt.mesh, nil
}

func (txt *Text) SetFont(fontAssetPath string) error {
	var err error
	txt.font, err = cache.GetFont(fontAssetPath)
	if err != nil {
		return err
	}

	// Set the texture to the font's page
	texturePath := path.Join(path.Dir(fontAssetPath), txt.font.Pages[0].File)
	if !strings.HasSuffix(texturePath, ".png") {
		return fmt.Errorf("font has invalid texture file type")
	}
	txt.texture = cache.GetTexture(texturePath)

	return nil
}

func (txt *Text) SetDest(dest math2.Rect) *Text {
	txt.dest = dest
	txt.transformDirty = true
	return txt
}

func (txt *Text) Dest() math2.Rect {
	return txt.dest
}

func (txt *Text) SetScale(scale float32) *Text {
	txt.scale = scale
	txt.transformDirty = true
	return txt
}

func (txt *Text) Scale() float32 {
	return txt.scale
}

func (txt *Text) Transform() mgl32.Mat4 {
	if txt.transformDirty {
		txt.transformDirty = false
		txt.transform = mgl32.Translate3D(txt.dest.X, txt.dest.Y, 0.0).Mul4(mgl32.Scale3D(txt.scale, txt.scale, 1.0))
	}
	return txt.transform
}
