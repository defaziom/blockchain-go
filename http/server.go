package http

import (
	"fmt"
	"github.com/defaziom/blockchain-go/blockchain"
	"github.com/defaziom/blockchain-go/tcp"
	"github.com/defaziom/blockchain-go/transaction"
	"github.com/defaziom/blockchain-go/wallet"
	"log"
	"net/http"
)

func StartServer(port int, pc chan tcp.Peer, bc blockchain.BlockChain, ts transaction.Service, ws *wallet.Service) {
	http.Handle("/blocks", LogMethodAndEndpoint(JsonResponse(BlocksHandler(bc))))
	http.Handle("/blocks/mine", LogMethodAndEndpoint(JsonResponse(MineBlockHandler(bc, pc))))
	http.Handle("/peers", LogMethodAndEndpoint(JsonResponse(PeersHandler())))
	http.Handle("/transactions", LogMethodAndEndpoint(JsonResponse(GetTxHandler(ts))))
	http.Handle("/transactions/mine", LogMethodAndEndpoint(JsonResponse(MineTxHandler(bc, ws, pc))))
	http.Handle("/wallet", LogMethodAndEndpoint(JsonResponse(WalletHandler(ws))))
	http.Handle("/wallet/sendAmount", LogMethodAndEndpoint(JsonResponse(WalletSendAmountHandler(ws, pc))))
	log.Println(fmt.Sprintf("Starting HTTP server on %d", port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
