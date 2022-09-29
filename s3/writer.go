package s3

import (
	"errors"
	"fmt"
	"io"
)

var _ io.WriterAt = (*sequentialWriter)(nil)

var ErrUnexpectedOffset error = errors.New("unexpected offset")

type sequentialWriter struct {
	w      io.Writer
	offset int64
}

func (sw *sequentialWriter) WriteAt(p []byte, offset int64) (int, error) {
	if offset != sw.offset {
		return 0, ErrUnexpectedOffset
	}

	var (
		err error

		n int
	)

	n, err = sw.w.Write(p)
	if err != nil {
		return n, fmt.Errorf("could not write: %w", err)
	}

	sw.offset += int64(n)

	return n, nil
}
