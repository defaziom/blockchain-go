package task

import (
	"errors"
	"fmt"
	"github.com/defaziom/blockchain-go/block"
	"github.com/defaziom/blockchain-go/blockchain"
	"github.com/defaziom/blockchain-go/tcp"
	"log"
)

func StartTasks(pc chan tcp.Peer, bc blockchain.BlockChain) {
	for peer := range pc {
		if peer.IsClosed() {
			continue
		}
		jobExecutor := PeerJobExecutor{
			Peer: peer,
			Job: &PeerJob{
				BlockChain: bc,
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
}

// PeerMsgTask is a Task created from a message from a peer
type PeerMsgTask struct {
	Msg *tcp.PeerMsg
	tcp.Peer
}

func (pje *PeerJobExecutor) Start() error {
	var task Task
	task, err := pje.Job.GetNextTask()
	if err != nil {
		return errors.New("Failed to get next task: " + err.Error())
	}
	for task != nil {
		err := task.Execute()
		if err != nil {
			log.Println("Task Failed: ", err.Error())
			log.Println("Closing peer")
			err = pje.Peer.ClosePeer()
			if err != nil {
				log.Println("Failed to close peer: ", err.Error())
			}
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

	switch msg.Type {
	case tcp.ACK:
		t = &Ack{
			Msg:  msg,
			Peer: pj.Peer,
		}
	case tcp.QUERY_ALL:
		t = &QueryAll{
			Blocks: pj.BlockChain.GetBlocks().ToSlice(),
			PeerMsgTask: &PeerMsgTask{
				Msg:  msg,
				Peer: pj.Peer,
			},
		}
	case tcp.QUERY_LATEST:
		t = &QueryLatest{
			PeerMsgTask: &PeerMsgTask{
				Msg:  msg,
				Peer: pj.Peer,
			},
			Block: pj.BlockChain.GetLatestBlock(),
		}
	case tcp.RESPONSE_BLOCKCHAIN:
		t = &ResponseBlockChain{
			BlockChain: pj.BlockChain,
			PeerMsgTask: &PeerMsgTask{
				Msg:  msg,
				Peer: pj.Peer,
			},
		}
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

type QueryLatest struct {
	Block *block.Block
	*PeerMsgTask
}

func (task *QueryLatest) Execute() error {
	// Send the latest block in the blockchain
	err := task.Peer.SendResponseBlockChainMsg([]*block.Block{task.Block})
	if err != nil {
		log.Println("Failed to send response blockchain msg", err.Error())
		return err
	}
	return nil
}

type QueryAll struct {
	Blocks []*block.Block
	*PeerMsgTask
}

func (task *QueryAll) Execute() error {
	// Send the entire blockchain
	log.Println("Sending entire blockchain")

	err := task.Peer.SendResponseBlockChainMsg(task.Blocks)
	if err != nil {
		log.Println("Failed to send response blockchain msg", err.Error())
		return err
	}
	return nil
}

type ResponseBlockChain struct {
	*PeerMsgTask
	blockchain.BlockChain
}

func (task *ResponseBlockChain) Execute() error {
	receivedBlocks := task.Msg.Data
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
				task.BlockChain.ReplaceChain(receivedBlockChain)
			}
		} else {
			log.Println("Received chain is not longer than our own chain. Do nothing.")
		}

	}

	// Send ACK message to notify the peer we are finished
	err := task.Peer.SendAckMsg()
	if err != nil {
		log.Println("Failed to send ack msg", err.Error())
		return err
	}
	return nil
}
