package main

// A pattern in luci `gen.go` files is to use `cproto` instead.
// I.e. our equivalent would be: `cproto -disable-grpc`
// However, given we do not need `grpc` we use protoc instead.

//go:generate protoc --go_out=../../.. inputprops.proto
