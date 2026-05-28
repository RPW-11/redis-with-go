package resp

import (
	"errors"
	"testing"
)

func TestSerialize_InvalidType(t *testing.T) {
	_, err := Serialize(&Value{Typ: DataType('?')})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestSerializeString(t *testing.T) {
	t.Run("simple string", func(t *testing.T) {
		s, err := Serialize(NewString("OK"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "+OK\r\n" {
			t.Fatalf("expected '+OK\\r\\n', got %q", s)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		s, err := Serialize(NewString(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "+\r\n" {
			t.Fatalf("expected '+\\r\\n', got %q", s)
		}
	})
}

func TestSerializeError(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		s, err := Serialize(NewError(errors.New("something went wrong")))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "-ERR something went wrong\r\n" {
			t.Fatalf("unexpected output: %q", s)
		}
	})
}

func TestSerializeBulkString(t *testing.T) {
	t.Run("simple bulk string", func(t *testing.T) {
		s, err := Serialize(NewBulkString("hello"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "$5\r\nhello\r\n" {
			t.Fatalf("expected '$5\\r\\nhello\\r\\n', got %q", s)
		}
	})

	t.Run("empty bulk string", func(t *testing.T) {
		s, err := Serialize(NewBulkString(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "$0\r\n\r\n" {
			t.Fatalf("expected '$0\\r\\n\\r\\n', got %q", s)
		}
	})
}

func TestSerializeInteger(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		s, err := Serialize(NewInteger(42))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != ":42\r\n" {
			t.Fatalf("expected ':42\\r\\n', got %q", s)
		}
	})

	t.Run("zero", func(t *testing.T) {
		s, err := Serialize(NewInteger(0))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != ":0\r\n" {
			t.Fatalf("expected ':0\\r\\n', got %q", s)
		}
	})

	t.Run("negative", func(t *testing.T) {
		s, err := Serialize(NewInteger(-99))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != ":-99\r\n" {
			t.Fatalf("expected ':-99\\r\\n', got %q", s)
		}
	})
}

func TestSerializeArray(t *testing.T) {
	t.Run("empty array", func(t *testing.T) {
		s, err := Serialize(NewArray([]*Value{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "*0\r\n" {
			t.Fatalf("expected '*0\\r\\n', got %q", s)
		}
	})

	t.Run("array of strings", func(t *testing.T) {
		s, err := Serialize(NewArray([]*Value{
			NewString("a"),
			NewString("b"),
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "*2\r\n+a\r\n+b\r\n" {
			t.Fatalf("unexpected output: %q", s)
		}
	})

	t.Run("nested array", func(t *testing.T) {
		inner := NewArray([]*Value{NewString("x")})
		s, err := Serialize(NewArray([]*Value{inner}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "*1\r\n*1\r\n+x\r\n" {
			t.Fatalf("unexpected output: %q", s)
		}
	})

	t.Run("array with invalid element returns error", func(t *testing.T) {
		_, err := Serialize(NewArray([]*Value{
			{Typ: DataType('?')},
		}))
		if err == nil {
			t.Fatal("expected error for invalid element type")
		}
	})
}

func TestSerializeNull(t *testing.T) {
	s, err := Serialize(NewNull())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != "_\r\n" {
		t.Fatalf("expected '_\\r\\n', got %q", s)
	}
}

func TestSerializeBoolean(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		s, err := Serialize(NewBoolean(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "#t\r\n" {
			t.Fatalf("expected '#t\\r\\n', got %q", s)
		}
	})

	t.Run("false", func(t *testing.T) {
		s, err := Serialize(NewBoolean(false))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "#f\r\n" {
			t.Fatalf("expected '#f\\r\\n', got %q", s)
		}
	})
}

func TestSerializeDouble(t *testing.T) {
	t.Run("float", func(t *testing.T) {
		s, err := Serialize(NewDouble(3.14))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != ",3.14\r\n" {
			t.Fatalf("expected ',3.14\\r\\n', got %q", s)
		}
	})

	t.Run("integer-like", func(t *testing.T) {
		s, err := Serialize(NewDouble(42))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != ",42\r\n" {
			t.Fatalf("expected ',42\\r\\n', got %q", s)
		}
	})

	t.Run("negative", func(t *testing.T) {
		s, err := Serialize(NewDouble(-1.5))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != ",-1.5\r\n" {
			t.Fatalf("expected ',-1.5\\r\\n', got %q", s)
		}
	})

	t.Run("zero", func(t *testing.T) {
		s, err := Serialize(NewDouble(0))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != ",0\r\n" {
			t.Fatalf("expected ',0\\r\\n', got %q", s)
		}
	})
}

// Round-trip: serialize then parse back and verify the value is intact.
func TestSerializeRoundTrip(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		s, _ := Serialize(NewString("hello"))
		v, err := Parse(newReader(s))
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if v.Str != "hello" {
			t.Fatalf("expected 'hello', got %q", v.Str)
		}
	})

	t.Run("bulk string", func(t *testing.T) {
		s, _ := Serialize(NewBulkString("world"))
		v, err := Parse(newReader(s))
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if string(v.Bytes) != "world" {
			t.Fatalf("expected 'world', got %q", v.Bytes)
		}
	})

	t.Run("integer", func(t *testing.T) {
		s, _ := Serialize(NewInteger(123))
		v, err := Parse(newReader(s))
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if v.Num != 123 {
			t.Fatalf("expected 123, got %d", v.Num)
		}
	})

	t.Run("boolean", func(t *testing.T) {
		s, _ := Serialize(NewBoolean(true))
		v, err := Parse(newReader(s))
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if !v.Bool {
			t.Fatal("expected Bool=true")
		}
	})

	t.Run("null", func(t *testing.T) {
		s, _ := Serialize(NewNull())
		v, err := Parse(newReader(s))
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if !v.IsNull {
			t.Fatal("expected IsNull=true")
		}
	})

	t.Run("double", func(t *testing.T) {
		s, _ := Serialize(NewDouble(2.5))
		v, err := Parse(newReader(s))
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if v.Float != 2.5 {
			t.Fatalf("expected 2.5, got %f", v.Float)
		}
	})

	t.Run("array of bulk strings", func(t *testing.T) {
		s, _ := Serialize(NewArray([]*Value{
			NewBulkString("SET"),
			NewBulkString("key"),
			NewBulkString("val"),
		}))
		v, err := Parse(newReader(s))
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if len(v.Arr) != 3 {
			t.Fatalf("expected 3 elements, got %d", len(v.Arr))
		}
		expected := []string{"SET", "key", "val"}
		for i, elem := range v.Arr {
			if string(elem.Bytes) != expected[i] {
				t.Fatalf("arr[%d]: expected %q, got %q", i, expected[i], elem.Bytes)
			}
		}
	})
}
