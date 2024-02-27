package textures

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type aseSpriteSheet struct {
	Frames []aseFrame
	Meta   aseMeta
}

type aseMeta struct {
	Image string
	Size  struct {
		W, H uint
	}
	FrameTags []aseTag
	Layers    []aseLayer
	Data      string
}

type aseFrame struct {
	FileName string
	Frame    Rect
	Duration uint
}

type Rect struct {
	X, Y, W, H int
}

type aseTag struct {
	Name     string
	From, To uint
	Repeat   string
	Data     string
}

type aseLayer struct {
	Name, Data string
}

func (af *aseFrame) getLayerName() (string, error) {
	firstSemi := strings.IndexRune(af.FileName, ';')
	if firstSemi > -1 && firstSemi < len(af.FileName)-1 {
		secondSemi := strings.IndexRune(af.FileName[firstSemi+1:], ';')
		if secondSemi > -1 {
			return af.FileName[firstSemi+1 : firstSemi+1+secondSemi], nil
		} else {
			return af.FileName[firstSemi+1:], nil
		}
	}
	return "", fmt.Errorf("could not get layer name; frame name did not have the correct format")
}

func (af *aseFrame) getFrameNo() (uint, error) {
	lastSemi := strings.LastIndexByte(af.FileName, byte(';'))
	if lastSemi > -1 {
		index, err := strconv.ParseUint(af.FileName[lastSemi+1:], 10, 32)
		if err != nil {
			return 0, err
		}
		return uint(index), nil
	}
	return 0, fmt.Errorf("could not get frame number; frame name did not have the correct format")
}

// Returns the path to the texture atlas relative to the directory of the sprite sheet .json file.
func (ss *aseSpriteSheet) atlasPath() string {
	return ss.Meta.Image
}

func (ss *aseSpriteSheet) loadFlags() ([]string, error) {
	if len(ss.Meta.Data) == 0 {
		return []string{}, nil
	}

	var flags []string
	err := json.Unmarshal([]byte(ss.Meta.Data), &flags)
	if err != nil {
		return nil, err
	}

	return flags, nil
}

func (ss *aseSpriteSheet) loadFrames(layer *aseLayer, tag *aseTag) (frames []Frame, err error) {
	capacity := len(ss.Frames)
	if tag != nil {
		capacity = int(tag.To) - int(tag.From) + 1
	}
	frames = make([]Frame, 0, capacity)

	for _, frame := range ss.Frames {
		if layer != nil {
			var layerName string
			layerName, err = frame.getLayerName()
			if err != nil {
				frames = nil
				return
			}
			if layerName != layer.Name {
				continue
			}
		}
		if tag != nil {
			var frameNo uint
			frameNo, err = frame.getFrameNo()
			if err != nil {
				frames = nil
				return
			}
			if frameNo < tag.From || frameNo > tag.To {
				continue
			}
		}
		frames = append(frames, frameFromAesprite(frame))
	}

	frames = slices.Clip(frames)
	return
}

func (ss *aseSpriteSheet) loadLayers() (map[string]Layer, error) {
	layers := make(map[string]Layer)
	for l := range ss.Meta.Layers {
		layer, err := layerFromAseprite(ss.Meta.Layers[l])
		if err != nil {
			return nil, err
		}
		layers[layer.Name] = layer
	}
	return layers, nil
}

// Loads texture animations from an Aseprite sprite sheet file.
//
// When there are no animations available, the returned slice is empty.
//
// When there are no layers, an animation is returned for each tag.
//
// When there are layers, an animation is returned for each combination of tag and layer.
func (ss *aseSpriteSheet) loadAnimations() (map[string]Animation, error) {
	if ss.Meta.FrameTags == nil || len(ss.Meta.FrameTags) == 0 || len(ss.Frames) <= 1 {
		// There are no animations to load
		return map[string]Animation{}, nil
	}

	anims := make(map[string]Animation)

	for t := range ss.Meta.FrameTags {
		anim := Animation{
			AtlasSize: [2]uint{
				ss.Meta.Size.W,
				ss.Meta.Size.H,
			},
			Loop: (len(ss.Meta.FrameTags[t].Repeat) == 0),
		}

		var err error
		for l := range ss.Meta.Layers {
			anim.Name = ss.Meta.FrameTags[t].Name + ";" + ss.Meta.Layers[l].Name
			// Load a separate animation for each layer
			anim.Frames, err = ss.loadFrames(&ss.Meta.Layers[l], &ss.Meta.FrameTags[t])
			if err != nil {
				return nil, err
			}
			anims[anim.Name] = anim
		}
		if ss.Meta.Layers == nil || len(ss.Meta.Layers) == 0 {
			anim.Name = ss.Meta.FrameTags[t].Name
			// When there are no layers, just load all of the frames in the tag as one animation
			anim.Frames, err = ss.loadFrames(nil, &ss.Meta.FrameTags[t])
			if err != nil {
				return nil, err
			}
			anims[anim.Name] = anim
		}
	}

	return anims, nil
}

// Converts an Aseprite animation frame to our Frame
func frameFromAesprite(af aseFrame) Frame {
	return Frame{
		Rect: math2.Rect{
			X:      float32(af.Frame.X),
			Y:      float32(af.Frame.Y),
			Width:  float32(af.Frame.W),
			Height: float32(af.Frame.H),
		},
		Duration: float32(af.Duration) / 1000.0,
	}
}

func layerFromAseprite(af aseLayer) (Layer, error) {
	var layerData struct {
		ViewRange, FlippedViewRange [2]int
	}

	err := json.Unmarshal([]byte(af.Data), &layerData)
	if err != nil {
		return Layer{}, err
	}

	return Layer{
		Name:             af.Name,
		ViewRange:        layerData.ViewRange,
		FlippedViewRange: layerData.FlippedViewRange,
	}, nil
}
