package respgo

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"strconv"
	"time"
)

var (
	// TypeSimpleStrings For Simple Strings the first byte of the reply is "+"
	TypeSimpleStrings byte = '+'
	// TypeErrors For Errors the first byte of the reply is "-"
	TypeErrors byte = '-'
	//TypeIntegers For Integers the first byte of the reply is ":"
	TypeIntegers byte = ':'
	// TypeBulkStrings For Bulk Strings the first byte of the reply is "$"
	TypeBulkStrings byte = '$'
	// TypeArrays For Arrays the first byte of the reply is "*"
	TypeArrays byte = '*'
)

var (
	//OK Common responses
	OK = EncodeString("OK")
	//PONG ...
	PONG = EncodeString("PONG")
	// CRLF is "\r\n"
	CRLF = []byte{'\r', '\n'}
)

func checkType(resp byte) (result byte) {
	result = resp
	if result == TypeSimpleStrings || result == TypeErrors || result == TypeIntegers || result == TypeBulkStrings || result == TypeArrays {
		return
	}
	return '0'
}

// Decode ...
func Decode(resp []byte) (msgType byte, result interface{}, err error) {
	err = checkError(resp)
	if err != nil {
		return
	}
	_, result, err = parseBuffer(resp)
	msgType = resp[0]
	return
}

// DecodeToString ...
func DecodeToString(resp []byte) (result string, err error) {
	_, val, err := Decode(resp)
	if err != nil {
		return
	}
	var OK bool
	if result, OK = val.(string); !OK {
		err = errors.New("invalid string or bulkstring type")
	}
	return
}

// DecodeToInt ...
func DecodeToInt(resp []byte) (result int, err error) {
	_, val, err := Decode(resp)
	if err != nil {
		return
	}
	var OK bool
	if result, OK = val.(int); !OK {
		err = errors.New("invalid int type")
	}
	return
}

// DecodeToArray ...
func DecodeToArray(resp []byte) (result []interface{}, err error) {
	_, val, err := Decode(resp)
	if err != nil {
		return
	}
	var OK bool
	if result, OK = val.([]interface{}); !OK {
		err = errors.New("invalid array type")
	}
	return
}

// EncodeString returns a simple string with the given contents.
func EncodeString(s string) []byte {
	var buf bytes.Buffer
	buf.WriteByte(TypeSimpleStrings)
	buf.WriteString(s)
	buf.Write(CRLF)
	return buf.Bytes()
}

// EncodeError ...
func EncodeError(s string) []byte {
	var buf bytes.Buffer
	buf.WriteByte(TypeErrors)
	buf.WriteString(s)
	buf.Write(CRLF)
	return buf.Bytes()
}

// EncodeBulkString a bulk string with the given contents.
func EncodeBulkString(s string) []byte {
	var buf bytes.Buffer
	buf.WriteByte(TypeBulkStrings)
	buf.WriteString(strconv.Itoa(len(s)))
	buf.Write(CRLF)
	buf.WriteString(s)
	buf.Write(CRLF)
	return buf.Bytes()
}

// EncodeInt ....
func EncodeInt(s int) []byte {
	var buf bytes.Buffer
	buf.WriteByte(TypeIntegers)
	buf.WriteString(strconv.Itoa(s))
	buf.Write(CRLF)
	return buf.Bytes()
}

// EncodeArray ...
func EncodeArray(s [][]byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte(TypeArrays)
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
	} else if checkType(resp[0]) == '0' {
		err = errors.New("invalid resp type")
	}
	return
}

func parseBuffer(resp []byte) (foward int, result interface{}, err error) {
	var line []byte
	switch resp[0] {
	case TypeSimpleStrings: // +  Simple string Prefix
		line, foward = readLine(resp)
		result = string(line)
	case TypeErrors: // -  error Prefix
		line, foward = readLine(resp)
		result = string(line)
	case TypeIntegers: // : integer Prefix
		line, foward = readLine(resp)
		result, err = strconv.Atoi(string(line))
	case TypeBulkStrings: //$ bulk String Prefix
		line, foward = readLine(resp)
		var length int
		length, err = strconv.Atoi(string(line))
		if length == -1 {
			result = ""
		} else {
			result = string(resp[foward : foward+length])
			foward = foward + length + 2
		}
	case TypeArrays: // * arrayPrefix
		line, foward = readLine(resp)
		var length int
		length, err = strconv.Atoi(string(line))
		if length == -1 {
			err = errors.New("invalid array type due to times out")
		} else if length == 0 {
			result = make([]interface{}, 0)
		} else {
			array := make([]interface{}, length)
			y := 0
			for x := 0; x < length; x++ {
				y, array[x], err = parseBuffer(resp[foward:])
				foward += y
				if err != nil {
					break
				}
			}
			result = array
		}
	}
	return
}

// Parse ...
func Parse(conn net.Conn, timeouts ...time.Duration) (msgType byte, result interface{}, err error) {
	if len(timeouts) > 0 {
		conn.SetReadDeadline(time.Now().Add(timeouts[0]))
		defer conn.SetReadDeadline(time.Time{})
	}
	reader := bufio.NewReader(conn)
	prefix := make([]byte, 1)
	_, err = reader.Read(prefix)
	if err != nil {
		return
	}
	result, err = parseReader(prefix, reader)
	msgType = prefix[0]
	return
}

func parseReader(prefix []byte, reader *bufio.Reader) (result interface{}, err error) {
	if prefix == nil {
		prefix = make([]byte, 1)
		reader.Read(prefix)
	}
	var str string
	switch prefix[0] {
	case TypeSimpleStrings: // +  Simple string Prefix
		str, err = reader.ReadString('\n')
		if err == nil {
			result = str[:len(str)-2]
		}
	case TypeErrors: // -  error Prefix
		str, err = reader.ReadString('\n')
		if err == nil {
			result = str[:len(str)-2]
		}
	case TypeIntegers: // : integer Prefix
		str, err = reader.ReadString('\n')
		if err == nil {
			result, err = strconv.Atoi(str[:len(str)-2])
		}
	case TypeBulkStrings: //$ bulk String Prefix
		str, err = reader.ReadString('\n')
		if err == nil {
			var length int
			length, err = strconv.Atoi(str[:len(str)-2])
			if length == -1 {
				result = ""
			} else {
				var realLength = length + 2
				var bulk []byte
				if realLength > 64 {
					bulk = make([]byte, 64)
				} else {
					bulk = make([]byte, realLength)
				}
				var jsonBuf bytes.Buffer
				var n int
				for {
					var num int
					num, err = reader.Read(bulk)
					n += num
					if err != nil {
						break
					}
					jsonBuf.Write(bulk[:num])
					if n >= realLength {
						break
					}
				}
				res := jsonBuf.Bytes()
				if len(res) > 1 {
					result = string(res[:len(res)-2])
				} else {
					result = ""
				}
			}
		}
	case TypeArrays: // * arrayPrefix
		str, err = reader.ReadString('\n')
		if err == nil {
			var length int
			length, err = strconv.Atoi(str[:len(str)-2])
			if length == -1 {
				err = errors.New("invalid array type due to times out")
			} else if length == 0 {
				result = make([]interface{}, 0)
			} else {
				array := make([]interface{}, length)
				for x := 0; x < length; x++ {
					array[x], err = parseReader(nil, reader)
					if err != nil {
						break
					}
				}
				result = array
			}
		}
	default:
		err = errors.New("invalid resp type")
	}
	return
}
