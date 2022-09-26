package tcp

import (
	"fmt"
	"github.com/defaziom/blockchain-go/database"
	"log"
	"net"
)

func GetPeers() ([]Peer, error) {
	infoList, err := database.GetAllPeerConnInfo()
	if err != nil {
		return nil, err
	}

	var peers []Peer
	for _, info := range infoList {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", info.Ip, info.Port))
		if err == nil {
			log.Println("Could not connect to peer", err.Error())
		}
		peerConn := &PeerConn{
			Conn: conn,
		}
		peer := Peer(peerConn)
		peers = append(peers, peer)
	}
	return peers, nil
}
