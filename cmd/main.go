package main

import (
	"github.com/RPW-11/redis-with-go/internal/server"
)

func main() {
	s, err := server.NewServer("8000", 1_000_000)
	if err != nil {
		return
	}
	err = s.Run()
	if err != nil {
		return
	}
}
