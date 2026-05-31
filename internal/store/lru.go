package store

import (
	"time"

	ds "github.com/RPW-11/redis-with-go/internal/data_structures"
)

type LRUEngine struct {
	cap int
	l   int
	m   map[string]*ds.DLNode[Data]
	dl  *ds.DoublyLinkedList[Data]
}

func NewLRUEngine(cap int) *LRUEngine {
	return &LRUEngine{
		cap: cap,
		m:   make(map[string]*ds.DLNode[Data]),
		dl:  ds.NewDoublyLinkedList[Data](),
	}
}

func (l *LRUEngine) Get(k string) (*ds.DLNode[Data], bool) {
	node, ok := l.m[k]
	return node, ok
}

func (l *LRUEngine) Promote(node *ds.DLNode[Data]) {
	l.dl.MoveToHead(node)
}

func (l *LRUEngine) Insert(data Data) {
	l.dl.InsertHead(data)
	l.l++
	l.m[data.Key] = l.dl.Head
}

func (l *LRUEngine) Full() bool {
	return l.l+1 > l.cap
}

func (l *LRUEngine) Evict() *ds.DLNode[Data] {
	node := l.dl.RemoveTail()
	if node == nil {
		return nil
	}
	delete(l.m, node.Val.Key)
	l.l--
	return node
}

func (l *LRUEngine) Unlink(node *ds.DLNode[Data]) {
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

func (l *LRUEngine) Exist(k string) bool {
	_, ok := l.m[k]
	return ok
}

func (l *LRUEngine) ExpiryOf(k string) (time.Time, bool) {
	node, ok := l.m[k]
	if !ok {
		return time.Time{}, false
	}
	return node.Val.Expiry, true
}

func (l *LRUEngine) SetExpiry(k string, t time.Time) bool {
	node, ok := l.m[k]
	if !ok {
		return false
	}
	node.Val.Expiry = t
	return true
}
