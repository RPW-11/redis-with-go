package command

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/RPW-11/redis-with-go/internal/lru"
	"github.com/RPW-11/redis-with-go/internal/resp"
)

type Command string

const (
	Set    Command = "SET"
	Get    Command = "GET"
	Del    Command = "DEL"
	Expire Command = "EXPIRE"
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

func Handle(conn net.Conn, lru *lru.LRUEngine) {
	rd := bufio.NewReader(conn)
	v, err := resp.Parse(rd)
	if err != nil {
		errBytes, _ := resp.Serialize(resp.NewError(err))
		conn.Write(errBytes)
		return
	}
	if v.Typ != resp.ArrayType {
		errBytes, _ := resp.Serialize(resp.NewError(errors.New("command must be in array type")))
		conn.Write(errBytes)
		return
	}
	if len(v.Arr) == 0 {
		errBytes, _ := resp.Serialize(resp.NewError(errors.New("empty payload")))
		conn.Write(errBytes)
		return
	}

	cmd := v.Arr[0].Bytes

	switch Command(cmd) {
	case Set:
		err = handleSetCmd(v.Arr, lru)
		if err != nil {
			errBytes, _ := resp.Serialize(resp.NewError(err))
			conn.Write(errBytes)
			return
		}
		res, _ := resp.Serialize(resp.NewString("OK"))
		conn.Write(res)
		return
	case Get:
		v, err := handleGetCmd(v.Arr, lru)
		if err != nil {
			errBytes, _ := resp.Serialize(resp.NewError(err))
			conn.Write(errBytes)
			return
		}
		if v == nil {
			nilBytes, _ := resp.Serialize(resp.NewNull())
			conn.Write(nilBytes)
			return
		}

		data, _ := resp.Serialize(resp.NewBulkString(v))
		conn.Write(data)
		return
	case Del:
		err = handleDelCmd(v.Arr, lru)
		if err != nil {
			errBytes, _ := resp.Serialize(resp.NewError(err))
			conn.Write(errBytes)
			return
		}
		res, _ := resp.Serialize(resp.NewString("OK"))
		conn.Write(res)
		return
	case Expire:
		ok, err := handleExpireCmd(v.Arr, lru)
		if err != nil {
			errBytes, _ := resp.Serialize(resp.NewError(err))
			conn.Write(errBytes)
			return
		}
		val := 0
		if ok {
			val = 1
		}
		res, _ := resp.Serialize(resp.NewInteger(val))
		conn.Write(res)
		return
	case Ttl:
		sec, err := handleTtlCmd(v.Arr, lru)
		if err != nil {
			errBytes, _ := resp.Serialize(resp.NewError(err))
			conn.Write(errBytes)
			return
		}

		res, _ := resp.Serialize(resp.NewInteger(sec))
		conn.Write(res)
		return
	}

	errBytes, _ := resp.Serialize(resp.NewError(errors.New("command is invalid")))
	conn.Write(errBytes)
}

func handleSetCmd(arr []*resp.Value, lru *lru.LRUEngine) error {
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

	lru.Set(string(key.Bytes), val.Bytes)

	return nil
}

func handleGetCmd(arr []*resp.Value, lru *lru.LRUEngine) ([]byte, error) {
	if len(arr) != 2 {
		return nil, fmt.Errorf("invalid get command")
	}
	if arr[1].Typ != resp.BulkStringType {
		return nil, fmt.Errorf("key must be a bulk string")
	}

	key := string(arr[1].Bytes)

	v, _ := lru.Get(key)

	return v, nil
}

func handleDelCmd(arr []*resp.Value, lru *lru.LRUEngine) error {
	if len(arr) != 2 {
		return fmt.Errorf("invalid del command")
	}
	if arr[1].Typ != resp.BulkStringType {
		return fmt.Errorf("key must be a bulk string")
	}

	key := string(arr[1].Bytes)

	lru.Delete(key)

	return nil
}

func handleExpireCmd(arr []*resp.Value, lru *lru.LRUEngine) (bool, error) {
	if len(arr) != 3 {
		return false, fmt.Errorf("invalid expire command")
	}
	if arr[1].Typ != resp.BulkStringType {
		return false, fmt.Errorf("key must be a bulk string")
	}
	if arr[2].Typ != resp.BulkStringType {
		return false, fmt.Errorf("key must be a bulk string")
	}

	key := string(arr[1].Bytes)
	sec, err := strconv.Atoi(string(arr[2].Bytes))
	if err != nil {
		return false, err
	}
	if sec < 0 {
		return false, fmt.Errorf("invalid expire time in 'expire' command")
	}

	t := time.Now().Add(time.Duration(sec) * time.Second)
	ok := lru.SetExpiry(key, t)

	if !ok {
		return false, nil
	}

	return true, nil
}

func handleTtlCmd(arr []*resp.Value, lru *lru.LRUEngine) (int, error) {
	if len(arr) != 2 {
		return 0, fmt.Errorf("invalid ttl command")
	}
	if arr[1].Typ != resp.BulkStringType {
		return 0, fmt.Errorf("key must be a bulk string")
	}

	key := string(arr[1].Bytes)
	t, ok := lru.ExpiryOf(key)
	if !ok {
		return -2, nil // indicates the key does not exist
	}
	if t.IsZero() {
		return -1, nil
	}

	return int(time.Until(t).Seconds()), nil
}
