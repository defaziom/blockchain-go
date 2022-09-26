package main

import (
	"github.com/defaziom/blockchain-go/http"
	"github.com/defaziom/blockchain-go/tcp"
)

func main() {
	tc := make(chan tcp.Peer)
	http.Start(tc)
	tcp.Start(tc)
}
