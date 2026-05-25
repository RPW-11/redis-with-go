package datastructures

import "sync"

type RedisMap struct {
	mu sync.RWMutex
	m  map[string][]byte
}

func NewRedisMap() *RedisMap {
	return &RedisMap{
		m: make(map[string][]byte),
	}
}

// Get gets the value based on the given key.
//
// It returns a copy of the value
func (rm *RedisMap) Get(k string) ([]byte, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	if v, ok := rm.m[k]; ok {
		cpy := make([]byte, len(v))
		copy(cpy, v)
		return cpy, true
	}
	return nil, false
}

// Set inserts the key-value pair to the map.
//
// It stores a copy of the value
func (rm *RedisMap) Set(k string, v []byte) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	cpy := make([]byte, len(v))
	copy(cpy, v)
	rm.m[k] = cpy
}

// SetIfNotExists inserts the key-value pair input to the map
// if the key does not exist.
//
// It stores a copy of the value
func (rm *RedisMap) SetIfNotExists(k string, v []byte) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, ok := rm.m[k]; ok {
		return false
	}

	cpy := make([]byte, len(v))
	copy(cpy, v)
	rm.m[k] = cpy
	return true
}

// Delete deletes the value from the map based on the key
func (rm *RedisMap) Delete(k string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.m, k)
}

// Len returns the length of the map
func (rm *RedisMap) Len() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return len(rm.m)
}
