package respgo_test

import (
	"testing"

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
)

func TestRespGo(t *testing.T) {
	// Simple "OK" string
	assert := assert.New(t)
	t.Run("respgo with Decode func that should be", func(t *testing.T) {

		result, err := respgo.Decode([]byte(respErrorText))
		assert.NotNil(err)
		assert.Equal(respError, result.(string))
		assert.Equal("Error message", err.Error())

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

		result, err := respgo.Decode([]byte(""))
		assert.Nil(result)
		if assert.NotNil(err) {
			assert.Equal("invalid resp length that shoud be >4", err.Error())
		}

		result, err = respgo.Decode([]byte("^dxx\r\n"))
		assert.Nil(result)
		if assert.NotNil(err) {
			assert.Equal("invalid resp type", err.Error())
		}

		result, err = respgo.DecodeToString([]byte(respIntegerText))
		assert.Empty(result)
		if assert.NotNil(err) {
			assert.Equal("invalid string or bulkstring type", err.Error())
		}

		num, err := respgo.DecodeToInt([]byte(respArrayText))
		assert.Equal(0, num)
		if assert.NotNil(err) {
			assert.Equal("invalid int type", err.Error())
		}

	})
}
