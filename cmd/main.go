package main

import (
	"fmt"

	datastructures "github.com/RPW-11/redis-with-go/internal/data_structures"
	"github.com/RPW-11/redis-with-go/internal/server"
)

func main() {
	dll := datastructures.NewDoublyLinkedList[int]()
	dll.InsertHead(2)
	dll.InsertHead(3)
	dll.InsertTail(5)

	if v, ok := dll.FindFunc(func(v int) bool {
		return v == 5
	}); ok {
		fmt.Printf("%d is found\n", v)
	}

	type Student struct {
		Id   string
		Name string
	}

	newDll := datastructures.NewDoublyLinkedList[Student]()

	newDll.InsertTail(Student{
		Id:   "1",
		Name: "Pedro",
	})
	newDll.InsertTail(Student{
		Id:   "2",
		Name: "Lucas",
	})

	if v, ok := newDll.FindFunc(func(v Student) bool {
		return v.Id == "2"
	}); ok {
		fmt.Printf("%s is found\n", v.Name)
		return
	}

	fmt.Println("Not found, running the server")

	s := server.NewServer("8000")
	err := s.Run()
	if err != nil {
		return
	}
}
