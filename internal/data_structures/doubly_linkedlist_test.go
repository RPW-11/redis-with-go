package datastructures

import "testing"

const maxTraversalIter = 1000

func collectForward[T comparable](t *testing.T, head *DLNode[T]) []T {
	t.Helper()
	var result []T
	curr := head
	for i := 0; curr != nil; i++ {
		if i >= maxTraversalIter {
			t.Fatalf("collectForward: exceeded %d iterations, possible cycle at value %v", maxTraversalIter, curr.Val)
		}
		result = append(result, curr.Val)
		curr = curr.Next
	}
	return result
}

func collectBackward[T comparable](t *testing.T, tail *DLNode[T]) []T {
	t.Helper()
	var result []T
	curr := tail
	for i := 0; curr != nil; i++ {
		if i >= maxTraversalIter {
			t.Fatalf("collectBackward: exceeded %d iterations, possible cycle at value %v", maxTraversalIter, curr.Val)
		}
		result = append(result, curr.Val)
		curr = curr.Prev
	}
	return result
}

func TestNewDoublyLinkedList(t *testing.T) {
	dl := NewDoublyLinkedList[int]()
	if dl == nil {
		t.Fatal("expected non-nil list")
	}
	if dl.Len() != 0 {
		t.Fatalf("expected length 0, got %d", dl.Len())
	}
	if dl.Head != nil || dl.Tail != nil {
		t.Fatal("expected nil Head and Tail on empty list")
	}
}

func TestLen(t *testing.T) {
	dl := NewDoublyLinkedList[int]()

	if dl.Len() != 0 {
		t.Fatalf("expected 0, got %d", dl.Len())
	}

	dl.InsertHead(1)
	if dl.Len() != 1 {
		t.Fatalf("expected 1, got %d", dl.Len())
	}

	dl.InsertTail(2)
	if dl.Len() != 2 {
		t.Fatalf("expected 2, got %d", dl.Len())
	}
}

func TestInsertHead(t *testing.T) {
	t.Run("single element", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertHead(42)

		if dl.Head == nil || dl.Tail == nil {
			t.Fatal("Head and Tail must not be nil")
		}
		if dl.Head != dl.Tail {
			t.Fatal("Head and Tail must point to the same node")
		}
		if dl.Head.Val != 42 {
			t.Fatalf("expected 42, got %d", dl.Head.Val)
		}
		if dl.Head.Prev != nil || dl.Head.Next != nil {
			t.Fatal("single node must have nil Prev and Next")
		}
	})

	t.Run("forward order", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		nums := []int{1, 2, 3, 4, 5}
		for _, v := range nums {
			dl.InsertHead(v)
		}

		if dl.Len() != len(nums) {
			t.Fatalf("expected length %d, got %d", len(nums), dl.Len())
		}

		forward := collectForward(t, dl.Head)
		for i, v := range forward {
			expected := nums[len(nums)-1-i]
			if v != expected {
				t.Fatalf("forward[%d]: expected %d, got %d", i, expected, v)
			}
		}
	})

	t.Run("backward links", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		nums := []int{1, 2, 3}
		for _, v := range nums {
			dl.InsertHead(v)
		}

		// forward: [3, 2, 1], so backward from tail: [1, 2, 3]
		backward := collectBackward(t, dl.Tail)
		for i, v := range backward {
			if v != nums[i] {
				t.Fatalf("backward[%d]: expected %d, got %d", i, nums[i], v)
			}
		}
	})

	t.Run("head prev is nil", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertHead(1)
		dl.InsertHead(2)
		if dl.Head.Prev != nil {
			t.Fatal("Head.Prev must be nil")
		}
	})

	t.Run("tail next is nil", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertHead(1)
		dl.InsertHead(2)
		if dl.Tail.Next != nil {
			t.Fatal("Tail.Next must be nil")
		}
	})
}

