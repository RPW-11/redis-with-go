package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func writeAOF(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, AofFilename), []byte(content), 0644); err != nil {
		t.Fatalf("writeAOF: %v", err)
	}
}

func respSet(key, val string) string {
	return fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(val), val)
}

func respDel(key string) string {
	return fmt.Sprintf("*2\r\n$3\r\nDEL\r\n$%d\r\n%s\r\n", len(key), key)
}

func respExpireAt(key string, ts int64) string {
	tsStr := strconv.FormatInt(ts, 10)
	return fmt.Sprintf("*3\r\n$8\r\nEXPIREAT\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(tsStr), tsStr)
}

func TestReplayAOF(t *testing.T) {
	t.Run("missing aof file is a no-op", func(t *testing.T) {
		s := newStore(t, 100)
		if err := s.ReplayAOF(t.TempDir()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("replays SET", func(t *testing.T) {
		dir := t.TempDir()
		writeAOF(t, dir, respSet("foo", "bar"))

		s := newStore(t, 100)
		if err := s.ReplayAOF(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := s.Get("foo")
		if !ok || string(v) != "bar" {
			t.Fatalf("expected foo=bar, got %q ok=%v", v, ok)
		}
	})

	t.Run("replays DEL", func(t *testing.T) {
		dir := t.TempDir()
		writeAOF(t, dir, respSet("foo", "bar")+respDel("foo"))

		s := newStore(t, 100)
		if err := s.ReplayAOF(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := s.Get("foo"); ok {
			t.Fatal("expected foo to be deleted")
		}
	})

	t.Run("replays EXPIREAT with future timestamp", func(t *testing.T) {
		dir := t.TempDir()
		future := time.Now().Add(time.Hour).Unix()
		writeAOF(t, dir, respSet("foo", "bar")+respExpireAt("foo", future))

		s := newStore(t, 100)
		if err := s.ReplayAOF(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		v, ok := s.Get("foo")
		if !ok || string(v) != "bar" {
			t.Fatalf("expected foo=bar still alive, got %q ok=%v", v, ok)
		}
	})

	t.Run("EXPIREAT with past timestamp removes key", func(t *testing.T) {
		dir := t.TempDir()
		past := time.Now().Add(-time.Hour).Unix()
		writeAOF(t, dir, respSet("foo", "bar")+respExpireAt("foo", past))

		s := newStore(t, 100)
		if err := s.ReplayAOF(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := s.Get("foo"); ok {
			t.Fatal("expected foo to be expired and removed")
		}
	})

	t.Run("does not log to AOF during replay", func(t *testing.T) {
		dir := t.TempDir()
		content := respSet("foo", "bar")
		writeAOF(t, dir, content)

		s := newStore(t, 100)
		if err := s.ReplayAOF(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		info, err := os.Stat(filepath.Join(dir, AofFilename))
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		if info.Size() != int64(len(content)) {
			t.Fatal("aof file was modified during replay — re-logging detected")
		}
	})

	t.Run("replays multiple commands", func(t *testing.T) {
		dir := t.TempDir()
		writeAOF(t, dir,
			respSet("a", "1")+
				respSet("b", "2")+
				respSet("c", "3")+
				respDel("b"),
		)

		s := newStore(t, 100)
		if err := s.ReplayAOF(dir); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v, ok := s.Get("a"); !ok || string(v) != "1" {
			t.Fatalf("expected a=1, got %q ok=%v", v, ok)
		}
		if _, ok := s.Get("b"); ok {
			t.Fatal("expected b to be deleted")
		}
		if v, ok := s.Get("c"); !ok || string(v) != "3" {
			t.Fatalf("expected c=3, got %q ok=%v", v, ok)
		}
	})
}
