package math2

// This file contains functions that wrap the standard math library to take either float32 or float64, because casting from float64 everywhere is freakin' annoying.

import "math"

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

func Acos[F Float](val F) F {
	return F(math.Acos(float64(val)))
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

func Inf32() float32 {
	return float32(math.Inf(1))
}

func Mod[F Float](a, b F) F {
	return F(math.Mod(float64(a), float64(b)))
}