func TestInsertTail(t *testing.T) {
	t.Run("single element", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(42)

		if dl.Head == nil || dl.Tail == nil {
			t.Fatal("Head and Tail must not be nil")
		}
		if dl.Head != dl.Tail {
			t.Fatal("Head and Tail must point to the same node")
		}
		if dl.Tail.Val != 42 {
			t.Fatalf("expected 42, got %d", dl.Tail.Val)
		}
		if dl.Tail.Prev != nil || dl.Tail.Next != nil {
			t.Fatal("single node must have nil Prev and Next")
		}
	})

	t.Run("forward order", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		nums := []int{1, 2, 3, 4, 5}
		for _, v := range nums {
			dl.InsertTail(v)
		}

		if dl.Len() != len(nums) {
			t.Fatalf("expected length %d, got %d", len(nums), dl.Len())
		}

		forward := collectForward(t, dl.Head)
		for i, v := range forward {
			if v != nums[i] {
				t.Fatalf("forward[%d]: expected %d, got %d", i, nums[i], v)
			}
		}
	})

	t.Run("backward links", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		nums := []int{1, 2, 3}
		for _, v := range nums {
			dl.InsertTail(v)
		}

		// forward: [1, 2, 3], so backward from tail: [3, 2, 1]
		backward := collectBackward(t, dl.Tail)
		for i, v := range backward {
			expected := nums[len(nums)-1-i]
			if v != expected {
				t.Fatalf("backward[%d]: expected %d, got %d", i, expected, v)
			}
		}
	})

	t.Run("tail next is nil", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.InsertTail(2)
		if dl.Tail.Next != nil {
			t.Fatal("Tail.Next must be nil")
		}
	})

	t.Run("head prev is nil", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.InsertTail(2)
		if dl.Head.Prev != nil {
			t.Fatal("Head.Prev must be nil")
		}
	})
}

func TestInsertAt(t *testing.T) {
	t.Run("zero pos delegates to head", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.InsertTail(2)
		dl.InsertAt(99, 0)

		if dl.Head.Val != 99 {
			t.Fatalf("expected Head 99, got %d", dl.Head.Val)
		}
		if dl.Len() != 3 {
			t.Fatalf("expected length 3, got %d", dl.Len())
		}
	})

	t.Run("negative pos delegates to head", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.InsertAt(99, -5)

		if dl.Head.Val != 99 {
			t.Fatalf("expected Head 99, got %d", dl.Head.Val)
		}
	})

	t.Run("pos at n-1 delegates to tail", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.InsertTail(2) // n=2, n-1=1
		dl.InsertAt(99, 1)

		if dl.Tail.Val != 99 {
			t.Fatalf("expected Tail 99, got %d", dl.Tail.Val)
		}
		if dl.Len() != 3 {
			t.Fatalf("expected length 3, got %d", dl.Len())
		}
	})

	t.Run("pos beyond n-1 delegates to tail", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.InsertTail(2)
		dl.InsertAt(99, 100)

		if dl.Tail.Val != 99 {
			t.Fatalf("expected Tail 99, got %d", dl.Tail.Val)
		}
	})

	t.Run("middle insertion forward links", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		dl.InsertAt(10, 1) // [1, 10, 2, 3]

		expected := []int{1, 10, 2, 3}
		forward := collectForward(t, dl.Head)

		if len(forward) != len(expected) {
			t.Fatalf("expected length %d, got %d", len(expected), len(forward))
		}
		for i, v := range forward {
			if v != expected[i] {
				t.Fatalf("forward[%d]: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("middle insertion backward links", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		dl.InsertAt(10, 1) // [1, 10, 2, 3], backward: [3, 2, 10, 1]

		expected := []int{3, 2, 10, 1}
		backward := collectBackward(t, dl.Tail)

		if len(backward) != len(expected) {
			t.Fatalf("expected length %d, got %d", len(expected), len(backward))
		}
		for i, v := range backward {
			if v != expected[i] {
				t.Fatalf("backward[%d]: expected %d, got %d", i, expected[i], v)
			}
		}
	})

	t.Run("length increments on middle insert", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3} {
			dl.InsertTail(v)
		}
		dl.InsertAt(10, 1)

		if dl.Len() != 4 {
			t.Fatalf("expected length 4, got %d", dl.Len())
		}
	})
}

