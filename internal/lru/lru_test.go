package lru

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func newEngine(t *testing.T, cap int) *LRUEngine {
	t.Helper()
	e, err := NewLRUEngine(cap)
	if err != nil {
		t.Fatalf("NewLRUEngine(%d): %v", cap, err)
	}
	return e
}

// --- Constructor ---

func TestNewLRUEngine(t *testing.T) {
	t.Run("valid capacity", func(t *testing.T) {
		e, err := NewLRUEngine(10)
		if err != nil || e == nil {
			t.Fatalf("expected valid engine, got err=%v", err)
		}
	})

	t.Run("zero capacity rejected", func(t *testing.T) {
		_, err := NewLRUEngine(0)
		if err == nil {
			t.Fatal("expected error for zero capacity")
		}
	})

	t.Run("negative capacity rejected", func(t *testing.T) {
		_, err := NewLRUEngine(-1)
		if err == nil {
			t.Fatal("expected error for negative capacity")
		}
	})

	t.Run("exceeds MaxCapacity rejected", func(t *testing.T) {
		_, err := NewLRUEngine(MaxCapacity + 1)
		if err == nil {
			t.Fatal("expected error for capacity exceeding MaxCapacity")
		}
	})

	t.Run("MaxCapacity accepted", func(t *testing.T) {
		_, err := NewLRUEngine(MaxCapacity)
		if err != nil {
			t.Fatalf("expected MaxCapacity to be valid, got: %v", err)
		}
	})
}

// --- Exist ---

func TestExist(t *testing.T) {
	t.Run("returns true for existing key", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Set("k", []byte("v"))
		if !e.Exist("k") {
			t.Fatal("expected true for existing key")
		}
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		e := newEngine(t, 10)
		if e.Exist("missing") {
			t.Fatal("expected false for missing key")
		}
	})

	t.Run("returns false after delete", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Set("k", []byte("v"))
		e.Delete("k")
		if e.Exist("k") {
			t.Fatal("expected false after delete")
		}
	})

	t.Run("returns false after eviction", func(t *testing.T) {
		e := newEngine(t, 1)
		e.Set("a", []byte("1"))
		e.Set("b", []byte("2")) // evicts a
		if e.Exist("a") {
			t.Fatal("expected false for evicted key")
		}
		if !e.Exist("b") {
			t.Fatal("expected true for current key")
		}
	})

	t.Run("does not promote to head", func(t *testing.T) {
		e := newEngine(t, 3)
		e.Set("a", []byte("1"))
		e.Set("b", []byte("2"))
		e.Set("c", []byte("3")) // head=c, tail=a
		e.Exist("a")            // should NOT move 'a' to head
		e.Set("d", []byte("4")) // should evict 'a' (still LRU)
		if e.Exist("a") {
			t.Fatal("expected 'a' to be evicted; Exist must not promote LRU order")
		}
	})
}

// --- SetExpiry ---

func TestSetExpiry(t *testing.T) {
	t.Run("returns true for existing key", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Set("k", []byte("v"))
		if !e.SetExpiry("k", time.Now().Add(10*time.Second)) {
			t.Fatal("expected true for existing key")
		}
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		e := newEngine(t, 10)
		if e.SetExpiry("missing", time.Now().Add(10*time.Second)) {
			t.Fatal("expected false for missing key")
		}
	})

	t.Run("expiry is stored on node", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Set("k", []byte("v"))
		expiry := time.Now().Add(10 * time.Second).Truncate(time.Second)
		e.SetExpiry("k", expiry)
		if got := e.m["k"].Val.Expiry; !got.Equal(expiry) {
			t.Fatalf("expected expiry %v, got %v", expiry, got)
		}
	})

	t.Run("overwrites previous expiry", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Set("k", []byte("v"))
		first := time.Now().Add(5 * time.Second)
		second := time.Now().Add(60 * time.Second)
		e.SetExpiry("k", first)
		e.SetExpiry("k", second)
		if got := e.m["k"].Val.Expiry; !got.Equal(second) {
			t.Fatalf("expected updated expiry %v, got %v", second, got)
		}
	})

	t.Run("does not promote LRU order", func(t *testing.T) {
		e := newEngine(t, 3)
		e.Set("a", []byte("1"))
		e.Set("b", []byte("2"))
		e.Set("c", []byte("3")) // head=c, tail=a
		e.SetExpiry("a", time.Now().Add(10*time.Second))
		e.Set("d", []byte("4")) // should evict 'a' (still LRU)
		if e.Exist("a") {
			t.Fatal("expected 'a' to be evicted; SetExpiry must not promote LRU order")
		}
	})

	t.Run("returns false after key is deleted", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Set("k", []byte("v"))
		e.Delete("k")
		if e.SetExpiry("k", time.Now().Add(10*time.Second)) {
			t.Fatal("expected false after key is deleted")
		}
	})
}

