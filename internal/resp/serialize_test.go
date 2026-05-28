package resp

import (
	"bytes"
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
		b, err := Serialize(NewString("OK"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte("+OK\r\n")) {
			t.Fatalf("expected '+OK\\r\\n', got %q", b)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		b, err := Serialize(NewString(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte("+\r\n")) {
			t.Fatalf("expected '+\\r\\n', got %q", b)
		}
	})
}

func TestSerializeError(t *testing.T) {
	t.Run("error message", func(t *testing.T) {
		b, err := Serialize(NewError(errors.New("something went wrong")))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte("-ERR something went wrong\r\n")) {
			t.Fatalf("unexpected output: %q", b)
		}
	})
}

func TestSerializeBulkString(t *testing.T) {
	t.Run("simple bulk string", func(t *testing.T) {
		b, err := Serialize(NewBulkString([]byte("hello")))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte("$5\r\nhello\r\n")) {
			t.Fatalf("expected '$5\\r\\nhello\\r\\n', got %q", b)
		}
	})

	t.Run("empty bulk string", func(t *testing.T) {
		b, err := Serialize(NewBulkString([]byte("")))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte("$0\r\n\r\n")) {
			t.Fatalf("expected '$0\\r\\n\\r\\n', got %q", b)
		}
	})
}

func TestSerializeInteger(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		b, err := Serialize(NewInteger(42))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte(":42\r\n")) {
			t.Fatalf("expected ':42\\r\\n', got %q", b)
		}
	})

	t.Run("zero", func(t *testing.T) {
		b, err := Serialize(NewInteger(0))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte(":0\r\n")) {
			t.Fatalf("expected ':0\\r\\n', got %q", b)
		}
	})

	t.Run("negative", func(t *testing.T) {
		b, err := Serialize(NewInteger(-99))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte(":-99\r\n")) {
			t.Fatalf("expected ':-99\\r\\n', got %q", b)
		}
	})
}

func TestSerializeArray(t *testing.T) {
	t.Run("empty array", func(t *testing.T) {
		b, err := Serialize(NewArray([]*Value{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte("*0\r\n")) {
			t.Fatalf("expected '*0\\r\\n', got %q", b)
		}
	})

	t.Run("array of strings", func(t *testing.T) {
		b, err := Serialize(NewArray([]*Value{
			NewString("a"),
			NewString("b"),
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte("*2\r\n+a\r\n+b\r\n")) {
			t.Fatalf("unexpected output: %q", b)
		}
	})

	t.Run("nested array", func(t *testing.T) {
		inner := NewArray([]*Value{NewString("x")})
		b, err := Serialize(NewArray([]*Value{inner}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte("*1\r\n*1\r\n+x\r\n")) {
			t.Fatalf("unexpected output: %q", b)
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
	b, err := Serialize(NewNull())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(b, []byte("_\r\n")) {
		t.Fatalf("expected '_\\r\\n', got %q", b)
	}
}

func TestSerializeBoolean(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		b, err := Serialize(NewBoolean(true))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte("#t\r\n")) {
			t.Fatalf("expected '#t\\r\\n', got %q", b)
		}
	})

	t.Run("false", func(t *testing.T) {
		b, err := Serialize(NewBoolean(false))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte("#f\r\n")) {
			t.Fatalf("expected '#f\\r\\n', got %q", b)
		}
	})
}

func TestSerializeDouble(t *testing.T) {
	t.Run("float", func(t *testing.T) {
		b, err := Serialize(NewDouble(3.14))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte(",3.14\r\n")) {
			t.Fatalf("expected ',3.14\\r\\n', got %q", b)
		}
	})

	t.Run("integer-like", func(t *testing.T) {
		b, err := Serialize(NewDouble(42))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte(",42\r\n")) {
			t.Fatalf("expected ',42\\r\\n', got %q", b)
		}
	})

	t.Run("negative", func(t *testing.T) {
		b, err := Serialize(NewDouble(-1.5))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte(",-1.5\r\n")) {
			t.Fatalf("expected ',-1.5\\r\\n', got %q", b)
		}
	})

	t.Run("zero", func(t *testing.T) {
		b, err := Serialize(NewDouble(0))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !bytes.Equal(b, []byte(",0\r\n")) {
			t.Fatalf("expected ',0\\r\\n', got %q", b)
		}
	})
}

// Round-trip: serialize then parse back and verify the value is intact.
func TestSerializeRoundTrip(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		b, _ := Serialize(NewString("hello"))
		v, err := Parse(newReader(string(b)))
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if v.Str != "hello" {
			t.Fatalf("expected 'hello', got %q", v.Str)
		}
	})

	t.Run("bulk string", func(t *testing.T) {
		b, _ := Serialize(NewBulkString([]byte("world")))
		v, err := Parse(newReader(string(b)))
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if string(v.Bytes) != "world" {
			t.Fatalf("expected 'world', got %q", v.Bytes)
		}
	})

	t.Run("integer", func(t *testing.T) {
		b, _ := Serialize(NewInteger(123))
		v, err := Parse(newReader(string(b)))
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if v.Num != 123 {
			t.Fatalf("expected 123, got %d", v.Num)
		}
	})

	t.Run("boolean", func(t *testing.T) {
		b, _ := Serialize(NewBoolean(true))
		v, err := Parse(newReader(string(b)))
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if !v.Bool {
			t.Fatal("expected Bool=true")
		}
	})

	t.Run("null", func(t *testing.T) {
		b, _ := Serialize(NewNull())
		v, err := Parse(newReader(string(b)))
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if !v.IsNull {
			t.Fatal("expected IsNull=true")
		}
	})

	t.Run("double", func(t *testing.T) {
		b, _ := Serialize(NewDouble(2.5))
		v, err := Parse(newReader(string(b)))
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if v.Float != 2.5 {
			t.Fatalf("expected 2.5, got %f", v.Float)
		}
	})

	t.Run("array of bulk strings", func(t *testing.T) {
		b, _ := Serialize(NewArray([]*Value{
			NewBulkString([]byte("SET")),
			NewBulkString([]byte("key")),
			NewBulkString([]byte("val")),
		}))
		v, err := Parse(newReader(string(b)))
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
