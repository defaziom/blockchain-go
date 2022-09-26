package http

import (
	"encoding/json"
	"github.com/defaziom/blockchain-go/block"
	"github.com/defaziom/blockchain-go/blockchain"
	"github.com/defaziom/blockchain-go/task"
	"github.com/defaziom/blockchain-go/tcp"
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

func MineBlockHandler(pc chan tcp.Peer) http.Handler {
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
		err = blockchain.TheBlockChain.AddBlock(newBlock)
		if err != nil {
			http.Error(w, "Error", http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusCreated)

		peer := tcp.GetPeer()
		t := task.SendNewBlock{
			Msg: &tcp.PeerMsg{
				Data: []*block.Block{newBlock},
			},
			Peer:    &peer,
			Channel: pc,
		}
		go t.Execute()
	})
}

func PeersHandler(w http.ResponseWriter, req *http.Request) {

}

func AddPeerHandler(w http.ResponseWriter, req *http.Request) {

}
