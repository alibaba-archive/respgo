package respgo_test

import (
	"net"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teambition/respgo"
)

var (
	respSimpleString     = "OK"
	respSimpleStringText = "+OK\r\n"

	respError     = "Error message"
	respErrorText = "-Error message\r\n"

	respInteger     = 1000
	respIntegerText = ":1000\r\n"

	respBulkString     = "foobar"
	respBulkStringText = "$6\r\nfoobar\r\n"

	respArrayText = "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"

	respArrayComplexText = "*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Foo\r\n-Bar\r\n"

	respArrayNullElementsText = "*3\r\n$3\r\nfoo\r\n$-1\r\n$3\r\nbar\r\n"
	respArrayEmptyArray       = "*0\r\n"
	respArrayNullArray        = "*-1\r\n"

	respErrorArrayText = "*2\r\n$3\r\nfoo\r\n:xxx\r\n"
)

func TestRespGo(t *testing.T) {
	// Simple "OK" string
	assert := assert.New(t)
	t.Run("respgo with Decode func that should be", func(t *testing.T) {

		msgtype, result, err := respgo.Decode([]byte(respErrorText))
		assert.Equal(msgtype, respgo.TypeErrors)
		assert.Nil(err)
		assert.Equal(respError, result.(string))

		str, err := respgo.DecodeToString([]byte(respSimpleStringText))
		assert.Nil(err)
		assert.Equal(respSimpleString, str)

		num, err := respgo.DecodeToInt([]byte(respIntegerText))
		assert.Nil(err)
		assert.Equal(respInteger, num)

		result, err = respgo.DecodeToString([]byte(respBulkStringText))
		assert.Nil(err)
		assert.Equal(respBulkString, result)

		arr, err := respgo.DecodeToArray([]byte(respArrayText))
		assert.Nil(err)
		assert.Equal(2, len(arr))

		arr, err = respgo.DecodeToArray([]byte(respArrayComplexText))
		assert.Nil(err)
		assert.Equal(2, len(arr))
		if assert.NotNil(arr[0]) {
			arr1 := arr[0].([]interface{})
			assert.Equal(3, len(arr1))
			assert.Equal(1, arr1[0].(int))
			assert.Equal(2, arr1[1].(int))
			assert.Equal(3, arr1[2].(int))
		}
		if assert.NotNil(arr[1]) {
			arr2 := arr[1].([]interface{})
			assert.Equal(2, len(arr2))
			assert.Equal("Foo", arr2[0].(string))
			assert.Equal("Bar", arr2[1].(string))
		}
	})
	t.Run("respgo with Null elements in Arrays or empty Array func that should be", func(t *testing.T) {
		arr, err := respgo.DecodeToArray([]byte(respArrayNullElementsText))
		assert.Nil(err)
		if assert.NotNil(arr) {
			assert.Equal(3, len(arr))
			assert.Equal("foo", arr[0].(string))
			assert.Equal("", arr[1].(string))
			assert.Equal("bar", arr[2].(string))
		}

		arr, err = respgo.DecodeToArray([]byte(respArrayEmptyArray))
		assert.Nil(err)
		if assert.NotNil(arr) {
			assert.Equal(0, len(arr))
		}

		arr, err = respgo.DecodeToArray([]byte(respArrayNullArray))
		assert.NotNil(err)
		assert.Nil(arr)
	})

	t.Run("respgo with Encode func that should be", func(t *testing.T) {

		str := respgo.EncodeString(respSimpleString)
		assert.Equal(respSimpleStringText, string(str))

		str = respgo.EncodeError(respError)
		assert.Equal(respErrorText, string(str))

		str = respgo.EncodeInt(respInteger)
		assert.Equal(respIntegerText, string(str))

		str = respgo.EncodeBulkString(respBulkString)
		assert.Equal(respBulkStringText, string(str))

		foo := respgo.EncodeBulkString("foo")
		bar := respgo.EncodeBulkString("bar")

		str = respgo.EncodeArray([][]byte{foo, bar})
		assert.Equal(respArrayText, string(str))
	})
	t.Run("respgo with error message that should be", func(t *testing.T) {

		_, result, err := respgo.Decode([]byte(""))
		assert.Nil(result)
		if assert.NotNil(err) {
			assert.Equal("invalid resp length that shoud be >4", err.Error())
		}

		_, result, err = respgo.Decode([]byte("^dxx\r\n"))
		assert.Nil(result)
		if assert.NotNil(err) {
			assert.Equal("invalid resp type", err.Error())
		}

		result, err = respgo.DecodeToString([]byte(respIntegerText))
		assert.Empty(result)
		if assert.NotNil(err) {
			assert.Equal("invalid string or bulkstring type", err.Error())
		}

		result, err = respgo.DecodeToString([]byte(""))
		assert.Empty(result)
		if assert.NotNil(err) {
			assert.Equal("invalid resp length that shoud be >4", err.Error())
		}

		num, err := respgo.DecodeToInt([]byte(respArrayText))
		assert.Equal(0, num)
		if assert.NotNil(err) {
			assert.Equal("invalid int type", err.Error())
		}
		num, err = respgo.DecodeToInt([]byte(""))
		assert.Equal(0, num)
		if assert.NotNil(err) {
			assert.Equal("invalid resp length that shoud be >4", err.Error())
		}

		array, err := respgo.DecodeToArray([]byte(""))
		assert.Equal(0, len(array))
		if assert.NotNil(err) {
			assert.Equal("invalid resp length that shoud be >4", err.Error())
		}

		array, err = respgo.DecodeToArray([]byte("xxxxxxxxxxxxxxxxx"))
		assert.Equal(0, len(array))
		if assert.NotNil(err) {
			assert.Equal("invalid resp type", err.Error())
		}
		array, err = respgo.DecodeToArray([]byte(respIntegerText))
		assert.Equal(0, len(array))
		if assert.NotNil(err) {
			assert.Equal("invalid array type", err.Error())
		}

		array, err = respgo.DecodeToArray([]byte(respErrorArrayText))
		assert.Equal(0, len(array))
		if assert.NotNil(err) {
			assert.Contains(err.Error(), "invalid syntax")
		}
	})
}
func TestParse(t *testing.T) {
	assert := assert.New(t)
	go func() {
		conn, err := net.Dial("tcp", ":3000")
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		conn.Write([]byte(respErrorText))
		time.Sleep(10 * time.Millisecond)
		conn.Write([]byte(respSimpleStringText))
		time.Sleep(10 * time.Millisecond)
		conn.Write([]byte(respIntegerText))
		time.Sleep(10 * time.Millisecond)
		conn.Write([]byte(respBulkStringText))
		time.Sleep(10 * time.Millisecond)
		conn.Write([]byte(respArrayText))
		time.Sleep(10 * time.Millisecond)
		conn.Write([]byte(respArrayComplexText))
		time.Sleep(10 * time.Millisecond)
		conn.Write([]byte(respArrayEmptyArray))
		time.Sleep(10 * time.Millisecond)
		conn.Write([]byte(respArrayNullArray))
		time.Sleep(10 * time.Millisecond)
		conn.Write([]byte(respArrayNullElementsText))
		time.Sleep(10 * time.Millisecond)
		conn.Write([]byte("xxxxxxxxxxx"))
		time.Sleep(10 * time.Millisecond)
		conn.Write([]byte(respErrorArrayText))
		time.Sleep(10 * time.Millisecond)
		conn.Write([]byte(respErrorArrayText))

	}()

	l, err := net.Listen("tcp", ":3000")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	_, result, err := respgo.Parse(conn, time.Minute)
	assert.Nil(err)
	assert.Equal(respError, result.(string))

	_, result, err = respgo.Parse(conn)
	assert.Nil(err)
	assert.Equal(respSimpleString, result)

	_, result, err = respgo.Parse(conn, time.Minute)
	assert.Nil(err)
	assert.Equal(respInteger, result)

	_, result, err = respgo.Parse(conn, time.Minute)
	assert.Nil(err)
	assert.Equal(respBulkString, result)

	_, result, err = respgo.Parse(conn)
	assert.Nil(err)
	assert.Equal(2, len(result.([]interface{})))

	_, result, err = respgo.Parse(conn, time.Minute)
	arr := result.([]interface{})
	assert.Nil(err)
	assert.Equal(2, len(arr))
	if assert.NotNil(arr[0]) {
		arr1 := arr[0].([]interface{})
		assert.Equal(3, len(arr1))
		assert.Equal(1, arr1[0].(int))
		assert.Equal(2, arr1[1].(int))
		assert.Equal(3, arr1[2].(int))
	}
	if assert.NotNil(arr[1]) {
		arr2 := arr[1].([]interface{})
		assert.Equal(2, len(arr2))
		assert.Equal("Foo", arr2[0].(string))
		assert.Equal("Bar", arr2[1].(string))
	}
	_, result, err = respgo.Parse(conn, time.Minute)
	assert.Nil(err)
	if assert.NotNil(arr) {
		assert.Equal(0, len(result.([]interface{})))
	}
	_, result, err = respgo.Parse(conn, time.Minute)
	assert.NotNil(err)
	assert.Nil(result)
	_, result, err = respgo.Parse(conn, time.Minute)
	arr = result.([]interface{})
	assert.Nil(err)
	if assert.NotNil(arr) {
		assert.Equal(3, len(arr))
		assert.Equal("foo", arr[0].(string))
		assert.Equal("", arr[1].(string))
		assert.Equal("bar", arr[2].(string))
	}

	_, result, err = respgo.Parse(conn, time.Minute)
	if assert.NotNil(err) {
		assert.Equal("invalid resp type", err.Error())
	}
	_, result, err = respgo.Parse(conn, time.Minute)
	arr = result.([]interface{})
	assert.Equal(2, len(arr))
	if assert.NotNil(err) {
		assert.Contains(err.Error(), "invalid syntax")
	}

	_, result, err = respgo.Parse(conn, time.Duration(0))
	arr, _ = result.([]interface{})
	assert.Equal(0, len(arr))
	if assert.NotNil(err) {
		assert.Contains(err.Error(), "i/o timeout")
	}
}
func TestParseConcurrent(t *testing.T) {
	assert := assert.New(t)
	go func() {
		conn, err := net.Dial("tcp", ":3000")
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		conn.Write([]byte(respErrorText))
		conn.Write([]byte(respSimpleStringText))
		conn.Write([]byte(respIntegerText))
		conn.Write([]byte(respBulkStringText))
		conn.Write([]byte(respArrayComplexText))

		conn.Write([]byte("$6\r\nfo"))
		conn.Write([]byte("obar\r\n"))
	}()

	l, err := net.Listen("tcp", ":3000")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	for index := 0; index < 5; index++ {

		msgtype, result, err := respgo.Parse(conn, time.Minute)
		switch msgtype {
		case respgo.TypeErrors:
			assert.Nil(err)
			assert.Equal(respError, result.(string))
		case respgo.TypeSimpleStrings:
			assert.Nil(err)
			assert.Equal(respSimpleString, result.(string))
		case respgo.TypeIntegers:
			assert.Nil(err)
			assert.Equal(respInteger, result.(int))
		case respgo.TypeBulkStrings:
			assert.Nil(err)
			assert.Equal(respBulkString, result.(string))
		case respgo.TypeArrays:
			arr := result.([]interface{})
			assert.Nil(err)
			assert.Equal(2, len(arr))
			if assert.NotNil(arr[0]) {
				arr1 := arr[0].([]interface{})
				assert.Equal(3, len(arr1))
				assert.Equal(1, arr1[0].(int))
				assert.Equal(2, arr1[1].(int))
				assert.Equal(3, arr1[2].(int))
			}
			if assert.NotNil(arr[1]) {
				arr2 := arr[1].([]interface{})
				assert.Equal(2, len(arr2))
				assert.Equal("Foo", arr2[0].(string))
				assert.Equal("Bar", arr2[1].(string))
			}

		}
	}
}
func TestPartPacket(t *testing.T) {
	assert := assert.New(t)

	go func() {
		conn, err := net.Dial("tcp", ":3000")
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		conn.Write([]byte("$6\r\nfo"))
		conn.Write([]byte("obar\r\n"))
		conn.Write([]byte("$66\r\n012345678901234567890123456789012345678901234567890123456789"))
		conn.Write([]byte("obara\r\n"))

	}()

	l, err := net.Listen("tcp", ":3000")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		return
	}
	defer conn.Close()
	for index := 0; index < 2; index++ {
		msgtype, result, _ := respgo.Parse(conn, time.Minute)
		switch msgtype {
		case respgo.TypeBulkStrings:
			str := result.(string)
			assert.Nil(err)
			if len(str) > 10 {
				assert.Equal("012345678901234567890123456789012345678901234567890123456789obara", result.(string))
			} else {
				assert.Equal(respBulkString, result.(string))
			}
		}
	}
}
