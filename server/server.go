package server

import (
	"log"
	"net/http"
)

func Start() {
	http.Handle("/blocks", JsonResponse(BlocksHandler()))
	http.Handle("/blocks/mine", JsonResponse(MineBlockHandler()))
	log.Println("Starting HTTP server on 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
