package http

import (
	"github.com/defaziom/blockchain-go/tcp"
	"log"
	"net/http"
)

func Start(pc chan tcp.Peer) {
	http.Handle("/blocks", JsonResponse(BlocksHandler()))
	http.Handle("/blocks/mine", JsonResponse(MineBlockHandler(pc)))
	log.Println("Starting HTTP http on 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
