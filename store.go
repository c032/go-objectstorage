package objectstorage

import (
	"io"
)

type Store interface {
	StoreReader
	StoreWriter
}

type StoreReader interface {
	Read(key string, w io.Writer) (int64, error)
}

type StoreWriter interface {
	Put(key string, r io.Reader) error
	Remove(path string) error
}
