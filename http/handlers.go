package http

import (
	"encoding/json"
	"fmt"
	"github.com/defaziom/blockchain-go/blockchain"
	"github.com/defaziom/blockchain-go/database"
	"github.com/defaziom/blockchain-go/tcp"
	"github.com/defaziom/blockchain-go/transaction"
	"github.com/defaziom/blockchain-go/wallet"
	"io"
	"log"
	"net/http"
)

type MineBlockRequest struct {
	Data string
}

// BlocksHandler GET /blocks
func BlocksHandler(bc blockchain.BlockChain) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}
		// Return list of all blocks stored on the chain
		resp, err := json.Marshal(bc.GetBlocks().ToSlice())
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
		}
		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
		}
	})
}

func MineBlockHandler(bc blockchain.BlockChain, pc chan tcp.Peer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("Failed to read request body: %s\n", err)
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}

		mineBlockRequest := &MineBlockRequest{}
		err = json.Unmarshal(body, mineBlockRequest)
		if err != nil {
			log.Printf("Failed to unmarshal request: %s\n", err)
			http.Error(w, "Error", http.StatusBadRequest)
			return
		}

		newBlock := bc.MineBlock(mineBlockRequest.Data)
		err = bc.AddBlock(newBlock)
		if err != nil {
			log.Printf("Failed to add new block to blockchain: %s\n", err)
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)

		log.Println("Successfully mined a new block!")

		peerConnList, err := database.GetAllPeerConnInfo()
		if err != nil {
			log.Printf("Failed to get list of registered peers: %s\n", err)
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}
		peers, err := tcp.GetPeers(peerConnList, &tcp.TcpDialer{})

		if err != nil {
			log.Println("Failed to get peers: " + err.Error())
			return
		}
		// Broadcast the newly mined block to all peers
		tcp.BroadCastBlockToPeers(newBlock, peers, pc)
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

			log.Println(fmt.Sprintf("Registered peer with IP=%s and port=%d", peerConnInfo.Ip, peerConnInfo.Port))
		default:
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}
	})
}

func GetTxHandler(ts transaction.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		// Return transaction pool
		resp, err := json.Marshal(ts.Pool())
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
		}
		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
		}
	})
}

func MineTxHandler(bc blockchain.BlockChain, ws *wallet.Service, pc chan tcp.Peer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		newBlockIndex := bc.GetLatestBlock().Index + 1
		coinbaseTx := transaction.CreateCoinbaseTransaction(ws.Address, newBlockIndex)
		txs := append(*ws.TxService.Pool(), coinbaseTx)

		err := ws.TxService.ProcessTransactions([]*transaction.TransactionIml{coinbaseTx}, newBlockIndex)

		blockData, err := json.Marshal(txs)
		if err != nil {
			log.Println("Failed to create blockdata: ", err.Error())
			http.Error(w, "Error creating blockdata", http.StatusInternalServerError)
			return
		}

		newBlock := bc.MineBlock(string(blockData))
		err = bc.AddBlock(newBlock)
		if err != nil {
			log.Printf("Failed to add new block to blockchain: %s\n", err)
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)

		log.Println("Successfully mined a new block!")

		// Update the transaction pool
		ws.TxService.Pool().Update(ws.TxService.GetUnspentTxOutList())

		peerConnList, err := database.GetAllPeerConnInfo()
		if err != nil {
			log.Printf("Failed to get list of registered peers: %s\n", err)
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}
		peers, err := tcp.GetPeers(peerConnList, &tcp.TcpDialer{})

		if err != nil {
			log.Println("Failed to get peers: " + err.Error())
			return
		}
		// Broadcast the newly mined block to all peers
		tcp.BroadCastBlockToPeers(newBlock, peers, pc)
	})
}

func WalletHandler(ws *wallet.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		// Return wallet info
		resp, err := json.Marshal(ws.GetInfo())
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
		}
		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
		}
	})
}

func WalletSendAmountHandler(ws *wallet.Service, pc chan tcp.Peer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
		}
		sendReq := &wallet.SendRequest{}
		err = json.Unmarshal(body, sendReq)
		if err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		}

		err = ws.SendToAddress(sendReq.Amount, sendReq.Address)
		if err != nil {
			log.Println("Failed to send amount to address: ", err.Error())
			http.Error(w, "Failed to send amount to address", http.StatusInternalServerError)
		}

		peerConnList, err := database.GetAllPeerConnInfo()
		if err != nil {
			log.Printf("Failed to get list of registered peers: %s\n", err)
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}
		peers, err := tcp.GetPeers(peerConnList, &tcp.TcpDialer{})

		if err != nil {
			log.Println("Failed to get peers: " + err.Error())
			return
		}

		tcp.BroadCastTransactionPoolToPeers(ws.TxService.Pool(), peers, pc)
	})
}
