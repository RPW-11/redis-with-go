package command

import (
	"bufio"
	"fmt"
	"net"

	ds "github.com/RPW-11/redis-with-go/internal/data_structures"
	"github.com/RPW-11/redis-with-go/internal/resp"
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

func Handle(conn net.Conn, m *ds.RedisMap) {
	rd := bufio.NewReader(conn)
	v, err := resp.Parse(rd)
	if err != nil {
		fmt.Fprintf(conn, "-ERR %v\r\n", err)
		return
	}
	if v.Typ != resp.ArrayType {
		fmt.Fprintf(conn, "-ERR command format must be an array\r\n")
		return
	}
	if len(v.Arr) == 0 {
		fmt.Fprintf(conn, "-ERR empty command\r\n")
		return
	}

	cmd := v.Arr[0].Bytes

	switch Command(cmd) {
	case Set:
		err = handleSetCmd(v.Arr, m)
		if err != nil {
			fmt.Fprintf(conn, "-ERR %v\r\n", err)
			return
		}
		conn.Write([]byte("+OK\r\n"))
		return
	case Get:
		v, err := handleGetCmd(v.Arr, m)
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
		err = handleDelCmd(v.Arr, m)
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

	fmt.Fprintf(conn, "-ERR command is invalid\r\n")
}

func handleSetCmd(arr []*resp.Value, m *ds.RedisMap) error {
	if len(arr) != 3 {
		return fmt.Errorf("invalid set command")
	}
	key, val := arr[1], arr[2]

	if key.Typ != resp.BulkStringType || val.Typ != resp.BulkStringType {
		return fmt.Errorf("key and value must be bulks string")
	}
	if len(key.Bytes) == 0 {
		return fmt.Errorf("key cannot be empty")
	}

	m.Set(string(key.Bytes), val.Bytes)

	return nil
}

func handleGetCmd(arr []*resp.Value, m *ds.RedisMap) ([]byte, error) {
	if len(arr) != 2 {
		return nil, fmt.Errorf("invalid get command")
	}
	if arr[1].Typ != resp.BulkStringType {
		return nil, fmt.Errorf("key must be a bulk string")
	}

	key := string(arr[1].Bytes)

	v, _ := m.Get(key)

	return v, nil
}

func handleDelCmd(arr []*resp.Value, m *ds.RedisMap) error {
	if len(arr) != 2 {
		return fmt.Errorf("invalid del command")
	}
	if arr[1].Typ != resp.BulkStringType {
		return fmt.Errorf("key must be a bulk string")
	}

	key := string(arr[1].Bytes)

	m.Delete(key)

	return nil
}

func handleExpireCmd() {

}

func handleTtlCmd() {

}
