package encryptedstore

import (
	"bytes"
	"fmt"
	"io"

	"github.com/minio/sio"
)

type decryptor struct {
	key []byte

	r *io.PipeReader
	w *io.PipeWriter

	nonce               []byte
	encryptedBuffer     *bytes.Buffer
	decryptionKey       [keySize]byte
	writerForDecryption io.WriteCloser
}

func (d *decryptor) Read(p []byte) (int, error) {
	return d.r.Read(p)
}

func (d *decryptor) Write(p []byte) (int, error) {
	var (
		err error

		n int
	)

	if d.writerForDecryption != nil {
		return d.writerForDecryption.Write(p)
	}

	n, err = d.encryptedBuffer.Write(p)
	if err != nil {
		return n, fmt.Errorf("could not write: %w", err)
	}

	if d.encryptedBuffer.Len() < nonceSize {
		return n, nil
	}

	d.nonce = d.encryptedBuffer.Next(nonceSize)

	var decryptionKey [keySize]byte

	decryptionKey, err = deriveKey(d.key, d.nonce)
	if err != nil {
		return n, fmt.Errorf("could not derive key: %w", err)
	}

	cfg := sio.Config{
		Key: decryptionKey[:],
	}

	var wfd io.WriteCloser

	wfd, err = sio.DecryptWriter(d.w, cfg)
	if err != nil {
		return n, fmt.Errorf("could not decrypt writer: %w", err)
	}

	d.writerForDecryption = wfd

	// Decrypt content already in buffer.
	if d.encryptedBuffer.Len() > 0 {
		_, err = d.encryptedBuffer.WriteTo(d.writerForDecryption)
		if err != nil {
			return n, err
		}
	}

	d.encryptedBuffer = nil

	return n, nil
}

func (d *decryptor) Close() error {
	var err error
	if d.writerForDecryption == nil {
		// TODO: Better error message.
		err = fmt.Errorf("not enough data to decrypt")

		return d.w.CloseWithError(err)
	}

	err = d.writerForDecryption.Close()
	if err != nil {
		return d.w.CloseWithError(err)
	}
	d.writerForDecryption = nil

	return nil
}

func newDecryptor(key []byte) *decryptor {
	r, w := io.Pipe()

	return &decryptor{
		key: key,

		r: r,
		w: w,

		encryptedBuffer: &bytes.Buffer{},
	}
}
