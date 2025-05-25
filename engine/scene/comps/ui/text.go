package ui

import (
	"fmt"
	"log"
	"strings"
	"text/scanner"
	"unicode"
	"unicode/utf8"

	"github.com/fzipp/bmfont"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/fonts"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

type TextAlign int

const (
	TEXT_ALIGN_LEFT TextAlign = iota
	TEXT_ALIGN_CENTER
	TEXT_ALIGN_RIGHT
)

// Holds text properties that will cause the underlying mesh to regenerate.
type TextSettings struct {
	Text         string
	Alignment    TextAlign
	ShadowColor  color.Color // Color of the drop shadow under the text. Set to transparent to disable.
	ShadowOffset mgl32.Vec2
	WrapWords    bool // Whether text wrapping around the boundary is done by word instead of by character
	Font         *fonts.Font
}

type Text struct {
	Transform
	Hidden       bool
	Color        color.Color
	Settings     TextSettings
	oldSettings  TextSettings
	oldTransform Transform
	mesh         *geom.Mesh
	matrix       mgl32.Mat4
	lineCount    int
}

var _ engine.HasDefault = (*Text)(nil)

func (txt *Text) InitDefault() {
	var zero Text
	*txt = zero
	txt.Color = color.White
	txt.Scale = 1.0
	txt.Settings.WrapWords = true
	txt.Settings.Font = cache.DefaultFont
}

func (txt *Text) SetText(newText string) *Text {
	txt.Settings.Text = newText
	return txt
}

func (txt *Text) Text() string {
	return txt.Settings.Text
}

func (txt *Text) ShadowEnabled() bool {
	return txt.Settings.ShadowColor != color.Transparent
}

func (txt *Text) SetShadow(color color.Color, offset mgl32.Vec2) *Text {
	txt.Settings.ShadowColor = color
	txt.Settings.ShadowOffset = offset
	return txt
}

// Returns the number of lines needed to fit the text into the destination box horizontally.
// Can be used to detect text overflow.
func (txt *Text) LineCount() int {
	txt.Mesh() // Calculates box positions to update line count
	return txt.lineCount
}

// Calculates positions for each character's rectangle
func (txt *Text) generateBoxes() ([]math2.Rect, []bmfont.Char) {
	txt.lineCount = 1
	var originX, originY float32 = 0.0, 0.0
	cursorX, cursorY := originX, originY
	var prevRune rune = scanner.EOF

	boxes := make([]math2.Rect, 0, len(txt.Text()))
	chars := make([]bmfont.Char, 0, cap(boxes))

	var scan scanner.Scanner
	scan.Init(strings.NewReader(txt.Text()))
	// Don't skip spaces or newlines
	scan.Whitespace ^= (1 << '\n') | (1 << ' ')
	scan.Mode = scanner.ScanIdents

	numCharsInLine := 0
	// Applies text alignment to all character boxes in the current line, then starts a new one
	newLine := func() {
		// Set cursor to next line position
		cursorX = originX
		cursorY += float32(txt.Settings.Font.Common.LineHeight)

		if txt.Settings.Alignment != TEXT_ALIGN_LEFT && len(boxes) > 0 {
			// This will be the last box added to this line.
			lastBox := boxes[len(boxes)-1]

			shiftAmount := ((txt.Dest.Width / txt.Scale) - (lastBox.X + lastBox.Width - originX)) // Amount of remaining space within the text's bounds
			if txt.Settings.Alignment == TEXT_ALIGN_CENTER {
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
			txt.lineCount += 1
			continue
		} else if unicode.IsSpace(token) {
			cursorX += 16
			continue
		}

		word := scan.TokenText()

		var firstBox math2.Rect
	restartPoint:
		runeIndex := 0
		for i, runeWidth := 0, 0; i < len(word); i += runeWidth {
			var rn rune
			rn, runeWidth = utf8.DecodeRuneInString(word[i:])
			char, ok := txt.Settings.Font.Chars[rn]
			if !ok {
				// Add blank space for unknown character
				cursorX += 16
				continue
			}

			// Find character position
			charRect := math2.Rect{
				X:      originX + cursorX + float32(char.XOffset),
				Y:      originY + cursorY - float32(txt.Settings.Font.Common.Base+char.YOffset),
				Width:  float32(char.Size().X),
				Height: float32(char.Size().Y),
			}
			if i == 0 {
				firstBox = charRect
			}

			// Stop drawing when out of bounds
			if charRect.Y+charRect.Height >= txt.Dest.Height {
				break
			}

			if i == len(word)-runeWidth || !txt.Settings.WrapWords {
				// On the last character in the word, determine if the word should go on a new line
				overflowsBounds := (charRect.X+charRect.Width >= txt.Dest.Width)
				firstWordOnLine := (firstBox.X == originX)

				if overflowsBounds && !firstWordOnLine {
					// Remove the previous letters in this word
					if txt.Settings.WrapWords {
						boxes = boxes[:len(boxes)-runeIndex]
						chars = chars[:len(chars)-runeIndex]
						numCharsInLine -= runeIndex
					} else {
						boxes = boxes[:len(boxes)-1]
						chars = chars[:len(chars)-1]
						numCharsInLine -= 1
					}
					txt.lineCount += 1
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
				pair := bmfont.CharPair{First: prevRune, Second: rn}
				kerning, ok := txt.Settings.Font.Kerning[pair]
				if ok {
					cursorX += float32(kerning.Amount)
				}
			}

			prevRune = rn
			runeIndex += 1
			numCharsInLine += 1
		}
	}
	// This will apply alignment to text that is made of only one line.
	newLine()

	return boxes, chars
}

// Retrieves the mesh corresponding to the text, regenerating if there have been any changes.
func (txt *Text) Mesh() (*geom.Mesh, error) {
	if txt.Settings.Font == nil {
		return nil, fmt.Errorf("font not assigned")
	}

	if txt.oldSettings != txt.Settings {
		txt.oldSettings = txt.Settings
		if txt.mesh != nil {
			txt.mesh.Free()
		}

		boxes, chars := txt.generateBoxes()
		if boxes == nil || len(boxes) == 0 || chars == nil || len(chars) == 0 {
			return nil, fmt.Errorf("generated empty mesh when rendering text")
		}

		// Regenerate text mesh
		numVertsGuess := len(txt.Text()) * 4
		if txt.ShadowEnabled() {
			numVertsGuess *= 2
		}
		verts := geom.Vertices{
			Pos:      make([]mgl32.Vec3, 0, numVertsGuess),
			TexCoord: make([]mgl32.Vec2, 0, numVertsGuess),
			Color:    make([]mgl32.Vec4, 0, numVertsGuess),
		}

		numIndsGuess := len(txt.Text()) * 2
		if txt.ShadowEnabled() {
			numIndsGuess *= 2
		}
		inds := make([]uint32, 0, numIndsGuess)

		// Generate mesh data for the boxes
		for b, charRect := range boxes {
			indexBase := uint32(len(verts.Pos))

			// Generate mesh data
			verts.Pos = append(verts.Pos,
				mgl32.Vec3{charRect.X, charRect.Y + charRect.Height, 0.0},
				mgl32.Vec3{charRect.X + charRect.Width, charRect.Y + charRect.Height, 0.0},
				mgl32.Vec3{charRect.X, charRect.Y, 0.0},
				mgl32.Vec3{charRect.X + charRect.Width, charRect.Y, 0.0},
			)

			pageW, pageH := float32(txt.Settings.Font.Common.ScaleW), float32(txt.Settings.Font.Common.ScaleH)

			srcRect := math2.Rect{
				X:      float32(chars[b].X+chars[b].Width) / pageW,
				Y:      1.0 - (float32(chars[b].Y) / pageH),
				Width:  float32(chars[b].Width) / pageW,
				Height: float32(chars[b].Height) / pageH,
			}

			verts.TexCoord = append(verts.TexCoord,
				mgl32.Vec2{srcRect.X - srcRect.Width, srcRect.Y - srcRect.Height},
				mgl32.Vec2{srcRect.X, srcRect.Y - srcRect.Height},
				mgl32.Vec2{srcRect.X - srcRect.Width, srcRect.Y},
				mgl32.Vec2{srcRect.X, srcRect.Y},
			)

			verts.Color = append(verts.Color,
				mgl32.Vec4{1.0, 1.0, 1.0, 1.0},
				mgl32.Vec4{1.0, 1.0, 1.0, 1.0},
				mgl32.Vec4{1.0, 1.0, 1.0, 1.0},
				mgl32.Vec4{1.0, 1.0, 1.0, 1.0},
			)

			if txt.ShadowEnabled() {
				// Duplicate the vertices at an offset for shadows
				for range 4 {
					verts.Pos = append(verts.Pos, verts.Pos[len(verts.Pos)-4].Add(txt.Settings.ShadowOffset.Vec3(-0.01)))
					verts.TexCoord = append(verts.TexCoord, verts.TexCoord[len(verts.TexCoord)-4])
					verts.Color = append(verts.Color, txt.Settings.ShadowColor.Vector())
				}
				inds = append(inds, indexBase+4, indexBase+6, indexBase+5, indexBase+6, indexBase+7, indexBase+5)
			}

			inds = append(inds, indexBase+0, indexBase+2, indexBase+1, indexBase+2, indexBase+3, indexBase+1)
		}

		txt.mesh = geom.CreateMesh(verts, inds)
	}
	return txt.mesh, nil
}

func (txt *Text) Matrix() mgl32.Mat4 {
	if txt.Transform != txt.oldTransform {
		txt.oldTransform = txt.Transform
		txt.matrix = mgl32.Translate3D(txt.Dest.X, txt.Dest.Y, txt.Depth).Mul4(mgl32.Scale3D(txt.Scale, txt.Scale, 1.0))
	}
	return txt.matrix
}

func (txt *Text) Render(context *render.Context) {
	if len(txt.Text()) == 0 || txt.Hidden || txt.Settings.Font == nil {
		return
	}

	cache.QuadMesh.Bind()
	shaders.UIShader.Use()
	// Set color
	_ = shaders.UIShader.SetUniformVec4(shaders.UniformDiffuseColor, txt.Color.Vector())
	_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, false)
	_ = shaders.UIShader.SetUniformBool(shaders.UniformFlipHorz, false)

	// Set texture
	cache.GetTexture(txt.Settings.Font.TexturePath()).Bind()
	_ = shaders.UIShader.SetUniformVec4(shaders.UniformSrcRect, mgl32.Vec4{0.0, 1.0, 1.0, 1.0})
	// Set transform
	_ = shaders.UIShader.SetUniformMatrix(shaders.UniformModelMatrix, txt.Matrix())
	// Draw
	if mesh, err := txt.Mesh(); err == nil && mesh != nil {
		mesh.Bind()
		mesh.DrawAll()
	} else if err != nil {
		log.Println(err)
	}

	failure.CheckOpenGLError()
}
