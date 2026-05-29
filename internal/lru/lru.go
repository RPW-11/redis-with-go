package lru

import (
	"fmt"
	"sync"

	ds "github.com/RPW-11/redis-with-go/internal/data_structures"
)

const MaxCapacity = 10_000_000

type LRUEngine struct {
	cap int
	l   int
	mu  sync.RWMutex
	m   map[string]*ds.DLNode[[]byte]
	dl  *ds.DoublyLinkedList[[]byte]
}

func NewLRUEngine(cap int) (*LRUEngine, error) {
	if cap <= 0 {
		return nil, fmt.Errorf("capacity must be greater than 0")
	}
	if cap > MaxCapacity {
		return nil, fmt.Errorf("capacity %d exceeds maximum allowed %d", cap, MaxCapacity)
	}
	return &LRUEngine{
		cap: cap,
		m:   make(map[string]*ds.DLNode[[]byte]),
		dl:  ds.NewDoublyLinkedList[[]byte](),
	}, nil
}

func (lru *LRUEngine) Delete(k string) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	node, ok := lru.m[k]
	if !ok {
		return
	}

	delete(lru.m, k)
	lru.l--

	if node == lru.dl.Head {
		lru.dl.RemoveHead()
		return
	}
	if node == lru.dl.Tail {
		lru.dl.RemoveTail()
		return
	}

	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
}

func (lru *LRUEngine) Get(k string) ([]byte, bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	node, ok := lru.m[k]
	if !ok {
		return nil, false
	}

	lru.dl.MoveToHead(node)

	cpy := make([]byte, len(node.Val))
	copy(cpy, node.Val)
	return cpy, true

}

func (lru *LRUEngine) Set(k string, v []byte) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	cpy := make([]byte, len(v))
	copy(cpy, v)

	// check if the key exist. If it does, update the value. Else, insert to the head
	if node, ok := lru.m[k]; ok {
		node.Val = cpy
		lru.dl.MoveToHead(node)
		return
	}

	// check if it exceeds the capacity, if it does, run eviction policy
	if lru.l+1 > lru.cap {
		node := lru.dl.RemoveTail()
		if node == nil {
			return
		}

		delete(lru.m, node.Id)
		lru.l--
	}

	lru.dl.InsertHead(k, cpy)
	lru.l++
	lru.m[k] = lru.dl.Head
}
