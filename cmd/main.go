package main

import (
	"flag"
	"log/slog"

	"github.com/RPW-11/redis-with-go/internal/server"
)

func main() {
	port := flag.String("port", "6379", "port to listen on")
	maxCap := flag.Int("max-cap", 1_000_000, "max capacity before LRU eviction")
	aofDir := flag.String("aof-dir", ".", "directory to AOF file")
	aofOff := flag.Bool("no-aof", false, "disable AOF persistence")
	flag.Parse()

	if *aofOff {
		*aofDir = ""
	}

	s, err := server.NewServer(*port, *aofDir, *maxCap)
	if err != nil {
		slog.Error("error creating the server:", "error", err, "port", *port)
		return
	}
	err = s.Run()
	if err != nil {
		slog.Error("error starting the server:", "error", err, "port", *port)
		return
	}
}
