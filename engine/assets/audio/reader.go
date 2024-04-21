package audio

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"sync"

	"tophatdemon.com/total-invasion-ii/engine/math2"
)

// This is a reader for an audio stream that will loop or adjust the panning of the audio in real time.
type Reader struct {
	byteReader bytes.Reader
	muty       sync.Mutex
	Loop       bool
	Pan        float32 // From -1.0 (left ear) to 1.0 (right ear)
}

var _ io.ReadSeeker = (*Reader)(nil)

func NewReader(buffer []byte, loop bool, pan float32) *Reader {
	return &Reader{
		byteReader: *bytes.NewReader(buffer),
		Loop:       loop,
		Pan:        pan,
	}
}

func (reader *Reader) Read(outBuffer []byte) (nBytesRead int, err error) {
	reader.muty.Lock()
	defer reader.muty.Unlock()

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
	if reader.Pan != 0.0 {
		// Balance the left and right channels according to reader.Pan
		for s := 0; s < len(rawSamples); s += 8 {
			leftSample := math.Float32frombits(binary.LittleEndian.Uint32(rawSamples[s:][:4]))
			rightSample := math.Float32frombits(binary.LittleEndian.Uint32(rawSamples[s:][4:8]))
			leftDest := outBuffer[s:][:4]
			rightDest := outBuffer[s:][4:8]

			leftFactor := (reader.Pan + 1.0) / 2.0
			rightFactor := 1.0 - leftFactor

			newLeftSample := max(-1.0, min(1.0, leftSample*math2.FastApproxSin(leftFactor)))
			newRightSample := max(-1.0, min(1.0, rightSample*math2.FastApproxSin(rightFactor)))
			binary.LittleEndian.PutUint32(leftDest, math.Float32bits(newLeftSample))
			binary.LittleEndian.PutUint32(rightDest, math.Float32bits(newRightSample))
		}
	} else {
		copy(outBuffer, rawSamples)
	}
	return
}

func (reader *Reader) Seek(offset int64, whence int) (int64, error) {
	return reader.byteReader.Seek(offset, whence)
}
