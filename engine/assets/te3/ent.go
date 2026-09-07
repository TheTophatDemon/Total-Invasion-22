package te3

import (
	"encoding/json/jsontext"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type (
	EntDisplay     uint8
	Ent[Props any] struct {
		Angles     [3]math2.Degrees `json:"angles"`
		Color      [3]int           `json:"color"`
		Position   [3]float32       `json:"position"`
		Radius     float32          `json:"radius"`
		Texture    string           `json:"texture"`
		Model      string           `json:"model"`
		Display    EntDisplay       `json:"display"`
		Properties Props            `json:"properties"`
	}
	// A type that unmarshals a JSON string property into an optional value.
	StringProp struct{ maybe.T[string] }
	// A type that unmarshals a JSON string into an optional bool
	BoolProp struct{ maybe.T[bool] }
	// A type that unmarshals a JSON string into an optional int
	IntProp struct{ maybe.T[int] }
	// A type that unmarshals a JSON string into an optional float32
	FloatProp struct{ maybe.T[float32] }
	// Type that unmarshals a list of 4 floats from a JSON string
	Vec4Prop struct{ maybe.T[mgl32.Vec4] }
	// Type that unmarshals a list of 3 floats from a JSON string
	Vec3Prop struct{ maybe.T[mgl32.Vec3] }
)

const (
	ENT_DISPLAY_SPHERE EntDisplay = iota
	ENT_DISPLAY_MODEL
	ENT_DISPLAY_SPRITE
)

func (ent *Ent[Props]) AnglesInRadians() [3]math2.Radians {
	return [3]math2.Radians{
		math2.Radians(ent.Angles[0]) * math.Pi / 180.0,
		math2.Radians(ent.Angles[1]) * math.Pi / 180.0,
		math2.Radians(ent.Angles[2]) * math.Pi / 180.0,
	}
}

func (ent *Ent[Props]) GridPosition() [3]int {
	return [3]int{
		int(ent.Position[0] / GridSpacing),
		int(ent.Position[1] / GridSpacing),
		int(ent.Position[2] / GridSpacing),
	}
}

// func (ent *Ent[Props]) MarshalJSONTo(encoder *jsontext.Encoder) error {
// 	var err error
// 	// Begin object
// 	err = encoder.WriteToken(jsontext.BeginObject)
// 	if err != nil {
// 		return err
// 	}

// 	// Angles
// 	err = encoder.WriteToken(jsontext.String("angles"))
// 	if err != nil {
// 		return err
// 	}
// 	err = json.MarshalEncode(encoder, ent.Angles)
// 	if err != nil {
// 		return err
// 	}

// 	// Color
// 	err = encoder.WriteToken(jsontext.String("color"))
// 	if err != nil {
// 		return err
// 	}
// 	err = json.MarshalEncode(encoder, ent.Color)
// 	if err != nil {
// 		return err
// 	}

// 	// Position
// 	err = encoder.WriteToken(jsontext.String("position"))
// 	if err != nil {
// 		return err
// 	}
// 	err = json.MarshalEncode(encoder, ent.Position)
// 	if err != nil {
// 		return err
// 	}

// 	// Radius
// 	err = encoder.WriteToken(jsontext.String("radius"))
// 	if err != nil {
// 		return err
// 	}
// 	err = json.MarshalEncode(encoder, ent.Radius)
// 	if err != nil {
// 		return err
// 	}

// 	// Texture
// 	err = encoder.WriteToken(jsontext.String("texture"))
// 	if err != nil {
// 		return err
// 	}
// 	err = json.MarshalEncode(encoder, ent.Texture)
// 	if err != nil {
// 		return err
// 	}

// 	// Model
// 	err = encoder.WriteToken(jsontext.String("model"))
// 	if err != nil {
// 		return err
// 	}
// 	err = json.MarshalEncode(encoder, ent.Model)
// 	if err != nil {
// 		return err
// 	}

// 	// Display
// 	err = encoder.WriteToken(jsontext.String("display"))
// 	if err != nil {
// 		return err
// 	}
// 	err = json.MarshalEncode(encoder, ent.Display)
// 	if err != nil {
// 		return err
// 	}

// 	// Properties
// 	err = encoder.WriteToken(jsontext.String("properties"))
// 	if err != nil {
// 		return err
// 	}
// 	err = encoder.WriteToken(jsontext.BeginObject)
// 	if err != nil {
// 		return err
// 	}

// 	propsType := reflect.TypeOf(ent.Properties)
// 	if propsType.Kind() != reflect.Struct {
// 		return fmt.Errorf("enitity properties type should be a struct, but is %v", propsType.Kind())
// 	}
// 	for field := range propsType.Fields() {
// 		if field.Type.
// 	}

// 	err = encoder.WriteToken(jsontext.EndObject)
// 	if err != nil {
// 		return err
// 	}

// 	// End object
// 	err = encoder.WriteToken(jsontext.EndObject)
// 	if err != nil {
// 		return err
// 	}
// }

func decodeProp[Inner any](decoder *jsontext.Decoder, parse func(string) (Inner, error)) (maybe.T[Inner], error) {
	var none maybe.T[Inner]
	token, err := decoder.ReadToken()
	if err != nil {
		return none, err
	}
	switch token.Kind() {
	case jsontext.KindNull:
		return none, nil
	case jsontext.KindString:
		str := token.String()
		var val Inner
		val, err = parse(str)
		if err != nil {
			return none, err
		}
		return maybe.Some(val), nil
	default:
		return none, fmt.Errorf("unexpected token kind %v", token.Kind())
	}
}

func encodeProp[Inner any](maybeVal maybe.T[Inner], encoder *jsontext.Encoder) error {
	value, ok := maybeVal.Value()
	if !ok {
		return encoder.WriteToken(jsontext.Null)
	}
	// Encode as a string with the stringified value inside of it.
	return encoder.WriteToken(jsontext.String(fmt.Sprintf("%v", value)))
}

func SomeString(str string) StringProp {
	return StringProp{maybe.Some(str)}
}

func NoneString() StringProp {
	return StringProp{}
}

func (sp *StringProp) UnmarshalJSONFrom(decoder *jsontext.Decoder) error {
	maybeVal, err := decodeProp(decoder, func(str string) (string, error) {
		return str, nil
	})
	if err != nil {
		return err
	}
	*sp = StringProp{maybeVal}
	return nil
}

func (sp *StringProp) MarshalJSONTo(encoder *jsontext.Encoder) error {
	return encodeProp(sp.T, encoder)
}

func SomeBool(boo bool) BoolProp {
	return BoolProp{maybe.Some(boo)}
}

func NoneBool() BoolProp {
	return BoolProp{}
}

func (bp *BoolProp) MarshalJSONTo(encoder *jsontext.Encoder) error {
	return encodeProp(bp.T, encoder)
}

func (bp *BoolProp) UnmarshalJSONFrom(decoder *jsontext.Decoder) error {
	maybeVal, err := decodeProp(decoder, strconv.ParseBool)
	if err != nil {
		return err
	}
	*bp = BoolProp{maybeVal}
	return nil
}

func SomeInt(i int) IntProp {
	return IntProp{maybe.Some(i)}
}

func NoneInt() IntProp {
	return IntProp{}
}

func (ip *IntProp) MarshalJSONTo(encoder *jsontext.Encoder) error {
	return encodeProp(ip.T, encoder)
}

func (ip *IntProp) UnmarshalJSONFrom(decoder *jsontext.Decoder) error {
	maybeVal, err := decodeProp(decoder, func(str string) (int, error) {
		i64, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return 0, err
		}
		return int(i64), nil
	})
	if err != nil {
		return err
	}
	*ip = IntProp{maybeVal}
	return nil
}

