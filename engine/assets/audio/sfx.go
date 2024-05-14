package audio

import (
	"encoding/binary"
	"io"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/go-audio/wav"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Sfx struct {
	Polyphony        uint8
	AttenuationPower float32
	loop             bool
	players          []sfxPlayer
}

type VoiceId struct {
	index, generation uint16
	sfx               *Sfx
}

type sfxPlayer struct {
	oto.Player
	reader     io.Reader
	maxVolume  float32
	playCount  uint16
	lastPlayed time.Time
}

type sfxMetadata struct {
	Loop             bool
	Polyphony        int
	AttenuationPower *float32
}

func (vid VoiceId) IsValid() bool {
	return vid.generation > 0 && vid.sfx != nil && vid.generation == vid.sfx.players[vid.index].playCount
}

func (vid VoiceId) Attenuate(sourcePos mgl32.Vec3, listenerTransform mgl32.Mat4) {
	if !vid.IsValid() {
		return
	}

	vid.sfx.Attenuate(vid, sourcePos, listenerTransform)
}

func (vid VoiceId) Stop() {
	if !vid.IsValid() {
		return
	}

	vid.sfx.Stop(vid)
}

func LoadSfx(assetPath string) (*Sfx, error) {
	file, err := assets.GetFile(assetPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Look for metadata file
	metaPath := strings.TrimSuffix(assetPath, ".wav") + ".json"
	metadata, err := assets.LoadAndUnmarshalJSON[sfxMetadata](metaPath)
	if _, ok := err.(*os.PathError); err != nil && !ok {
		// The file is optional, so print errors that aren't 'file not found'.
		log.Printf("Could not parse metadata for %s; %s\n", assetPath, err)
	}

	sfx := &Sfx{}

	decoder := *wav.NewDecoder(file)

	if metadata != nil && metadata.Polyphony > 0 && metadata.Polyphony < 128 {
		sfx.Polyphony = uint8(metadata.Polyphony)
	} else {
		sfx.Polyphony = 8
	}

	if metadata != nil && metadata.AttenuationPower != nil {
		sfx.AttenuationPower = *metadata.AttenuationPower
	} else {
		sfx.AttenuationPower = 1.5
	}

	if metadata != nil {
		sfx.loop = metadata.Loop
	}

	wavBuffer, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, err
	}
	floatBuffer := wavBuffer.AsFloat32Buffer()

	sampleTimes := 1
	if decoder.NumChans == 1 {
		// Mono samples have each data point doubled to make them "stereo".
		sampleTimes = 2
	}

	wavBytes := make([]byte, 0, len(wavBuffer.Data)*2)
	for i := range floatBuffer.Data {
		sample := math.Float32bits(floatBuffer.Data[i])
		for range sampleTimes {
			wavBytes = binary.LittleEndian.AppendUint32(wavBytes, sample)
		}
	}
	sfx.players = make([]sfxPlayer, sfx.Polyphony)
	for i := range sfx.players {
		var reader io.Reader = NewReader(wavBytes, sfx.loop, 0.0)
		sfx.players[i] = sfxPlayer{
			Player:    *context.NewPlayer(reader),
			reader:    reader,
			playCount: 0,
			maxVolume: 1.0,
		}
		// The buffer size must be small so that the panning can be updated while the sound is playing
		// This size should hold about 1/10 of a second of audio
		sfx.players[i].SetBufferSize(binary.Size(float32(0)) * 2 * SAMPLE_RATE / 10)
	}
	log.Printf("Sfx loaded at %v.\n", assetPath)
	return sfx, nil
}

func (sfx *Sfx) Play() VoiceId {
	playIt := func(player *sfxPlayer) {
		player.Seek(0, io.SeekStart)
		player.Play()
		player.playCount++
		player.lastPlayed = time.Now()
	}

	var oldestPlayer *sfxPlayer
	var oldestPlayerIndex int = -1
	for i := range sfx.players {
		if !sfx.players[i].IsPlaying() {
			playIt(&sfx.players[i])
			return VoiceId{
				generation: sfx.players[i].playCount,
				index:      uint16(i),
				sfx:        sfx,
			}
		} else if oldestPlayer == nil || sfx.players[i].lastPlayed.Compare(oldestPlayer.lastPlayed) < 0 {
			oldestPlayer = &sfx.players[i]
			oldestPlayerIndex = i
		}
	}
	// If all the players are occupied, cut short the one that was playing for the longest
	if oldestPlayer != nil {
		playIt(oldestPlayer)
		return VoiceId{
			generation: oldestPlayer.playCount,
			index:      uint16(oldestPlayerIndex),
			sfx:        sfx,
		}
	}
	return VoiceId{}
}

func (sfx *Sfx) SetVolume(pid VoiceId, targetVolume float32) {
	if !pid.IsValid() {
		return
	}
	player := &sfx.players[pid.index]
	if player.IsPlaying() {
		player.maxVolume = targetVolume
		player.SetVolume(float64(targetVolume))
	}
}

func (sfx *Sfx) Attenuate(pid VoiceId, sourcePos mgl32.Vec3, listener mgl32.Mat4) {
	if !pid.IsValid() {
		return
	}
	player := &sfx.players[pid.index]
	if player.IsPlaying() {
		// Set volume based on distance
		diff := sourcePos.Sub(listener.Col(3).Vec3())
		distance := diff.Len()
		if distance == 0.0 {
			player.SetVolume(1.0)
			return
		}
		newVolume := float64(player.maxVolume / math2.Pow(max(1.0, distance), sfx.AttenuationPower))
		if newVolume < 0.01 {
			newVolume = 0.0
		}
		player.SetVolume(newVolume)

		// Set panning based on angle
		if dynReader, ok := player.reader.(*SfxReader); ok {
			dynReader.muty.Lock()
			defer dynReader.muty.Unlock()
			right := mgl32.TransformNormal(math2.Vec3Right(), listener)
			dynReader.Pan = -right.Dot(diff.Mul(1.0 / distance))
		}
	}
}

func (sfx *Sfx) Stop(pid VoiceId) {
	if !pid.IsValid() {
		return
	}
	if sfx.players[pid.index].playCount == pid.generation {
		sfx.players[pid.index].Pause()
	}
}

func (sfx *Sfx) Free() {
	for i := range sfx.players {
		err := sfx.players[i].Close()
		if err != nil {
			log.Println(err)
		}
	}
}
