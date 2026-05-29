package datastructures

// DLNode is a node in a doubly linked list. Id mirrors the map key for O(1) map deletion.
type DLNode[T any] struct {
	Id   string
	Val  T
	Prev *DLNode[T]
	Next *DLNode[T]
}

// DoublyLinkedList is a generic doubly linked list.
type DoublyLinkedList[T any] struct {
	n    int
	Head *DLNode[T]
	Tail *DLNode[T]
}

// Len returns the number of elements in the list.
func (dl *DoublyLinkedList[T]) Len() int {
	return dl.n
}

// InsertHead inserts a new node at the front of the list.
func (dl *DoublyLinkedList[T]) InsertHead(id string, v T) {
	dl.n++
	node := &DLNode[T]{
		Id:  id,
		Val: v,
	}

	if dl.Head == nil {
		dl.Head = node
		dl.Tail = node
		return
	}

	node.Next = dl.Head
	dl.Head.Prev = node
	dl.Head = node
}

// InsertTail inserts a new node at the back of the list.
func (dl *DoublyLinkedList[T]) InsertTail(id string, v T) {
	dl.n++
	node := &DLNode[T]{
		Id:  id,
		Val: v,
	}

	if dl.Head == nil {
		dl.Head = node
		dl.Tail = node
		return
	}

	dl.Tail.Next = node
	node.Prev = dl.Tail
	dl.Tail = node
}

// InsertAt inserts a new node at position pos (0-indexed). Clamps to head/tail if out of range.
func (dl *DoublyLinkedList[T]) InsertAt(id string, v T, pos int) {
	if pos <= 0 {
		dl.InsertHead(id, v)
		return
	}

	if pos >= dl.n-1 {
		dl.InsertTail(id, v)
		return
	}

	dl.n++
	node := &DLNode[T]{
		Id:  id,
		Val: v,
	}

	temp := dl.Head
	for pos > 0 {
		pos--
		temp = temp.Next
	}

	temp.Prev.Next = node
	node.Prev = temp.Prev
	node.Next = temp
	temp.Prev = node
}

// FindFunc returns the first element matching f, searching from both ends simultaneously.
func (dl *DoublyLinkedList[T]) FindFunc(f func(v T) bool) (T, bool) {
	var zero T

	if dl.Len() == 0 {
		return zero, false
	}

	h, t := dl.Head, dl.Tail

	for h != t && t.Next != h {
		if f(h.Val) {
			return h.Val, true
		}
		if f(t.Val) {
			return t.Val, true
		}

		h = h.Next
		t = t.Prev
	}

	if f(h.Val) {
		return h.Val, true
	}
	if f(t.Val) {
		return t.Val, true
	}

	return zero, false
}

// RemoveTail removes the tail node and returns it. Returns nil if the list is empty.
func (dl *DoublyLinkedList[T]) RemoveTail() *DLNode[T] {
	if dl.Tail == nil {
		return nil
	}
	removed := dl.Tail
	if dl.Tail.Prev == nil {
		dl.Head = nil
		dl.Tail = nil
		return removed
	}

	dl.Tail = dl.Tail.Prev
	dl.Tail.Next = nil
	dl.n--
	return removed
}

// RemoveHead removes the head node and returns it. Returns nil if the list is empty.
func (dl *DoublyLinkedList[T]) RemoveHead() *DLNode[T] {
	if dl.Head == nil {
		return nil
	}
	removed := dl.Head
	if dl.Head.Next == nil {
		dl.Tail = nil
		dl.Head = nil
		return removed
	}

	dl.Head = dl.Head.Next
	dl.Head.Prev = nil
	dl.n--
	return removed
}

// MoveToHead moves node to the front. No-op if node is nil, the only element, or already the head.
func (dl *DoublyLinkedList[T]) MoveToHead(node *DLNode[T]) {
	if node == nil || dl.Head.Next == nil || dl.Head == node {
		return
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		dl.Tail = node.Prev
	}

	node.Prev.Next = node.Next
	node.Prev = nil

	dl.Head.Prev = node
	node.Next = dl.Head
	dl.Head = node
}

// NewDoublyLinkedList returns an empty doubly linked list.
func NewDoublyLinkedList[T any]() *DoublyLinkedList[T] {
	return &DoublyLinkedList[T]{}
}
