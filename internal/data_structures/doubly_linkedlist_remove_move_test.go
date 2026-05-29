package datastructures

import "testing"

func TestRemoveTail(t *testing.T) {
	t.Run("empty list does not panic", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.RemoveTail()
		if dl.Head != nil || dl.Tail != nil {
			t.Fatal("expected nil Head and Tail after removing from empty list")
		}
	})

	t.Run("single element clears head and tail", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertHead(1)
		dl.RemoveTail()
		if dl.Head != nil {
			t.Fatal("expected nil Head after removing sole element")
		}
		if dl.Tail != nil {
			t.Fatal("expected nil Tail after removing sole element")
		}
	})

	t.Run("length decrements", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.InsertTail(2)
		dl.InsertTail(3)
		dl.RemoveTail()
		if dl.Len() != 2 {
			t.Fatalf("expected length 2, got %d", dl.Len())
		}
	})

	t.Run("new tail next is nil", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.InsertTail(2)
		dl.RemoveTail()
		if dl.Tail.Next != nil {
			t.Fatal("new Tail.Next must be nil after RemoveTail")
		}
	})

	t.Run("forward traversal stops at new tail", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		dl.RemoveTail() // removes 3; list should be [1, 2]

		forward := collectForward(t, dl.Head)
		expected := []int{1, 2}
		if len(forward) != len(expected) {
			t.Fatalf("expected %v, got %v", expected, forward)
		}
		for i, v := range forward {
			if v != expected[i] {
				t.Fatalf("forward[%d]: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("backward traversal intact after removal", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		dl.RemoveTail() // list should be [1, 2]

		backward := collectBackward(t, dl.Tail)
		expected := []int{2, 1}
		if len(backward) != len(expected) {
			t.Fatalf("expected %v, got %v", expected, backward)
		}
		for i, v := range backward {
			if v != expected[i] {
				t.Fatalf("backward[%d]: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("correct tail value after removal", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.InsertTail(2)
		dl.InsertTail(3)
		dl.RemoveTail()
		if dl.Tail.Val != 2 {
			t.Fatalf("expected Tail.Val=2, got %d", dl.Tail.Val)
		}
	})
}

func TestRemoveHead(t *testing.T) {
	t.Run("empty list does not panic", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.RemoveHead()
		if dl.Head != nil || dl.Tail != nil {
			t.Fatal("expected nil Head and Tail after removing from empty list")
		}
	})

	t.Run("single element clears head and tail", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertHead(1)
		dl.RemoveHead()
		if dl.Head != nil {
			t.Fatal("expected nil Head after removing sole element")
		}
		if dl.Tail != nil {
			t.Fatal("expected nil Tail after removing sole element")
		}
	})

	t.Run("length decrements", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.InsertTail(2)
		dl.InsertTail(3)
		dl.RemoveHead()
		if dl.Len() != 2 {
			t.Fatalf("expected length 2, got %d", dl.Len())
		}
	})

	t.Run("new head prev is nil", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.InsertTail(2)
		dl.RemoveHead()
		if dl.Head.Prev != nil {
			t.Fatal("new Head.Prev must be nil after RemoveHead")
		}
	})

	t.Run("forward traversal starts at new head", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		dl.RemoveHead() // removes 1; list should be [2, 3]

		forward := collectForward(t, dl.Head)
		expected := []int{2, 3}
		if len(forward) != len(expected) {
			t.Fatalf("expected %v, got %v", expected, forward)
		}
		for i, v := range forward {
			if v != expected[i] {
				t.Fatalf("forward[%d]: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("backward traversal intact after removal", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		dl.RemoveHead() // list should be [2, 3]

		backward := collectBackward(t, dl.Tail)
		expected := []int{3, 2}
		if len(backward) != len(expected) {
			t.Fatalf("expected %v, got %v", expected, backward)
		}
		for i, v := range backward {
			if v != expected[i] {
				t.Fatalf("backward[%d]: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("correct head value after removal", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.InsertTail(2)
		dl.InsertTail(3)
		dl.RemoveHead()
		if dl.Head.Val != 2 {
			t.Fatalf("expected Head.Val=2, got %d", dl.Head.Val)
		}
	})
}

func TestMoveToHead(t *testing.T) {
	t.Run("nil node does not panic", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.MoveToHead(nil)
	})

	t.Run("single element node does not panic", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.MoveToHead(dl.Head)
		if dl.Head.Val != 1 {
			t.Fatalf("expected Head.Val=1, got %d", dl.Head.Val)
		}
		if dl.Tail.Val != 1 {
			t.Fatalf("expected Tail.Val=1, got %d", dl.Tail.Val)
		}
	})

	t.Run("node already at head is a no-op", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		dl.MoveToHead(dl.Head) // move 1, already head

		forward := collectForward(t, dl.Head)
		expected := []int{1, 2, 3}
		if len(forward) != len(expected) {
			t.Fatalf("expected %v, got %v", expected, forward)
		}
		for i, v := range forward {
			if v != expected[i] {
				t.Fatalf("forward[%d]: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("moving tail updates Tail pointer", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		tail := dl.Tail
		dl.MoveToHead(tail) // move 3 to head; list should be [3, 1, 2]

		if dl.Tail.Val != 2 {
			t.Fatalf("expected Tail.Val=2 after moving tail to head, got %d", dl.Tail.Val)
		}
		if dl.Tail.Next != nil {
			t.Fatal("new Tail.Next must be nil")
		}
	})

	t.Run("moving tail does not panic", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		dl.MoveToHead(dl.Tail)
	})

	t.Run("middle node moved to head: forward links correct", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		middle := dl.Head.Next // node with value 2
		dl.MoveToHead(middle)  // list should be [2, 1, 3]

		forward := collectForward(t, dl.Head)
		expected := []int{2, 1, 3}
		if len(forward) != len(expected) {
			t.Fatalf("expected %v, got %v", expected, forward)
		}
		for i, v := range forward {
			if v != expected[i] {
				t.Fatalf("forward[%d]: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("middle node moved to head: backward links correct", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		middle := dl.Head.Next // node with value 2
		dl.MoveToHead(middle)  // list should be [2, 1, 3], backward: [3, 1, 2]

		backward := collectBackward(t, dl.Tail)
		expected := []int{3, 1, 2}
		if len(backward) != len(expected) {
			t.Fatalf("expected %v, got %v", expected, backward)
		}
		for i, v := range backward {
			if v != expected[i] {
				t.Fatalf("backward[%d]: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("tail moved to head: forward links correct", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		dl.MoveToHead(dl.Tail) // list should be [3, 1, 2]

		forward := collectForward(t, dl.Head)
		expected := []int{3, 1, 2}
		if len(forward) != len(expected) {
			t.Fatalf("expected %v, got %v", expected, forward)
		}
		for i, v := range forward {
			if v != expected[i] {
				t.Fatalf("forward[%d]: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("tail moved to head: backward links correct", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		dl.MoveToHead(dl.Tail) // list should be [3, 1, 2], backward: [2, 1, 3]

		backward := collectBackward(t, dl.Tail)
		expected := []int{2, 1, 3}
		if len(backward) != len(expected) {
			t.Fatalf("expected %v, got %v", expected, backward)
		}
		for i, v := range backward {
			if v != expected[i] {
				t.Fatalf("backward[%d]: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("length unchanged after move", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		dl.MoveToHead(dl.Head.Next)
		if dl.Len() != 3 {
			t.Fatalf("expected length 3, got %d", dl.Len())
		}
	})

	t.Run("new head prev is nil after move", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		dl.MoveToHead(dl.Head.Next)
		if dl.Head.Prev != nil {
			t.Fatal("Head.Prev must be nil after MoveToHead")
		}
	})
}
