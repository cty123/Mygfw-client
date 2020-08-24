package proxy

import (
	"github.com/gorilla/websocket"
)

type MyWritter struct {
	Conn *websocket.Conn
}

func (w *MyWritter) Write(p []byte) (n int, err error) {
	if err := w.Conn.WriteMessage(websocket.BinaryMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}
