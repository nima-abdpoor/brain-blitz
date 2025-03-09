package websocket

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

type OpCode byte

type Handshake struct {
	Protocol string
}

type WebSocket interface {
	Upgrade(r *http.Request, w http.ResponseWriter) (*net.Conn, *bufio.ReadWriter, Handshake, error)
	ReadClientData(rw io.ReadWriter) (string, OpCode, error)
	WriteServerData(rw io.Writer, code OpCode, message string) error
}