// --- Get ---

func TestGet(t *testing.T) {
	t.Run("missing key returns false", func(t *testing.T) {
		e := newEngine(t, 10)
		v, ok := e.Get("missing")
		if ok || v != nil {
			t.Fatal("expected nil, false for missing key")
		}
	})

	t.Run("returns correct value", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Set("k", []byte("hello"))
		v, ok := e.Get("k")
		if !ok {
			t.Fatal("expected key to exist")
		}
		if string(v) != "hello" {
			t.Fatalf("expected 'hello', got %q", v)
		}
	})

	t.Run("returns a copy", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Set("k", []byte("hello"))
		v, _ := e.Get("k")
		v[0] = 'X'
		v2, _ := e.Get("k")
		if string(v2) != "hello" {
			t.Fatal("Get should return a copy, not a reference")
		}
	})

	t.Run("get moves key to head", func(t *testing.T) {
		e := newEngine(t, 3)
		e.Set("a", []byte("1"))
		e.Set("b", []byte("2"))
		e.Set("c", []byte("3")) // head=c, tail=a
		e.Get("a")              // a accessed, should move to head; tail=b
		e.Set("d", []byte("4")) // evicts tail (b)
		_, ok := e.Get("b")
		if ok {
			t.Fatal("expected 'b' to be evicted after 'a' was accessed")
		}
		_, ok = e.Get("a")
		if !ok {
			t.Fatal("expected 'a' to survive since it was recently accessed")
		}
	})
}

// --- Set ---

func TestSet(t *testing.T) {
	t.Run("stores and retrieves value", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Set("k", []byte("v"))
		v, ok := e.Get("k")
		if !ok || string(v) != "v" {
			t.Fatalf("expected 'v', got %q ok=%v", v, ok)
		}
	})

	t.Run("stores a copy of value", func(t *testing.T) {
		e := newEngine(t, 10)
		b := []byte("hello")
		e.Set("k", b)
		b[0] = 'X'
		v, _ := e.Get("k")
		if string(v) != "hello" {
			t.Fatal("Set should store a copy, not a reference")
		}
	})

	t.Run("update existing key changes value", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Set("k", []byte("old"))
		e.Set("k", []byte("new"))
		v, ok := e.Get("k")
		if !ok || string(v) != "new" {
			t.Fatalf("expected 'new', got %q", v)
		}
	})

	t.Run("update does not increment count", func(t *testing.T) {
		e := newEngine(t, 2)
		e.Set("a", []byte("1"))
		e.Set("b", []byte("2"))
		e.Set("a", []byte("updated")) // update, not insert
		e.Set("c", []byte("3"))       // should evict b, not fail due to wrong count
		_, ok := e.Get("b")
		if ok {
			t.Fatal("expected 'b' to be evicted")
		}
		_, ok = e.Get("a")
		if !ok {
			t.Fatal("expected 'a' to still exist after update")
		}
	})

	t.Run("update moves key to head", func(t *testing.T) {
		e := newEngine(t, 3)
		e.Set("a", []byte("1"))
		e.Set("b", []byte("2"))
		e.Set("c", []byte("3")) // head=c, tail=a
		e.Set("a", []byte("X")) // update a, moves to head; tail=b
		e.Set("d", []byte("4")) // evicts tail (b)
		_, ok := e.Get("b")
		if ok {
			t.Fatal("expected 'b' to be evicted")
		}
	})

	t.Run("evicts least recently used on overflow", func(t *testing.T) {
		e := newEngine(t, 3)
		e.Set("a", []byte("1"))
		e.Set("b", []byte("2"))
		e.Set("c", []byte("3"))
		e.Set("d", []byte("4")) // evicts a
		_, ok := e.Get("a")
		if ok {
			t.Fatal("expected 'a' to be evicted")
		}
		_, ok = e.Get("d")
		if !ok {
			t.Fatal("expected 'd' to exist")
		}
	})

	t.Run("evicted key is removed from map", func(t *testing.T) {
		e := newEngine(t, 1)
		e.Set("a", []byte("1"))
		e.Set("b", []byte("2")) // evicts a
		if e.l != 1 {
			t.Fatalf("expected count=1, got %d", e.l)
		}
		if _, ok := e.m["a"]; ok {
			t.Fatal("evicted key 'a' should be removed from map")
		}
	})
}

