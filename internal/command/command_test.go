package command

import (
	"testing"

	ds "github.com/RPW-11/redis-with-go/internal/data_structures"
	"github.com/RPW-11/redis-with-go/internal/resp"
)

func bulkVal(s string) *resp.Value {
	return &resp.Value{
		Typ:   resp.BulkStringType,
		Bytes: []byte(s),
	}
}

func cmdArr(cmd string, args ...string) []*resp.Value {
	arr := make([]*resp.Value, 0, 1+len(args))
	arr = append(arr, bulkVal(cmd))
	for _, a := range args {
		arr = append(arr, bulkVal(a))
	}
	return arr
}

func TestHandleSetCmd(t *testing.T) {
	t.Run("set key and value", func(t *testing.T) {
		m := ds.NewRedisMap()
		err := handleSetCmd(cmdArr("SET", "hello", "world"), m)
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
		err := handleSetCmd(cmdArr("SET", "hello", "new"), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, _ := m.Get("hello")
		if string(v) != "new" {
			t.Fatalf("expected 'new', got %q", v)
		}
	})

	// Redis allows SET key "" — an empty value is valid
	t.Run("empty value is valid per Redis spec", func(t *testing.T) {
		m := ds.NewRedisMap()
		err := handleSetCmd(cmdArr("SET", "k", ""), m)
		if err != nil {
			t.Fatalf("empty value should be allowed, got error: %v", err)
		}
		v, ok := m.Get("k")
		if !ok {
			t.Fatal("expected key to exist")
		}
		if len(v) != 0 {
			t.Fatalf("expected empty value, got %q", v)
		}
	})

	t.Run("too few arguments", func(t *testing.T) {
		m := ds.NewRedisMap()
		err := handleSetCmd(cmdArr("SET", "hello"), m)
		if err == nil {
			t.Fatal("expected error for missing value argument")
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		m := ds.NewRedisMap()
		err := handleSetCmd(cmdArr("SET", "hello", "world", "extra"), m)
		if err == nil {
			t.Fatal("expected error for extra argument")
		}
	})

	t.Run("empty key", func(t *testing.T) {
		m := ds.NewRedisMap()
		err := handleSetCmd(cmdArr("SET", "", "world"), m)
		if err == nil {
			t.Fatal("expected error for empty key")
		}
	})

	t.Run("key is not a bulk string", func(t *testing.T) {
		m := ds.NewRedisMap()
		arr := []*resp.Value{
			bulkVal("SET"),
			{Typ: resp.StringType, Str: "hello"},
			bulkVal("world"),
		}
		err := handleSetCmd(arr, m)
		if err == nil {
			t.Fatal("expected error for non-bulk-string key")
		}
	})

	t.Run("value is not a bulk string", func(t *testing.T) {
		m := ds.NewRedisMap()
		arr := []*resp.Value{
			bulkVal("SET"),
			bulkVal("hello"),
			{Typ: resp.StringType, Str: "world"},
		}
		err := handleSetCmd(arr, m)
		if err == nil {
			t.Fatal("expected error for non-bulk-string value")
		}
	})
}

func TestHandleGetCmd(t *testing.T) {
	t.Run("key exists", func(t *testing.T) {
		m := ds.NewRedisMap()
		m.Set("hello", []byte("world"))
		v, err := handleGetCmd(cmdArr("GET", "hello"), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(v) != "world" {
			t.Fatalf("expected 'world', got %q", v)
		}
	})

	t.Run("key does not exist returns nil", func(t *testing.T) {
		m := ds.NewRedisMap()
		v, err := handleGetCmd(cmdArr("GET", "missing"), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != nil {
			t.Fatalf("expected nil for missing key, got %q", v)
		}
	})

	t.Run("too few arguments", func(t *testing.T) {
		m := ds.NewRedisMap()
		_, err := handleGetCmd(cmdArr("GET"), m)
		if err == nil {
			t.Fatal("expected error for missing key argument")
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		m := ds.NewRedisMap()
		_, err := handleGetCmd(cmdArr("GET", "k1", "k2"), m)
		if err == nil {
			t.Fatal("expected error for extra argument")
		}
	})

	t.Run("key is not a bulk string", func(t *testing.T) {
		m := ds.NewRedisMap()
		arr := []*resp.Value{
			bulkVal("GET"),
			{Typ: resp.StringType, Str: "hello"},
		}
		_, err := handleGetCmd(arr, m)
		if err == nil {
			t.Fatal("expected error for non-bulk-string key")
		}
	})
}

func TestHandleDelCmd(t *testing.T) {
	t.Run("delete existing key", func(t *testing.T) {
		m := ds.NewRedisMap()
		m.Set("hello", []byte("world"))
		err := handleDelCmd(cmdArr("DEL", "hello"), m)
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
		err := handleDelCmd(cmdArr("DEL", "ghost"), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("too few arguments", func(t *testing.T) {
		m := ds.NewRedisMap()
		err := handleDelCmd(cmdArr("DEL"), m)
		if err == nil {
			t.Fatal("expected error for missing key argument")
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		m := ds.NewRedisMap()
		err := handleDelCmd(cmdArr("DEL", "k1", "k2"), m)
		if err == nil {
			t.Fatal("expected error for extra argument")
		}
	})

	t.Run("key is not a bulk string", func(t *testing.T) {
		m := ds.NewRedisMap()
		arr := []*resp.Value{
			bulkVal("DEL"),
			{Typ: resp.StringType, Str: "hello"},
		}
		err := handleDelCmd(arr, m)
		if err == nil {
			t.Fatal("expected error for non-bulk-string key")
		}
	})
}
