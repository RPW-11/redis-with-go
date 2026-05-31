package store

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func newStore(t *testing.T, cap int) *Store {
	t.Helper()
	s, err := NewStore(cap, "")
	if err != nil {
		t.Fatalf("NewStore(%d): %v", cap, err)
	}
	return s
}

// --- Constructor ---

func TestNewStore(t *testing.T) {
	t.Run("valid capacity", func(t *testing.T) {
		s, err := NewStore(10, "")
		if err != nil || s == nil {
			t.Fatalf("expected valid store, got err=%v", err)
		}
	})

	t.Run("zero capacity rejected", func(t *testing.T) {
		_, err := NewStore(0, "")
		if err == nil {
			t.Fatal("expected error for zero capacity")
		}
	})

	t.Run("negative capacity rejected", func(t *testing.T) {
		_, err := NewStore(-1, "")
		if err == nil {
			t.Fatal("expected error for negative capacity")
		}
	})

	t.Run("exceeds MaxCapacity rejected", func(t *testing.T) {
		_, err := NewStore(MaxCapacity+1, "")
		if err == nil {
			t.Fatal("expected error for capacity exceeding MaxCapacity")
		}
	})

	t.Run("MaxCapacity accepted", func(t *testing.T) {
		_, err := NewStore(MaxCapacity, "")
		if err != nil {
			t.Fatalf("expected MaxCapacity to be valid, got: %v", err)
		}
	})
}

// --- Exist ---

func TestExist(t *testing.T) {
	t.Run("returns true for existing key", func(t *testing.T) {
		s := newStore(t, 10)
		s.Set("k", []byte("v"))
		if !s.Exist("k") {
			t.Fatal("expected true for existing key")
		}
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		s := newStore(t, 10)
		if s.Exist("missing") {
			t.Fatal("expected false for missing key")
		}
	})

	t.Run("returns false after delete", func(t *testing.T) {
		s := newStore(t, 10)
		s.Set("k", []byte("v"))
		s.Delete("k")
		if s.Exist("k") {
			t.Fatal("expected false after delete")
		}
	})

	t.Run("returns false after eviction", func(t *testing.T) {
		s := newStore(t, 1)
		s.Set("a", []byte("1"))
		s.Set("b", []byte("2")) // evicts a
		if s.Exist("a") {
			t.Fatal("expected false for evicted key")
		}
		if !s.Exist("b") {
			t.Fatal("expected true for current key")
		}
	})

	t.Run("does not promote to head", func(t *testing.T) {
		s := newStore(t, 3)
		s.Set("a", []byte("1"))
		s.Set("b", []byte("2"))
		s.Set("c", []byte("3")) // head=c, tail=a
		s.Exist("a")            // should NOT move 'a' to head
		s.Set("d", []byte("4")) // should evict 'a' (still LRU)
		if s.Exist("a") {
			t.Fatal("expected 'a' to be evicted; Exist must not promote LRU order")
		}
	})
}

// --- SetExpiry ---

func TestSetExpiry(t *testing.T) {
	t.Run("returns true for existing key", func(t *testing.T) {
		s := newStore(t, 10)
		s.Set("k", []byte("v"))
		if !s.SetExpiry("k", time.Now().Add(10*time.Second)) {
			t.Fatal("expected true for existing key")
		}
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		s := newStore(t, 10)
		if s.SetExpiry("missing", time.Now().Add(10*time.Second)) {
			t.Fatal("expected false for missing key")
		}
	})

	t.Run("expiry is stored on node", func(t *testing.T) {
		s := newStore(t, 10)
		s.Set("k", []byte("v"))
		expiry := time.Now().Add(10 * time.Second).Truncate(time.Second)
		s.SetExpiry("k", expiry)
		if got := s.lru.m["k"].Val.Expiry; !got.Equal(expiry) {
			t.Fatalf("expected expiry %v, got %v", expiry, got)
		}
	})

	t.Run("overwrites previous expiry", func(t *testing.T) {
		s := newStore(t, 10)
		s.Set("k", []byte("v"))
		first := time.Now().Add(5 * time.Second)
		second := time.Now().Add(60 * time.Second)
		s.SetExpiry("k", first)
		s.SetExpiry("k", second)
		if got := s.lru.m["k"].Val.Expiry; !got.Equal(second) {
			t.Fatalf("expected updated expiry %v, got %v", second, got)
		}
	})

	t.Run("does not promote LRU order", func(t *testing.T) {
		s := newStore(t, 3)
		s.Set("a", []byte("1"))
		s.Set("b", []byte("2"))
		s.Set("c", []byte("3")) // head=c, tail=a
		s.SetExpiry("a", time.Now().Add(10*time.Second))
		s.Set("d", []byte("4")) // should evict 'a' (still LRU)
		if s.Exist("a") {
			t.Fatal("expected 'a' to be evicted; SetExpiry must not promote LRU order")
		}
	})

	t.Run("returns false after key is deleted", func(t *testing.T) {
		s := newStore(t, 10)
		s.Set("k", []byte("v"))
		s.Delete("k")
		if s.SetExpiry("k", time.Now().Add(10*time.Second)) {
			t.Fatal("expected false after key is deleted")
		}
	})
}

