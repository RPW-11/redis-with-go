package resp

import (
	"fmt"
	"strconv"
	"strings"
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

func NewBulkString(s string) *Value {
	return &Value{
		Typ:   BulkStringType,
		Bytes: []byte(s),
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

func Serialize(v *Value) (string, error) {
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

	return "", fmt.Errorf("invalid type")
}

func serializeString(v *Value) string {
	return "+" + v.Str + "\r\n"
}

func serializeError(v *Value) string {
	return "-ERR " + v.Str + "\r\n"
}

func serializeBulkString(v *Value) string {
	return "$" + strconv.Itoa(len(v.Bytes)) + "\r\n" + string(v.Bytes) + "\r\n"
}

func serializeInteger(v *Value) string {
	return ":" + strconv.Itoa(v.Num) + "\r\n"
}

func serializeArray(v *Value) (string, error) {
	var sb strings.Builder
	sb.WriteString("*" + strconv.Itoa(len(v.Arr)) + "\r\n")
	for _, elem := range v.Arr {
		s, err := Serialize(elem)
		if err != nil {
			return "", err
		}
		sb.WriteString(s)
	}
	return sb.String(), nil
}

func serializeNull() string {
	return "_\r\n"
}

func serializeBoolean(v *Value) string {
	if v.Bool {
		return "#t\r\n"
	}
	return "#f\r\n"
}

func serializeDouble(v *Value) string {
	return "," + strconv.FormatFloat(v.Float, 'f', -1, 64) + "\r\n"
}
