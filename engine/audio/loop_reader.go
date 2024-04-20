package audio

import (
	"bytes"
	"io"
)

// A bytes.Reader that, when reaching the end of the buffer, reads from the beginning in an infinite loop.
type LoopReader struct {
	byteReader bytes.Reader
}

var _ io.ReadSeeker = (*LoopReader)(nil)

func NewLoopReader(buffer []byte) *LoopReader {
	return &LoopReader{
		byteReader: *bytes.NewReader(buffer),
	}
}

func (lr *LoopReader) Read(p []byte) (n int, err error) {
	n, _ = lr.byteReader.Read(p)
	if n < len(p) {
		lr.byteReader.Seek(0, io.SeekStart)
		n2, _ := lr.byteReader.Read(p[n:])
		n += n2
	}
	return n, nil
}

func (lr *LoopReader) Seek(offset int64, whence int) (int64, error) {
	return lr.byteReader.Seek(offset, whence)
}
