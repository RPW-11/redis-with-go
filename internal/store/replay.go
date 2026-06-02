package store

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/RPW-11/redis-with-go/internal/resp"
)

func (s *Store) ReplayAOF(dir string) error {
	path := filepath.Join(dir, AofFilename)
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("replay: %w", err)
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	for {
		if _, err := rd.Peek(1); errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return fmt.Errorf("replay: %w", err)
		}

		v, err := resp.Parse(rd)
		if err != nil {
			return fmt.Errorf("replay: %w", err)
		}
		if v.Typ != resp.ArrayType || len(v.Arr) == 0 {
			return fmt.Errorf("replay: expected non-empty array")
		}

		cmd := string(v.Arr[0].Bytes)
		switch cmd {
		case "SET":
			if len(v.Arr) != 3 {
				return fmt.Errorf("replay: SET expects 3 elements")
			}
			s.Set(string(v.Arr[1].Bytes), v.Arr[2].Bytes)

		case "DEL":
			if len(v.Arr) != 2 {
				return fmt.Errorf("replay: DEL expects 2 elements")
			}
			s.Delete(string(v.Arr[1].Bytes))

		case "EXPIREAT":
			if len(v.Arr) != 3 {
				return fmt.Errorf("replay: EXPIREAT expects 3 elements")
			}
			ts, err := strconv.ParseInt(string(v.Arr[2].Bytes), 10, 64)
			if err != nil {
				return fmt.Errorf("replay: invalid EXPIREAT timestamp: %w", err)
			}
			s.SetExpiry(string(v.Arr[1].Bytes), time.Unix(ts, 0))

		default:
			return fmt.Errorf("replay: unknown command %q", cmd)
		}
	}
}
