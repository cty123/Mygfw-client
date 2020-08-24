package proxy

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

var (
	scheme = "ws"
	host   = ":9000"
	path   = "/test"
)

type HTTPProxy struct {
	port string
}

func New(localPort string) HTTPProxy {
	proxy := HTTPProxy{port: localPort}
	return proxy
}

func ForwardRequest(socket *websocket.Conn, host string) {
	log.Println("Forwarding the request to server")
	request := &Request{
		Uid:  uuid.New().String(),
		Host: host,
	}
	data, err := proto.Marshal(request)
	if err != nil {
		log.Println("Failed to serialize the http request")
	}
	socket.WriteMessage(websocket.BinaryMessage, data)
}

func GetServerResponse(socket *websocket.Conn) error {
	_, b, err := socket.ReadMessage()
	if err != nil {
		log.Println(err)
	}
	response := &Response{}
	if err := proto.Unmarshal(b, response); err != nil {
		log.Println("Failed to parse response message", err)
		return err
	}
	if !response.Status {
		return errors.New("Server failed to establish the connection")
	}
	return nil
}

func initWebsocket() *websocket.Conn {
	u := url.URL{Scheme: scheme, Host: host, Path: path}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	return conn
}

func (proxy HTTPProxy) serverHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Send the request URL to the server so that the server
	// can establish the connection
	conn := initWebsocket()

	ForwardRequest(conn, r.Host)

	// Step 2: Wait for the server to establish the connection and reply
	if err := GetServerResponse(conn); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusRequestTimeout)
		return
	}
	w.WriteHeader(http.StatusOK)

	// Step 3: Hijack the request
	log.Println("Server established the connection")
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	// Step 4: Pass the request body to the server
	pr, pw := io.Pipe()

	go proxy.transfer(pw, clientConn)

	go proxy.transportRequestBodyToRemote(conn, pr)

	go proxy.passResponseToClient(conn, clientConn)
}

func (proxy HTTPProxy) transportRequestBodyToRemote(conn *websocket.Conn, reader io.Reader) {
	defer conn.Close()
	writer := &MyWritter{Conn: conn}
	io.Copy(writer, reader)
}

func (proxy HTTPProxy) passResponseToClient(conn *websocket.Conn, clientConn net.Conn) {
	defer conn.Close()
	defer clientConn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading the messag")
			return
		}
		io.Copy(clientConn, bytes.NewReader(msg))
	}
}

func (proxy HTTPProxy) handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		proxy.serverHandler(w, r)
	} else {
		// proxy.handleHTTP(w, r)
	}
}

func (proxy HTTPProxy) transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

// func (proxy HTTPProxy) handleHTTP(w http.ResponseWriter, r *http.Request) {
// 	response, err := http.DefaultTransport.RoundTrip(r)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusServiceUnavailable)
// 		return
// 	}
// 	defer response.Body.Close()

// 	proxy.copyHeader(w.Header(), response.Header)
// 	w.WriteHeader(response.StatusCode)
// 	io.Copy(w, response.Body)
// }

func (proxy HTTPProxy) handleTunneling(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	prServer, pwServer := io.Pipe()
	prClient, pwClient := io.Pipe()
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go proxy.transfer(pwServer, clientConn)
	go proxy.transfer(destConn, prServer)
	go proxy.transfer(pwClient, destConn)
	go proxy.transfer(clientConn, prClient)
}

// func (proxy HTTPProxy) copyHeader(dst, src http.Header) {
// 	for k, vv := range src {
// 		for _, v := range vv {
// 			dst.Add(k, v)
// 		}
// 	}
// }

func (proxy HTTPProxy) StartProxy() {

	server := &http.Server{
		Addr:         proxy.port,
		Handler:      http.HandlerFunc(proxy.handler),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	server.ListenAndServe()
}
