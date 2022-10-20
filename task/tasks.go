package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/defaziom/blockchain-go/block"
	"github.com/defaziom/blockchain-go/blockchain"
	"github.com/defaziom/blockchain-go/tcp"
	"github.com/defaziom/blockchain-go/transaction"
	"log"
)

func StartTasks(pc chan tcp.Peer, bc blockchain.BlockChain, ts transaction.Service) {
	for peer := range pc {
		if peer.IsClosed() {
			continue
		}
		jobExecutor := PeerJobExecutor{
			Peer: peer,
			Job: &PeerJob{
				BlockChain: bc,
				TxService:  ts,
				Peer:       peer,
			},
		}
		go func() {
			err := jobExecutor.Start()
			if err != nil {
				log.Println(err.Error())
			}
		}()
	}
}

// JobExecutor starts a Job
type JobExecutor interface {
	Start() error
}

// Job represents a series of tasks to be performed
type Job interface {
	GetNextTask() (Task, error)
}

// Task contains a business thread to be run.
type Task interface {
	// Execute performs the task
	Execute() error
}

// PeerJobExecutor contains one Job to perform
type PeerJobExecutor struct {
	JobExecutor
	Job
	tcp.Peer
}

// PeerJob is a Job that interacts with a Peer
type PeerJob struct {
	tcp.Peer
	blockchain.BlockChain
	TxService transaction.Service
}

// PeerMsgTask is a Task created from a message from a peer
type PeerMsgTask struct {
	Msg *tcp.PeerMsg
	tcp.Peer
	blockchain.BlockChain
	TxService transaction.Service
}

func (pje *PeerJobExecutor) Start() error {
	defer pje.Peer.ClosePeer()

	var task Task
	task, err := pje.Job.GetNextTask()
	if err != nil {
		return errors.New("Failed to get next task: " + err.Error())
	}
	for task != nil {
		err = task.Execute()
		if err != nil {
			log.Println("Task Failed: ", err.Error())
			return errors.New("job failed due to failed task: " + err.Error())
		}
		task, err = pje.Job.GetNextTask()
		if err != nil {
			return errors.New("Failed to get next task: " + err.Error())
		}
	}
	log.Println("Job complete!")
	return nil
}

func (pj *PeerJob) GetNextTask() (Task, error) {

	// No more tasks if peer is closed
	if pj.Peer.IsClosed() {
		return nil, nil
	}

	// Get the next message from the peer
	msg, err := pj.Peer.ReceiveMsg()
	if err != nil {
		log.Println("Failed to receive msg from peer")
		return nil, err
	}
	if msg == nil {
		return nil, nil
	}

	var t Task
	pMsgT := PeerMsgTask{
		msg,
		pj.Peer,
		pj.BlockChain,
		pj.TxService,
	}

	switch msg.Type {
	case tcp.ACK:
		ack := Ack(pMsgT)
		t = &ack
	case tcp.QUERY_ALL:
		qa := QueryAll(pMsgT)
		t = &qa
	case tcp.QUERY_LATEST:
		ql := QueryLatest(pMsgT)
		t = &ql
	case tcp.RESPONSE_BLOCKCHAIN:
		rbc := ResponseBlockChain(pMsgT)
		t = &rbc
	case tcp.QUERY_TRANSACTION_POOL:
		qtp := QueryTransactionPool(pMsgT)
		t = &qtp
	case tcp.RESPONSE_TRANSACTION_POOL:
		rtp := ResponseTransactionPool(pMsgT)
		t = &rtp
	default:
		return nil, errors.New("received unknown msg type")
	}

	return t, nil
}

type Ack PeerMsgTask

func (task *Ack) Execute() error {
	log.Println("Received ACK!")
	err := task.Peer.ClosePeer()
	if err != nil {
		log.Println("Failed to close peer: ", err.Error())
	}

	return nil
}

type QueryLatest PeerMsgTask

func (task *QueryLatest) Execute() error {
	// Send the latest block in the blockchain
	b := task.BlockChain.GetLatestBlock()
	err := task.Peer.SendResponseBlockChainMsg([]*block.Block{b})
	if err != nil {
		log.Println("Failed to send response blockchain msg", err.Error())
		return err
	}
	return nil
}

type QueryAll PeerMsgTask

func (task *QueryAll) Execute() error {
	// Send the entire blockchain
	log.Println("Sending entire blockchain")

	blocks := task.BlockChain.GetBlocks().ToSlice()
	err := task.Peer.SendResponseBlockChainMsg(blocks)
	if err != nil {
		log.Println("Failed to send response blockchain msg", err.Error())
		return err
	}
	return nil
}

