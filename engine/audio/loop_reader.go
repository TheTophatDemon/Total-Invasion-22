package audio

import (
	"bytes"
	"io"
)

// A bytes.Reader that, when reaching the end of the buffer, reads from the beginning in an infinite loop.
type LoopReader struct {
	byteReader *bytes.Reader
}

var _ io.Reader = (*LoopReader)(nil)

func NewLoopReader(buffer []byte) *LoopReader {
	return &LoopReader{
		byteReader: bytes.NewReader(buffer),
	}
}

func (lr *LoopReader) Read(p []byte) (n int, err error) {
	n, _ = lr.byteReader.Read(p)
	if n < len(p) {
		lr.byteReader.Seek(0, io.SeekStart)
		n, _ = lr.byteReader.Read(p[n:])
	}
	return n, nil
}
