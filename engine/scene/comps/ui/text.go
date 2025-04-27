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
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
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

type Text struct {
	Hidden         bool
	alignment      TextAlign
	color          color.Color
	text           string
	shadowEnabled  bool
	shadowColor    color.Color
	shadowOffset   mgl32.Vec2
	textDirty      bool
	mesh           *geom.Mesh
	texture        *textures.Texture
	font           *fonts.Font
	dest           math2.Rect
	scale, depth   float32
	transform      mgl32.Mat4
	transformDirty bool
	lineCount      int
	wrapWords      bool // Whether text wrapping around the boundary is done by word instead of by character
}

var _ engine.HasDefault = (*Text)(nil)

func (txt *Text) InitDefault() {
	var zero Text
	*txt = zero
	txt.color = color.White
	txt.scale = 1.0
	txt.wrapWords = true
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

func (txt *Text) WrapWords() bool {
	return txt.wrapWords
}

func (txt *Text) SetWrapWords(newValue bool) {
	txt.wrapWords = newValue
	txt.textDirty = true
}

func (txt *Text) SetShadow(color color.Color, offset mgl32.Vec2) *Text {
	txt.textDirty = true
	txt.shadowEnabled = true
	txt.shadowColor = color
	txt.shadowOffset = offset
	return txt
}

func (txt *Text) DisableShadow() {
	txt.shadowEnabled = false
	txt.textDirty = true
}

func (txt *Text) Depth() float32 {
	return txt.depth
}

func (txt *Text) SetDepth(value float32) *Text {
	txt.depth = value
	txt.transformDirty = true
	return txt
}

// Returns the number of lines needed to fit the text into the destination box horizontally.
// Can be used to detect text overflow.
func (txt *Text) LineCount() int {
	_, _ = txt.Mesh() // Calculates box positions to update line count
	return txt.lineCount
}

// Calculates positions for each character's rectangle
func (txt *Text) generateBoxes() ([]math2.Rect, []bmfont.Char) {
	txt.lineCount = 1
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

		if txt.alignment != TEXT_ALIGN_LEFT && len(boxes) > 0 {
			// This will be the last box added to this line.
			lastBox := boxes[len(boxes)-1]

			shiftAmount := ((txt.dest.Width / txt.scale) - (lastBox.X + lastBox.Width - originX)) // Amount of remaining space within the text's bounds
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
			char, ok := txt.font.Chars[rn]
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

			if i == len(word)-runeWidth || !txt.wrapWords {
				// On the last character in the word, determine if the word should go on a new line
				overflowsBounds := (charRect.X+charRect.Width >= txt.dest.Width)
				firstWordOnLine := (firstBox.X == originX)

				if overflowsBounds && !firstWordOnLine {
					// Remove the previous letters in this word
					if txt.wrapWords {
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
				kerning, ok := txt.font.Kerning[pair]
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
	if txt.font == nil {
		return nil, fmt.Errorf("font not assigned")
	}

	if txt.textDirty {
		txt.textDirty = false
		if txt.mesh != nil {
			txt.mesh.Free()
		}

		boxes, chars := txt.generateBoxes()
		if boxes == nil || len(boxes) == 0 || chars == nil || len(chars) == 0 {
			return nil, fmt.Errorf("generated empty mesh when rendering text")
		}

		// Regenerate text mesh
		numVertsGuess := len(txt.text) * 4
		if txt.shadowEnabled {
			numVertsGuess *= 2
		}
		verts := geom.Vertices{
			Pos:      make([]mgl32.Vec3, 0, numVertsGuess),
			TexCoord: make([]mgl32.Vec2, 0, numVertsGuess),
			Color:    make([]mgl32.Vec4, 0, numVertsGuess),
		}

		numIndsGuess := len(txt.text) * 2
		if txt.shadowEnabled {
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

			pageW, pageH := float32(txt.font.Common.ScaleW), float32(txt.font.Common.ScaleH)

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

			if txt.shadowEnabled {
				// Duplicate the vertices at an offset for shadows
				for range 4 {
					verts.Pos = append(verts.Pos, verts.Pos[len(verts.Pos)-4].Add(txt.shadowOffset.Vec3(-0.01)))
					verts.TexCoord = append(verts.TexCoord, verts.TexCoord[len(verts.TexCoord)-4])
					verts.Color = append(verts.Color, txt.shadowColor.Vector())
				}
				inds = append(inds, indexBase+4, indexBase+6, indexBase+5, indexBase+6, indexBase+7, indexBase+5)
			}

			inds = append(inds, indexBase+0, indexBase+2, indexBase+1, indexBase+2, indexBase+3, indexBase+1)
		}

		txt.mesh = geom.CreateMesh(verts, inds)
	}
	return txt.mesh, nil
}

func (txt *Text) SetFont(fontAssetPath string) *Text {
	txt.font, _ = cache.GetFont(fontAssetPath)
	txt.texture = cache.GetTexture(txt.font.TexturePath())
	return txt
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
		txt.transform = mgl32.Translate3D(txt.dest.X, txt.dest.Y, txt.depth).Mul4(mgl32.Scale3D(txt.Scale(), txt.Scale(), 1.0))
	}
	return txt.transform
}

func (txt *Text) Render(context *render.Context) {
	if len(txt.text) == 0 || txt.Hidden {
		return
	}

	failure.CheckOpenGLError()
	cache.QuadMesh.Bind()
	shaders.UIShader.Use()
	// Set color
	_ = shaders.UIShader.SetUniformVec4(shaders.UniformDiffuseColor, txt.Color().Vector())
	failure.CheckOpenGLError()
	_ = shaders.UIShader.SetUniformBool(shaders.UniformNoTexture, false)
	failure.CheckOpenGLError()
	_ = shaders.UIShader.SetUniformBool(shaders.UniformFlipHorz, false)
	failure.CheckOpenGLError()

	// Set texture
	txt.texture.Bind()
	failure.CheckOpenGLError()
	_ = shaders.UIShader.SetUniformVec4(shaders.UniformSrcRect, mgl32.Vec4{0.0, 1.0, 1.0, 1.0})
	failure.CheckOpenGLError()
	// Set transform
	_ = shaders.UIShader.SetUniformMatrix(shaders.UniformModelMatrix, txt.Transform())
	failure.CheckOpenGLError()
	// Draw
	if mesh, err := txt.Mesh(); err == nil && mesh != nil {
		mesh.Bind()
		mesh.DrawAll()
	} else if err != nil {
		log.Println(err)
	}

	failure.CheckOpenGLError()
}
