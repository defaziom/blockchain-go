package http

import (
	"fmt"
	"github.com/defaziom/blockchain-go/blockchain"
	"github.com/defaziom/blockchain-go/tcp"
	"log"
	"net/http"
)

func StartServer(port int, pc chan tcp.Peer, bc blockchain.BlockChain) {
	http.Handle("/blocks", LogMethodAndEndpoint(JsonResponse(BlocksHandler(bc))))
	http.Handle("/blocks/mine", LogMethodAndEndpoint(JsonResponse(MineBlockHandler(bc, pc))))
	http.Handle("/peers", LogMethodAndEndpoint(JsonResponse(PeersHandler())))
	log.Println(fmt.Sprintf("Starting HTTP server on %d", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
