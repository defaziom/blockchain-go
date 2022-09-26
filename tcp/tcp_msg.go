package tcp

import (
	"encoding/json"
	"github.com/defaziom/blockchain-go/block"
	"net"
)

type PeerMsgType int

const (
	ACK                 PeerMsgType = iota // Signals end of communication with a Peer
	QUERY_LATEST                           // Asks for the latest block held by a Peer
	QUERY_ALL                              // Ask for the entire blockchain held by a Peer
	RESPONSE_BLOCKCHAIN                    // Contains a single block, or an entire blockchain
)

// PeerMsg is a message from a blockchain peer
type PeerMsg struct {
	Type PeerMsgType
	Data []*block.Block
}

// Peer represents a blockchain peer with methods to interact with
type Peer interface {
	ClosePeer() error
	IsClosed() bool
	ReceiveMsg() (*PeerMsg, error)
	SendResponseBlockChainMsg(blocks []*block.Block) error
	SendQueryAllMsg() error
	SendAckMsg() error
}

// PeerConn is a Peer with an underlying TCP connection
type PeerConn struct {
	net.Conn
	Peer
	Closed bool
}

const readBufferSizeBytes = 1024

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
	data, err := ReadData(pc)

	if err != nil {
		if err.Error() == "EOF" {
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
	_, err = pc.Write(append(dataToSend, byte('\n')))
	if err != nil {
		return err
	}

	return nil
}

func (pc *PeerConn) SendResponseBlockChainMsg(blocks []*block.Block) error {
	return pc.SendResp(&PeerMsg{
		Type: RESPONSE_BLOCKCHAIN,
		Data: blocks,
	})
}

func (pc *PeerConn) SendQueryAllMsg() error {
	return pc.SendResp(&PeerMsg{
		Type: QUERY_ALL,
		Data: []*block.Block{},
	})
}

func (pc *PeerConn) SendAckMsg() error {
	return pc.SendResp(&PeerMsg{
		Type: ACK,
		Data: []*block.Block{},
	})
}
