package math2 //Get ready for MATH 2: Revenge of the Quaternions, coming to theatres this Pi Day.

import (
	"image/color"
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

const HALF_PI = 3.14159 / 2.0

type (
	Number interface {
		int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
	}

	Float interface {
		float32 | float64
	}

	Rect struct {
		X, Y, Width, Height float32
	}

	Triangle [3]mgl32.Vec3

	Plane struct {
		Normal mgl32.Vec3
		Dist   float32
	}

	Box struct {
		Min, Max mgl32.Vec3
	}
)

// Generate a rectangle that wraps around all of the given points (there must be at least 2).
func RectFromPoints(point0, point1 mgl32.Vec2, points ...mgl32.Vec2) Rect {
	minX, minY := min(point0.X(), point1.X()), min(point0.Y(), point1.Y())
	maxX, maxY := max(point0.X(), point1.X()), max(point0.Y(), point1.Y())
	for _, p := range points {
		minX = min(minX, p.X())
		minY = min(minY, p.Y())
		maxX = max(maxX, p.X())
		maxY = max(maxY, p.Y())
	}
	return Rect{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}

func (r Rect) Center() (float32, float32) {
	return (r.X + r.Width/2.0), (r.Y + r.Height/2.0)
}

func (r Rect) Vec4() mgl32.Vec4 {
	return mgl32.Vec4{r.X, r.Y, r.Width, r.Height}
}

func (t Triangle) Plane() Plane {
	normal := t[1].Sub(t[0]).Cross(t[2].Sub(t[0])).Normalize()
	return Plane{
		Normal: normal,
		Dist:   -normal.Dot(t[0]),
	}
}

func BoxFromExtents(halfWidth, halfHeight, halfLength float32) Box {
	return Box{
		Max: mgl32.Vec3{halfWidth, halfHeight, halfLength},
		Min: mgl32.Vec3{-halfWidth, -halfHeight, -halfLength},
	}
}

func BoxFromRadius(radius float32) Box {
	return Box{
		Max: mgl32.Vec3{radius, radius, radius},
		Min: mgl32.Vec3{-radius, -radius, -radius},
	}
}

// Returns true if two boxes intersect.
// Each box is described by an origin position and a half-size vector.
func (box Box) Intersects(other Box) bool {
	return box.Max[0] > other.Min[0] && box.Max[1] > other.Min[1] && box.Max[2] > other.Min[2] &&
		other.Max[0] > box.Min[0] && other.Max[1] > box.Min[1] && other.Max[2] > box.Min[2]
}

func (box Box) Translate(offset mgl32.Vec3) Box {
	return Box{
		Min: box.Min.Add(offset),
		Max: box.Max.Add(offset),
	}
}

func (box Box) Size() mgl32.Vec3 {
	return box.Max.Sub(box.Min)
}

func (box Box) Center() mgl32.Vec3 {
	return box.Min.Add(box.Size().Mul(0.5))
}

func ClosestPointOnLine(lineStart, lineEnd, point mgl32.Vec3) mgl32.Vec3 {
	lineDir := lineEnd.Sub(lineStart)
	t := point.Sub(lineStart).Dot(lineDir) / lineDir.Dot(lineDir)
	return lineStart.Add(lineDir.Mul(Clamp(t, 0.0, 1.0)))
}

func Clamp[N Number](val, min, max N) N {
	if val < min {
		return min
	} else if val > max {
		return max
	}
	return val
}

func Abs[N Number](val N) N {
	if val > 0 {
		return val
	} else {
		return -val
	}
}

func Cos[F Float](val F) F {
	return F(math.Cos(float64(val)))
}

func Sin[F Float](val F) F {
	return F(math.Sin(float64(val)))
}

func Asin[F Float](val F) F {
	return F(math.Asin(float64(val)))
}

func Atan2[F Float](y F, x F) F {
	return F(math.Atan2(float64(y), float64(x)))
}

func CopySign[F Float](mag F, sign F) F {
	return F(math.Copysign(float64(mag), float64(sign)))
}

func Pow[F Float](base F, exp F) F {
	return F(math.Pow(float64(base), float64(exp)))
}

func Sqrt[F Float](x F) F {
	return F(math.Sqrt(float64(x)))
}

func Vec3Up() mgl32.Vec3 {
	return mgl32.Vec3{0.0, 1.0, 0.0}
}

func Vec3Zero() mgl32.Vec3 {
	return mgl32.Vec3{0.0, 0.0, 0.0}
}

func Vec3One() mgl32.Vec3 {
	return mgl32.Vec3{1.0, 1.0, 1.0}
}

// Returns a vector with the element-wise minimum value on each axis.
func Vec3Min(a, b mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		min(a[0], b[0]),
		min(a[1], b[1]),
		min(a[2], b[2]),
	}
}

// Returns a vector with the element-wise maximum value on each axis.
func Vec3Max(a, b mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		max(a[0], b[0]),
		max(a[1], b[1]),
		max(a[2], b[2]),
	}
}

