package memorystore

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/c032/go-objectstorage"
)

var _ objectstorage.Store = (*MemoryStore)(nil)

type MemoryStore struct {
	mu sync.RWMutex

	isInitialized bool

	objects    map[string][]byte
	bufferPool *sync.Pool
}

func (ms *MemoryStore) init() {
	if ms.isInitialized {
		return
	}

	ms.objects = map[string][]byte{}

	ms.bufferPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 1024))
		},
	}

	ms.isInitialized = true
}

func (ms *MemoryStore) Read(key string, w io.Writer) (int64, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	b, ok := ms.objects[key]
	if !ok {
		return 0, objectstorage.ErrNotFound
	}

	buf := bytes.NewReader(b)

	return buf.WriteTo(w)
}

func (ms *MemoryStore) Put(key string, r io.Reader) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.init()

	buf := ms.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer ms.bufferPool.Put(buf)

	_, err := io.Copy(buf, r)
	if err != nil {
		return fmt.Errorf("could not put: %w", err)
	}

	ms.objects[key] = buf.Bytes()

	return nil
}

func (ms *MemoryStore) Remove(key string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	delete(ms.objects, key)

	return nil
}

func New() objectstorage.Store {
	return &MemoryStore{}
}
