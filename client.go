package main

import (
	"mygfw.com/client/proxy"
)

var (
	proxyPort = ":8090"
)

func main() {
	proxy := proxy.New(proxyPort)
	proxy.StartProxy()
}
