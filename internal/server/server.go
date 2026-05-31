package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/RPW-11/redis-with-go/internal/command"
	"github.com/RPW-11/redis-with-go/internal/lru"
)

type Server struct {
	Port string
	m    *lru.LRUEngine
}

func (s *Server) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", s.Port))
	if err != nil {
		slog.Error("couldn't listen to network")
		return fmt.Errorf("couldn't listen to network\n")
	}

	defer func() {
		if err := l.Close(); err != nil {
			slog.Error("error closing the network")
		}
	}()

	slog.Info(fmt.Sprintf("Server is running on port: %s\n", s.Port))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		conn, err := l.Accept()
		if err != nil {
			slog.Error(fmt.Sprintf("err while accept: %v\n", err))
			return fmt.Errorf("couldn't listen to network: %v\n", err)
		}

		go s.handle(ctx, conn)
	}
}

func (s *Server) handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	slog.Info(fmt.Sprintf("Request arrive from: %s", conn.RemoteAddr().String()))

	for {
		select {
		case <-ctx.Done():
			return
		default:
			command.Handle(conn, s.m)
			return
		}
	}
}

func NewServer(port string, cap int) (*Server, error) {
	m, err := lru.NewLRUEngine(cap, ".")
	if err != nil {
		return nil, err
	}
	return &Server{
		Port: port,
		m:    m,
	}, nil
}
