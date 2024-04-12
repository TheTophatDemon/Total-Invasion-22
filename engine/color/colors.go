package color

import "github.com/go-gl/mathgl/mgl32"

type Color struct {
	R, G, B, A float32
}

var (
	Red         = Color{R: 1.0, A: 1.0}
	Green       = Color{G: 1.0, A: 1.0}
	Blue        = Color{B: 1.0, A: 1.0}
	Magenta     = Color{R: 1.0, B: 1.0, A: 1.0}
	White       = Color{R: 1.0, G: 1.0, B: 1.0, A: 1.0}
	Black       = Color{A: 1.0}
	Transparent = Color{}
)

func (c Color) Fade(amount float32) Color {
	return Color{
		R: c.R, G: c.G, B: c.B,
		A: max(c.A-amount, 0.0),
	}
}

func (c Color) WithAlpha(alpha float32) Color {
	return Color{
		R: c.R, G: c.G, B: c.B,
		A: alpha,
	}
}

func (c Color) Vector() mgl32.Vec4 {
	return mgl32.Vec4{c.R, c.G, c.B, c.A}
}
