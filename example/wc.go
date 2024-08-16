package main

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"io"
	"net/http"
)

func main() {
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			fmt.Println("error in connecting...", err)
		}

		go readMessage(conn)
		//err = wsutil.WriteServerMessage(conn, opCode, msg)
		//if err != nil {
		//	fmt.Println("error in writing message...", err)
		//} ()
	}))
}

func readMessage(rw io.ReadWriter) {
	for {
		msg, opCode, err := wsutil.ReadClientData(rw)
		if err != nil {
			fmt.Println("error in reading message...", err)
		}

		fmt.Println("message:", string(msg), "opCODE:", opCode)
	}
}
