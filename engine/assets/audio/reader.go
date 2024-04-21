package audio

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"sync"
)

// This is a reader for an audio stream that will loop or adjust the panning of the audio in real time.
type Reader struct {
	byteReader bytes.Reader
	muty       sync.Mutex
	Loop       bool
	Pan        float32
}

var _ io.ReadSeeker = (*Reader)(nil)

func NewReader(buffer []byte, loop bool, pan float32) *Reader {
	return &Reader{
		byteReader: *bytes.NewReader(buffer),
		Loop:       loop,
		Pan:        pan,
	}
}

func (reader *Reader) Read(p []byte) (n int, err error) {
	reader.muty.Lock()
	defer reader.muty.Unlock()

	rawSamples := make([]byte, len(p))
	n, err = reader.byteReader.Read(rawSamples)
	// Loop the samples if the buffer reaches the end
	if reader.Loop {
		for n < len(rawSamples) {
			reader.byteReader.Seek(0, io.SeekStart)
			n2, _ := reader.byteReader.Read(rawSamples[n:])
			n += n2
		}
	}
	if reader.Pan != 0.0 {
		// Balance the left and right channels according to reader.Pan
		for s := 0; s < len(rawSamples); s += 4 {
			leftSample := float32(int16(binary.LittleEndian.Uint16(rawSamples[s:][:2]))) / math.MaxInt16
			rightSample := float32(int16(binary.LittleEndian.Uint16(rawSamples[s:][2:4]))) / math.MaxInt16
			leftDest := p[s:][:2]
			rightDest := p[s:][2:4]
			var newLeftSample, newRightSample float32
			if reader.Pan < 0.0 {
				offset := -reader.Pan * rightSample
				newLeftSample = leftSample + offset
				newRightSample = rightSample - offset
			} else {
				offset := reader.Pan * leftSample
				newLeftSample = leftSample - offset
				newRightSample = rightSample + offset
			}
			newLeftSample = max(-1.0, min(1.0, newLeftSample))
			newRightSample = max(-1.0, min(1.0, newRightSample))
			binary.LittleEndian.PutUint16(leftDest, uint16(int16(newLeftSample*math.MaxInt16)))
			binary.LittleEndian.PutUint16(rightDest, uint16(int16(newRightSample*math.MaxInt16)))
		}
	} else {
		copy(p, rawSamples)
	}
	return
}

func (reader *Reader) Seek(offset int64, whence int) (int64, error) {
	return reader.byteReader.Seek(offset, whence)
}
