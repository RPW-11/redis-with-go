package store

import (
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"
)

const MaxCapacity = 10_000_000

type Data struct {
	Key    string
	Bytes  []byte
	Expiry time.Time
}

type Store struct {
	mu  sync.Mutex
	lru *LRUEngine
	aof *AofLogger
}

func NewStore(cap int, aofDir string) (*Store, error) {
	if cap <= 0 {
		return nil, fmt.Errorf("capacity must be greater than 0")
	}
	if cap > MaxCapacity {
		return nil, fmt.Errorf("capacity %d exceeds maximum allowed %d", cap, MaxCapacity)
	}

	var aof *AofLogger
	if aofDir != "" {
		var err error
		aof, err = NewAofLogger(aofDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create aof logger: %w", err)
		}
	}

	return &Store{
		lru: NewLRUEngine(cap),
		aof: aof,
	}, nil
}

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
