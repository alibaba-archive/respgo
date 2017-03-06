package main

import "github.com/teambition/respgo"

var (
	respSimpleString     = "OK"
	respSimpleStringText = "+OK\r\n"

	respError     = "Error message"
	respErrorText = "-Error message\r\n"

	respInteger     = "1000"
	respIntegerText = ":1000\r\n"

	respArrayText = "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"

	respArrayComplexText = "*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Foo\r\n-Bar\r\n"

	respArrayNullElementsText = "*3\r\n$3\r\nfoo\r\n$-1\r\n$3\r\nbar\r\n"
)

func main() {
	respgo.DecodeToArray([]byte(respArrayNullElementsText))
}
