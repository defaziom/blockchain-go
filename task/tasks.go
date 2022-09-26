package task

import (
	"fmt"
	"github.com/defaziom/blockchain-go/block"
	"github.com/defaziom/blockchain-go/blockchain"
	"github.com/defaziom/blockchain-go/tcp"
	"log"
)

func StartTasks(pc chan tcp.Peer) {
	for peer := range pc {
		msg, err := peer.ReceiveMsg()
		if err != nil {
			log.Println("Failed to receive msg from peer")
			continue
		}

		go ProcessMsg(msg, &peer, pc)
	}
}

func ProcessMsg(msg *tcp.PeerMsg, peer *tcp.Peer, c chan tcp.Peer) {
	var tsk Task

	switch msg.Type {
	case tcp.ACK:
		t := &Ack{
			Msg:     msg,
			Peer:    peer,
			Channel: c,
		}
		tsk = Task(t)
	case tcp.QUERY_ALL:
		t := &QueryAll{
			Msg:     msg,
			Peer:    peer,
			Channel: c,
		}
		tsk = Task(t)
	case tcp.QUERY_LATEST:
		t := &QueryLatest{
			Msg:     msg,
			Peer:    peer,
			Channel: c,
		}
		tsk = Task(t)
	case tcp.RESPONSE_BLOCKCHAIN:
		t := &ResponseBlockChain{
			Msg:     msg,
			Peer:    peer,
			Channel: c,
		}
		tsk = Task(t)
	default:
		log.Println("Received unknown msg type")
		return
	}

	tsk.Execute()
}

type Task interface {
	Execute()
	Continue()
}

type PeerMsgTask struct {
	Msg     *tcp.PeerMsg
	Peer    *tcp.Peer
	Channel chan tcp.Peer
	Task
}

type Ack PeerMsgTask

func (task *Ack) Execute() {
	log.Println("Received ACK!")
	(*task.Peer).ClosePeer()
}

type QueryLatest PeerMsgTask

func (task *QueryLatest) Execute() {
	// Send the latest block in the blockchain
	latestBlock := blockchain.GetBlockChain().GetLatestBlock()

	err := (*task.Peer).SendResponseBlockChainMsg([]*block.Block{latestBlock})
	if err != nil {
		log.Println("Failed to send response blockchain msg", err.Error())
		(*task.Peer).ClosePeer()
		return
	}

	task.Continue()
}
func (task *QueryLatest) Continue() {
	task.Channel <- *task.Peer
}

type QueryAll PeerMsgTask

func (task *QueryAll) Execute() {
	// Send the entire blockchain
	log.Println("Sending entire blockchain")
	blocks := blockchain.GetBlockChain().Blocks.ToSlice()

	err := (*task.Peer).SendResponseBlockChainMsg(blocks)
	if err != nil {
		log.Println("Failed to send response blockchain msg", err.Error())
		(*task.Peer).ClosePeer()
		return
	}

	task.Continue()
}
func (task *QueryAll) Continue() {
	task.Channel <- *task.Peer
}

type ResponseBlockChain PeerMsgTask

func (task *ResponseBlockChain) Execute() {
	receivedBlocks := task.Msg.Data
	log.Println("Got blockchain: " + fmt.Sprint(receivedBlocks))

	if len(receivedBlocks) == 0 {
		log.Println("Got zero blocks")
		return
	}

	latestBlockReceived := receivedBlocks[len(receivedBlocks)-1]
	latestBlockHeld := blockchain.GetBlockChain().GetLatestBlock()
	if latestBlockReceived.Index > latestBlockHeld.Index {
		log.Println(fmt.Sprintf("Blockchain possible behind. We got: %d Peer got: %d", latestBlockHeld.Index,
			latestBlockReceived.Index))

		if latestBlockHeld.BlockHash == latestBlockReceived.PrevBlockHash {
			// The block received is the next block in the chain
			log.Println("Adding block to the blockchain")
			err := blockchain.GetBlockChain().AddBlock(latestBlockReceived)
			if err != nil {
				log.Println("Received invalid block: " + err.Error())
			}
		} else if len(receivedBlocks) == 1 {
			// We have to query the chain from our peer
			log.Println("Querying peer for entire blockchain")
			err := (*task.Peer).SendQueryAllMsg()
			if err != nil {
				log.Println("Failed to query peer for entire chain: " + err.Error())
				(*task.Peer).ClosePeer()
				return
			}
			task.Continue()
			return
		} else {
			// Received chain is longer than our own chain
			log.Println("Replacing blockchain")
			receivedChainList := blockchain.DoublyLinkedBlockListCreateFromSlice(receivedBlocks)
			receivedBlockChain := &blockchain.BlockChain{Blocks: receivedChainList}
			blockchain.GetBlockChain().ReplaceChain(receivedBlockChain)
		}
	} else {
		log.Println("Received chain is not longer than our own chain. Do nothing.")
	}

	// Send ACK message to notify the peer we are finished
	err := (*task.Peer).SendAckMsg()
	if err != nil {
		log.Println("Failed to send ack msg", err.Error())
		(*task.Peer).ClosePeer()
	}

}
func (task *ResponseBlockChain) Continue() {
	task.Channel <- *task.Peer
}

type SendNewBlock PeerMsgTask

func (task *SendNewBlock) Execute() {
	log.Println("Sending new block to peer")
	err := (*task.Peer).SendResponseBlockChainMsg(task.Msg.Data)
	if err != nil {
		log.Println("Failed to send new block to peer")
		(*task.Peer).ClosePeer()
	}
	task.Continue()
}
func (task *SendNewBlock) Continue() {
	task.Channel <- *task.Peer
}
