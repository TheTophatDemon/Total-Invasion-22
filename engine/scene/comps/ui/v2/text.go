package ui

import (
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
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type TextConfig struct {
	Align         TextAlign
	Color         maybe.T[color.Color]
	DisableShadow bool
	WrapWords     bool
	Font          *fonts.Font
}

func NewText(transform Transform, text string, config TextConfig) Element {
	elem := Element{
		transform:  transform,
		textConfig: maybe.Some(config),
	}
	elem.SetText(text)
	return elem
}

// Calculates positions for each character's rectangle
func (txt *Element) generateTextBoxes() ([]math2.Rect, []bmfont.Char) {
	var cursorX, cursorY float32
	var prevRune rune = scanner.EOF

	boxes := make([]math2.Rect, 0, len(txt.text))
	chars := make([]bmfont.Char, 0, cap(boxes))

	config, hasConfig := txt.textConfig.Get()
	if txt.text == "" || !hasConfig {
		return boxes, chars
	}

	font := config.Font
	if font == nil {
		font = cache.DefaultFont
	}

	var scan scanner.Scanner
	scan.Init(strings.NewReader(txt.text))
	// Don't skip spaces or newlines
	scan.Whitespace ^= (1 << '\n') | (1 << ' ')
	scan.Mode = scanner.ScanIdents

	numCharsInLine := 0
	// Applies text alignment to all character boxes in the current line, then starts a new one
	newLine := func() {
		// Set cursor to next line position
		cursorX = 0.0
		cursorY += float32(font.Common.LineHeight)

		if (config.Align&(TextAlignCenterH|TextAlignRight)) != 0 && len(boxes) > 0 {
			// This will be the last box added to this line.
			lastBox := boxes[len(boxes)-1]

			shiftAmount := (txt.Size()[0] - (lastBox.X + lastBox.Width)) // Amount of remaining space within the text's bounds
			if (config.Align & TextAlignCenterH) != 0 {
				shiftAmount *= 0.5
			}

			// Shift all characters in the line to the right depending on text alignment
			for i := range numCharsInLine {
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
		for i, runeWidth := 0, 0; i < len(word); i += runeWidth {
			var rn rune
			rn, runeWidth = utf8.DecodeRuneInString(word[i:])
			char, ok := font.Chars[rn]

			if !ok {
				// Add blank space for unknown character
				cursorX += 16
				continue
			}

			// Find character position
			charRect := math2.Rect{
				X:      cursorX + float32(char.XOffset),
				Y:      cursorY - float32(font.Common.Base+char.YOffset),
				Width:  float32(char.Size().X),
				Height: float32(char.Size().Y),
			}
			if i == 0 {
				firstBox = charRect
			}

			// Stop drawing when out of bounds
			if charRect.Y+charRect.Height > txt.Size()[1] {
				break
			}

			if i == len(word)-runeWidth || !config.WrapWords {
				// Determine if the word should go on a new line
				overflowsBounds := (charRect.X+charRect.Width > txt.Size()[0])
				firstWordOnLine := (firstBox.X == 0.0)

				if overflowsBounds && !firstWordOnLine {
					// Remove the previous letters in this word
					if config.WrapWords {
						boxes = boxes[:len(boxes)-runeIndex]
						chars = chars[:len(chars)-runeIndex]
						numCharsInLine -= runeIndex
					} else {
						word = word[i:]
					}
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
				kerning, ok := font.Kerning[pair]
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
	// `cursorY`` now defines the overall height of the text.

	// Apply vertical alignment
	if (config.Align&TextAlignCenterV) != 0 && cursorY < txt.Size()[1] {
		shift := txt.Size()[1] - (cursorY / 2.0)
		for i := range boxes {
			boxes[i].Y += shift
		}
	}

	return boxes, chars
}

// Retrieves the mesh corresponding to the text, regenerating if there have been any changes.
// Returns false and logs error if there is a failure.
func (txt *Element) generateTextMesh() (*geom.Mesh, bool) {
	config, hasConfig := txt.textConfig.Get()
	if !hasConfig {
		return nil, false
	}
	if config.Font == nil {
		config.Font = cache.DefaultFont
	}

	if txt.textMesh != nil {
		txt.textMesh.Free()
	}

	boxes, chars := txt.generateTextBoxes()
	if len(boxes) != len(chars) {
		failure.LogErrWithLocation("boxes(%v) and chars(%v) arrays mismatch for \"%v\"", len(boxes), len(chars), txt.text)
		return nil, false
	} else if len(boxes) == 0 && txt.text != "" && txt.Size() != (mgl32.Vec2{}) {
		failure.LogWarningWithLocation("non-empty text element could not generate any vertices")
	}

	// Regenerate text mesh
	numVertsGuess := len(txt.text) * 4
	verts := geom.Vertices{
		Pos:      make([]mgl32.Vec3, 0, numVertsGuess),
		TexCoord: make([]mgl32.Vec2, 0, numVertsGuess),
		Color:    make([]mgl32.Vec4, 0, numVertsGuess),
	}

	numIndsGuess := len(txt.text) * 2
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

		pageW, pageH := float32(config.Font.Common.ScaleW), float32(config.Font.Common.ScaleH)

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

		inds = append(inds, indexBase+0, indexBase+2, indexBase+1, indexBase+2, indexBase+3, indexBase+1)
	}

	txt.textMesh = geom.CreateMesh(verts, inds)
	return txt.textMesh, true
}

func (el *Element) FitText() {
	if el.TextMesh() == nil || !el.HasText() {
		return
	}
	// Expand the box first to fit all of the text in one line
	scrW, scrH := engine.ScreenSize()
	el.SetSize(mgl32.Vec2{
		float32(len(el.Text()) * scrW),
		float32(len(el.Text()) * scrH),
	})
	el.SetSize(el.TextMesh().BoundingBox().Size().Vec2())
}