func ColorToVec4(col color.Color) mgl32.Vec4 {
	r, g, b, a := col.RGBA()
	return mgl32.Vec4{
		float32(r) / float32(0xffff),
		float32(g) / float32(0xffff),
		float32(b) / float32(0xffff),
		float32(a) / float32(0xffff),
	}
}

func QuatToEulerAngles(q mgl32.Quat) mgl32.Vec3 {
	//https://en.wikipedia.org/wiki/Conversion_between_quaternions_and_Euler_angles
	sinr_cosp := 2.0 * (q.W*q.X() + q.Y()*q.Z())
	cosr_cosp := 1.0 - 2.0*(q.X()*q.X()+q.Y()*q.Y())
	roll := Atan2(sinr_cosp, cosr_cosp)

	sinp := 2.0 * (q.W*q.Y() - q.Z()*q.X())
	var pitch float32
	if Abs(sinp) >= 1.0 {
		pitch = CopySign(math.Pi/2.0, sinp)
	} else {
		pitch = Asin(sinp)
	}

	siny_cosp := 2.0 * (q.W*q.Z() + q.X()*q.Y())
	cosy_cosp := 1.0 - 2.0*(q.Y()*q.Y()+q.Z()*q.Z())
	yaw := Atan2(siny_cosp, cosy_cosp)

	return mgl32.Vec3{pitch, yaw, roll}
}

// Returns the pitch, yaw, and roll of a Mat4 as a vector of Euler angles (in radians).
func Mat4EulerAngles(m *mgl32.Mat4) mgl32.Vec3 {
	//Referencing http://eecs.qmul.ac.uk/~gslabaugh/publications/euler.pdf
	var theta, psi, fi float64

	//Change the handedness of the matrix
	matx := mgl32.Mat4FromCols(
		m.Col(0), m.Col(1), m.Col(2).Mul(-1.0), m.Col(3))

	r33 := float64(matx.At(2, 2))
	r21 := float64(matx.At(1, 0))
	r22 := float64(matx.At(1, 1))
	r23 := float64(matx.At(1, 2))
	r11 := float64(matx.At(0, 0))
	r12 := float64(matx.At(0, 1))
	r13 := float64(matx.At(0, 2))

	if Abs(r23) != 1.0 {
		psi = -math.Asin(r23)
		cosPsi := math.Cos(psi)
		theta = math.Atan2(r13/cosPsi, r33/cosPsi)
		fi = math.Atan2(r21/cosPsi, r22/cosPsi)
	} else {
		fi = 0.0
		if r23 == -1 {
			psi = math.Pi / 2.0
			theta = fi + math.Atan2(r12, r11)
		} else {
			psi = -math.Pi / 2.0
			theta = -fi + math.Atan2(-r12, -r11)
		}
	}

	return mgl32.Vec3{float32(psi), float32(theta), float32(fi)}
}

func LookAtV(eye, center, up mgl32.Vec3) mgl32.Mat4 {
	f := center.Sub(eye).Normalize()
	s := f.Cross(up.Normalize()).Normalize()
	u := s.Cross(f)

	return mgl32.Mat4{
		s[0], s[1], s[2], 0,
		u[0], u[1], u[2], 0,
		-f[0], -f[1], -f[2], 0,
		eye[0], eye[1], eye[2], 1,
	}
}
