package audio

import (
	"bytes"
	"fmt"
	"io"
)

// This is a reader for an audio stream that will loop or adjust the panning of the audio in real time.
type DynamicReader struct {
	byteReader bytes.Reader
	Loop       bool
	Pan        float32
}

var _ io.ReadSeeker = (*DynamicReader)(nil)

func NewDynamicReader(buffer []byte, loop bool) *DynamicReader {
	return &DynamicReader{
		byteReader: *bytes.NewReader(buffer),
		Loop:       loop,
		Pan:        0.0,
	}
}

func (reader *DynamicReader) Read(p []byte) (n int, err error) {
	n, err = reader.byteReader.Read(p)
	if reader.Loop && n < len(p) {
		reader.byteReader.Seek(0, io.SeekStart)
		n2, _ := reader.byteReader.Read(p[n:])
		n += n2
	}
	if len(p)%4 != 0 {
		fmt.Println("AAAA")
	}
	return
}

func (reader *DynamicReader) Seek(offset int64, whence int) (int64, error) {
	return reader.byteReader.Seek(offset, whence)
}
