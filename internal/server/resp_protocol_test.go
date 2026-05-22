package server

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"testing"
)

func newReader(s string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(s))
}

func TestParse_InvalidType(t *testing.T) {
	_, err := parse(newReader("@unknown\r\n"))
	if !errors.Is(err, ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
}

func TestParse_EOF(t *testing.T) {
	_, err := parse(newReader(""))
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}

func TestParseString(t *testing.T) {
	t.Run("simple string", func(t *testing.T) {
		v, err := parse(newReader("+OK\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.typ != StringType {
			t.Fatalf("expected StringType, got '%c'", v.typ)
		}
		if v.str != "OK" {
			t.Fatalf("expected 'OK', got %q", v.str)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		v, err := parse(newReader("+\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.str != "" {
			t.Fatalf("expected empty string, got %q", v.str)
		}
	})

	t.Run("string with spaces", func(t *testing.T) {
		v, err := parse(newReader("+hello world\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.str != "hello world" {
			t.Fatalf("expected 'hello world', got %q", v.str)
		}
	})

	t.Run("truncated input", func(t *testing.T) {
		_, err := parse(newReader("+OK"))
		if err == nil {
			t.Fatal("expected error on truncated input")
		}
	})
}

func TestParseInteger(t *testing.T) {
	t.Run("positive integer", func(t *testing.T) {
		v, err := parse(newReader(":42\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.typ != IntegerType {
			t.Fatalf("expected IntegerType, got '%c'", v.typ)
		}
		if v.num != 42 {
			t.Fatalf("expected 42, got %d", v.num)
		}
	})

	t.Run("zero", func(t *testing.T) {
		v, err := parse(newReader(":0\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.num != 0 {
			t.Fatalf("expected 0, got %d", v.num)
		}
	})

	t.Run("negative integer", func(t *testing.T) {
		v, err := parse(newReader(":-99\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.num != -99 {
			t.Fatalf("expected -99, got %d", v.num)
		}
	})

	t.Run("non-numeric value", func(t *testing.T) {
		_, err := parse(newReader(":abc\r\n"))
		if err == nil {
			t.Fatal("expected error for non-numeric integer")
		}
	})
}

func TestParseBString(t *testing.T) {
	t.Run("simple bulk string", func(t *testing.T) {
		v, err := parse(newReader("$5\r\nhello\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.typ != BStringType {
			t.Fatalf("expected BStringType, got '%c'", v.typ)
		}
		if len(v.arr) != 1 {
			t.Fatalf("expected arr length 1, got %d", len(v.arr))
		}
		if v.arr[0].str != "hello" {
			t.Fatalf("expected 'hello', got %q", v.arr[0].str)
		}
	})

	t.Run("empty bulk string", func(t *testing.T) {
		v, err := parse(newReader("$0\r\n\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(v.arr) != 1 {
			t.Fatalf("expected arr length 1, got %d", len(v.arr))
		}
		if v.arr[0].str != "" {
			t.Fatalf("expected empty string segment, got %q", v.arr[0].str)
		}
	})

	// payload containing CRLF is split into multiple arr segments
	t.Run("payload with CRLF splits into segments", func(t *testing.T) {
		v, err := parse(newReader("$12\r\nhello\r\nworld\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(v.arr) != 2 {
			t.Fatalf("expected arr length 2, got %d", len(v.arr))
		}
		if v.arr[0].str != "hello" {
			t.Fatalf("expected arr[0]='hello', got %q", v.arr[0].str)
		}
		if v.arr[1].str != "world" {
			t.Fatalf("expected arr[1]='world', got %q", v.arr[1].str)
		}
	})

	t.Run("non-numeric length", func(t *testing.T) {
		_, err := parse(newReader("$abc\r\nhello\r\n"))
		if err == nil {
			t.Fatal("expected error for non-numeric length")
		}
	})

	t.Run("truncated payload", func(t *testing.T) {
		// declares 10 bytes but stream only has 2
		_, err := parse(newReader("$10\r\nhi\r\n"))
		if err == nil {
			t.Fatal("expected error for truncated payload")
		}
	})
}

func TestParseArray(t *testing.T) {
	t.Run("empty array", func(t *testing.T) {
		v, err := parse(newReader("*0\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.typ != ArrayType {
			t.Fatalf("expected ArrayType, got '%c'", v.typ)
		}
		if len(v.arr) != 0 {
			t.Fatalf("expected empty arr, got len %d", len(v.arr))
		}
	})

	t.Run("array of strings", func(t *testing.T) {
		v, err := parse(newReader("*2\r\n+hello\r\n+world\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(v.arr) != 2 {
			t.Fatalf("expected arr length 2, got %d", len(v.arr))
		}
		if v.arr[0].str != "hello" || v.arr[1].str != "world" {
			t.Fatalf("unexpected values: %q, %q", v.arr[0].str, v.arr[1].str)
		}
	})

	t.Run("array of integers", func(t *testing.T) {
		v, err := parse(newReader("*3\r\n:1\r\n:2\r\n:3\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []int{1, 2, 3}
		for i, elem := range v.arr {
			if elem.num != expected[i] {
				t.Fatalf("arr[%d]: expected %d, got %d", i, expected[i], elem.num)
			}
		}
	})

	t.Run("nested array", func(t *testing.T) {
		v, err := parse(newReader("*1\r\n*2\r\n+a\r\n+b\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(v.arr) != 1 {
			t.Fatalf("expected outer arr length 1, got %d", len(v.arr))
		}
		inner := v.arr[0]
		if inner.typ != ArrayType {
			t.Fatalf("expected inner ArrayType, got '%c'", inner.typ)
		}
		if len(inner.arr) != 2 {
			t.Fatalf("expected inner arr length 2, got %d", len(inner.arr))
		}
		if inner.arr[0].str != "a" || inner.arr[1].str != "b" {
			t.Fatalf("unexpected inner values: %q, %q", inner.arr[0].str, inner.arr[1].str)
		}
	})

	t.Run("non-numeric count", func(t *testing.T) {
		_, err := parse(newReader("*abc\r\n"))
		if err == nil {
			t.Fatal("expected error for non-numeric count")
		}
	})

	t.Run("truncated array", func(t *testing.T) {
		// declares 3 elements but only 2 are present
		_, err := parse(newReader("*3\r\n+a\r\n+b\r\n"))
		if err == nil {
			t.Fatal("expected error for truncated array")
		}
	})
}

func TestParseBoolean(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		v, err := parse(newReader("#t\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.typ != BooleanType {
			t.Fatalf("expected BooleanType, got '%c'", v.typ)
		}
		if !v.bl {
			t.Fatal("expected bl=true")
		}
	})

	t.Run("false", func(t *testing.T) {
		v, err := parse(newReader("#f\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.bl {
			t.Fatal("expected bl=false")
		}
	})

	t.Run("invalid value", func(t *testing.T) {
		_, err := parse(newReader("#x\r\n"))
		if err == nil {
			t.Fatal("expected error for invalid boolean value")
		}
	})
}

func TestParseNull(t *testing.T) {
	t.Run("valid null", func(t *testing.T) {
		v, err := parse(newReader("_\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !v.isNull {
			t.Fatal("expected isNull=true")
		}
	})

	t.Run("null consumes CRLF leaving stream intact", func(t *testing.T) {
		rd := newReader("_\r\n+OK\r\n")
		_, err := parse(rd)
		if err != nil {
			t.Fatalf("unexpected error parsing null: %v", err)
		}
		v, err := parse(rd)
		if err != nil {
			t.Fatalf("expected next parse to succeed, got: %v", err)
		}
		if v.str != "OK" {
			t.Fatalf("expected 'OK' after null, got %q", v.str)
		}
	})
}
