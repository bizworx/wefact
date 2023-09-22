package main

import (
	"net"

	"github.com/armon/go-socks5"
)

func main() {
	socksConf := &socks5.Config{}
	socksServer, err := socks5.New(socksConf)
	if err != nil {
		panic(err)
	}

	l, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go socksServer.ServeConn(conn)
	}
}
