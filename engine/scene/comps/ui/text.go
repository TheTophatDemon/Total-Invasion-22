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
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

var GlobalTextScale float32 = 1.0

type TextConfig struct {
	Align         TextAlign
	Color         color.Color
	DisableShadow bool
	WrapWords     bool
	Font          *fonts.Font
	Scale         float32
	Padding       mgl32.Vec2
}

func DefaultTextConfig() TextConfig {
	return TextConfig{
		Align:         TextAlignTopLeft,
		Color:         color.White,
		DisableShadow: false,
		WrapWords:     true,
		Font:          cache.DefaultFont,
		Scale:         1,
		Padding:       mgl32.Vec2{},
	}
}

func (config TextConfig) SetAlign(value TextAlign) TextConfig {
	config.Align = value
	return config
}
func (config TextConfig) SetColor(value color.Color) TextConfig {
	config.Color = value
	return config
}
func (config TextConfig) SetDisableShadow(value bool) TextConfig {
	config.DisableShadow = value
	return config
}
func (config TextConfig) SetWrapWords(value bool) TextConfig {
	config.WrapWords = value
	return config
}
func (config TextConfig) SetFont(value *fonts.Font) TextConfig {
	config.Font = value
	return config
}
func (config TextConfig) SetScale(value float32) TextConfig {
	config.Scale = value
	return config
}
func (config TextConfig) SetPadding(value mgl32.Vec2) TextConfig {
	config.Padding = value
	return config
}

func (el *Element) IsText() bool {
	return el.textConfig.Font != nil
}

func (el *Element) SetText(text string) {
	if !el.IsText() {
		el.textConfig = DefaultTextConfig()
	}
	if el.text != text {
		el.text = text
		el.textClean = false
	}
}

func (el *Element) SetTextWith(text string, config TextConfig) {
	el.textConfig = config
	el.text = text
	el.textClean = false
}

func (el *Element) Text() string {
	return el.text
}

func (el *Element) TextAlign() TextAlign {
	return el.textConfig.Align
}

func (el *Element) TextColor() color.Color {
	return el.textConfig.Color
}

func (el *Element) TextDisableShadow() bool {
	return el.textConfig.DisableShadow
}

func (el *Element) TextWrapWords() bool {
	return el.textConfig.WrapWords
}

func (el *Element) TextFont() *fonts.Font {
	return el.textConfig.Font
}

func (el *Element) TextScale() float32 {
	return el.textConfig.Scale
}

func (el *Element) TextPadding() mgl32.Vec2 {
	return el.textConfig.Padding
}

func (el *Element) SetTextAlign(value TextAlign) *Element {
	if value != el.textConfig.Align {
		el.textConfig.Align = value
		el.textClean = false
	}
	return el
}

func (el *Element) SetTextColor(value color.Color) *Element {
	if value != el.textConfig.Color {
		el.textConfig.Color = value
		el.textClean = false
	}
	return el
}

func (el *Element) SetTextDisableShadow(value bool) *Element {
	if value != el.textConfig.DisableShadow {
		el.textConfig.DisableShadow = value
		el.textClean = false
	}
	return el
}

func (el *Element) SetTextWrapWords(value bool) *Element {
	if value != el.textConfig.WrapWords {
		el.textConfig.WrapWords = value
		el.textClean = false
	}
	return el
}

func (el *Element) SetTextFont(value *fonts.Font) *Element {
	if value != el.textConfig.Font {
		el.textConfig.Font = value
		el.textClean = false
	}
	return el
}

func (el *Element) SetTextScale(value float32) *Element {
	if value != el.textConfig.Scale {
		el.textConfig.Scale = value
		el.textClean = false
	}
	return el
}

func (el *Element) SetTextPadding(value mgl32.Vec2) *Element {
	if value != el.textConfig.Padding {
		el.textConfig.Padding = value
		el.textClean = false
	}
	return el
}

func NewText(transform Transform, text string, config TextConfig) Element {
	elem := Element{
		transform:  transform,
		textConfig: config,
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

	if txt.text == "" || txt.TextFont() == nil {
		return boxes, chars
	}

	font := txt.TextFont()
	if font == nil {
		font = cache.DefaultFont
	}

	scale := txt.TextScale() * GlobalTextScale
	availableSpace := txt.Size().Sub(txt.TextPadding().Mul(2.0))

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
		cursorY += float32(font.Common.LineHeight) * scale

		if (txt.TextAlign()&(TextAlignCenterH|TextAlignRight)) != 0 && len(boxes) > 0 {
			// This will be the last box added to this line.
			lastBox := boxes[len(boxes)-1]

			shiftAmount := (availableSpace[0] - (lastBox.X + lastBox.Width)) // Amount of remaining space within the text's bounds
			if (txt.TextAlign() & TextAlignCenterH) != 0 {
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
			cursorX += 16 * scale
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
				cursorX += 16 * scale
				continue
			}

			// Find character position
			charRect := math2.Rect{
				X:      cursorX + float32(char.XOffset)*scale,
				Y:      cursorY - float32(font.Common.Base+char.YOffset),
				Width:  float32(char.Size().X) * scale,
				Height: float32(char.Size().Y) * scale,
			}
			if i == 0 {
				firstBox = charRect
			}

			// Stop drawing when out of bounds
			if charRect.Y+charRect.Height > availableSpace[1] {
				break
			}

			if i == len(word)-runeWidth || !txt.TextWrapWords() {
				// Determine if the word should go on a new line
				overflowsBounds := (charRect.X+charRect.Width > availableSpace[0])
				firstWordOnLine := (firstBox.X == 0.0)

				if overflowsBounds && !firstWordOnLine {
					// Remove the previous letters in this word
					if txt.TextWrapWords() {
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

			cursorX += float32(char.XAdvance) * scale

			// Add kerning
			if prevRune != scanner.EOF {
				pair := bmfont.CharPair{First: prevRune, Second: rn}
				kerning, ok := font.Kerning[pair]
				if ok {
					cursorX += float32(kerning.Amount) * scale
				}
			}

			prevRune = rn
			runeIndex += 1
			numCharsInLine += 1
		}
	}
	// This will apply alignment to text that is made of only one line.
	newLine()

	if len(boxes) > 0 {
		lastBox := boxes[len(boxes)-1]
		charsHeight := lastBox.Height + lastBox.Y
		// Apply vertical alignment
		if (txt.TextAlign()&TextAlignCenterV) != 0 && charsHeight < availableSpace[1] {
			shift := (availableSpace[1] - charsHeight) / 2.0
			for i := range boxes {
				boxes[i].Y += shift
			}
		} else if (txt.TextAlign() & TextAlignBottom) != 0 {
			shift := availableSpace[1] - charsHeight
			for i := range boxes {
				boxes[i].Y += shift
			}
		}
		// Apply offset due to padding
		if txt.TextPadding() != (mgl32.Vec2{}) {
			for i := range boxes {
				boxes[i].X += txt.TextPadding()[0]
				boxes[i].Y += txt.TextPadding()[1]
			}
		}
	}

	return boxes, chars
}

// Retrieves the mesh corresponding to the text, regenerating if there have been any changes.
// Returns false and logs error if there is a failure.
func (txt *Element) generateTextMesh() (*geom.Mesh, bool) {
	if txt.TextFont() == nil {
		return nil, false
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

		pageW, pageH := float32(txt.TextFont().Common.ScaleW), float32(txt.TextFont().Common.ScaleH)

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
	if el.TextMesh() == nil || !el.IsText() {
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
