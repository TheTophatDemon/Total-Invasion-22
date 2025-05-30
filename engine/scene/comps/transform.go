package comps

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Transform struct {
	pos, rot, scale mgl32.Vec3
	matrix          mgl32.Mat4
	upToDate        bool // True if the matrix currently matches the position, rotation, and scale.
}

func TransformFromMatrix(matrix mgl32.Mat4) *Transform {
	return &Transform{
		pos:      matrix.Col(3).Vec3(),
		rot:      math2.Mat4EulerAngles(&matrix),
		scale:    mgl32.Vec3{matrix[0], matrix[5], matrix[10]},
		matrix:   matrix,
		upToDate: true,
	}
}

func TransformFromTranslation(position mgl32.Vec3) Transform {
	return Transform{
		pos:      position,
		rot:      math2.Vec3Zero(),
		scale:    math2.Vec3One(),
		upToDate: false,
	}
}

func TransformFromTranslationAngles(position mgl32.Vec3, angles mgl32.Vec3) Transform {
	return Transform{
		pos:      position,
		rot:      angles,
		scale:    math2.Vec3One(),
		upToDate: false,
	}
}

func TransformFromTranslationAnglesScale(position, angles, scale mgl32.Vec3) Transform {
	return Transform{
		pos:      position,
		rot:      angles,
		scale:    scale,
		upToDate: false,
	}
}

func TransformFromTE3Ent(ent te3.Ent, scaleByRadius, stayOnFloor bool) Transform {
	angles := ent.AnglesInRadians()
	if scaleByRadius {
		pos := mgl32.Vec3(ent.Position)
		if stayOnFloor {
			pos = pos.Add(mgl32.Vec3{0.0, ent.Radius - 1.0, 0.0})
		}
		return TransformFromTranslationAnglesScale(
			pos, angles, mgl32.Vec3{ent.Radius, ent.Radius, ent.Radius},
		)
	} else {
		return TransformFromTranslationAngles(
			ent.Position, angles,
		)
	}
}

func (t *Transform) SetPosition(newPos mgl32.Vec3) {
	t.pos = newPos
	t.upToDate = false
}

func (t *Transform) Position() mgl32.Vec3 {
	return t.pos
}

func (t *Transform) SetScale(x, y, z float32) {
	t.SetScaleV(mgl32.Vec3{x, y, z})
}

func (t *Transform) SetScaleUniform(scale float32) {
	t.SetScale(scale, scale, scale)
}

func (t *Transform) SetScaleV(scale mgl32.Vec3) {
	t.scale = scale
	t.upToDate = false
}

func (t *Transform) Scale() mgl32.Vec3 {
	return t.scale
}

func (t *Transform) Translate(x, y, z float32) {
	t.pos[0] += x
	t.pos[1] += y
	t.pos[2] += z
	t.upToDate = false
}

func (t *Transform) TranslateV(offset mgl32.Vec3) {
	t.Translate(offset[0], offset[1], offset[2])
}

// Sets the rotation in euler angles (radians)
func (t *Transform) SetRotation(pitch, yaw, roll float32) {
	t.rot = mgl32.Vec3{pitch, yaw, roll}
	t.upToDate = false
}

func (t *Transform) SetRotationV(v mgl32.Vec3) {
	t.rot = v
	t.upToDate = false
}

// Returns the euler angles of rotation (pitch, yaw, roll)
func (t *Transform) Rotation() mgl32.Vec3 {
	return t.rot
}

func (t *Transform) Yaw() float32 {
	return t.rot[1]
}

func (t *Transform) Pitch() float32 {
	return t.rot[0]
}

func (t *Transform) Roll() float32 {
	return t.rot[2]
}

func (t *Transform) Forward() mgl32.Vec3 {
	fwd := mgl32.TransformNormal(math2.Vec3Forward(), t.Matrix())
	if fwd.LenSqr() == 0.0 {
		return mgl32.Vec3{}
	}
	return fwd.Normalize()
}

func (t *Transform) Rotate(pitch, yaw, roll float32) {
	t.rot[0] += pitch
	t.rot[1] += yaw
	t.rot[2] += roll
	t.upToDate = false
}

func (t *Transform) Matrix() mgl32.Mat4 {
	if !t.upToDate {
		t.matrix = mgl32.Scale3D(t.scale.X(), t.scale.Y(), t.scale.Z())
		t.matrix = mgl32.HomogRotate3DZ(t.rot[2]).Mul4(t.matrix)
		t.matrix = mgl32.HomogRotate3DX(t.rot[0]).Mul4(t.matrix)
		t.matrix = mgl32.HomogRotate3DY(t.rot[1]).Mul4(t.matrix)
		t.matrix = mgl32.Translate3D(t.pos.X(), t.pos.Y(), t.pos.Z()).Mul4(t.matrix)
		t.upToDate = true
	}
	return t.matrix
}
