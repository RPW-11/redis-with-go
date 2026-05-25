package datastructures

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewRedisMap(t *testing.T) {
	rm := NewRedisMap()
	if rm == nil {
		t.Fatal("expected non-nil RedisMap")
	}
	if rm.Len() != 0 {
		t.Fatalf("expected length 0, got %d", rm.Len())
	}
}

func TestRedisMapGet(t *testing.T) {
	t.Run("existing key", func(t *testing.T) {
		rm := NewRedisMap()
		rm.Set("key", []byte("value"))
		v, ok := rm.Get("key")
		if !ok {
			t.Fatal("expected ok=true")
		}
		if string(v) != "value" {
			t.Fatalf("expected 'value', got %q", v)
		}
	})

	t.Run("non-existing key", func(t *testing.T) {
		rm := NewRedisMap()
		v, ok := rm.Get("missing")
		if ok {
			t.Fatal("expected ok=false")
		}
		if v != nil {
			t.Fatalf("expected nil, got %v", v)
		}
	})

	t.Run("returns a copy — mutation does not affect stored value", func(t *testing.T) {
		rm := NewRedisMap()
		rm.Set("key", []byte("hello"))
		v, _ := rm.Get("key")
		v[0] = 'X'
		v2, _ := rm.Get("key")
		if string(v2) != "hello" {
			t.Fatalf("stored value was mutated: got %q", v2)
		}
	})

	t.Run("empty value", func(t *testing.T) {
		rm := NewRedisMap()
		rm.Set("key", []byte{})
		v, ok := rm.Get("key")
		if !ok {
			t.Fatal("expected ok=true")
		}
		if len(v) != 0 {
			t.Fatalf("expected empty slice, got %v", v)
		}
	})

	t.Run("empty key", func(t *testing.T) {
		rm := NewRedisMap()
		rm.Set("", []byte("empty-key"))
		v, ok := rm.Get("")
		if !ok {
			t.Fatal("expected ok=true for empty key")
		}
		if string(v) != "empty-key" {
			t.Fatalf("expected 'empty-key', got %q", v)
		}
	})
}

func TestRedisMapSet(t *testing.T) {
	t.Run("set and retrieve", func(t *testing.T) {
		rm := NewRedisMap()
		rm.Set("k", []byte("v"))
		v, ok := rm.Get("k")
		if !ok || string(v) != "v" {
			t.Fatalf("expected 'v', got %q ok=%v", v, ok)
		}
	})

	t.Run("overwrites existing key", func(t *testing.T) {
		rm := NewRedisMap()
		rm.Set("k", []byte("first"))
		rm.Set("k", []byte("second"))
		v, _ := rm.Get("k")
		if string(v) != "second" {
			t.Fatalf("expected 'second', got %q", v)
		}
	})

	t.Run("stores a copy — mutation of input does not affect stored value", func(t *testing.T) {
		rm := NewRedisMap()
		input := []byte("hello")
		rm.Set("k", input)
		input[0] = 'X'
		v, _ := rm.Get("k")
		if string(v) != "hello" {
			t.Fatalf("stored value was affected by input mutation: got %q", v)
		}
	})
}

func TestRedisMapSetIfNotExists(t *testing.T) {
	t.Run("key does not exist — sets and returns true", func(t *testing.T) {
		rm := NewRedisMap()
		ok := rm.SetIfNotExists("k", []byte("v"))
		if !ok {
			t.Fatal("expected true for new key")
		}
		v, exists := rm.Get("k")
		if !exists || string(v) != "v" {
			t.Fatalf("expected 'v', got %q", v)
		}
	})

	t.Run("key exists — does not overwrite and returns false", func(t *testing.T) {
		rm := NewRedisMap()
		rm.Set("k", []byte("original"))
		ok := rm.SetIfNotExists("k", []byte("new"))
		if ok {
			t.Fatal("expected false for existing key")
		}
		v, _ := rm.Get("k")
		if string(v) != "original" {
			t.Fatalf("expected 'original' to be unchanged, got %q", v)
		}
	})
}

func TestRedisMapDelete(t *testing.T) {
	t.Run("delete existing key", func(t *testing.T) {
		rm := NewRedisMap()
		rm.Set("k", []byte("v"))
		rm.Delete("k")
		_, ok := rm.Get("k")
		if ok {
			t.Fatal("expected key to be deleted")
		}
	})

	t.Run("delete non-existing key does not panic", func(t *testing.T) {
		rm := NewRedisMap()
		rm.Delete("ghost")
	})

	t.Run("len decrements after delete", func(t *testing.T) {
		rm := NewRedisMap()
		rm.Set("a", []byte("1"))
		rm.Set("b", []byte("2"))
		rm.Delete("a")
		if rm.Len() != 1 {
			t.Fatalf("expected length 1, got %d", rm.Len())
		}
	})
}

func TestRedisMapLen(t *testing.T) {
	t.Run("empty map", func(t *testing.T) {
		rm := NewRedisMap()
		if rm.Len() != 0 {
			t.Fatalf("expected 0, got %d", rm.Len())
		}
	})

	t.Run("after sets", func(t *testing.T) {
		rm := NewRedisMap()
		rm.Set("a", []byte("1"))
		rm.Set("b", []byte("2"))
		rm.Set("c", []byte("3"))
		if rm.Len() != 3 {
			t.Fatalf("expected 3, got %d", rm.Len())
		}
	})

	t.Run("overwrite does not increase len", func(t *testing.T) {
		rm := NewRedisMap()
		rm.Set("k", []byte("v1"))
		rm.Set("k", []byte("v2"))
		if rm.Len() != 1 {
			t.Fatalf("expected 1, got %d", rm.Len())
		}
	})
}

func TestRedisMapConcurrent(t *testing.T) {
	t.Run("concurrent reads and writes", func(t *testing.T) {
		rm := NewRedisMap()
		var wg sync.WaitGroup
		const goroutines = 100

		for i := range goroutines {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				rm.Set(fmt.Sprintf("key%d", i), []byte(fmt.Sprintf("val%d", i)))
			}(i)
		}

		for i := range goroutines {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				rm.Get(fmt.Sprintf("key%d", i))
			}(i)
		}

		wg.Wait()
	})

	t.Run("concurrent SetIfNotExists — only one wins per key", func(t *testing.T) {
		rm := NewRedisMap()
		var wg sync.WaitGroup
		wins := make(chan bool, 100)

		for range 100 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ok := rm.SetIfNotExists("shared", []byte("value"))
				if ok {
					wins <- true
				}
			}()
		}

		wg.Wait()
		close(wins)

		count := len(wins)
		if count != 1 {
			t.Fatalf("expected exactly 1 goroutine to win SetIfNotExists, got %d", count)
		}
	})

	t.Run("concurrent deletes and reads", func(t *testing.T) {
		rm := NewRedisMap()
		for i := range 50 {
			rm.Set(fmt.Sprintf("key%d", i), []byte("v"))
		}

		var wg sync.WaitGroup
		for i := range 50 {
			wg.Add(2)
			go func(i int) {
				defer wg.Done()
				rm.Delete(fmt.Sprintf("key%d", i))
			}(i)
			go func(i int) {
				defer wg.Done()
				rm.Get(fmt.Sprintf("key%d", i))
			}(i)
		}
		wg.Wait()
	})
}
