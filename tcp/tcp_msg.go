package tcp

import (
	"encoding/json"
	"errors"
	"github.com/defaziom/blockchain-go/block"
	"github.com/defaziom/blockchain-go/transaction"
	"io"
	"net"
)

type PeerMsgType int

const (
	ACK                       PeerMsgType = iota // Signals end of communication with a Peer
	QUERY_LATEST                                 // Asks for the latest block held by a Peer
	QUERY_ALL                                    // Ask for the entire blockchain held by a Peer
	RESPONSE_BLOCKCHAIN                          // Contains a single block, or an entire blockchain
	QUERY_TRANSACTION_POOL                       // Asks a Peer for the transaction pool it holds
	RESPONSE_TRANSACTION_POOL                    // Contains a list of pending transactions
)

// PeerMsg is a message from a blockchain peer
type PeerMsg struct {
	Type PeerMsgType
	Data []byte
}

// Peer represents a blockchain peer with methods to interact with
type Peer interface {
	ClosePeer() error
	IsClosed() bool
	ReceiveMsg() (*PeerMsg, error)
	SendResponseBlockChainMsg(blocks []*block.Block) error
	SendResponseTransactionPoolMsg(txs *transaction.PoolSlice) error
	SendQueryAllMsg() error
	SendQueryTransactionPoolMsg() error
	SendAckMsg() error
}

// PeerConn is a Peer with an underlying TCP connection
type PeerConn struct {
	Conn   net.Conn
	Closed bool
}

const readBufferSizeBytes = 1024

func (m PeerMsg) GetBlocks() ([]*block.Block, error) {
	var blocks []*block.Block
	err := json.Unmarshal(m.Data, &blocks)
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

func (m PeerMsg) GetTransactions() ([]*transaction.TransactionIml, error) {
	var txs []*transaction.TransactionIml
	err := json.Unmarshal(m.Data, &txs)
	if err != nil {
		return nil, err
	}
	return txs, nil
}

// ReadData reads data from a TCP connection until it receives a '\n' and returns it.
func ReadData(conn net.Conn) ([]byte, error) {

	buf := make([]byte, readBufferSizeBytes)
	data := make([]byte, 0)

	bytesRx, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	data = append(data, buf[:bytesRx]...)

	// Keep reading until we get a \n char
	for data[len(data)-1] != '\n' {
		bytesRx, err = conn.Read(buf)
		if err != nil {
			return nil, err
		}
		data = append(data, buf[:bytesRx]...)
	}

	return data, nil
}

// ClosePeer closes the underlying net.Conn
func (pc *PeerConn) ClosePeer() error {
	err := pc.Conn.Close()
	pc.Closed = true
	if err != nil {
		return err
	}
	return nil
}

// IsClosed returns if the Peer connection has been closed
func (pc *PeerConn) IsClosed() bool {
	return pc.Closed
}

// ReceiveMsg reads data from the TCP connection and unmarshals it into a PeerMsg
func (pc *PeerConn) ReceiveMsg() (*PeerMsg, error) {
	data, err := ReadData(pc.Conn)

	if err != nil {
		if errors.Is(err, io.EOF) {
			// Connection has been closed gracefully
			return nil, nil
		} else {
			return nil, err
		}
	}

	msg := &PeerMsg{}
	err = json.Unmarshal(data, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// SendResp sends a PeerMsg to a Peer
func (pc *PeerConn) SendResp(msg *PeerMsg) error {
	dataToSend, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = pc.Conn.Write(append(dataToSend, byte('\n')))
	if err != nil {
		return err
	}

	return nil
}

func (pc *PeerConn) SendResponseBlockChainMsg(blocks []*block.Block) error {
	data, err := json.Marshal(blocks)
	if err != nil {
		return err
	}
	return pc.SendResp(&PeerMsg{
		Type: RESPONSE_BLOCKCHAIN,
		Data: data,
	})
}

func (pc *PeerConn) SendQueryAllMsg() error {

	return pc.SendResp(&PeerMsg{
		Type: QUERY_ALL,
		Data: nil,
	})
}

func (pc *PeerConn) SendAckMsg() error {
	return pc.SendResp(&PeerMsg{
		Type: ACK,
		Data: nil,
	})
}

func (pc *PeerConn) SendQueryTransactionPoolMsg() error {
	return pc.SendResp(&PeerMsg{
		Type: QUERY_TRANSACTION_POOL,
		Data: nil,
	})
}

func (pc *PeerConn) SendResponseTransactionPoolMsg(tx *transaction.PoolSlice) error {
	data, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	return pc.SendResp(&PeerMsg{
		Type: QUERY_TRANSACTION_POOL,
		Data: data,
	})
}
