package audio

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
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
	Cooldown         float32
	Polyphony        uint8
	AttenuationPower float32
	loop             bool
	decoder          wav.Decoder
	players          []sfxPlayer
	lastPlayed       time.Time
}

type VoiceId uint32

type sfxPlayer struct {
	oto.Player
	maxVolume float32
	playCount uint16
}

type sfxMetadata struct {
	Loop             bool
	Cooldown         float32
	Polyphony        int
	AttenuationPower float32
}

func (pid VoiceId) index() uint16 {
	return uint16(pid & 0xFF00)
}

func (pid VoiceId) generation() uint16 {
	return uint16(pid & 0x00FF)
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

	sfx := &Sfx{
		decoder: *wav.NewDecoder(file),
	}

	if metadata != nil && metadata.Polyphony > 0 && metadata.Polyphony < 128 {
		sfx.Polyphony = uint8(metadata.Polyphony)
	} else {
		sfx.Polyphony = 8
	}

	if metadata != nil && metadata.Cooldown > 0.0 {
		sfx.Cooldown = metadata.Cooldown
	} else {
		sfx.Cooldown = 0.1
	}

	if metadata != nil && metadata.AttenuationPower > 0.01 && metadata.AttenuationPower < 10.0 {
		sfx.AttenuationPower = metadata.AttenuationPower
	} else {
		sfx.AttenuationPower = 1.0
	}

	if metadata != nil {
		sfx.loop = metadata.Loop
	}

	wavBuffer, err := sfx.decoder.FullPCMBuffer()
	if err != nil {
		return nil, err
	}

	sampleTimes := 1
	if sfx.decoder.NumChans == 1 {
		// Mono samples have each data point doubled to make them "stereo".
		sampleTimes = 2
	}

	wavBytes := make([]byte, 0, len(wavBuffer.Data)*2)
	for i := range wavBuffer.Data {
		for range sampleTimes {
			wavBytes = binary.LittleEndian.AppendUint16(wavBytes, uint16(wavBuffer.Data[i]))
		}
	}
	sfx.players = make([]sfxPlayer, sfx.Polyphony)
	for i := range sfx.players {
		var reader io.Reader
		if sfx.loop {
			reader = NewLoopReader(wavBytes)
		} else {
			reader = bytes.NewReader(wavBytes)
		}
		sfx.players[i] = sfxPlayer{
			Player:    *context.NewPlayer(reader),
			playCount: 0,
			maxVolume: 1.0,
		}
	}
	log.Printf("Sfx loaded at %v.\n", assetPath)
	return sfx, nil
}

func (sfx *Sfx) Play() VoiceId {
	if time.Since(sfx.lastPlayed).Seconds() < float64(sfx.Cooldown) {
		return VoiceId(0)
	}

	for i := range sfx.players {
		if !sfx.players[i].IsPlaying() {
			sfx.players[i].Seek(0, io.SeekStart)
			sfx.players[i].Play()
			sfx.lastPlayed = time.Now()
			sfx.players[i].playCount++
			pid := VoiceId(uint32(sfx.players[i].playCount&0x00FF) | uint32(i)<<8)
			return pid
		}
	}
	return VoiceId(0)
}

func (sfx *Sfx) SetVolume(pid VoiceId, targetVolume float32) {
	if pid == 0 {
		return
	}
	player := &sfx.players[pid.index()]
	if player.IsPlaying() {
		player.maxVolume = targetVolume
		player.SetVolume(float64(targetVolume))
	}
}

func (sfx *Sfx) Attenuate(pid VoiceId, sourcePos, listenPos mgl32.Vec3) {
	if pid == 0 {
		return
	}
	player := &sfx.players[pid.index()]
	if player.IsPlaying() {
		distance := max(1.0, sourcePos.Sub(listenPos).Len())
		player.SetVolume(float64(player.maxVolume / math2.Pow(distance, sfx.AttenuationPower)))
	}
}

func (sfx *Sfx) Stop(pid VoiceId) {
	if pid == 0 {
		return
	}
	idx := pid.index()
	if sfx.players[idx].playCount == pid.generation() {
		sfx.players[idx].Pause()
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
