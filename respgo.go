package respgo

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
)

// Data type
const (
	TypeSimpleStrings = "+"
	TypeErrors        = "-"
	TypeIntegers      = ":"
	TypeBulkStrings   = "$"
	TypeArrays        = "*"
)

const (
	// CRLF ...
	CRLF = "\r\n"
)

// EncodeString encodes a simple string
func EncodeString(s string) []byte {
	return []byte(TypeSimpleStrings + s + CRLF)
}

// EncodeError encodes a error string
func EncodeError(s string) []byte {
	return []byte(TypeErrors + s + CRLF)
}

// EncodeInt encodes an int
func EncodeInt(s int64) []byte {
	return []byte(TypeIntegers + strconv.FormatInt(s, 10) + CRLF)
}

// EncodeBulkString encodes a bulk string
func EncodeBulkString(s string) []byte {
	return []byte(TypeBulkStrings + strconv.Itoa(len(s)) + CRLF + s + CRLF)
}

//EncodeNull encodes null value
func EncodeNull() []byte {
	return []byte(TypeBulkStrings + "-1" + CRLF)
}

// EncodeNullArray encodes null array
func EncodeNullArray() []byte {
	return []byte(TypeArrays + "-1" + CRLF)
}

// EncodeArray encode a slice of byte slice. It accepts the results of other encode method including itself.
// For example: EncodeArray([][]byte{EncodeInt(1), EncodeNull()})
func EncodeArray(s [][]byte) []byte {
	var buf bytes.Buffer
	buf.WriteString(TypeArrays)
	buf.WriteString(strconv.Itoa(len(s)))
	buf.WriteString(CRLF)
	for _, val := range s {
		buf.Write(val)
	}
	return buf.Bytes()
}

// Decode decode from reader
func Decode(reader *bufio.Reader) (result interface{}, err error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	if len(line) < 3 {
		err = errors.New("line is too short: " + line)
		return
	}
	msgType, line := string(line[0]), line[1:len(line)-2]
	switch msgType {
	case TypeSimpleStrings:
		result = line
	case TypeErrors:
		result = errors.New(line)
	case TypeIntegers:
		result, err = strconv.ParseInt(line, 10, 64)
	case TypeBulkStrings:
		var length int
		length, err = strconv.Atoi(line)
		if err != nil {
			return
		}
		if length == -1 {
			return
		}
		buff := make([]byte, length+2)
		_, err = io.ReadFull(reader, buff)
		if err != nil {
			return
		}
		result = string(buff[:length])
	case TypeArrays:
		var length int
		length, err = strconv.Atoi(line)
		if length == -1 {
			return
		}
		array := make([]interface{}, length)
		for i := 0; i < length; i++ {
			array[i], err = Decode(reader)
			if err != nil {
				return
			}
		}
		result = array
	default:
		err = errors.New("invalid type: " + msgType)
	}
	return
}
