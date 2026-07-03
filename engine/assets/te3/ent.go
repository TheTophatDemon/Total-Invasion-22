package te3

import (
	"fmt"
	"math"
	"strconv"

	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type EntDisplay uint8

const (
	ENT_DISPLAY_SPHERE EntDisplay = iota
	ENT_DISPLAY_MODEL
	ENT_DISPLAY_SPRITE
)

type Ent struct {
	Angles     [3]math2.Degrees  `json:"angles"`
	Color      [3]uint8          `json:"color"`
	Position   [3]float32        `json:"position"`
	Radius     float32           `json:"radius"`
	Texture    string            `json:"texture"`
	Model      string            `json:"model"`
	Display    EntDisplay        `json:"display"`
	Properties map[string]string `json:"properties"`
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

func (ent *Ent) FloatPropertyOr(key string, defaultValue float32) float32 {
	f, err := ent.FloatProperty(key)
	if err != nil {
		return defaultValue
	}
	return f
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

func (ent *Ent) IntPropertyOr(key string, defaultValue int) int {
	i, err := ent.IntProperty(key)
	if err != nil {
		return defaultValue
	}
	return i
}

func (ent *Ent) BoolProperty(key string) (bool, error) {
	prop, ok := ent.Properties[key]
	if !ok {
		return false, PropNotFoundError(key)
	}
	return strconv.ParseBool(prop)
}

func (ent *Ent) BoolPropertyOr(key string, defaultValue bool) bool {
	b, err := ent.BoolProperty(key)
	if err != nil {
		return defaultValue
	}
	return b
}

func (ent *Ent) AnglesInRadians() [3]math2.Radians {
	return [3]math2.Radians{
		math2.Radians(ent.Angles[0]) * math.Pi / 180.0,
		math2.Radians(ent.Angles[1]) * math.Pi / 180.0,
		math2.Radians(ent.Angles[2]) * math.Pi / 180.0,
	}
}

func (ent *Ent) GridPosition() [3]int {
	return [3]int{
		int(ent.Position[0] / GridSpacing),
		int(ent.Position[1] / GridSpacing),
		int(ent.Position[2] / GridSpacing),
	}
}
