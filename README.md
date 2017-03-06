# respgo


[![Build Status](https://travis-ci.org/teambition/respgo.svg?branch=master)](https://travis-ci.org/teambition/respgo)
[![Coverage Status](http://img.shields.io/coveralls/teambition/respgo.svg?style=flat-square)](https://coveralls.io/r/teambition/respgo)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/teambition/respgo/master/LICENSE)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/teambition/respgo)


## Installation
```go
go get github.com/teambition/respgo
```
## Examples
```go
go run examples/main.go
```
## API
### respgo.Decode
Parse any resp bytes
```go
respError     = "Error message"
respErrorText = "-Error message\r\n"
result,_ := respgo.Decode([]byte(respErrorText))

result == respError
```
### respgo.DecodeToString
Parse Simple String or Bulk String text
```go
respSimpleStringText = "+OK\r\n"
respSimpleString     = "OK"
result,_ := respgo.DecodeToString([]byte(respSimpleStringText))

result == respSimpleString

respBulkString     = "foobar"
respBulkStringText = "$6\r\nfoobar\r\n"
result,_ = respgo.DecodeToString([]byte(respBulkStringText))

result == respBulkString
```