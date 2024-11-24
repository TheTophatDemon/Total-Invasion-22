// Get ready for MATH 2: Revenge of the Quaternions, coming to theatres this Pi Day.
package math2

import (
	"math"
	"math/rand"

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
)

func Clamp[N Number](val, min, max N) N {
	if val < min {
		return min
	} else if val > max {
		return max
	}
	return val
}

// Returns 1 if the number is positive and -1 if it is negative.
func Signum[N Number](val N) N {
	if val == 0 {
		return 0
	}
	return Abs(val) / val
}

// Calculates the sign using a fast sin approximation accurate from 0 to 1
func FastApproxSin[F Float](val F) F {
	return 0.5 * val * (3.0 - val*val)
}

func Vec3Up() mgl32.Vec3 {
	return mgl32.Vec3{0.0, 1.0, 0.0}
}

func Vec3Down() mgl32.Vec3 {
	return mgl32.Vec3{0.0, -1.0, 0.0}
}

func Vec3Zero() mgl32.Vec3 {
	return mgl32.Vec3{0.0, 0.0, 0.0}
}

func Vec3One() mgl32.Vec3 {
	return mgl32.Vec3{1.0, 1.0, 1.0}
}

func Vec3Forward() mgl32.Vec3 {
	return mgl32.Vec3{0.0, 0.0, -1.0}
}

func Vec3Right() mgl32.Vec3 {
	return mgl32.Vec3{1.0, 0.0, 0.0}
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

func Vec3WithY(vec mgl32.Vec3, newY float32) mgl32.Vec3 {
	return mgl32.Vec3{vec.X(), newY, vec.Z()}
}

func Vec3Reflect(direction, normal mgl32.Vec3) mgl32.Vec3 {
	return direction.Sub(normal.Mul(2 * direction.Dot(normal)))
}

// Returns a unit vector pointing in a random direction.
func RandomDir() mgl32.Vec3 {
	dir := mgl32.Vec3{
		(rand.Float32() - 0.5) * 2.0,
		(rand.Float32() - 0.5) * 2.0,
		(rand.Float32() - 0.5) * 2.0,
	}
	if dir.LenSqr() == 0.0 {
		return mgl32.Vec3{1.0, 0.0, 0.0}
	}
	return dir.Normalize()
}

func DegToRadVec3(degrees mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		mgl32.DegToRad(degrees[0]),
		mgl32.DegToRad(degrees[1]),
		mgl32.DegToRad(degrees[2]),
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

func Lerp[F Float](min, max, ratio F) F {
	if ratio >= 1.0 {
		return max
	}
	if ratio <= 0.0 {
		return min
	}
	return min + (max-min)*ratio
}
