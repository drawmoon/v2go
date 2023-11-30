package proxyctl

import (
	"net"
	"time"
)

func pickFreeTcpPort() uint32 {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		return pickFreeTcpPort()
	}
	defer listener.Close()

	localAddr := listener.Addr().(*net.TCPAddr)
	return uint32(localAddr.Port)
}
