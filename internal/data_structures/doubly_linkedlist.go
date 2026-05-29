package datastructures

// DLNode defines the node for the doubly linkedlist.
// It accepts generic any type
type DLNode[T any] struct {
	Val  T
	Prev *DLNode[T]
	Next *DLNode[T]
}

// DoublyLinkedList defines the linkedlist datastructure.
// It accepts generic any type
type DoublyLinkedList[T any] struct {
	n    int
	Head *DLNode[T]
	Tail *DLNode[T]
}

// Len returns the length of the linkedlist
func (dl *DoublyLinkedList[T]) Len() int {
	return dl.n
}

// InsertHead inserts the value at the head of the linkedlist
func (dl *DoublyLinkedList[T]) InsertHead(v T) {
	dl.n++
	node := &DLNode[T]{
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

// InsertTail inserts the value at the end of the linkedlist
func (dl *DoublyLinkedList[T]) InsertTail(v T) {
	dl.n++
	node := &DLNode[T]{
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

// InsertAt inserts the value at the `pos` position
func (dl *DoublyLinkedList[T]) InsertAt(v T, pos int) {
	if pos <= 0 {
		dl.InsertHead(v)
		return
	}

	if pos >= dl.n-1 {
		dl.InsertTail(v)
		return
	}

	dl.n++
	node := &DLNode[T]{
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

// FindFunc finds the element in the list with the specified function criteria and returns it
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

// RemoveTail removes the data pointed by the tail
func (dl *DoublyLinkedList[T]) RemoveTail() {
	if dl.Tail == nil {
		return
	}
	if dl.Tail.Prev == nil {
		dl.Head = nil
		dl.Tail = nil
		return
	}

	dl.Tail = dl.Tail.Prev
	dl.Tail.Next = nil
	dl.n--
}

// RemoveHead removes the data pointed by the head
func (dl *DoublyLinkedList[T]) RemoveHead() {
	if dl.Head == nil {
		return
	}
	if dl.Head.Next == nil {
		dl.Tail = nil
		dl.Head = nil
		return
	}

	dl.Head = dl.Head.Next
	dl.Head.Prev = nil
	dl.n--
}

// MoveToHead moves the given input node to the head of list
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

// NewDoublyLinkedList creates a new doubly linked list with the specified type
func NewDoublyLinkedList[T any]() *DoublyLinkedList[T] {
	return &DoublyLinkedList[T]{}
}
