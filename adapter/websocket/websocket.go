package websocket

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

type OpCode byte

const (
	OpContinuation OpCode = 0x0
	OpText         OpCode = 0x1
	OpBinary       OpCode = 0x2
	OpClose        OpCode = 0x8
	OpPing         OpCode = 0x9
	OpPong         OpCode = 0xa
)

type Handshake struct {
	Protocol string
}

type WebSocket interface {
	Upgrade(r *http.Request, w http.ResponseWriter) (*net.Conn, *bufio.ReadWriter, Handshake, error)
	ReadClientData(rw io.ReadWriter) (string, OpCode, error)
	WriteServerData(rw io.Writer, code OpCode, message string) error
}
