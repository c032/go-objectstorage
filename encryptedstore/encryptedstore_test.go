package encryptedstore_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/c032/go-objectstorage"
	"github.com/c032/go-objectstorage/encryptedstore"
	"github.com/c032/go-objectstorage/memorystore"
)

func TestEncryptedStore(t *testing.T) {
	ms := memorystore.New()
	es := encryptedstore.New([]byte("correct horse battery staple"), ms)

	const key = "test"
	buf := &bytes.Buffer{}

	var err error

	_, err = es.Read(key, buf)
	if err == nil {
		t.Fatalf("es.Read(key, buf) returns nil error; want non-nil error")
	} else if !errors.Is(err, objectstorage.ErrNotFound) {
		t.Fatal(err)
	}

	const testContent = "Hello, world!"

	_, err = buf.WriteString(testContent)
	if err != nil {
		t.Fatal(err)
	}

	err = es.Put(key, buf)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := buf.Len(), 0; got != want {
		t.Fatalf("buf.Len() = %#v; want %#v", got, want)
	}

	buf.Reset()

	_, err = es.Read(key, buf)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := buf.String(), testContent; got != want {
		t.Fatalf("buf.String() = %#v; want %#v", got, want)
	}

	buf.Reset()

	_, err = ms.Read(key, buf)
	if err != nil {
		t.Fatal(err)
	}

	if got, notWant := buf.String(), testContent; got == notWant {
		t.Fatalf("buf.String() = %#v; want a different value", got)
	}

	err = es.Remove(key)
	if err != nil {
		t.Fatal(err)
	}

	buf.Reset()

	_, err = es.Read(key, buf)
	if err == nil {
		t.Fatalf("es.Read(key, buf) returns nil error; want non-nil error")
	} else if !errors.Is(err, objectstorage.ErrNotFound) {
		t.Fatal(err)
	}
}
