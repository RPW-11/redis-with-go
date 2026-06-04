package resp

import (
	"bytes"
	"fmt"
	"strconv"
)

func NewString(s string) *Value {
	return &Value{
		Typ: StringType,
		Str: s,
	}
}

func NewError(err error) *Value {
	return &Value{
		Typ: ErrorType,
		Str: err.Error(),
	}
}

func NewBulkString(b []byte) *Value {
	return &Value{
		Typ:   BulkStringType,
		Bytes: b,
	}
}

func NewInteger(n int) *Value {
	return &Value{
		Typ: IntegerType,
		Num: n,
	}
}

func NewArray(arr []*Value) *Value {
	return &Value{
		Typ: ArrayType,
		Arr: arr,
	}
}

func NewNull() *Value {
	return &Value{
		Typ:    NullType,
		IsNull: true,
	}
}

func NewBoolean(b bool) *Value {
	return &Value{
		Typ:  BooleanType,
		Bool: b,
	}
}

func NewDouble(f float64) *Value {
	return &Value{
		Typ:   DoubleType,
		Float: f,
	}
}

// Serialize encodes v into its RESP3 byte representation.
func Serialize(v *Value) ([]byte, error) {
	switch v.Typ {
	case StringType:
		return serializeString(v), nil
	case ErrorType:
		return serializeError(v), nil
	case BulkStringType:
		return serializeBulkString(v), nil
	case IntegerType:
		return serializeInteger(v), nil
	case ArrayType:
		return serializeArray(v)
	case NullType:
		return serializeNull(), nil
	case BooleanType:
		return serializeBoolean(v), nil
	case DoubleType:
		return serializeDouble(v), nil
	}

	return nil, fmt.Errorf("invalid type")
}

func serializeString(v *Value) []byte {
	return []byte("+" + v.Str + "\r\n")
}

func serializeError(v *Value) []byte {
	return []byte("-ERR " + v.Str + "\r\n")
}

func serializeBulkString(v *Value) []byte {
	prefix := []byte("$" + strconv.Itoa(len(v.Bytes)) + "\r\n")
	suffix := []byte("\r\n")
	out := make([]byte, 0, len(prefix)+len(v.Bytes)+len(suffix))
	out = append(out, prefix...)
	out = append(out, v.Bytes...)
	out = append(out, suffix...)
	return out
}

func serializeInteger(v *Value) []byte {
	return []byte(":" + strconv.Itoa(v.Num) + "\r\n")
}

func serializeArray(v *Value) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(len(v.Arr)) + "\r\n")
	for _, elem := range v.Arr {
		b, err := Serialize(elem)
		if err != nil {
			return nil, err
		}
		buf.Write(b)
	}
	return buf.Bytes(), nil
}

func serializeNull() []byte {
	return []byte("_\r\n")
}

func serializeBoolean(v *Value) []byte {
	if v.Bool {
		return []byte("#t\r\n")
	}
	return []byte("#f\r\n")
}

func serializeDouble(v *Value) []byte {
	return []byte("," + strconv.FormatFloat(v.Float, 'f', -1, 64) + "\r\n")
}
