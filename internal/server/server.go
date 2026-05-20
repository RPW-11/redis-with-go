package server

import (
	"fmt"
	"io"
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

	for {
		b := [1024]byte{}
		n, err := conn.Read(b[:])
		if err != nil {
			if err != io.EOF {
				slog.Error(fmt.Sprintf("error reading data from the connection: %v\n", err))
			}
			return
		}

		fmt.Printf("Data from conn: %s", string(b[:n]))

		conn.Write([]byte("Your message has been received\n"))
	}
}

func NewServer(port string) *Server {
	return &Server{
		Port: port,
	}
}
