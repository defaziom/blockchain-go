package http

import (
	"fmt"
	"github.com/defaziom/blockchain-go/tcp"
	"log"
	"net/http"
)

func StartServer(port int, pc chan tcp.Peer) {
	http.Handle("/blocks", JsonResponse(BlocksHandler()))
	http.Handle("/blocks/mine", JsonResponse(MineBlockHandler(pc)))
	http.Handle("/peers", JsonResponse(PeersHandler()))
	log.Println(fmt.Sprintf("Starting HTTP server on %d", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
