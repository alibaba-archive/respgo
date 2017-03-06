package respgo

import (
	"bytes"
	"errors"
	"strconv"
)

const (
	// SimpleStrings For Simple Strings the first byte of the reply is "+"
	SimpleStrings = '+'
	// Errors For Errors the first byte of the reply is "-"
	Errors = '-'
	//Integers For Integers the first byte of the reply is ":"
	Integers = ':'
	// BulkStrings For Bulk Strings the first byte of the reply is "$"
	BulkStrings = '$'
	// Arrays For Arrays the first byte of the reply is "*"
	Arrays = '*'
)

var (
	//OK Common responses
	OK = EncodeString("OK")
	//PONG ...
	PONG = EncodeString("PONG")
	// CRLF is "\r\n"
	CRLF = []byte{'\r', '\n'}
)

// CheckType ...
func CheckType(resp []byte) (result int) {
	result = int(resp[0])
	if result == SimpleStrings || result == Errors || result == Integers || result == BulkStrings || result == Arrays {
		return
	}
	return -1
}

// Decode ...
func Decode(resp []byte) (result interface{}, err error) {
	err = checkError(resp)
	if err != nil {
		return
	}
	_, result, err = parseBuffer(resp)
	return
}

// DecodeToString ...
func DecodeToString(resp []byte) (result string, err error) {
	val, err := Decode(resp)
	var OK bool
	if result, OK = val.(string); !OK {
		err = errors.New("invalid string or bulkstring type")
	}
	return
}

// DecodeToInt ...
func DecodeToInt(resp []byte) (result int, err error) {
	val, err := Decode(resp)
	var OK bool
	if result, OK = val.(int); !OK {
		err = errors.New("invalid int type")
	}
	return
}

// DecodeToArray ...
func DecodeToArray(resp []byte) (result []interface{}, err error) {
	val, err := Decode(resp)
	var OK bool
	if result, OK = val.([]interface{}); !OK {
		err = errors.New("invalid array type")
	}
	return
}

// EncodeString returns a simple string with the given contents.
func EncodeString(s string) []byte {
	var buf bytes.Buffer
	buf.WriteByte(SimpleStrings)
	buf.WriteString(s)
	buf.Write(CRLF)
	return buf.Bytes()
}

// EncodeError ...
func EncodeError(s string) []byte {
	var buf bytes.Buffer
	buf.WriteByte(Errors)
	buf.WriteString(s)
	buf.Write(CRLF)
	return buf.Bytes()
}

// EncodeBulkString a bulk string with the given contents.
func EncodeBulkString(s string) []byte {
	var buf bytes.Buffer
	buf.WriteByte(BulkStrings)
	buf.WriteString(strconv.Itoa(len(s)))
	buf.Write(CRLF)
	buf.WriteString(s)
	buf.Write(CRLF)
	return buf.Bytes()
}

// EncodeInt ....
func EncodeInt(s int) []byte {
	var buf bytes.Buffer
	buf.WriteByte(Integers)
	buf.WriteString(strconv.Itoa(s))
	buf.Write(CRLF)
	return buf.Bytes()
}

// EncodeArray ...
func EncodeArray(s [][]byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte(Arrays)
	buf.WriteString(strconv.Itoa(len(s)))
	buf.Write(CRLF)
	for _, val := range s {
		buf.Write(val)
	}
	return buf.Bytes()
}
func readLine(resp []byte) (c []byte, foward int) {
	index := 0
	for i := range resp {
		if resp[i] == 13 && resp[i+1] == 10 {
			break
		}
		index++
	}
	return resp[1:index], index + 2
}
func checkError(resp []byte) (err error) {
	if len(resp) < 4 {
		err = errors.New("invalid resp length that shoud be >4")
	} else if CheckType(resp[:1]) == -1 {
		err = errors.New("invalid resp type")
	}
	return
}

func parseBuffer(resp []byte) (foward int, result interface{}, err error) {
	var line []byte
	switch resp[0] {
	case SimpleStrings: // +  Simple string Prefix
		line, foward = readLine(resp)
		result = string(line)
	case Errors: // -  error Prefix
		line, foward = readLine(resp)
		result = string(line)
		err = errors.New(string(line))
	case Integers: // : integer Prefix
		line, foward = readLine(resp)
		result, err = strconv.Atoi(string(line))
	case BulkStrings: //$ bulk String Prefix
		line, foward = readLine(resp)
		var length int
		length, err = strconv.Atoi(string(line))
		if length == -1 {
			result = ""
		} else {
			result = string(resp[foward : foward+length])
			foward = foward + length + 2
		}
	case Arrays: // * arrayPrefix
		line, foward = readLine(resp)
		i, err := strconv.Atoi(string(line))
		if i == -1 {
			err = errors.New("invalid array type due to times out")
		} else if i == 0 {
			result = make([]interface{}, 0)
		} else {
			array := make([]interface{}, i)
			y := 0
			for x := 0; x < i; x++ {
				y, array[x], err = parseBuffer(resp[foward:])
				foward += y
				if err != nil {
					break
				}
			}
			result = array
		}
	default:
		err = errors.New("invalid resp type")
	}

	return
}
