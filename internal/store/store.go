// Package store implements the key-value store with LRU eviction and AOF persistence.
package store

import (
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"
)

const MaxCapacity = 10_000_000

// Data is the value type held by each LRU node.
type Data struct {
	Key    string
	Bytes  []byte
	Expiry time.Time // zero means no expiry
}

// Store is the public interface to the cache. It owns the mutex, LRU engine,
// and AOF logger — all writes go through here.
type Store struct {
	mu  sync.Mutex
	lru *lruEngine
	aof *aofLogger // nil when AOF is disabled or during replay
}

// NewStore creates a Store with the given capacity. Pass aofDir="" to disable AOF.
func NewStore(cap int, aofDir string) (*Store, error) {
	if cap <= 0 {
		return nil, fmt.Errorf("capacity must be greater than 0")
	}
	if cap > MaxCapacity {
		return nil, fmt.Errorf("capacity %d exceeds maximum allowed %d", cap, MaxCapacity)
	}

	var aof *aofLogger
	if aofDir != "" {
		var err error
		aof, err = newAofLogger(aofDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create aof logger: %w", err)
		}
	}

	return &Store{
		lru: newLRUEngine(cap),
		aof: aof,
	}, nil
}

// AttachAOF wires up the AOF logger after replay is complete.
// Until this is called, s.aof is nil and writes are not logged.
func (s *Store) AttachAOF(dir string) error {
	aof, err := newAofLogger(dir)
	if err != nil {
		return fmt.Errorf("failed to create aof logger: %w", err)
	}
	s.mu.Lock()
	s.aof = aof
	s.mu.Unlock()
	return nil
}

// Get returns a copy of the value for the given key. The key is promoted to
// most-recently-used first, and any expired key is removed on access.
func (s *Store) Get(k string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, ok := s.lru.Get(k)
	if !ok {
		return nil, false
	}

	s.lru.Promote(node)
	if !node.Val.Expiry.IsZero() && time.Now().After(node.Val.Expiry) {
		s.lru.Unlink(node)
		return nil, false
	}

	cpy := make([]byte, len(node.Val.Bytes))
	copy(cpy, node.Val.Bytes)
	return cpy, true
}

func (s *Store) Set(k string, v []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cpy := make([]byte, len(v))
	copy(cpy, v)

	if node, ok := s.lru.Get(k); ok {
		node.Val.Bytes = cpy
		s.lru.Promote(node)
		if s.aof != nil {
			if err := s.aof.LogSet(k, cpy); err != nil {
				slog.Error("aof log set failed", "err", err)
			}
		}
		return
	}

	if s.lru.Full() {
		if evicted := s.lru.Evict(); evicted == nil {
			return
		}
	}

	s.lru.Insert(Data{Key: k, Bytes: cpy})

	if s.aof != nil {
		if err := s.aof.LogSet(k, cpy); err != nil {
			slog.Error("aof log set failed", "err", err)
		}
	}
}

func (s *Store) Delete(k string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, ok := s.lru.Get(k)
	if !ok {
		return
	}

	s.lru.Unlink(node)

	if s.aof != nil {
		if err := s.aof.LogDelete(k, nil); err != nil {
			slog.Error("aof log delete failed", "err", err)
		}
	}
}

func (s *Store) Exist(k string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lru.Exist(k)
}

func (s *Store) ExpiryOf(k string) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lru.ExpiryOf(k)
}

func (s *Store) SetExpiry(k string, t time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.lru.SetExpiry(k, t) {
		return false
	}

	if s.aof != nil {
		ts := []byte(strconv.FormatInt(t.Unix(), 10))
		if err := s.aof.LogExpireAt(k, ts); err != nil {
			slog.Error("aof log expireat failed", "err", err)
		}
	}

	return true
}