// --- Get ---

func TestGet(t *testing.T) {
	t.Run("missing key returns false", func(t *testing.T) {
		s := newStore(t, 10)
		v, ok := s.Get("missing")
		if ok || v != nil {
			t.Fatal("expected nil, false for missing key")
		}
	})

	t.Run("returns correct value", func(t *testing.T) {
		s := newStore(t, 10)
		s.Set("k", []byte("hello"))
		v, ok := s.Get("k")
		if !ok {
			t.Fatal("expected key to exist")
		}
		if string(v) != "hello" {
			t.Fatalf("expected 'hello', got %q", v)
		}
	})

	t.Run("returns a copy", func(t *testing.T) {
		s := newStore(t, 10)
		s.Set("k", []byte("hello"))
		v, _ := s.Get("k")
		v[0] = 'X'
		v2, _ := s.Get("k")
		if string(v2) != "hello" {
			t.Fatal("Get should return a copy, not a reference")
		}
	})

	t.Run("get moves key to head", func(t *testing.T) {
		s := newStore(t, 3)
		s.Set("a", []byte("1"))
		s.Set("b", []byte("2"))
		s.Set("c", []byte("3")) // head=c, tail=a
		s.Get("a")              // a accessed, should move to head; tail=b
		s.Set("d", []byte("4")) // evicts tail (b)
		_, ok := s.Get("b")
		if ok {
			t.Fatal("expected 'b' to be evicted after 'a' was accessed")
		}
		_, ok = s.Get("a")
		if !ok {
			t.Fatal("expected 'a' to survive since it was recently accessed")
		}
	})
}

// --- Set ---

func TestSet(t *testing.T) {
	t.Run("stores and retrieves value", func(t *testing.T) {
		s := newStore(t, 10)
		s.Set("k", []byte("v"))
		v, ok := s.Get("k")
		if !ok || string(v) != "v" {
			t.Fatalf("expected 'v', got %q ok=%v", v, ok)
		}
	})

	t.Run("stores a copy of value", func(t *testing.T) {
		s := newStore(t, 10)
		b := []byte("hello")
		s.Set("k", b)
		b[0] = 'X'
		v, _ := s.Get("k")
		if string(v) != "hello" {
			t.Fatal("Set should store a copy, not a reference")
		}
	})

	t.Run("update existing key changes value", func(t *testing.T) {
		s := newStore(t, 10)
		s.Set("k", []byte("old"))
		s.Set("k", []byte("new"))
		v, ok := s.Get("k")
		if !ok || string(v) != "new" {
			t.Fatalf("expected 'new', got %q", v)
		}
	})

	t.Run("update does not increment count", func(t *testing.T) {
		s := newStore(t, 2)
		s.Set("a", []byte("1"))
		s.Set("b", []byte("2"))
		s.Set("a", []byte("updated")) // update, not insert
		s.Set("c", []byte("3"))       // should evict b, not fail due to wrong count
		_, ok := s.Get("b")
		if ok {
			t.Fatal("expected 'b' to be evicted")
		}
		_, ok = s.Get("a")
		if !ok {
			t.Fatal("expected 'a' to still exist after update")
		}
	})

	t.Run("update moves key to head", func(t *testing.T) {
		s := newStore(t, 3)
		s.Set("a", []byte("1"))
		s.Set("b", []byte("2"))
		s.Set("c", []byte("3")) // head=c, tail=a
		s.Set("a", []byte("X")) // update a, moves to head; tail=b
		s.Set("d", []byte("4")) // evicts tail (b)
		_, ok := s.Get("b")
		if ok {
			t.Fatal("expected 'b' to be evicted")
		}
	})

	t.Run("evicts least recently used on overflow", func(t *testing.T) {
		s := newStore(t, 3)
		s.Set("a", []byte("1"))
		s.Set("b", []byte("2"))
		s.Set("c", []byte("3"))
		s.Set("d", []byte("4")) // evicts a
		_, ok := s.Get("a")
		if ok {
			t.Fatal("expected 'a' to be evicted")
		}
		_, ok = s.Get("d")
		if !ok {
			t.Fatal("expected 'd' to exist")
		}
	})

	t.Run("evicted key is removed from map", func(t *testing.T) {
		s := newStore(t, 1)
		s.Set("a", []byte("1"))
		s.Set("b", []byte("2")) // evicts a
		if s.lru.l != 1 {
			t.Fatalf("expected count=1, got %d", s.lru.l)
		}
		if _, ok := s.lru.m["a"]; ok {
			t.Fatal("evicted key 'a' should be removed from map")
		}
	})
}

