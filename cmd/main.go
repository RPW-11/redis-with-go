package main

import (
	"github.com/RPW-11/redis-with-go/internal/server"
)

func main() {
	s := server.NewServer("8000")
	err := s.Run()
	if err != nil {
		return
	}
}
