package http

import (
	"encoding/json"
	"github.com/defaziom/blockchain-go/block"
	"github.com/defaziom/blockchain-go/blockchain"
	"github.com/defaziom/blockchain-go/database"
	"github.com/defaziom/blockchain-go/tcp"
	"io"
	"log"
	"net/http"
)

type MineBlockRequest struct {
	Data string
}

// BlocksHandler GET /blocks
func BlocksHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}
		// Return list of all blocks stored on the chain
		resp, err := json.Marshal(blockchain.GetBlockChain().Blocks.ToSlice())
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
		}
		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
		}
	})
}

func MineBlockHandler(pc chan tcp.Peer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}

		mineBlockRequest := &MineBlockRequest{}
		err = json.Unmarshal(body, mineBlockRequest)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Error", http.StatusBadRequest)
			return
		}

		newBlock := blockchain.GetBlockChain().MineBlock(mineBlockRequest.Data)
		err = blockchain.GetBlockChain().AddBlock(newBlock)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)

		// Broadcast the newly mined block to all peers
		peers, err := tcp.GetPeers()
		if err != nil {
			log.Println("Failed to get peers: " + err.Error())
			return
		}
		log.Println("Sending block to peers")
		for _, peer := range peers {
			err = peer.SendResponseBlockChainMsg([]*block.Block{newBlock})
			if err != nil {
				log.Println("Failed to send block to peer: " + err.Error())
			} else {
				// Place the peer in the channel to continue processing
				pc <- peer
			}
		}
	})
}

func PeersHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			// Get list of all peer connection info
			peers, err := database.GetAllPeerConnInfo()
			if err != nil {
				log.Println(err.Error())
				http.Error(w, "Error", http.StatusInternalServerError)
				return
			}

			// Write the data into the response
			resp, err := json.Marshal(peers)
			if err != nil {
				log.Println(err.Error())
				http.Error(w, "Error", http.StatusInternalServerError)
				return
			}
			_, err = w.Write(resp)
			if err != nil {
				log.Println(err.Error())
				http.Error(w, "Error", http.StatusInternalServerError)
				return
			}
		case http.MethodPost:
			// Read the request body
			body, err := io.ReadAll(req.Body)
			if err != nil {
				log.Println(err.Error())
				http.Error(w, "Error", http.StatusInternalServerError)
			}

			// Convert the request body to a PeerConnInfo
			peerConnInfo := &database.PeerConnInfo{}
			err = json.Unmarshal(body, peerConnInfo)
			if err != nil {
				log.Println(err.Error())
				http.Error(w, "Error", http.StatusBadRequest)
				return
			}

			// Save the info in the db
			err = database.InsertPeerConnInfo(peerConnInfo)
			if err != nil {
				log.Println(err.Error())
				http.Error(w, "Error", http.StatusInternalServerError)
			}

			w.WriteHeader(http.StatusCreated)
		default:
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}
	})
}
