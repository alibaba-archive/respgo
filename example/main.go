package main

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/teambition/respgo"
)

func main() {
	str := string(respgo.EncodeString("OK"))
	fmt.Println(str)
	// +OK\r\n
	reader := bufio.NewReader(strings.NewReader("+OK\r\n"))
	result, _ := respgo.Decode(reader)
	fmt.Println(result)
	// OK

	str = string(respgo.EncodeError("Error message"))
	fmt.Println(str)
	// -Error message\r\n
	reader = bufio.NewReader(strings.NewReader("-Error message\r\n"))
	result, _ = respgo.Decode(reader)
	fmt.Println(result)
	// Error message

	str = string(respgo.EncodeInt(1000))
	fmt.Println(str)
	// :1000\r\n
	reader = bufio.NewReader(strings.NewReader(":1000\r\n"))
	result, _ = respgo.Decode(reader)
	fmt.Println(result)
	// 1000

	str = string(respgo.EncodeBulkString("foobar"))
	fmt.Println(str)
	// $6\r\nfoobar\r\n
	reader = bufio.NewReader(strings.NewReader("$6\r\nfoobar\r\n"))
	result, _ = respgo.Decode(reader)
	fmt.Println(result)
	// foobar

	str = string(respgo.EncodeNull())
	fmt.Println(str)
	// $-1\r\n
	reader = bufio.NewReader(strings.NewReader("$-1\r\n"))
	result, _ = respgo.Decode(reader)
	fmt.Println(result)
	// <nil>

	str = string(respgo.EncodeArray([][]byte{
		respgo.EncodeBulkString("foo"),
		respgo.EncodeBulkString("bar")}))
	fmt.Println(str)
	// *2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n
	reader = bufio.NewReader(strings.NewReader("*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"))
	result, _ = respgo.Decode(reader)
	fmt.Println(result)
	// [foo bar]

	str = string(respgo.EncodeNullArray())
	fmt.Println(str)
	// *-1\r\n
	reader = bufio.NewReader(strings.NewReader("*-1\r\n"))
	result, _ = respgo.Decode(reader)
	fmt.Println(result)
	// <nil>
}
