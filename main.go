package main

import (
	"github.com/defaziom/blockchain-go/blockchain"
	"github.com/defaziom/blockchain-go/database"
	"github.com/defaziom/blockchain-go/http"
	"github.com/defaziom/blockchain-go/task"
	"github.com/defaziom/blockchain-go/tcp"
	"github.com/defaziom/blockchain-go/transaction"
	"github.com/defaziom/blockchain-go/wallet"
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
	theBlockChain := blockchain.CreateBlockChain()

	// Init transaction service
	txOuts := transaction.UnspentTxOutSlice(make([]*transaction.UnspentTxOut, 0))
	pool := transaction.PoolSlice(make([]*transaction.TransactionIml, 0))
	ts := &transaction.ServiceIml{
		UnspentTxOuts: &txOuts,
		PoolSlice:     &pool,
		Validator:     &transaction.TxValidator{},
	}

	// Create wallet
	key, err := transaction.GeneratePrivateKey()
	if err != nil {
		log.Fatalln("Failed to create private key: ", err.Error())
	}
	ws, err := wallet.GetWalletFromPrivateKey(key, ts)
	if err != nil {
		log.Fatalln("Failed to create wallet: ", err.Error())
	}

	pc := make(chan tcp.Peer)

	// Init database
	_ = database.GetDatabase()

	// Startup
	go tcp.StartServer(tcpPort, pc)
	go task.StartTasks(pc, theBlockChain, ts)
	http.StartServer(httpPort, pc, theBlockChain, ts, ws)
}
