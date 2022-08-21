package comps

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Transform struct {
	pos, rot, scale mgl32.Vec3
	matrix          mgl32.Mat4
	dirty           bool
}

func TransformFromMatrix(matrix mgl32.Mat4) Transform {
	return Transform{
		pos: matrix.Col(3).Vec3(),
		rot: math2.Mat4EulerAngles(&matrix),
		scale: mgl32.Vec3{ matrix[0], matrix[5], matrix[10] },
		matrix: matrix,
		dirty: false,
	}
}

func (t *Transform) IsDirty() bool {
	return t.dirty
}

func (t *Transform) SetPosition(newPos mgl32.Vec3) {
	t.pos = newPos
	t.dirty = true
}

func (t *Transform) GetPosition() mgl32.Vec3 {
	return t.pos
}

func (t *Transform) Translate(x, y, z float32) {
	t.pos[0] += x
	t.pos[1] += y
	t.pos[2] += z
	t.dirty = true
}

//Sets the rotation in euler angles (radians)
func (t *Transform) SetRotation(pitch, yaw, roll float32) {
	t.rot = mgl32.Vec3{pitch, yaw, roll}
	t.dirty = true
}

//Returns the euler angles of rotation (pitch, yaw, roll)
func (t *Transform) GetRotation() mgl32.Vec3 {
	return t.rot
}

func (t *Transform) GetYaw() float32 {
	return t.rot[1]
}

func (t *Transform) GetPitch() float32 {
	return t.rot[0]
}

func (t *Transform) GetRoll() float32 {
	return t.rot[2]
}

func (t *Transform) Rotate(pitch, yaw, roll float32) {
	t.rot[0] += pitch
	t.rot[1] += yaw
	t.rot[2] += roll
	t.dirty = true
}

func (t *Transform) GetMatrix() mgl32.Mat4 {
	if t.IsDirty() {
		t.matrix = mgl32.Translate3D(t.pos.X(), t.pos.Y(), t.pos.Z()).Mul4(
			mgl32.HomogRotate3DY(t.rot[1]).Mul4(
				mgl32.HomogRotate3DX(t.rot[0]).Mul4(
					mgl32.HomogRotate3DZ(t.rot[2]))))
	}
	return t.matrix
}