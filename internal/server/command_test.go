package server

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	ds "github.com/RPW-11/redis-with-go/internal/data_structures"
)

func bulkStr(s string) string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)
}

func newCommandReader(s string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(s))
}

func TestHandleSetCmd(t *testing.T) {
	t.Run("set key and value", func(t *testing.T) {
		m := ds.NewRedisMap()
		rd := newCommandReader(bulkStr("hello") + bulkStr("world"))
		err := handleSetCmd(rd, m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := m.Get("hello")
		if !ok {
			t.Fatal("expected key 'hello' to exist")
		}
		if string(v) != "world" {
			t.Fatalf("expected 'world', got %q", v)
		}
	})

	t.Run("overwrites existing key", func(t *testing.T) {
		m := ds.NewRedisMap()
		m.Set("hello", []byte("old"))
		rd := newCommandReader(bulkStr("hello") + bulkStr("new"))
		err := handleSetCmd(rd, m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, _ := m.Get("hello")
		if string(v) != "new" {
			t.Fatalf("expected 'new', got %q", v)
		}
	})

	t.Run("empty value", func(t *testing.T) {
		m := ds.NewRedisMap()
		rd := newCommandReader(bulkStr("k") + bulkStr(""))
		err := handleSetCmd(rd, m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := m.Get("k")
		if !ok {
			t.Fatal("expected key to exist")
		}
		if len(v) != 0 {
			t.Fatalf("expected empty value, got %q", v)
		}
	})

	t.Run("missing key bulk string", func(t *testing.T) {
		m := ds.NewRedisMap()
		rd := newCommandReader("")
		err := handleSetCmd(rd, m)
		if err == nil {
			t.Fatal("expected error for missing key")
		}
	})

	t.Run("missing value bulk string", func(t *testing.T) {
		m := ds.NewRedisMap()
		rd := newCommandReader(bulkStr("hello"))
		err := handleSetCmd(rd, m)
		if err == nil {
			t.Fatal("expected error for missing value")
		}
	})

	t.Run("negative length", func(t *testing.T) {
		m := ds.NewRedisMap()
		rd := newCommandReader("$-1\r\n")
		err := handleSetCmd(rd, m)
		if err == nil {
			t.Fatal("expected error for negative length")
		}
	})

	t.Run("length exceeds max", func(t *testing.T) {
		m := ds.NewRedisMap()
		rd := newCommandReader(fmt.Sprintf("$%d\r\n", MaxBulkStringLength+1))
		err := handleSetCmd(rd, m)
		if err == nil {
			t.Fatal("expected error for length exceeding max")
		}
	})

	t.Run("wrong type — not a bulk string", func(t *testing.T) {
		m := ds.NewRedisMap()
		rd := newCommandReader("+hello\r\n")
		err := handleSetCmd(rd, m)
		if err == nil {
			t.Fatal("expected error for non-bulk-string type")
		}
	})
}

func TestHandleGetCmd(t *testing.T) {
	t.Run("key exists", func(t *testing.T) {
		m := ds.NewRedisMap()
		m.Set("hello", []byte("world"))
		rd := newCommandReader(bulkStr("hello"))
		v, err := handleGetCmd(rd, m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(v) != "world" {
			t.Fatalf("expected 'world', got %q", v)
		}
	})

	t.Run("key does not exist", func(t *testing.T) {
		m := ds.NewRedisMap()
		rd := newCommandReader(bulkStr("missing"))
		v, err := handleGetCmd(rd, m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != nil {
			t.Fatalf("expected nil for missing key, got %q", v)
		}
	})

	t.Run("missing key bulk string", func(t *testing.T) {
		m := ds.NewRedisMap()
		rd := newCommandReader("")
		_, err := handleGetCmd(rd, m)
		if err == nil {
			t.Fatal("expected error for missing key")
		}
	})

	t.Run("negative length", func(t *testing.T) {
		m := ds.NewRedisMap()
		rd := newCommandReader("$-1\r\n")
		_, err := handleGetCmd(rd, m)
		if err == nil {
			t.Fatal("expected error for negative length")
		}
	})

	t.Run("wrong type — not a bulk string", func(t *testing.T) {
		m := ds.NewRedisMap()
		rd := newCommandReader("+hello\r\n")
		_, err := handleGetCmd(rd, m)
		if err == nil {
			t.Fatal("expected error for non-bulk-string type")
		}
	})
}

func TestHandleDelCmd(t *testing.T) {
	t.Run("delete existing key", func(t *testing.T) {
		m := ds.NewRedisMap()
		m.Set("hello", []byte("world"))
		rd := newCommandReader(bulkStr("hello"))
		err := handleDelCmd(rd, m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, ok := m.Get("hello")
		if ok {
			t.Fatal("expected key to be deleted")
		}
	})

	t.Run("delete non-existing key does not error", func(t *testing.T) {
		m := ds.NewRedisMap()
		rd := newCommandReader(bulkStr("ghost"))
		err := handleDelCmd(rd, m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("missing key bulk string", func(t *testing.T) {
		m := ds.NewRedisMap()
		rd := newCommandReader("")
		err := handleDelCmd(rd, m)
		if err == nil {
			t.Fatal("expected error for missing key")
		}
	})

	t.Run("negative length", func(t *testing.T) {
		m := ds.NewRedisMap()
		rd := newCommandReader("$-1\r\n")
		err := handleDelCmd(rd, m)
		if err == nil {
			t.Fatal("expected error for negative length")
		}
	})

	t.Run("wrong type — not a bulk string", func(t *testing.T) {
		m := ds.NewRedisMap()
		rd := newCommandReader("+hello\r\n")
		err := handleDelCmd(rd, m)
		if err == nil {
			t.Fatal("expected error for non-bulk-string type")
		}
	})
}
