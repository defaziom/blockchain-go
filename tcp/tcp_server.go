package tcp

import (
	"fmt"
	"log"
	"net"
)

func Start(pc chan Peer) {
	ln, err := net.Listen("tcp", ":4343")
	if err != nil {
		log.Println("Error listening:", err.Error())
	}
	defer ln.Close()
	log.Println("TCP http Listening on :4343")
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
