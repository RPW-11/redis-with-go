package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"

	ds "github.com/RPW-11/redis-with-go/internal/data_structures"
)

type Command string

const (
	Set    Command = "SET"
	Get    Command = "GET"
	Del    Command = "DEL"
	Expire Command = "Expire"
	Ttl    Command = "TTL"
)

const (
	MaxBulkStringLength = 5_000_000
	MaxCmdLen           = 4
)

type Request struct {
	cmd     Command
	payload []byte
}

var (
	ErrInvalidMessageType = errors.New("command type must start with * followed by $")
	ErrInvalidCommand     = errors.New("invalid command")
	ErrMaxCmdLen          = errors.New("command length must be <= 4")
)

func readCRLF(rd *bufio.Reader) ([]byte, error) {
	b, err := rd.ReadBytes(LF)
	if err != nil {
		return nil, err
	}

	if len(b) < 2 || b[len(b)-2] != '\r' {
		return nil, errors.New("protocol invalid")
	}

	return b[:len(b)-2], nil
}

func readBulkString(rd *bufio.Reader) ([]byte, error) {
	typ, err := rd.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("readBulkString: %w", err)
	}
	if DataType(typ) != BStringType {
		return nil, fmt.Errorf("readBulkString: invalid message type")
	}

	b, err := readCRLF(rd)
	if err != nil {
		return nil, err
	}

	l, err := strconv.Atoi(string(b))
	if err != nil {
		return nil, err
	}
	if l < 0 {
		return nil, fmt.Errorf("invalid bulk string length: %d", l)
	}
	if l > MaxBulkStringLength {
		return nil, fmt.Errorf("bulk string length exceeds %d", MaxBulkStringLength)
	}

	b = make([]byte, l)
	_, err = io.ReadFull(rd, b)
	if err != nil {
		return nil, err
	}
	rd.ReadByte()
	rd.ReadByte()

	return b, nil
}

func parseCommand(rd *bufio.Reader) (string, error) {
	// Read the message type. It has to be an array for a command
	typ, err := rd.ReadByte()
	if err != nil {
		return "", fmt.Errorf("parseCommand: %w", err)
	}
	if DataType(typ) != ArrayType {
		return "", fmt.Errorf("parseCommand: %w", ErrInvalidMessageType)
	}

	// Read the array length
	b, err := readCRLF(rd)
	if err != nil {
		return "", err
	}

	arrLength, err := strconv.Atoi(string(b))
	if err != nil {
		return "", err
	}
	if arrLength > MaxCmdLen {
		return "", ErrMaxCmdLen
	}

	// Read bulk string of the command
	b, err = readBulkString(rd)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func handleCommand(conn net.Conn, m *ds.RedisMap) {
	rd := bufio.NewReader(conn)
	cmd, err := parseCommand(rd)
	if err != nil {
		fmt.Fprintf(conn, "-ERR %v\r\n", err)
		return
	}

	switch Command(cmd) {
	case Set:
		err = handleSetCmd(rd, m)
		if err != nil {
			fmt.Fprintf(conn, "-ERR %v\r\n", err)
			return
		}
		conn.Write([]byte("+OK\r\n"))
		return
	case Get:
		v, err := handleGetCmd(rd, m)
		if err != nil {
			fmt.Fprintf(conn, "-ERR %v\r\n", err)
			return
		}
		if v == nil {
			conn.Write([]byte("_\r\n"))
			return
		}

		data := fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
		conn.Write([]byte(data))
		return
	case Del:
		err = handleDelCmd(rd, m)
		if err != nil {
			fmt.Fprintf(conn, "-ERR %v\r\n", err)
			return
		}
		conn.Write([]byte("+OK\r\n"))
		return
	case Expire:
		handleExpireCmd()
		return
	case Ttl:
		handleTtlCmd()
		return
	}

	fmt.Fprintf(conn, "-ERR %v\r\n", ErrInvalidCommand)
}

func handleSetCmd(rd *bufio.Reader, m *ds.RedisMap) error {
	b, err := readBulkString(rd)
	if err != nil {
		return err
	}
	key := string(b)

	val, err := readBulkString(rd)
	if err != nil {
		return err
	}

	m.Set(key, val)

	return nil
}

func handleGetCmd(rd *bufio.Reader, m *ds.RedisMap) ([]byte, error) {
	b, err := readBulkString(rd)
	if err != nil {
		return nil, err
	}
	key := string(b)

	v, _ := m.Get(key)

	return v, nil
}

func handleDelCmd(rd *bufio.Reader, m *ds.RedisMap) error {
	b, err := readBulkString(rd)
	if err != nil {
		return err
	}
	key := string(b)

	m.Delete(key)

	return nil
}

func handleExpireCmd() {

}

func handleTtlCmd() {

}