// --- Delete ---

func TestDelete(t *testing.T) {
	t.Run("deletes existing key", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Set("k", []byte("v"))
		e.Delete("k")
		_, ok := e.Get("k")
		if ok {
			t.Fatal("expected key to be deleted")
		}
	})

	t.Run("delete non-existing key does not error", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Delete("ghost")
	})

	t.Run("decrements count", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Set("a", []byte("1"))
		e.Set("b", []byte("2"))
		e.Delete("a")
		if e.l != 1 {
			t.Fatalf("expected count=1, got %d", e.l)
		}
	})

	t.Run("delete head", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Set("a", []byte("1"))
		e.Set("b", []byte("2")) // head=b
		e.Delete("b")
		_, ok := e.Get("b")
		if ok {
			t.Fatal("expected head to be deleted")
		}
		_, ok = e.Get("a")
		if !ok {
			t.Fatal("expected 'a' to still exist")
		}
	})

	t.Run("delete tail", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Set("a", []byte("1")) // tail=a
		e.Set("b", []byte("2"))
		e.Delete("a")
		_, ok := e.Get("a")
		if ok {
			t.Fatal("expected tail to be deleted")
		}
		_, ok = e.Get("b")
		if !ok {
			t.Fatal("expected 'b' to still exist")
		}
	})

	t.Run("delete middle node", func(t *testing.T) {
		e := newEngine(t, 10)
		e.Set("a", []byte("1"))
		e.Set("b", []byte("2"))
		e.Set("c", []byte("3")) // head=c, mid=b, tail=a
		e.Delete("b")
		_, ok := e.Get("b")
		if ok {
			t.Fatal("expected 'b' to be deleted")
		}
		_, ok = e.Get("a")
		if !ok {
			t.Fatal("expected 'a' to still exist")
		}
		_, ok = e.Get("c")
		if !ok {
			t.Fatal("expected 'c' to still exist")
		}
	})
}

// --- Eviction order ---

func TestEvictionOrder(t *testing.T) {
	t.Run("LRU order is maintained", func(t *testing.T) {
		e := newEngine(t, 3)
		e.Set("a", []byte("1"))
		e.Set("b", []byte("2"))
		e.Set("c", []byte("3"))
		e.Get("a") // a is now most recently used; order: a, c, b
		e.Get("c") // c is now most recently used; order: c, a, b
		e.Set("d", []byte("4")) // evicts b (LRU)
		_, ok := e.Get("b")
		if ok {
			t.Fatal("expected 'b' to be evicted as LRU")
		}
	})
}

// --- Concurrency ---

func TestConcurrency(t *testing.T) {
	t.Run("concurrent set and get", func(t *testing.T) {
		e := newEngine(t, 1000)
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := fmt.Sprintf("key%d", i)
				e.Set(key, []byte(key))
				e.Get(key)
			}(i)
		}
		wg.Wait()
	})

	t.Run("concurrent set triggers eviction safely", func(t *testing.T) {
		e := newEngine(t, 10)
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				e.Set(fmt.Sprintf("key%d", i), []byte("v"))
			}(i)
		}
		wg.Wait()
		if e.l > e.cap {
			t.Fatalf("count %d exceeds capacity %d", e.l, e.cap)
		}
	})
}
