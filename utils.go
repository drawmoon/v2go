package main

import (
	"net"
	"time"
)

func pickFreeTcpPort() uint16 {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return pickFreeTcpPort()
	}
	defer listener.Close()

	localAddr := listener.Addr().(*net.TCPAddr)
	return uint16(localAddr.Port)
}
