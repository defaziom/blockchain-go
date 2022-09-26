package tcp

import (
	"log"
	"net"
)

func GetPeer() Peer {
	conn, err := net.Dial("tcp", ":9999")
	if err != nil {
		log.Println("Could not connect to peer", err.Error())
	}
	peerConn := &PeerConn{
		Conn: conn,
	}
	return peerConn
}
