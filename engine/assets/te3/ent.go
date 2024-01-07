package te3

import (
	"fmt"
	"math"
	"strconv"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"
)

type EntDisplay uint8

const (
	ENT_DISPLAY_SPHERE EntDisplay = iota
	ENT_DISPLAY_MODEL
	ENT_DISPLAY_SPRITE
)

type Ent struct {
	Angles         [3]float32
	Color          [3]uint8
	Position       [3]float32
	Radius         float32
	Texture, Model string
	Display        EntDisplay
	Properties     map[string]string
}

// Extracts and parses the value of a float property.
func (ent *Ent) FloatProperty(key string) (float32, error) {
	prop, ok := ent.Properties[key]
	if !ok {
		return 0.0, fmt.Errorf("ent property not found: %v", key)
	}
	valF64, err := strconv.ParseFloat(prop, 32)
	if err != nil {
		return 0.0, err
	}
	return float32(valF64), nil
}

func (ent *Ent) BoolProperty(key string) (bool, error) {
	prop, ok := ent.Properties[key]
	if !ok {
		return false, fmt.Errorf("ent property not found: %v", prop)
	}
	return strconv.ParseBool(prop)
}

func (ent *Ent) Transform(scaleByRadius, stayOnFloor bool) comps.Transform {
	angles := mgl32.Vec3(ent.Angles).Mul(math.Pi / 180.0)
	if scaleByRadius {
		pos := mgl32.Vec3(ent.Position)
		if stayOnFloor {
			pos = pos.Add(mgl32.Vec3{0.0, ent.Radius - 1.0, 0.0})
		}
		return comps.TransformFromTranslationAnglesScale(
			pos, angles, mgl32.Vec3{ent.Radius, ent.Radius, ent.Radius},
		)
	} else {
		return comps.TransformFromTranslationAngles(
			ent.Position, angles,
		)
	}
}
