package server

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
)

type Server struct {
	Port string
}

func (s *Server) Run() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", s.Port))
	if err != nil {
		slog.Error("couldn't listen to network")
		return fmt.Errorf("couldn't listen to network\n")
	}

	slog.Info(fmt.Sprintf("Server is running on port: %s\n", s.Port))

	for {
		conn, err := l.Accept()
		if err != nil {
			slog.Error(fmt.Sprintf("err while accept: %v\n", err))
			return fmt.Errorf("couldn't listen to network: %v\n", err)
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	slog.Info(fmt.Sprintf("Request arrive from: %s", conn.RemoteAddr().String()))
	defer conn.Close()

	rd := bufio.NewReader(conn)
	v, err := parse(rd)
	if err != nil {
		slog.Error(fmt.Sprintf("processing conn: %v\n", err))
		conn.Write([]byte("invalid type\n"))
		return
	}

	fmt.Println("Received value:", v)

	conn.Write([]byte("Your type is valid\n"))
}

func NewServer(port string) *Server {
	return &Server{
		Port: port,
	}
}
