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
	TypeSimpleStrings = "+"
	// TypeErrors For Errors the first byte of the reply is "-"
	TypeErrors = "-"
	//TypeIntegers For Integers the first byte of the reply is ":"
	TypeIntegers = ":"
	// TypeBulkStrings For Bulk Strings the first byte of the reply is "$"
	TypeBulkStrings = "$"
	// TypeArrays For Arrays the first byte of the reply is "*"
	TypeArrays = "*"
)

var (
	//OK Common responses
	OK = EncodeString("OK")
	//PONG ...
	PONG = EncodeString("PONG")
	// CRLF is "\r\n"
	CRLF = "\r\n"
)

func checkType(resp byte) (result string) {
	result = string(resp)
	if result == TypeSimpleStrings || result == TypeErrors || result == TypeIntegers || result == TypeBulkStrings || result == TypeArrays {
		return
	}
	return ""
}

// Decode ...
func Decode(resp []byte) (msgType string, result interface{}, err error) {
	err = checkError(resp)
	if err != nil {
		return
	}
	_, result, err = parseBuffer(resp)
	msgType = string(resp[0])
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

func newBuf(data string) []byte {
	var buf bytes.Buffer
	buf.WriteString(data)
	return buf.Bytes()
}

//EncodeNull return RESP's Null value to RESP buffer.
// $-1\r\n
func EncodeNull() []byte {
	return newBuf(TypeBulkStrings + "-1" + CRLF)
}

// EncodeNullArray ...
func EncodeNullArray() []byte {
	return newBuf(TypeArrays + "-1" + CRLF)
}

// EncodeString returns a simple string with the given contents.
func EncodeString(s string) []byte {
	return newBuf(TypeSimpleStrings + s + CRLF)
}

// EncodeError ...
func EncodeError(s string) []byte {
	return newBuf(TypeErrors + s + CRLF)
}

// EncodeBulkString a bulk string with the given contents.
func EncodeBulkString(s string) []byte {
	return newBuf(TypeBulkStrings + strconv.Itoa(len(s)) + CRLF + s + CRLF)
}

// EncodeInt ....
func EncodeInt(s int) []byte {
	return newBuf(TypeIntegers + strconv.Itoa(s) + CRLF)
}

// EncodeBulkBuffer ...
func EncodeBulkBuffer(s []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString(TypeBulkStrings)
	buf.WriteString(strconv.Itoa(len(s)))
	buf.WriteString(CRLF)
	buf.Write(s)
	return buf.Bytes()
}

// EncodeArray ...
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
	} else if checkType(resp[0]) == "" {
		err = errors.New("invalid resp type")
	}
	return
}

func parseBuffer(resp []byte) (foward int, result interface{}, err error) {
	var line []byte
	switch string(resp[0]) {
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
func Parse(conn net.Conn, timeouts ...time.Duration) (msgType string, result interface{}, err error) {
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
	msgType = string(prefix[0])
	return
}

func parseReader(prefix []byte, reader *bufio.Reader) (result interface{}, err error) {
	if prefix == nil {
		prefix = make([]byte, 1)
		reader.Read(prefix)
	}
	var str string
	switch string(prefix[0]) {
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
			if length != -1 && err == nil {
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
				}
			}
		}
		if result == nil {
			result = ""
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
