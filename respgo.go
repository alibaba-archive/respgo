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
	typeSimpleStrings = "+"
	typeErrors        = "-"
	typeIntegers      = ":"
	typeBulkStrings   = "$"
	typeArrays        = "*"
)

const (
	crlf                = "\r\n"
	bulkStringMaxLength = 512 * 1024 * 1024
)

// EncodeString encodes a simple string
func EncodeString(s string) []byte {
	return []byte(typeSimpleStrings + s + crlf)
}

// EncodeError encodes a error string
func EncodeError(s string) []byte {
	return []byte(typeErrors + s + crlf)
}

// EncodeInt encodes an int
func EncodeInt(s int64) []byte {
	return []byte(typeIntegers + strconv.FormatInt(s, 10) + crlf)
}

// EncodeBulkString encodes a bulk string
func EncodeBulkString(s string) []byte {
	if len(s) > bulkStringMaxLength {
		panic("BulkString is over 512 MB")
	}
	return []byte(typeBulkStrings + strconv.Itoa(len(s)) + crlf + s + crlf)
}

//EncodeNull encodes null value
func EncodeNull() []byte {
	return []byte(typeBulkStrings + "-1" + crlf)
}

// EncodeNullArray encodes null array
func EncodeNullArray() []byte {
	return []byte(typeArrays + "-1" + crlf)
}

// EncodeArray encode a slice of byte slice. It accepts the results of other encode method including itself.
// For example: EncodeArray([][]byte{EncodeInt(1), EncodeNull()})
func EncodeArray(s [][]byte) []byte {
	var buf bytes.Buffer
	buf.WriteString(typeArrays)
	buf.WriteString(strconv.Itoa(len(s)))
	buf.WriteString(crlf)
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
	case typeSimpleStrings:
		result = line
	case typeErrors:
		result = errors.New(line)
	case typeIntegers:
		result, err = strconv.ParseInt(line, 10, 64)
	case typeBulkStrings:
		var length int
		length, err = strconv.Atoi(line)
		if err != nil {
			return
		}
		if length > bulkStringMaxLength {
			err = errors.New("BulkString is over 512 MB")
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
	case typeArrays:
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
