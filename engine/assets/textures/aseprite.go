package textures

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"tophatdemon.com/total-invasion-ii/engine/math2"
)

const (
	META_SLICE_NAME   = "meta"
	DEFAULT_ANIM_FLAG = "default"
)

type (
	aseSpriteSheet struct {
		Frames []aseFrame
		Meta   aseMeta
	}

	aseMeta struct {
		Image string
		Size  struct {
			W, H uint
		}
		FrameTags []aseTag
		Layers    []aseLayer
		Slices    []aseSlice
	}

	aseSlice struct {
		Name string
		Data string
		Keys []aseSliceKey
	}

	aseSliceKey struct {
		Frame  int
		Bounds Rect
	}

	aseFrame struct {
		FileName string
		Frame    Rect
		Duration uint
	}

	Rect struct {
		X, Y, W, H int
	}

	aseTag struct {
		Name     string
		From, To uint
		Repeat   string
		Data     string
	}

	aseLayer struct {
		Name, Data string
	}
)

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

// Loads texture flags from a slice called "meta" in the ase file.
// Once Aseprite updates to export its per-sprite metadata, this can
// be changed to use that instead.
// The flags are separated by space characters and can otherwise be
// any non-whitespace sequence.
// Returns an empty slice if no flags are found.
func (ss *aseSpriteSheet) loadFlags() []string {
	for _, slice := range ss.Meta.Slices {
		if slice.Name == META_SLICE_NAME {
			flags := strings.Split(slice.Data, " ")
			for i := range flags {
				flags[i] = strings.TrimSpace(flags[i])
			}
			return flags
		}
	}

	return []string{}
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

		if ss.Meta.FrameTags[t].Data == DEFAULT_ANIM_FLAG {
			anim.Default = true
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
