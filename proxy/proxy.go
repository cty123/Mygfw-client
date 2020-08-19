package proxy

import (
	"crypto/tls"
	"log"
	"net/http"

	"mygfw.com/client/socket"
)

type HTTPProxy struct {
	port   string
	socket socket.SocketForwarder
}

func New(localPort string, socket socket.SocketForwarder) HTTPProxy {
	proxy := HTTPProxy{port: localPort, socket: socket}
	return proxy
}

func (proxy HTTPProxy) handler(w http.ResponseWriter, r *http.Request) {
	// if r.Method == http.MethodConnect {
	// 	// proxy.handleTunneling(w, r)
	// } else {
	// 	// proxy.handleHTTP(w, r)
	// }
	log.Println("Forwarding the http request")
	proxy.socket.ForwardRequest("test")
}

// func (proxy HTTPProxy) transfer(destination io.WriteCloser, source io.ReadCloser) {
// 	defer destination.Close()
// 	defer source.Close()
// 	io.Copy(destination, source)
// }

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

// func (proxy HTTPProxy) handleTunneling(w http.ResponseWriter, r *http.Request) {
// 	log.Println(r.URL)
// 	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusServiceUnavailable)
// 		return
// 	}
// 	w.WriteHeader(http.StatusOK)
// 	hijacker, ok := w.(http.Hijacker)
// 	if !ok {
// 		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
// 		return
// 	}
// 	clientConn, _, err := hijacker.Hijack()
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusServiceUnavailable)
// 	}
// 	go proxy.transfer(destConn, clientConn)
// 	go proxy.transfer(clientConn, destConn)
// }

func (proxy HTTPProxy) copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (proxy HTTPProxy) StartProxy() {
	server := &http.Server{
		Addr:         proxy.port,
		Handler:      http.HandlerFunc(proxy.handler),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	server.ListenAndServe()
}
