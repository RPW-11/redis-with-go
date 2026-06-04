// Package resp implements a RESP3 parser and value types.
package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// DataType is the single-byte prefix that identifies a RESP3 type.
type DataType byte

// Value holds the parsed result of a single RESP3 element.
type Value struct {
	Typ    DataType
	Str    string
	Bytes  []byte
	Num    int
	Bool   bool
	Float  float64
	Arr    []*Value
	IsNull bool
}

var (
	ErrInvalidType = errors.New("invalid type")
)

const LF = '\n'

const (
	MaxBulkStringLength = 5_000_000
	MaxArrayLength      = 4_000_000
)

// RESP3 type prefixes as defined in the protocol spec.
const (
	StringType     DataType = '+'
	ErrorType      DataType = '-'
	IntegerType    DataType = ':'
	BulkStringType DataType = '$'
	ArrayType      DataType = '*'
	NullType       DataType = '_'
	BooleanType    DataType = '#'
	DoubleType     DataType = ','
)

func readCRLF(rd *bufio.Reader) ([]byte, error) {
	b, err := rd.ReadBytes(LF)
	if err != nil {
		return nil, err
	}

	if len(b) < 2 || b[len(b)-2] != '\r' {
		return nil, errors.New("protocol invalid")
	}

	return b[:len(b)-2], nil
}

// Parse reads one RESP3 value from rd. The first byte determines the type
// (e.g. '*' for array, '$' for bulk string) and the rest is parsed accordingly.
func Parse(rd *bufio.Reader) (Value, error) {
	typ, err := rd.ReadByte()
	if err != nil {
		return Value{}, fmt.Errorf("parse: %w", err)
	}

	switch DataType(typ) {
	case StringType:
		return ParseString(rd)
	case IntegerType:
		return ParseIntegerType(rd)
	case BulkStringType:
		return ParseBulkString(rd)
	case ArrayType:
		return ParseArrayType(rd)
	case NullType:
		return ParseNullType(rd)
	case BooleanType:
		return ParseBooleanType(rd)
	case DoubleType:
		return ParseDoubleType(rd)
	}

	return Value{}, fmt.Errorf("parse: %w", ErrInvalidType)
}

func ParseString(rd *bufio.Reader) (Value, error) {
	b, err := readCRLF(rd)
	if err != nil {
		return Value{}, err
	}

	return Value{
		Typ: StringType,
		Str: string(b),
	}, nil
}

func ParseBulkString(rd *bufio.Reader) (Value, error) {
	b, err := readCRLF(rd)
	if err != nil {
		return Value{}, err
	}

	l, err := strconv.Atoi(string(b))
	if err != nil {
		return Value{}, err
	}
	if l < 0 {
		return Value{}, fmt.Errorf("negative bulk string length: %d", l)
	}
	if l > MaxBulkStringLength {
		return Value{}, fmt.Errorf("bulk string length exceeds %d", MaxBulkStringLength)
	}

	payload := make([]byte, l)
	_, err = io.ReadFull(rd, payload)
	if err != nil {
		return Value{}, err
	}

	rd.ReadByte()
	rd.ReadByte()

	return Value{
		Typ:   BulkStringType,
		Bytes: payload,
	}, nil
}

func ParseIntegerType(rd *bufio.Reader) (Value, error) {
	b, err := readCRLF(rd)
	if err != nil {
		return Value{}, err
	}

	num, err := strconv.Atoi(string(b))
	if err != nil {
		return Value{}, err
	}

	return Value{
		Typ: IntegerType,
		Num: num,
	}, nil
}

func ParseArrayType(rd *bufio.Reader) (Value, error) {
	b, err := readCRLF(rd)
	if err != nil {
		return Value{}, err
	}

	l, err := strconv.Atoi(string(b))
	if err != nil {
		return Value{}, err
	}
	if l < 0 {
		return Value{}, fmt.Errorf("negative array length: %d", l)
	}
	if l > MaxArrayLength {
		return Value{}, fmt.Errorf("array length exceeds %d", MaxArrayLength)
	}

	arr := make([]*Value, 0, l)
	for l > 0 {
		data, err := Parse(rd)
		if err != nil {
			return Value{}, err
		}

		arr = append(arr, &data)
		l--
	}

	return Value{
		Typ: ArrayType,
		Arr: arr,
	}, nil
}

func ParseNullType(rd *bufio.Reader) (Value, error) {
	b, err := readCRLF(rd)
	if err != nil {
		return Value{}, err
	}

	if len(b) != 0 {
		return Value{}, errors.New("null length must be 0")
	}

	return Value{
		Typ:    NullType,
		IsNull: true,
	}, nil
}

func ParseBooleanType(rd *bufio.Reader) (Value, error) {
	b, err := readCRLF(rd)
	if err != nil {
		return Value{}, err
	}

	if len(b) != 1 || b[0] != 't' && b[0] != 'f' {
		return Value{}, fmt.Errorf("payload of boolean must be 't' or 'f'\n")
	}

	return Value{
		Typ:  BooleanType,
		Bool: b[0] == 't',
	}, nil
}

func ParseDoubleType(rd *bufio.Reader) (Value, error) {
	b, err := readCRLF(rd)
	if err != nil {
		return Value{}, err
	}

	fl, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return Value{}, err
	}

	return Value{
		Typ:   DoubleType,
		Float: fl,
	}, nil
}
