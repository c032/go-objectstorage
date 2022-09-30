package encryptedstore

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/minio/sio"
	"golang.org/x/crypto/hkdf"

	"github.com/c032/go-objectstorage"
)

const (
	keySize   = 32
	nonceSize = 32
)

var _ objectstorage.Store = (*EncryptedStore)(nil)

type EncryptedStore struct {
	key   []byte
	store objectstorage.Store
}

func (es *EncryptedStore) Read(key string, w io.Writer) (int64, error) {
	var (
		copyErr error
		copyN   int64

		storeErr error
		dErr     error
	)

	d := newDecryptor(es.key)

	chCopy := make(chan struct{})

	go func(key string, d *decryptor) {
		copyN, copyErr = io.Copy(w, d)
		close(chCopy)
	}(key, d)

	_, storeErr = es.store.Read(key, d)
	if storeErr != nil {
		_ = d.Close()
		<-chCopy

		return copyN, storeErr
	}
	dErr = d.Close()
	<-chCopy

	var err error

	// TODO: Create custom struct containing both errors.
	if copyErr != nil {
		err = copyErr
	} else if storeErr != nil {
		err = storeErr
	} else if dErr != nil {
		err = dErr
	}

	return copyN, err
}

func (es *EncryptedStore) Put(objectKey string, r io.Reader) error {
	var (
		err error

		nonce [nonceSize]byte
	)

	nonce, err = generateNonce()
	if err != nil {
		return fmt.Errorf("could not generate nonce: %w", err)
	}

	var encryptionKey [keySize]byte

	encryptionKey, err = deriveKey(es.key, nonce[:])
	if err != nil {
		return fmt.Errorf("could not derive key: %w", err)
	}

	cfg := sio.Config{
		Key: encryptionKey[:],
	}

	encrypted, err := sio.EncryptReader(r, cfg)
	if err != nil {
		return fmt.Errorf("could not encrypt reader: %w", err)
	}

	contentReader := io.MultiReader(
		bytes.NewBuffer(nonce[:]),
		encrypted,
	)

	return es.store.Put(objectKey, contentReader)
}

func (es *EncryptedStore) Remove(path string) error {
	return es.store.Remove(path)
}

func generateNonce() ([nonceSize]byte, error) {
	var nonce [nonceSize]byte

	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return nonce, fmt.Errorf("could not read random data: %w", err)
	}

	return nonce, nil
}

func deriveKey(key []byte, nonce []byte) ([keySize]byte, error) {
	var encryptionKey [keySize]byte

	kdf := hkdf.New(sha256.New, key, nonce, nil)
	_, err := io.ReadFull(kdf, key[:])
	if err != nil {
		return encryptionKey, fmt.Errorf("could not derive key: %w", err)
	}

	return encryptionKey, nil
}

func New(key []byte, backend objectstorage.Store) objectstorage.Store {
	es := &EncryptedStore{
		key:   key,
		store: backend,
	}

	return es
}
