package resp

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
	_, err := Parse(newReader("@unknown\r\n"))
	if !errors.Is(err, ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
}

func TestParse_EOF(t *testing.T) {
	_, err := Parse(newReader(""))
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF, got %v", err)
	}
}

func TestParseString(t *testing.T) {
	t.Run("simple string", func(t *testing.T) {
		v, err := Parse(newReader("+OK\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Typ != StringType {
			t.Fatalf("expected StringType, got '%c'", v.Typ)
		}
		if v.Str != "OK" {
			t.Fatalf("expected 'OK', got %q", v.Str)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		v, err := Parse(newReader("+\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Str != "" {
			t.Fatalf("expected empty string, got %q", v.Str)
		}
	})

	t.Run("string with spaces", func(t *testing.T) {
		v, err := Parse(newReader("+hello world\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Str != "hello world" {
			t.Fatalf("expected 'hello world', got %q", v.Str)
		}
	})

	t.Run("truncated input", func(t *testing.T) {
		_, err := Parse(newReader("+OK"))
		if err == nil {
			t.Fatal("expected error on truncated input")
		}
	})
}

func TestParseInteger(t *testing.T) {
	t.Run("positive integer", func(t *testing.T) {
		v, err := Parse(newReader(":42\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Typ != IntegerType {
			t.Fatalf("expected IntegerType, got '%c'", v.Typ)
		}
		if v.Num != 42 {
			t.Fatalf("expected 42, got %d", v.Num)
		}
	})

	t.Run("zero", func(t *testing.T) {
		v, err := Parse(newReader(":0\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Num != 0 {
			t.Fatalf("expected 0, got %d", v.Num)
		}
	})

	t.Run("negative integer", func(t *testing.T) {
		v, err := Parse(newReader(":-99\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Num != -99 {
			t.Fatalf("expected -99, got %d", v.Num)
		}
	})

	t.Run("non-numeric value", func(t *testing.T) {
		_, err := Parse(newReader(":abc\r\n"))
		if err == nil {
			t.Fatal("expected error for non-numeric integer")
		}
	})
}

func TestParseBString(t *testing.T) {
	t.Run("simple bulk string", func(t *testing.T) {
		v, err := Parse(newReader("$5\r\nhello\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Typ != BulkStringType {
			t.Fatalf("expected BulkStringType, got '%c'", v.Typ)
		}
		if string(v.Bytes) != "hello" {
			t.Fatalf("expected 'hello', got %q", string(v.Bytes))
		}
	})

	t.Run("empty bulk string", func(t *testing.T) {
		v, err := Parse(newReader("$0\r\n\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Typ != BulkStringType {
			t.Fatalf("expected BulkStringType, got '%c'", v.Typ)
		}
		if string(v.Bytes) != "" {
			t.Fatalf("expected empty string, got %q", string(v.Bytes))
		}
	})

	t.Run("payload preserved as raw string", func(t *testing.T) {
		v, err := Parse(newReader("$11\r\nhello world\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(v.Bytes) != "hello world" {
			t.Fatalf("expected 'hello world', got %q", string(v.Bytes))
		}
	})

	t.Run("non-numeric length", func(t *testing.T) {
		_, err := Parse(newReader("$abc\r\nhello\r\n"))
		if err == nil {
			t.Fatal("expected error for non-numeric length")
		}
	})

	t.Run("truncated payload", func(t *testing.T) {
		// declares 10 bytes but stream only has 2
		_, err := Parse(newReader("$10\r\nhi\r\n"))
		if err == nil {
			t.Fatal("expected error for truncated payload")
		}
	})
}

func TestParseArray(t *testing.T) {
	t.Run("empty array", func(t *testing.T) {
		v, err := Parse(newReader("*0\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Typ != ArrayType {
			t.Fatalf("expected ArrayType, got '%c'", v.Typ)
		}
		if len(v.Arr) != 0 {
			t.Fatalf("expected empty arr, got len %d", len(v.Arr))
		}
	})

	t.Run("array of strings", func(t *testing.T) {
		v, err := Parse(newReader("*2\r\n+hello\r\n+world\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(v.Arr) != 2 {
			t.Fatalf("expected arr length 2, got %d", len(v.Arr))
		}
		if v.Arr[0].Str != "hello" || v.Arr[1].Str != "world" {
			t.Fatalf("unexpected values: %q, %q", v.Arr[0].Str, v.Arr[1].Str)
		}
	})

	t.Run("array of integers", func(t *testing.T) {
		v, err := Parse(newReader("*3\r\n:1\r\n:2\r\n:3\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []int{1, 2, 3}
		for i, elem := range v.Arr {
			if elem.Num != expected[i] {
				t.Fatalf("arr[%d]: expected %d, got %d", i, expected[i], elem.Num)
			}
		}
	})

	t.Run("nested array", func(t *testing.T) {
		v, err := Parse(newReader("*1\r\n*2\r\n+a\r\n+b\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(v.Arr) != 1 {
			t.Fatalf("expected outer arr length 1, got %d", len(v.Arr))
		}
		inner := v.Arr[0]
		if inner.Typ != ArrayType {
			t.Fatalf("expected inner ArrayType, got '%c'", inner.Typ)
		}
		if len(inner.Arr) != 2 {
			t.Fatalf("expected inner arr length 2, got %d", len(inner.Arr))
		}
		if inner.Arr[0].Str != "a" || inner.Arr[1].Str != "b" {
			t.Fatalf("unexpected inner values: %q, %q", inner.Arr[0].Str, inner.Arr[1].Str)
		}
	})

	t.Run("non-numeric count", func(t *testing.T) {
		_, err := Parse(newReader("*abc\r\n"))
		if err == nil {
			t.Fatal("expected error for non-numeric count")
		}
	})

	t.Run("truncated array", func(t *testing.T) {
		// declares 3 elements but only 2 are present
		_, err := Parse(newReader("*3\r\n+a\r\n+b\r\n"))
		if err == nil {
			t.Fatal("expected error for truncated array")
		}
	})
}

func TestParseBoolean(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		v, err := Parse(newReader("#t\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Typ != BooleanType {
			t.Fatalf("expected BooleanType, got '%c'", v.Typ)
		}
		if !v.Bool {
			t.Fatal("expected bl=true")
		}
	})

	t.Run("false", func(t *testing.T) {
		v, err := Parse(newReader("#f\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Bool {
			t.Fatal("expected bl=false")
		}
	})

	t.Run("invalid value", func(t *testing.T) {
		_, err := Parse(newReader("#x\r\n"))
		if err == nil {
			t.Fatal("expected error for invalid boolean value")
		}
	})
}

func TestParseDouble(t *testing.T) {
	t.Run("positive float", func(t *testing.T) {
		v, err := Parse(newReader(",3.14\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Typ != DoubleType {
			t.Fatalf("expected DoubleType, got '%c'", v.Typ)
		}
		if v.Float != 3.14 {
			t.Fatalf("expected 3.14, got %f", v.Float)
		}
	})

	t.Run("negative float", func(t *testing.T) {
		v, err := Parse(newReader(",-1.5\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Float != -1.5 {
			t.Fatalf("expected -1.5, got %f", v.Float)
		}
	})

	t.Run("zero", func(t *testing.T) {
		v, err := Parse(newReader(",0\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Float != 0 {
			t.Fatalf("expected 0, got %f", v.Float)
		}
	})

	t.Run("integer-like value", func(t *testing.T) {
		v, err := Parse(newReader(",42\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Float != 42.0 {
			t.Fatalf("expected 42.0, got %f", v.Float)
		}
	})

	t.Run("non-numeric value", func(t *testing.T) {
		_, err := Parse(newReader(",abc\r\n"))
		if err == nil {
			t.Fatal("expected error for non-numeric double")
		}
	})

	t.Run("truncated input", func(t *testing.T) {
		_, err := Parse(newReader(",3.14"))
		if err == nil {
			t.Fatal("expected error on truncated input")
		}
	})
}

func TestParseNull(t *testing.T) {
	t.Run("valid null", func(t *testing.T) {
		v, err := Parse(newReader("_\r\n"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !v.IsNull {
			t.Fatal("expected isNull=true")
		}
	})

	t.Run("null consumes CRLF leaving stream intact", func(t *testing.T) {
		rd := newReader("_\r\n+OK\r\n")
		_, err := Parse(rd)
		if err != nil {
			t.Fatalf("unexpected error parsing null: %v", err)
		}
		v, err := Parse(rd)
		if err != nil {
			t.Fatalf("expected next parse to succeed, got: %v", err)
		}
		if v.Str != "OK" {
			t.Fatalf("expected 'OK' after null, got %q", v.Str)
		}
	})
}