func SomeFloat(i float32) FloatProp {
	return FloatProp{maybe.Some(i)}
}

func NoneFloat() FloatProp {
	return FloatProp{}
}

func (fp *FloatProp) MarshalJSONTo(encoder *jsontext.Encoder) error {
	return encodeProp(fp.T, encoder)
}

func (fp *FloatProp) UnmarshalJSONFrom(decoder *jsontext.Decoder) error {
	maybeVal, err := decodeProp(decoder, func(str string) (float32, error) {
		f64, err := strconv.ParseFloat(str, 32)
		if err != nil {
			return 0, err
		}
		return float32(f64), nil
	})
	if err != nil {
		return err
	}
	*fp = FloatProp{maybeVal}
	return nil
}

func SomeVec4(v4 mgl32.Vec4) Vec4Prop {
	return Vec4Prop{maybe.Some(v4)}
}

func NoneVec4() Vec4Prop {
	return Vec4Prop{}
}

func (v4p *Vec4Prop) MarshalJSONTo(encoder *jsontext.Encoder) error {
	if vec, ok := v4p.Value(); ok {
		maybeStr := maybe.Some(fmt.Sprintf("%v,%v,%v,%v", vec[0], vec[1], vec[2], vec[3]))
		return encodeProp(maybeStr, encoder)
	}
	return encodeProp(v4p.T, encoder)
}

func decodeVec(decoder *jsontext.Decoder, vec []float32) (bool, error) {
	token, err := decoder.ReadToken()
	if err != nil {
		return false, err
	}
	switch token.Kind() {
	case jsontext.KindNull:
		return false, nil
	case jsontext.KindString:
		var index int = 0
		str := token.String()
		for token := range strings.SplitSeq(str, ",") {
			f64, err := strconv.ParseFloat(strings.TrimSpace(token), 32)
			if err != nil {
				return false, err
			}
			if index > len(vec) {
				return false, fmt.Errorf("expected %v elements in vector value: %v", len(vec), str)
			}
			vec[index] = float32(f64)
			index += 1
		}
		if index != len(vec) {
			return false, fmt.Errorf("expected %v elements in vector value: %v", len(vec), str)
		}
		return true, nil
	default:
		return false, fmt.Errorf("unexpected token kind %v", token.Kind())
	}
}

func (v4 *Vec4Prop) UnmarshalJSONFrom(decoder *jsontext.Decoder) error {
	var v4Val mgl32.Vec4
	isSome, err := decodeVec(decoder, v4Val[:])
	if err != nil {
		return err
	}
	if !isSome {
		return nil
	}
	*v4 = Vec4Prop{maybe.Some(v4Val)}
	return nil
}

func SomeVec3(v3 mgl32.Vec3) Vec3Prop {
	return Vec3Prop{maybe.Some(v3)}
}

func NoneVec3() Vec3Prop {
	return Vec3Prop{}
}

func (v3p *Vec3Prop) MarshalJSONTo(encoder *jsontext.Encoder) error {
	if vec, ok := v3p.Value(); ok {
		maybeStr := maybe.Some(fmt.Sprintf("%v,%v,%v", vec[0], vec[1], vec[2]))
		return encodeProp(maybeStr, encoder)
	}
	return encodeProp(v3p.T, encoder)
}

func (v3 *Vec3Prop) UnmarshalJSONFrom(decoder *jsontext.Decoder) error {
	var v3Val mgl32.Vec3
	isSome, err := decodeVec(decoder, v3Val[:])
	if err != nil {
		return err
	}
	if !isSome {
		return nil
	}
	*v3 = Vec3Prop{maybe.Some(v3Val)}
	return nil
}
