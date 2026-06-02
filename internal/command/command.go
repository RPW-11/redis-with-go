package command

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/RPW-11/redis-with-go/internal/resp"
	"github.com/RPW-11/redis-with-go/internal/store"
)

type Command string

const (
	Set    Command = "SET"
	Get    Command = "GET"
	Del    Command = "DEL"
	Expire   Command = "EXPIRE"
	ExpireAt Command = "EXPIREAT"
	Ttl      Command = "TTL"
)

const (
	MaxBulkStringLength = 5_000_000
	MaxCmdLen           = 4
)

type Request struct {
	cmd     Command
	payload []byte
}

func Handle(ctx context.Context, conn net.Conn, lru *store.Store) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	rd := bufio.NewReader(conn)
	v, err := resp.Parse(rd)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		errBytes, _ := resp.Serialize(resp.NewError(err))
		conn.Write(errBytes)
		return err
	}
	if v.Typ != resp.ArrayType {
		errBytes, _ := resp.Serialize(resp.NewError(errors.New("command must be in array type")))
		conn.Write(errBytes)
		return nil
	}
	if len(v.Arr) == 0 {
		errBytes, _ := resp.Serialize(resp.NewError(errors.New("empty payload")))
		conn.Write(errBytes)
		return nil
	}

	cmd := v.Arr[0].Bytes

	switch Command(cmd) {
	case Set:
		err = handleSetCmd(v.Arr, lru)
		if err != nil {
			errBytes, _ := resp.Serialize(resp.NewError(err))
			conn.Write(errBytes)
			return nil
		}
		res, _ := resp.Serialize(resp.NewString("OK"))
		conn.Write(res)
		return nil
	case Get:
		v, err := handleGetCmd(v.Arr, lru)
		if err != nil {
			errBytes, _ := resp.Serialize(resp.NewError(err))
			conn.Write(errBytes)
			return nil
		}
		if v == nil {
			nilBytes, _ := resp.Serialize(resp.NewNull())
			conn.Write(nilBytes)
			return nil
		}
		data, _ := resp.Serialize(resp.NewBulkString(v))
		conn.Write(data)
		return nil
	case Del:
		err = handleDelCmd(v.Arr, lru)
		if err != nil {
			errBytes, _ := resp.Serialize(resp.NewError(err))
			conn.Write(errBytes)
			return nil
		}
		res, _ := resp.Serialize(resp.NewString("OK"))
		conn.Write(res)
		return nil
	case Expire:
		ok, err := handleExpireCmd(v.Arr, lru)
		if err != nil {
			errBytes, _ := resp.Serialize(resp.NewError(err))
			conn.Write(errBytes)
			return nil
		}
		val := 0
		if ok {
			val = 1
		}
		res, _ := resp.Serialize(resp.NewInteger(val))
		conn.Write(res)
		return nil
	case ExpireAt:
		ok, err := handleExpireAtCmd(v.Arr, lru)
		if err != nil {
			errBytes, _ := resp.Serialize(resp.NewError(err))
			conn.Write(errBytes)
			return nil
		}
		val := 0
		if ok {
			val = 1
		}
		res, _ := resp.Serialize(resp.NewInteger(val))
		conn.Write(res)
		return nil
	case Ttl:
		sec, err := handleTtlCmd(v.Arr, lru)
		if err != nil {
			errBytes, _ := resp.Serialize(resp.NewError(err))
			conn.Write(errBytes)
			return nil
		}
		res, _ := resp.Serialize(resp.NewInteger(sec))
		conn.Write(res)
		return nil
	}

	errBytes, _ := resp.Serialize(resp.NewError(errors.New("command is invalid")))
	conn.Write(errBytes)
	return nil
}

func handleSetCmd(arr []*resp.Value, lru *store.Store) error {
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

func handleGetCmd(arr []*resp.Value, lru *store.Store) ([]byte, error) {
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

func handleDelCmd(arr []*resp.Value, lru *store.Store) error {
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

func applyExpireAt(key string, expiry time.Time, lru *store.Store) (bool, error) {
	if time.Now().After(expiry) {
		_, ok := lru.Get(key)
		if !ok {
			return false, nil
		}
		lru.Delete(key)
		return true, nil
	}
	return lru.SetExpiry(key, expiry), nil
}

func handleExpireCmd(arr []*resp.Value, lru *store.Store) (bool, error) {
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
	sec, err := strconv.ParseInt(string(arr[2].Bytes), 10, 64)
	if err != nil {
		return false, err
	}
	if sec < 0 {
		return false, fmt.Errorf("invalid expire time in 'expire' command")
	}

	return applyExpireAt(key, time.Now().Add(time.Duration(sec)*time.Second), lru)
}

func handleExpireAtCmd(arr []*resp.Value, lru *store.Store) (bool, error) {
	if len(arr) != 3 {
		return false, fmt.Errorf("invalid expireat command")
	}
	if arr[1].Typ != resp.BulkStringType {
		return false, fmt.Errorf("key must be a bulk string")
	}
	if arr[2].Typ != resp.BulkStringType {
		return false, fmt.Errorf("timestamp must be a bulk string")
	}

	key := string(arr[1].Bytes)
	ts, err := strconv.ParseInt(string(arr[2].Bytes), 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid unix timestamp in 'expireat' command")
	}
	if ts < 0 {
		return false, fmt.Errorf("invalid expire time in 'expireat' command")
	}

	return applyExpireAt(key, time.Unix(ts, 0), lru)
}

func handleTtlCmd(arr []*resp.Value, lru *store.Store) (int, error) {
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
