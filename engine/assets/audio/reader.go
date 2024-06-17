package audio

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"sync"

	"github.com/jfreymuth/oggvorbis"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

// This is a reader for an audio stream that will loop or adjust the panning of the audio in real time.
type SfxReader struct {
	byteReader bytes.Reader
	muty       sync.Mutex
	Loop       bool
	Pan        float32 // From -1.0 (left ear) to 1.0 (right ear)
}

var _ io.ReadSeeker = (*SfxReader)(nil)

func NewReader(buffer []byte, loop bool, pan float32) *SfxReader {
	return &SfxReader{
		byteReader: *bytes.NewReader(buffer),
		Loop:       loop,
		Pan:        pan,
	}
}

func (reader *SfxReader) Read(outBuffer []byte) (nBytesRead int, err error) {
	reader.muty.Lock()
	defer reader.muty.Unlock()

	busVolume := float32(sfxVolume.Load()) / BUS_VOLUME_MAX

	rawSamples := make([]byte, len(outBuffer))
	nBytesRead, err = reader.byteReader.Read(rawSamples)
	// Loop the samples if the buffer reaches the end
	if reader.Loop {
		for nBytesRead < len(rawSamples) {
			reader.byteReader.Seek(0, io.SeekStart)
			nBytesRead2, _ := reader.byteReader.Read(rawSamples[nBytesRead:])
			nBytesRead += nBytesRead2
		}
	}
	// Balance the left and right channels according to reader.Pan
	for s := 0; s < len(rawSamples); s += 8 {
		leftSample := math.Float32frombits(binary.LittleEndian.Uint32(rawSamples[s:][:4]))
		rightSample := math.Float32frombits(binary.LittleEndian.Uint32(rawSamples[s:][4:8]))
		leftDest := outBuffer[s:][:4]
		rightDest := outBuffer[s:][4:8]

		leftFactor := (reader.Pan + 1.0) / 2.0
		rightFactor := 1.0 - leftFactor

		newLeftSample := leftSample * math2.FastApproxSin(leftFactor) * busVolume
		newRightSample := rightSample * math2.FastApproxSin(rightFactor) * busVolume
		binary.LittleEndian.PutUint32(leftDest, math.Float32bits(newLeftSample))
		binary.LittleEndian.PutUint32(rightDest, math.Float32bits(newRightSample))
	}
	return
}

func (reader *SfxReader) Seek(offset int64, whence int) (int64, error) {
	return reader.byteReader.Seek(offset, whence)
}

// This reader will stream audio from an OGG vorbis file and optionally loop it.
type SongReader struct {
	oggReader oggvorbis.Reader
	loop      bool
}

var _ io.ReadSeeker = (*SongReader)(nil)

func NewSongReader(oggReader oggvorbis.Reader, loop bool) SongReader {
	return SongReader{
		oggReader, loop,
	}
}

func (reader *SongReader) Read(outBuffer []byte) (nBytesRead int, err error) {
	busVolume := float32(musicVolume.Load()) / BUS_VOLUME_MAX

	// Read floats one at a time from the ogg file, then turn them into bytes and append to the output buffer.
	var floatBuffer [2]float32
	var n int
	floatSize := binary.Size(float32(0))
	stride := floatSize * 2
	for byteIdx := 0; byteIdx < len(outBuffer); byteIdx += stride {
		n, err = reader.oggReader.Read(floatBuffer[:])
		nBytesRead += (n / reader.oggReader.Channels()) * stride
		if err == io.EOF && reader.loop {
			err = reader.oggReader.SetPosition(0)
			if err == nil && n == 0 {
				n, err = reader.oggReader.Read(floatBuffer[:])
				nBytesRead += (n / reader.oggReader.Channels()) * stride
			}
		}
		if err != nil {
			return nBytesRead, err
		}
		if reader.oggReader.Channels() < 2 {
			floatBuffer[1] = floatBuffer[0]
		}
		floatBuffer[0] *= busVolume
		floatBuffer[1] *= busVolume
		binary.LittleEndian.PutUint32(outBuffer[byteIdx:], math.Float32bits(floatBuffer[0]))
		binary.LittleEndian.PutUint32(outBuffer[byteIdx+floatSize:], math.Float32bits(floatBuffer[1]))
	}
	return
}

func (reader *SongReader) Seek(offset int64, whence int) (int64, error) {
	var newPos int64
	offset /= int64(binary.Size(float32(0)) * reader.oggReader.Channels()) // Change offset from bytes to samples
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = reader.oggReader.Position() + offset
	case io.SeekEnd:
		newPos = reader.oggReader.Length() + offset
	}
	err := reader.oggReader.SetPosition(newPos)
	return newPos, err
}
