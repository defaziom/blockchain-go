package server

import (
	"encoding/json"
	"github.com/defaziom/blockchain-go/blockchain"
	"io"
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
		jsonBlocks, err := json.Marshal(blockchain.TheBlockChain.Blocks.ToSlice())
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
		}
		_, err = w.Write(jsonBlocks)
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
		}
	})
}

func MineBlockHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
		}

		mineBlockRequest := &MineBlockRequest{}
		err = json.Unmarshal(body, mineBlockRequest)
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
		}

		newBlock := blockchain.TheBlockChain.MineBlock(mineBlockRequest.Data)
		_, err = blockchain.TheBlockChain.AddBlock(newBlock)
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusCreated)
	})
}

func PeersHandler(w http.ResponseWriter, req *http.Request) {

}

func AddPeerHandler(w http.ResponseWriter, req *http.Request) {

}
