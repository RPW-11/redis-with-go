package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Value struct {
	typ    byte
	str    string
	num    int
	bl     bool
	arr    []*Value
	isNull bool
}

var (
	ErrInvalidType = errors.New("parse: invalid type")
)

const (
	CRLF = "\r\n"
	LF   = byte('\n')
)

const (
	StringType  = byte('+')
	ErrorType   = byte('-')
	IntegerType = byte(':')
	BStringType = byte('$')
	ArrayType   = byte('*')
	NullType    = byte('_')
	BooleanType = byte('#')
	DoubleType  = byte(',')
)

func parse(rd *bufio.Reader) (Value, error) {
	typ, err := rd.ReadByte()
	if err != nil {
		return Value{}, fmt.Errorf("parse: %w", err)
	}

	switch typ {
	case StringType:
		return parseString(rd)
	case IntegerType:
		return parseIntegerType(rd)
	case BStringType:
		return parseBString(rd)
	case ArrayType:
		return parseArrayType(rd)
	case NullType:
		return parseNullType(rd)
	case BooleanType:
		return parseBooleanType(rd)
	case DoubleType:
		return parseDoubleType(rd)
	}

	return Value{}, fmt.Errorf("parse: %w", ErrInvalidType)
}

func parseString(rd *bufio.Reader) (Value, error) {
	s, err := rd.ReadString(LF)
	if err != nil {
		return Value{}, err
	}

	s = strings.TrimRight(s, CRLF)

	return Value{
		typ: StringType,
		str: s,
	}, nil
}

func parseBString(rd *bufio.Reader) (Value, error) {
	s, err := rd.ReadString(LF)
	if err != nil {
		return Value{}, err
	}

	s = strings.TrimRight(s, CRLF)

	l, err := strconv.Atoi(s)
	if err != nil {
		return Value{}, err
	}

	payload := make([]byte, l)
	_, err = io.ReadFull(rd, payload)
	if err != nil {
		return Value{}, err
	}

	arr := make([]*Value, 0, l)
	for v := range strings.SplitSeq(string(payload), CRLF) {
		arr = append(arr, &Value{
			typ: StringType,
			str: v,
		})
	}

	rd.ReadByte()
	rd.ReadByte()

	return Value{
		typ: BStringType,
		arr: arr,
	}, nil
}

func parseIntegerType(rd *bufio.Reader) (Value, error) {
	s, err := rd.ReadString(LF)
	if err != nil {
		return Value{}, err
	}

	s = strings.TrimRight(s, CRLF)

	num, err := strconv.Atoi(s)
	if err != nil {
		return Value{}, err
	}

	return Value{
		typ: IntegerType,
		num: num,
	}, nil
}

func parseArrayType(rd *bufio.Reader) (Value, error) {
	s, err := rd.ReadString(LF)
	if err != nil {
		return Value{}, err
	}

	s = strings.TrimRight(s, CRLF)

	l, err := strconv.Atoi(s)
	if err != nil {
		return Value{}, err
	}

	arr := make([]*Value, 0, l)
	for l > 0 {
		data, err := parse(rd)
		if err != nil {
			return Value{}, err
		}

		arr = append(arr, &data)
		l--
	}

	return Value{
		typ: ArrayType,
		arr: arr,
	}, nil
}

func parseNullType(rd *bufio.Reader) (Value, error) {
	_, err := rd.ReadString(LF)
	if err != nil {
		return Value{}, err
	}

	return Value{
		typ:    NullType,
		isNull: true,
	}, nil
}

func parseBooleanType(rd *bufio.Reader) (Value, error) {
	s, err := rd.ReadString(LF)
	if err != nil {
		return Value{}, err
	}

	s = strings.TrimRight(s, CRLF)

	if len(s) == 0 || s[0] != byte('t') && s[0] != byte('f') {
		return Value{}, fmt.Errorf("payload of boolean must be 't' or 'f'\n")
	}

	return Value{
		typ: BooleanType,
		bl:  s[0] == byte('t'),
	}, nil
}

// TODO: Implement this
func parseDoubleType(rd *bufio.Reader) (Value, error) {
	return Value{}, nil
}
