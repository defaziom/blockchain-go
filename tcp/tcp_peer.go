package tcp

import (
	"fmt"
	"github.com/defaziom/blockchain-go/block"
	"github.com/defaziom/blockchain-go/database"
	"log"
	"net"
)

// GetPeers takes a list of database.PeerConnInfo and establishes a connection with the peer,
// returning a list of Peer to interact with
func GetPeers(peerConnInfoList []*database.PeerConnInfo) ([]Peer, error) {
	var peers []Peer
	for _, info := range peerConnInfoList {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", info.Ip, info.Port))
		if err != nil {
			log.Printf("Could not connect to peer: %s\n", err)
			continue
		}
		peerConn := &PeerConn{
			Conn: conn,
		}
		peer := Peer(peerConn)
		peers = append(peers, peer)
	}
	return peers, nil
}

// BroadCastBlockToPeers sends a block.Block all peers in the list of Peer. After sending the block,
// the Peer is placed in a Peer channel to continue the interaction.
func BroadCastBlockToPeers(b *block.Block, peers []Peer, pc chan Peer) {
	log.Println("Sending block to peers")
	for _, peer := range peers {
		err := peer.SendResponseBlockChainMsg([]*block.Block{b})
		if err != nil {
			log.Printf("Failed to send block to peer: %s\n", err)
		} else {
			// Place the peer in the channel to continue processing
			pc <- peer
		}
	}
}
