package main

import (
	"github.com/defaziom/blockchain-go/database"
	"github.com/defaziom/blockchain-go/http"
	"github.com/defaziom/blockchain-go/tcp"
	"log"
	"os"
	"strconv"
)

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		log.Fatalln("Usage: blockchain-go http_port tcp_port")
	}
	httpPort, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalln("HTTP port must be int")
	}
	tcpPort, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatalln("TCP port must be int")
	}
	pc := make(chan tcp.Peer)
	_ = database.GetDatabase()
	go tcp.StartServer(tcpPort, pc)
	http.StartServer(httpPort, pc)
}
