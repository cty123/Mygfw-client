package main

import (
	"mygfw.com/client/proxy"

	"mygfw.com/client/socket"
)

var (
	proxyPort = ":8090"
	scheme    = "ws"
	host      = "localhost:9000"
	path      = "/test"
)

func main() {
	sf := socket.New(scheme, host, path)
	proxy := proxy.New(proxyPort, sf)
	proxy.StartProxy()
}
