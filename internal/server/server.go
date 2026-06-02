package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os/signal"
	"syscall"

	"github.com/RPW-11/redis-with-go/internal/command"
	"github.com/RPW-11/redis-with-go/internal/resp"
	"github.com/RPW-11/redis-with-go/internal/store"
)

type Server struct {
	Port string
	m    *store.Store
}

func (s *Server) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", s.Port))
	if err != nil {
		slog.Error("couldn't listen to network")
		return fmt.Errorf("couldn't listen to network")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	defer func() {
		if err := l.Close(); err != nil && ctx.Err() == nil {
			slog.Error("error closing the network")
		}
	}()

	slog.Info(fmt.Sprintf("Server is running on port: %s\n", s.Port))

	go func() {
		<-ctx.Done()
		l.Close()
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			if ctx.Err() != nil {
				slog.Info("server shutting down")
				return nil
			}
			slog.Error(fmt.Sprintf("err while accept: %v", err))
			return fmt.Errorf("couldn't listen to network: %v", err)
		}

		go s.handle(ctx, conn)
	}
}

func (s *Server) handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	slog.Info(fmt.Sprintf("Request arrive from: %s", conn.RemoteAddr().String()))

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-done:
		}
	}()

	for {
		v := command.Handle(ctx, conn, s.m)
		b, _ := resp.Serialize(v)
		if _, err := conn.Write(b); err != nil {
			return
		}
	}
}

func NewServer(port, aofPath string, cap int) (*Server, error) {
	m, err := store.NewStore(cap, aofPath)
	if err != nil {
		return nil, err
	}
	return &Server{
		Port: port,
		m:    m,
	}, nil
}