func TestFindFunc(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		v, ok := dl.FindFunc(func(v int) bool { return v == 1 })
		if ok {
			t.Fatalf("expected not found, got %d", v)
		}
		if v != 0 {
			t.Fatalf("expected zero value, got %d", v)
		}
	})

	t.Run("single element found", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertHead(42)
		v, ok := dl.FindFunc(func(v int) bool { return v == 42 })
		if !ok {
			t.Fatal("expected found")
		}
		if v != 42 {
			t.Fatalf("expected 42, got %d", v)
		}
	})

	t.Run("single element not found", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertHead(42)
		v, ok := dl.FindFunc(func(v int) bool { return v == 99 })
		if ok {
			t.Fatalf("expected not found, got %d", v)
		}
	})

	t.Run("found at head", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3, 4, 5} {
			dl.InsertTail(v)
		}
		v, ok := dl.FindFunc(func(v int) bool { return v == 1 })
		if !ok {
			t.Fatal("expected found")
		}
		if v != 1 {
			t.Fatalf("expected 1, got %d", v)
		}
	})

	t.Run("found at tail", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3, 4, 5} {
			dl.InsertTail(v)
		}
		v, ok := dl.FindFunc(func(v int) bool { return v == 5 })
		if !ok {
			t.Fatal("expected found")
		}
		if v != 5 {
			t.Fatalf("expected 5, got %d", v)
		}
	})

	t.Run("found in middle of odd-length list", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3, 4, 5} {
			dl.InsertTail(v)
		}
		v, ok := dl.FindFunc(func(v int) bool { return v == 3 })
		if !ok {
			t.Fatal("expected found")
		}
		if v != 3 {
			t.Fatalf("expected 3, got %d", v)
		}
	})

	t.Run("found in middle of even-length list", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3, 4} {
			dl.InsertTail(v)
		}
		v, ok := dl.FindFunc(func(v int) bool { return v == 2 })
		if !ok {
			t.Fatal("expected found")
		}
		if v != 2 {
			t.Fatalf("expected 2, got %d", v)
		}
	})

	t.Run("not found", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3, 4, 5} {
			dl.InsertTail(v)
		}
		v, ok := dl.FindFunc(func(v int) bool { return v == 99 })
		if ok {
			t.Fatalf("expected not found, got %d", v)
		}
		if v != 0 {
			t.Fatalf("expected zero value, got %d", v)
		}
	})

	t.Run("predicate matches multiple elements returns first from head", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		for _, v := range []int{1, 2, 3, 4, 5} {
			dl.InsertTail(v)
		}
		// predicate matches 2 and 4; h reaches 2 before t reaches 4
		v, ok := dl.FindFunc(func(v int) bool { return v == 2 || v == 4 })
		if !ok {
			t.Fatal("expected found")
		}
		if v != 2 {
			t.Fatalf("expected 2 (head-side match), got %d", v)
		}
	})

	t.Run("works with string type", func(t *testing.T) {
		dl := NewDoublyLinkedList[string]()
		for _, v := range []string{"a", "b", "c"} {
			dl.InsertTail(v)
		}
		v, ok := dl.FindFunc(func(v string) bool { return v == "b" })
		if !ok {
			t.Fatal("expected found")
		}
		if v != "b" {
			t.Fatalf("expected b, got %s", v)
		}
	})
}

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

	t.Run("returns removed node", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.InsertTail(2)
		dl.InsertTail(3)
		node := dl.RemoveTail()
		if node == nil {
			t.Fatal("expected non-nil removed node")
		}
		if node.Val != 3 {
			t.Fatalf("expected removed Val=3, got %d", node.Val)
		}
	})

	t.Run("returns nil on empty list", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		if dl.RemoveTail() != nil {
			t.Fatal("expected nil from empty list")
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

	t.Run("returns removed node", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		dl.InsertTail(1)
		dl.InsertTail(2)
		dl.InsertTail(3)
		node := dl.RemoveHead()
		if node == nil {
			t.Fatal("expected non-nil removed node")
		}
		if node.Val != 1 {
			t.Fatalf("expected removed Val=1, got %d", node.Val)
		}
	})

	t.Run("returns nil on empty list", func(t *testing.T) {
		dl := NewDoublyLinkedList[int]()
		if dl.RemoveHead() != nil {
			t.Fatal("expected nil from empty list")
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
