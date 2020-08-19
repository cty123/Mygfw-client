package socket

import (
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

type SocketForwarder struct {
	scheme string
	host   string
	path   string
	conn   *websocket.Conn
}

func New(scheme string, host string, path string) SocketForwarder {
	socketForwarder := &SocketForwarder{
		scheme: scheme,
		host:   host,
		path:   path,
	}
	socketForwarder.initializeConnection()
	return *socketForwarder
}

func (sf *SocketForwarder) initializeConnection() {
	u := url.URL{Scheme: sf.scheme, Host: sf.host, Path: sf.path}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	sf.conn = c
}

func (sf *SocketForwarder) ForwardRequest(msg string) {
	log.Println("Forwarding the request to the server: ", msg)
	sf.conn.WriteMessage(websocket.TextMessage, []byte("Hello we have a request"))
}

func main() {
	u := url.URL{Scheme: "ws", Host: "localhost:9000", Path: "/test"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
}
