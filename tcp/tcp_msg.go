package tcp

import (
	"encoding/json"
	"github.com/defaziom/blockchain-go/block"
	"log"
	"net"
)

type PeerMsgType int

const (
	ACK PeerMsgType = iota
	QUERY_LATEST
	QUERY_ALL
	RESPONSE_BLOCKCHAIN
)

type PeerMsg struct {
	Type PeerMsgType
	Data []*block.Block
}

type Peer interface {
	ClosePeer()
	ReceiveMsg() (*PeerMsg, error)
	SendResponseBlockChainMsg(blocks []*block.Block) error
	SendQueryAllMsg() error
	SendAckMsg() error
}

type PeerConn struct {
	net.Conn
	Peer
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

// LoadMsg reads data from the TCP connection and unmarshals it into a PeerMsg
func (connMsg *PeerConn) LoadMsg() (*PeerMsg, error) {
	data, err := ReadData(connMsg)

	if err != nil {
		return nil, err
	}

	msg := &PeerMsg{}
	err = json.Unmarshal(data, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (connMsg *PeerConn) SendResp(msg *PeerMsg) error {
	dataToSend, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = connMsg.Write(append(dataToSend, byte('\n')))
	if err != nil {
		return err
	}

	return nil
}

func (connMsg *PeerConn) CloseConn() {
	err := connMsg.Conn.Close()
	if err != nil {
		log.Fatalln("Failed to close connection", err)
	}
}

func (connMsg *PeerConn) SendResponseBlockChainMsg(blocks []*block.Block) error {
	return connMsg.SendResp(&PeerMsg{
		Type: RESPONSE_BLOCKCHAIN,
		Data: blocks,
	})
}

func (connMsg *PeerConn) SendQueryAllMsg() error {
	return connMsg.SendResp(&PeerMsg{
		Type: QUERY_ALL,
		Data: []*block.Block{},
	})
}

func (connMsg *PeerConn) SendAckMsg() error {
	return connMsg.SendResp(&PeerMsg{
		Type: ACK,
		Data: []*block.Block{},
	})
}
