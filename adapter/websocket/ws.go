package websocket

import (
	"bufio"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"io"
	"net"
	"net/http"
)

type Config struct{}

type WS struct {
	config Config
}

func NewWS(config Config) WebSocket {
	return WS{
		config: config,
	}
}

func (websocket WS) Upgrade(r *http.Request, rw http.ResponseWriter) (*net.Conn, *bufio.ReadWriter, Handshake, error) {
	conn, readWriter, handShake, err := ws.UpgradeHTTP(r, rw)
	return &conn, readWriter, Handshake{
		Protocol: handShake.Protocol,
	}, err
}

func (websocket WS) ReadClientData(rw io.ReadWriter) (string, OpCode, error) {
	msg, opCode, err := wsutil.ReadClientData(rw)
	code := OpCode(opCode)
	return string(msg), code, err
}

func (websocket WS) WriteServerData(rw io.Writer, code OpCode, message string) error {
	return wsutil.WriteServerMessage(rw, ws.OpCode(code), []byte(message))
}
