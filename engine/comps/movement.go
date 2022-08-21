package comps

type Movement struct {
	MaxSpeed, Accel, Friction float32
	InputForward, InputStrafe float32
	YawAngle, PitchAngle      float32
}