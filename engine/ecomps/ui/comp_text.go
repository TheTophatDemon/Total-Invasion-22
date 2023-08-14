package ui

import (
	"fmt"
	"image/color"
	"path"
	"strings"

	"github.com/fzipp/bmfont"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Text struct {
	color          color.Color
	text           string
	textDirty      bool
	mesh           *assets.Mesh
	texture        *assets.Texture
	font           *assets.Font
	dest           math2.Rect
	scale          float32
	transform      mgl32.Mat4
	transformDirty bool
}

func NewText(fontPath, text string) (*Text, error) {
	txt := &Text{
		color:          color.White,
		text:           text,
		textDirty:      true,
		dest:           math2.Rect{},
		scale:          1.0,
		transform:      mgl32.Ident4(),
		transformDirty: true,
	}
	if err := txt.SetFont(fontPath); err != nil {
		return nil, err
	}
	return txt, nil
}

func (txt *Text) SetColor(col color.Color) *Text {
	txt.color = col
	return txt
}

func (txt *Text) Color() color.Color {
	return txt.color
}

func (txt *Text) SetText(newText string) *Text {
	txt.text = newText
	txt.textDirty = true
	return txt
}

func (txt *Text) Text() string {
	return txt.text
}

// Retrieves the mesh corresponding to the text, regenerating if there have been any changes.
func (txt *Text) Mesh() (*assets.Mesh, error) {
	if txt.font == nil {
		return nil, fmt.Errorf("font not assigned")
	}

	if txt.textDirty {
		txt.textDirty = false
		if txt.mesh != nil {
			txt.mesh.Free()
		}

		// Regenerate text mesh
		// TODO: Text wrapping

		numVertsGuess := len(txt.text) * 4
		verts := assets.Vertices{
			Pos:      make([]mgl32.Vec3, 0, numVertsGuess),
			TexCoord: make([]mgl32.Vec2, 0, numVertsGuess),
			Color:    make([]mgl32.Vec4, 0, numVertsGuess),
		}

		inds := make([]uint32, 0, len(txt.text)*2)

		var originX, originY float32 = 0.0, 0.0
		cursorX, cursorY := originX, originY
		var prevRune rune
		for i, r := range txt.text {
			// Handle newline
			if r == '\n' {
				cursorX = originX
				cursorY += float32(txt.font.Common.LineHeight)
				continue
			}

			char, ok := txt.font.Chars[r]
			if !ok {
				continue
			}

			// Find character position
			charRect := math2.Rect{
				X:      originX + cursorX + float32(char.XOffset),
				Y:      originY + cursorY - float32(txt.font.Common.Base+char.YOffset),
				Width:  float32(char.Size().X),
				Height: float32(char.Size().Y),
			}

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
				X:      1.0 - (float32(char.X+char.Width) / pageW),
				Y:      float32(char.Y) / pageH,
				Width:  float32(char.Width) / pageW,
				Height: float32(char.Height) / pageH,
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

			cursorX += float32(char.XAdvance)

			// Add kerning
			if i > 0 {
				pair := bmfont.CharPair{First: prevRune, Second: r}
				kerning, ok := txt.font.Kerning[pair]
				if ok {
					cursorX += float32(kerning.Amount)
				}
			}

			prevRune = r
		}

		txt.mesh = assets.CreateMesh(verts, inds)
	}
	return txt.mesh, nil
}

func (txt *Text) SetFont(fontAssetPath string) error {
	var err error
	txt.font, err = assets.GetFont(fontAssetPath)
	if err != nil {
		return err
	}

	// Set the texture to the font's page
	texturePath := path.Join(path.Dir(fontAssetPath), txt.font.Pages[0].File)
	if !strings.HasSuffix(texturePath, ".png") {
		return fmt.Errorf("font has invalid texture file type")
	}
	txt.texture = assets.GetTexture(texturePath)

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
