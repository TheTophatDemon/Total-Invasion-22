package te3

import (
	"fmt"
	"math"
	"strconv"

	"github.com/go-gl/mathgl/mgl32"
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

type PropNotFoundError string

func (err PropNotFoundError) Error() string {
	return fmt.Sprintf("ent property not found: %v", string(err))
}

// Extracts and parses the value of a float property.
func (ent *Ent) FloatProperty(key string) (float32, error) {
	prop, ok := ent.Properties[key]
	if !ok {
		return 0.0, PropNotFoundError(key)
	}
	valF64, err := strconv.ParseFloat(prop, 32)
	if err != nil {
		return 0.0, err
	}
	return float32(valF64), nil
}

func (ent *Ent) IntProperty(key string) (int, error) {
	prop, ok := ent.Properties[key]
	if !ok {
		return 0, PropNotFoundError(key)
	}
	valI64, err := strconv.ParseInt(prop, 10, 32)
	if err != nil {
		return 0, err
	}
	return int(valI64), nil
}

func (ent *Ent) BoolProperty(key string) (bool, error) {
	prop, ok := ent.Properties[key]
	if !ok {
		return false, PropNotFoundError(key)
	}
	return strconv.ParseBool(prop)
}

func (ent *Ent) AnglesInRadians() mgl32.Vec3 {
	return mgl32.Vec3(ent.Angles).Mul(math.Pi / 180.0)
}

func (ent *Ent) GridPosition() [3]int {
	return [3]int{
		int(ent.Position[0] / GRID_SPACING),
		int(ent.Position[1] / GRID_SPACING),
		int(ent.Position[2] / GRID_SPACING),
	}
}