// --- Delete ---

func TestDelete(t *testing.T) {
	t.Run("deletes existing key", func(t *testing.T) {
		s := newStore(t, 10)
		s.Set("k", []byte("v"))
		s.Delete("k")
		_, ok := s.Get("k")
		if ok {
			t.Fatal("expected key to be deleted")
		}
	})

	t.Run("delete non-existing key does not error", func(t *testing.T) {
		s := newStore(t, 10)
		s.Delete("ghost")
	})

	t.Run("decrements count", func(t *testing.T) {
		s := newStore(t, 10)
		s.Set("a", []byte("1"))
		s.Set("b", []byte("2"))
		s.Delete("a")
		if s.lru.l != 1 {
			t.Fatalf("expected count=1, got %d", s.lru.l)
		}
	})

	t.Run("delete head", func(t *testing.T) {
		s := newStore(t, 10)
		s.Set("a", []byte("1"))
		s.Set("b", []byte("2")) // head=b
		s.Delete("b")
		_, ok := s.Get("b")
		if ok {
			t.Fatal("expected head to be deleted")
		}
		_, ok = s.Get("a")
		if !ok {
			t.Fatal("expected 'a' to still exist")
		}
	})

	t.Run("delete tail", func(t *testing.T) {
		s := newStore(t, 10)
		s.Set("a", []byte("1")) // tail=a
		s.Set("b", []byte("2"))
		s.Delete("a")
		_, ok := s.Get("a")
		if ok {
			t.Fatal("expected tail to be deleted")
		}
		_, ok = s.Get("b")
		if !ok {
			t.Fatal("expected 'b' to still exist")
		}
	})

	t.Run("delete middle node", func(t *testing.T) {
		s := newStore(t, 10)
		s.Set("a", []byte("1"))
		s.Set("b", []byte("2"))
		s.Set("c", []byte("3")) // head=c, mid=b, tail=a
		s.Delete("b")
		_, ok := s.Get("b")
		if ok {
			t.Fatal("expected 'b' to be deleted")
		}
		_, ok = s.Get("a")
		if !ok {
			t.Fatal("expected 'a' to still exist")
		}
		_, ok = s.Get("c")
		if !ok {
			t.Fatal("expected 'c' to still exist")
		}
	})
}

// --- Eviction order ---

func TestEvictionOrder(t *testing.T) {
	t.Run("LRU order is maintained", func(t *testing.T) {
		s := newStore(t, 3)
		s.Set("a", []byte("1"))
		s.Set("b", []byte("2"))
		s.Set("c", []byte("3"))
		s.Get("a") // a is now most recently used; order: a, c, b
		s.Get("c") // c is now most recently used; order: c, a, b
		s.Set("d", []byte("4")) // evicts b (LRU)
		_, ok := s.Get("b")
		if ok {
			t.Fatal("expected 'b' to be evicted as LRU")
		}
	})
}

// --- Concurrency ---

func TestConcurrency(t *testing.T) {
	t.Run("concurrent set and get", func(t *testing.T) {
		s := newStore(t, 1000)
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := fmt.Sprintf("key%d", i)
				s.Set(key, []byte(key))
				s.Get(key)
			}(i)
		}
		wg.Wait()
	})

	t.Run("concurrent set triggers eviction safely", func(t *testing.T) {
		s := newStore(t, 10)
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				s.Set(fmt.Sprintf("key%d", i), []byte("v"))
			}(i)
		}
		wg.Wait()
		if s.lru.l > s.lru.cap {
			t.Fatalf("count %d exceeds capacity %d", s.lru.l, s.lru.cap)
		}
	})
}
