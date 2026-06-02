package command

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/RPW-11/redis-with-go/internal/resp"
	"github.com/RPW-11/redis-with-go/internal/store"
)

func newTestEngine(t *testing.T) *store.Store {
	t.Helper()
	m, err := store.NewStore(100, "")
	if err != nil {
		t.Fatalf("failed to create LRUEngine: %v", err)
	}
	return m
}

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
		m := newTestEngine(t)
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
		m := newTestEngine(t)
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
		m := newTestEngine(t)
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
		m := newTestEngine(t)
		err := handleSetCmd(cmdArr("SET", "hello"), m)
		if err == nil {
			t.Fatal("expected error for missing value argument")
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		m := newTestEngine(t)
		err := handleSetCmd(cmdArr("SET", "hello", "world", "extra"), m)
		if err == nil {
			t.Fatal("expected error for extra argument")
		}
	})

	t.Run("empty key", func(t *testing.T) {
		m := newTestEngine(t)
		err := handleSetCmd(cmdArr("SET", "", "world"), m)
		if err == nil {
			t.Fatal("expected error for empty key")
		}
	})

	t.Run("key is not a bulk string", func(t *testing.T) {
		m := newTestEngine(t)
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
		m := newTestEngine(t)
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
		m := newTestEngine(t)
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
		m := newTestEngine(t)
		v, err := handleGetCmd(cmdArr("GET", "missing"), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != nil {
			t.Fatalf("expected nil for missing key, got %q", v)
		}
	})

	t.Run("too few arguments", func(t *testing.T) {
		m := newTestEngine(t)
		_, err := handleGetCmd(cmdArr("GET"), m)
		if err == nil {
			t.Fatal("expected error for missing key argument")
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		m := newTestEngine(t)
		_, err := handleGetCmd(cmdArr("GET", "k1", "k2"), m)
		if err == nil {
			t.Fatal("expected error for extra argument")
		}
	})

	t.Run("key is not a bulk string", func(t *testing.T) {
		m := newTestEngine(t)
		arr := []*resp.Value{
			bulkVal("GET"),
			{Typ: resp.StringType, Str: "hello"},
		}
		_, err := handleGetCmd(arr, m)
		if err == nil {
			t.Fatal("expected error for non-bulk-string key")
		}
	})

	t.Run("expired key returns nil", func(t *testing.T) {
		m := newTestEngine(t)
		m.Set("k", []byte("v"))
		m.SetExpiry("k", time.Now().Add(-1*time.Second))
		v, err := handleGetCmd(cmdArr("GET", "k"), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != nil {
			t.Fatalf("expected nil for expired key, got %q", v)
		}
	})

	t.Run("non-expired key returns value", func(t *testing.T) {
		m := newTestEngine(t)
		m.Set("k", []byte("v"))
		m.SetExpiry("k", time.Now().Add(10*time.Second))
		v, err := handleGetCmd(cmdArr("GET", "k"), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(v) != "v" {
			t.Fatalf("expected 'v', got %q", v)
		}
	})
}

func TestHandleDelCmd(t *testing.T) {
	t.Run("delete existing key", func(t *testing.T) {
		m := newTestEngine(t)
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
		m := newTestEngine(t)
		err := handleDelCmd(cmdArr("DEL", "ghost"), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("too few arguments", func(t *testing.T) {
		m := newTestEngine(t)
		err := handleDelCmd(cmdArr("DEL"), m)
		if err == nil {
			t.Fatal("expected error for missing key argument")
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		m := newTestEngine(t)
		err := handleDelCmd(cmdArr("DEL", "k1", "k2"), m)
		if err == nil {
			t.Fatal("expected error for extra argument")
		}
	})

	t.Run("key is not a bulk string", func(t *testing.T) {
		m := newTestEngine(t)
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

func TestHandleExpireCmd(t *testing.T) {
	t.Run("sets expiry on existing key", func(t *testing.T) {
		m := newTestEngine(t)
		m.Set("k", []byte("v"))
		before := time.Now()
		ok, err := handleExpireCmd(cmdArr("EXPIRE", "k", "10"), m)
		after := time.Now()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatal("expected true for existing key")
		}
		expiry, _ := m.ExpiryOf("k")
		if expiry.Before(before.Add(10*time.Second)) || expiry.After(after.Add(10*time.Second)) {
			t.Fatalf("expiry %v out of expected range", expiry)
		}
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		m := newTestEngine(t)
		ok, err := handleExpireCmd(cmdArr("EXPIRE", "missing", "10"), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatal("expected false for missing key")
		}
	})

	t.Run("zero seconds is valid", func(t *testing.T) {
		m := newTestEngine(t)
		m.Set("k", []byte("v"))
		ok, err := handleExpireCmd(cmdArr("EXPIRE", "k", "0"), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatal("expected true for existing key")
		}
	})

	t.Run("negative seconds rejected", func(t *testing.T) {
		m := newTestEngine(t)
		m.Set("k", []byte("v"))
		_, err := handleExpireCmd(cmdArr("EXPIRE", "k", "-1"), m)
		if err == nil {
			t.Fatal("expected error for negative seconds")
		}
	})

	t.Run("non-integer seconds rejected", func(t *testing.T) {
		m := newTestEngine(t)
		m.Set("k", []byte("v"))
		_, err := handleExpireCmd(cmdArr("EXPIRE", "k", "abc"), m)
		if err == nil {
			t.Fatal("expected error for non-integer seconds")
		}
	})

	t.Run("too few arguments", func(t *testing.T) {
		m := newTestEngine(t)
		_, err := handleExpireCmd(cmdArr("EXPIRE", "k"), m)
		if err == nil {
			t.Fatal("expected error for missing seconds argument")
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		m := newTestEngine(t)
		_, err := handleExpireCmd(cmdArr("EXPIRE", "k", "10", "extra"), m)
		if err == nil {
			t.Fatal("expected error for extra argument")
		}
	})

	t.Run("key is not a bulk string", func(t *testing.T) {
		m := newTestEngine(t)
		arr := []*resp.Value{
			bulkVal("EXPIRE"),
			{Typ: resp.StringType, Str: "k"},
			bulkVal("10"),
		}
		_, err := handleExpireCmd(arr, m)
		if err == nil {
			t.Fatal("expected error for non-bulk-string key")
		}
	})

	t.Run("seconds is not a bulk string", func(t *testing.T) {
		m := newTestEngine(t)
		arr := []*resp.Value{
			bulkVal("EXPIRE"),
			bulkVal("k"),
			{Typ: resp.StringType, Str: "10"},
		}
		_, err := handleExpireCmd(arr, m)
		if err == nil {
			t.Fatal("expected error for non-bulk-string seconds")
		}
	})
}

func unixStr(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10)
}

func TestHandleExpireAtCmd(t *testing.T) {
	t.Run("sets expiry on existing key", func(t *testing.T) {
		m := newTestEngine(t)
		m.Set("k", []byte("v"))
		expireAt := time.Now().Add(10 * time.Second)
		ok, err := handleExpireAtCmd(cmdArr("EXPIREAT", "k", unixStr(expireAt)), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatal("expected true for existing key")
		}
		expiry, _ := m.ExpiryOf("k")
		if expiry.Unix() != expireAt.Unix() {
			t.Fatalf("expected expiry %v, got %v", expireAt.Unix(), expiry.Unix())
		}
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		m := newTestEngine(t)
		ok, err := handleExpireAtCmd(cmdArr("EXPIREAT", "missing", unixStr(time.Now().Add(10*time.Second))), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatal("expected false for missing key")
		}
	})

	t.Run("past timestamp deletes existing key", func(t *testing.T) {
		m := newTestEngine(t)
		m.Set("k", []byte("v"))
		ok, err := handleExpireAtCmd(cmdArr("EXPIREAT", "k", unixStr(time.Now().Add(-1*time.Second))), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatal("expected true for existing key with past timestamp")
		}
		_, exists := m.Get("k")
		if exists {
			t.Fatal("expected key to be deleted after past timestamp")
		}
	})

	t.Run("past timestamp returns false for missing key", func(t *testing.T) {
		m := newTestEngine(t)
		ok, err := handleExpireAtCmd(cmdArr("EXPIREAT", "missing", unixStr(time.Now().Add(-1*time.Second))), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatal("expected false for missing key with past timestamp")
		}
	})

	t.Run("negative timestamp rejected", func(t *testing.T) {
		m := newTestEngine(t)
		m.Set("k", []byte("v"))
		_, err := handleExpireAtCmd(cmdArr("EXPIREAT", "k", "-1"), m)
		if err == nil {
			t.Fatal("expected error for negative timestamp")
		}
	})

	t.Run("non-integer timestamp rejected", func(t *testing.T) {
		m := newTestEngine(t)
		m.Set("k", []byte("v"))
		_, err := handleExpireAtCmd(cmdArr("EXPIREAT", "k", "abc"), m)
		if err == nil {
			t.Fatal("expected error for non-integer timestamp")
		}
	})

	t.Run("too few arguments", func(t *testing.T) {
		m := newTestEngine(t)
		_, err := handleExpireAtCmd(cmdArr("EXPIREAT", "k"), m)
		if err == nil {
			t.Fatal("expected error for missing timestamp argument")
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		m := newTestEngine(t)
		_, err := handleExpireAtCmd(cmdArr("EXPIREAT", "k", unixStr(time.Now().Add(10*time.Second)), "extra"), m)
		if err == nil {
			t.Fatal("expected error for extra argument")
		}
	})

	t.Run("key is not a bulk string", func(t *testing.T) {
		m := newTestEngine(t)
		arr := []*resp.Value{
			bulkVal("EXPIREAT"),
			{Typ: resp.StringType, Str: "k"},
			bulkVal(unixStr(time.Now().Add(10 * time.Second))),
		}
		_, err := handleExpireAtCmd(arr, m)
		if err == nil {
			t.Fatal("expected error for non-bulk-string key")
		}
	})

	t.Run("timestamp is not a bulk string", func(t *testing.T) {
		m := newTestEngine(t)
		arr := []*resp.Value{
			bulkVal("EXPIREAT"),
			bulkVal("k"),
			{Typ: resp.StringType, Str: unixStr(time.Now().Add(10 * time.Second))},
		}
		_, err := handleExpireAtCmd(arr, m)
		if err == nil {
			t.Fatal("expected error for non-bulk-string timestamp")
		}
	})
}

func TestHandleTtlCmd(t *testing.T) {
	t.Run("returns -2 for missing key", func(t *testing.T) {
		m := newTestEngine(t)
		sec, err := handleTtlCmd(cmdArr("TTL", "missing"), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sec != -2 {
			t.Fatalf("expected -2 for missing key, got %d", sec)
		}
	})

	t.Run("returns -1 for key with no expiry", func(t *testing.T) {
		m := newTestEngine(t)
		m.Set("k", []byte("v"))
		sec, err := handleTtlCmd(cmdArr("TTL", "k"), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sec != -1 {
			t.Fatalf("expected -1 for key with no expiry, got %d", sec)
		}
	})

	t.Run("returns remaining seconds for key with expiry", func(t *testing.T) {
		m := newTestEngine(t)
		m.Set("k", []byte("v"))
		handleExpireCmd(cmdArr("EXPIRE", "k", "10"), m)
		sec, err := handleTtlCmd(cmdArr("TTL", "k"), m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sec < 9 || sec > 10 {
			t.Fatalf("expected TTL ~10, got %d", sec)
		}
	})

	t.Run("too few arguments", func(t *testing.T) {
		m := newTestEngine(t)
		_, err := handleTtlCmd(cmdArr("TTL"), m)
		if err == nil {
			t.Fatal("expected error for missing key argument")
		}
	})

	t.Run("too many arguments", func(t *testing.T) {
		m := newTestEngine(t)
		_, err := handleTtlCmd(cmdArr("TTL", "k", "extra"), m)
		if err == nil {
			t.Fatal("expected error for extra argument")
		}
	})

	t.Run("key is not a bulk string", func(t *testing.T) {
		m := newTestEngine(t)
		arr := []*resp.Value{
			bulkVal("TTL"),
			{Typ: resp.StringType, Str: "k"},
		}
		_, err := handleTtlCmd(arr, m)
		if err == nil {
			t.Fatal("expected error for non-bulk-string key")
		}
	})
}

func TestHandle(t *testing.T) {
	t.Run("cancelled context returns error value", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		m := newTestEngine(t)
		v := Handle(ctx, strings.NewReader(""), m)
		if v == nil || v.Typ != resp.ErrorType {
			t.Fatalf("expected error value for cancelled ctx, got %v", v)
		}
	})

	t.Run("active context processes SET and returns OK", func(t *testing.T) {
		ctx := context.Background()
		m := newTestEngine(t)

		r := strings.NewReader("*3\r\n$3\r\nSET\r\n$5\r\nhello\r\n$5\r\nworld\r\n")
		v := Handle(ctx, r, m)
		if v == nil || v.Typ != resp.StringType || v.Str != "OK" {
			t.Fatalf("expected +OK response, got %v", v)
		}
		val, ok := m.Get("hello")
		if !ok || string(val) != "world" {
			t.Fatalf("expected key 'hello'='world' in store after SET")
		}
	})
}
