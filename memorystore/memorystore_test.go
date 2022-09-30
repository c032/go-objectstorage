package memorystore_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/c032/go-objectstorage"
	"github.com/c032/go-objectstorage/memorystore"
)

func TestMemoryStore(t *testing.T) {
	ms := memorystore.New()

	const key = "test"
	buf := &bytes.Buffer{}

	var err error

	_, err = ms.Read(key, buf)
	if err == nil {
		t.Fatalf("ms.Read(key, buf) returns nil error; want non-nil error")
	} else if !errors.Is(err, objectstorage.ErrNotFound) {
		t.Fatal(err)
	}

	buf.Reset()

	const testContent = "Hello, world!"

	_, err = buf.WriteString(testContent)
	if err != nil {
		t.Fatal(err)
	}

	err = ms.Put(key, buf)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := buf.Len(), 0; got != want {
		t.Fatalf("buf.Len() = %#v; want %#v", got, want)
	}

	buf.Reset()

	_, err = ms.Read(key, buf)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := buf.String(), testContent; got != want {
		t.Fatalf("buf.String() = %#v; want %#v", got, want)
	}

	err = ms.Remove(key)
	if err != nil {
		t.Fatal(err)
	}

	buf.Reset()

	_, err = ms.Read(key, buf)
	if err == nil {
		t.Fatalf("ms.Read(key, buf) returns nil error; want non-nil error")
	} else if !errors.Is(err, objectstorage.ErrNotFound) {
		t.Fatal(err)
	}
}