type ResponseBlockChain PeerMsgTask

func (task *ResponseBlockChain) Execute() error {
	receivedBlocks, err := task.Msg.GetBlocks()
	if err != nil {
		return fmt.Errorf("failed to parse msg data: %w", err)
	}
	log.Println("Got blockchain: " + fmt.Sprint(receivedBlocks))

	if len(receivedBlocks) == 0 {
		log.Println("Got zero blocks")
	} else {
		latestBlockReceived := receivedBlocks[len(receivedBlocks)-1]
		latestBlockHeld := task.BlockChain.GetLatestBlock()
		if latestBlockReceived.Index > latestBlockHeld.Index {
			log.Println(fmt.Sprintf("Blockchain possible behind. We got: %d Peer got: %d", latestBlockHeld.Index,
				latestBlockReceived.Index))

			if latestBlockHeld.BlockHash == latestBlockReceived.PrevBlockHash {
				// The block received is the next block in the chain
				log.Println("Adding block to the blockchain")
				err := task.BlockChain.AddBlock(latestBlockReceived)
				if err != nil {
					log.Println("Received invalid block: " + err.Error())
				}
				var txs []*transaction.TransactionIml
				err = json.Unmarshal([]byte(latestBlockReceived.Data), &txs)

				// process tx
				txsToProcess := make([]*transaction.TransactionIml, 0)
				for _, tx := range txs {
					if !task.TxService.Pool().Contains(tx.Id) {
						txsToProcess = append(txsToProcess, tx)
					}
				}
				err = task.TxService.ProcessTransactions(txsToProcess, latestBlockReceived.Index)
				if err != nil {
					return err
				}

				// Update transactions in pool
				task.TxService.Pool().Update(task.TxService.GetUnspentTxOutList())

			} else if len(receivedBlocks) == 1 {
				// We have to query the chain from our peer
				log.Println("Querying peer for entire blockchain")
				err := task.Peer.SendQueryAllMsg()
				if err != nil {
					log.Println("Failed to query peer for entire chain: " + err.Error())
					return err
				}
				return nil
			} else {
				// Received chain is longer than our own chain
				log.Println("Replacing blockchain")
				receivedChainList := blockchain.DoublyLinkedBlockListCreateFromSlice(receivedBlocks)
				receivedBlockChain := &blockchain.BlockChainIml{Blocks: receivedChainList}
				task.BlockChain.ReplaceChain(receivedBlockChain, task.TxService)
				task.TxService.Pool().Update(task.TxService.GetUnspentTxOutList())
			}
		} else {
			log.Println("Received chain is not longer than our own chain. Do nothing.")
		}

	}

	// Send ACK message to notify the peer we are finished
	err = task.Peer.SendAckMsg()
	if err != nil {
		log.Println("Failed to send ack msg", err.Error())
		return err
	}
	return nil
}

type ResponseTransactionPool PeerMsgTask

func (task *ResponseTransactionPool) Execute() error {
	defer func(Peer tcp.Peer) {
		err := Peer.SendAckMsg()
		if err != nil {
			log.Println("Failed to send ack msg", err.Error())
		}
	}(task.Peer)
	// process transaction pool received by peer
	txs, err := task.Msg.GetTransactions()
	if err != nil {
		return fmt.Errorf("failed to read transactions from data: %w", err)
	}
	var txIns []*transaction.TxIn
	for _, tx := range txs {
		txIns = append(txIns, tx.TxIns...)
	}
	v := transaction.TxValidator{}

	addedTx := make([]*transaction.TransactionIml, 0)
	txPool := task.TxService.Pool()
	for _, tx := range txs {
		if txPool.Contains(tx.Id) {
			// Tx is already in the pool
			continue
		}
		if valid, reason := task.TxService.ValidateTransaction(tx); !valid {
			log.Println("Invalid transaction received: ", reason)
			continue
		}
		if !v.ValidateTxForPool(tx, task.TxService.Pool()) {
			log.Println("Invalid transaction received: A TxIn references an UnspentTxOut in the transaction pool")
			continue
		}
		task.TxService.Pool().Add(tx)
		addedTx = append(addedTx, tx)
	}

	task.TxService.UpdateUnspentTxOuts(addedTx)

	return nil
}

type QueryTransactionPool PeerMsgTask

func (task *QueryTransactionPool) Execute() error {
	// Send the transaction pool to the peer
	err := task.Peer.SendResponseTransactionPoolMsg(task.TxService.Pool())
	if err != nil {
		return fmt.Errorf("failed to send ResponseTransactionPool message: %w", err)
	}
	return nil
}
