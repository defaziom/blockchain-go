package tcp

import (
	"fmt"
	"log"
	"net"
)

func StartServer(port int, pc chan Peer) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Println("Error listening:", err.Error())
	}
	defer ln.Close()
	log.Println(fmt.Sprintf("Starting TCP server on :%d", port))
	for {
		// Listen for an incoming connection.
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
		}
		peerConn := &PeerConn{
			Conn: conn,
		}
		pc <- peerConn
	}
}
