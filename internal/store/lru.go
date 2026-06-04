package store

import (
	"time"

	ds "github.com/RPW-11/redis-with-go/internal/data_structures"
)

// lruEngine is pure LRU mechanics.
type lruEngine struct {
	cap int                         // maximum number of keys
	l   int                         // current number of keys
	m   map[string]*ds.DLNode[Data] // key → node lookup
	dl  *ds.DoublyLinkedList[Data]  // head = MRU, tail = LRU
}

func newLRUEngine(cap int) *lruEngine {
	return &lruEngine{
		cap: cap,
		m:   make(map[string]*ds.DLNode[Data]),
		dl:  ds.NewDoublyLinkedList[Data](),
	}
}

func (l *lruEngine) Get(k string) (*ds.DLNode[Data], bool) {
	node, ok := l.m[k]
	return node, ok
}

// Promote moves the node to the MRU position.
func (l *lruEngine) Promote(node *ds.DLNode[Data]) {
	l.dl.MoveToHead(node)
}

func (l *lruEngine) Insert(data Data) {
	l.dl.InsertHead(data)
	l.l++
	l.m[data.Key] = l.dl.Head
}

func (l *lruEngine) Full() bool {
	return l.l+1 > l.cap
}

// Evict removes the LRU node (tail) to make room for a new entry.
func (l *lruEngine) Evict() *ds.DLNode[Data] {
	node := l.dl.RemoveTail()
	if node == nil {
		return nil
	}
	delete(l.m, node.Val.Key)
	l.l--
	return node
}

// Unlink removes a specific node from the list and map.
func (l *lruEngine) Unlink(node *ds.DLNode[Data]) {
	switch node {
	case l.dl.Head:
		l.dl.RemoveHead()
	case l.dl.Tail:
		l.dl.RemoveTail()
	default:
		node.Prev.Next = node.Next
		node.Next.Prev = node.Prev
	}
	delete(l.m, node.Val.Key)
	l.l--
}

func (l *lruEngine) Exist(k string) bool {
	_, ok := l.m[k]
	return ok
}

func (l *lruEngine) ExpiryOf(k string) (time.Time, bool) {
	node, ok := l.m[k]
	if !ok {
		return time.Time{}, false
	}
	return node.Val.Expiry, true
}

func (l *lruEngine) SetExpiry(k string, t time.Time) bool {
	node, ok := l.m[k]
	if !ok {
		return false
	}
	node.Val.Expiry = t
	return true
}
