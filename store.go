package objectstorage

import (
	"errors"
	"io"
)

var ErrNotFound error = errors.New("not found")

type Store interface {
	StoreReader
	StoreWriter
}

type StoreReader interface {
	Read(key string, w io.Writer) (int64, error)
}

type StoreWriter interface {
	Put(key string, r io.Reader) error
	Remove(key string) error
}
