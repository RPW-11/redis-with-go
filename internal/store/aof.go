package store

import (
	"fmt"
	"os"
	"path/filepath"
)

const AofFilename = "data.aof"

// aofLogger appends RESP-encoded commands to data.aof.
type aofLogger struct {
	Dir string
	f   *os.File
}

func newAofLogger(dir string) (*aofLogger, error) {
	path := filepath.Join(dir, AofFilename)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &aofLogger{Dir: dir, f: f}, nil
}

func (aof *aofLogger) Close() error {
	return aof.f.Close()
}

func (aof *aofLogger) LogSet(key string, val []byte) error {
	_, err := fmt.Fprintf(aof.f, "*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(val), val)
	return err
}

func (aof *aofLogger) LogDelete(key string, _ []byte) error {
	_, err := fmt.Fprintf(aof.f, "*2\r\n$3\r\nDEL\r\n$%d\r\n%s\r\n", len(key), key)
	return err
}

func (aof *aofLogger) LogExpireAt(key string, unixTs []byte) error {
	_, err := fmt.Fprintf(aof.f, "*3\r\n$8\r\nEXPIREAT\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(unixTs), unixTs)
	return err
}
