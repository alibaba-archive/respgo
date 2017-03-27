package respgo

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
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
	if strings.ContainsAny(s, "\r\n") {
		panic("SimpleString should not contain \\r\\n")
	}
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
	lineLen := len(line)
	if lineLen < 3 {
		err = fmt.Errorf(`line is too short: %#v`, line)
		return
	}
	if line[lineLen-2] != '\r' || line[lineLen-1] != '\n' {
		err = fmt.Errorf("invalid CRLF: %#v", line)
		return
	}
	msgType, line := string(line[0]), line[1:lineLen-2]
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
		if length == -1 {
			return
		}
		if length > bulkStringMaxLength || length < -1 {
			err = fmt.Errorf("invalid Bulk Strings length: %#v", length)
			return
		}
		buff := make([]byte, length+2)
		_, err = io.ReadFull(reader, buff)
		if err != nil {
			return
		}
		if buff[length] != '\r' || buff[length+1] != '\n' {
			err = fmt.Errorf("invalid CRLF: %#v", buff)
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
		err = fmt.Errorf("invalid RESP type: %#v", msgType)
	}
	return
}
