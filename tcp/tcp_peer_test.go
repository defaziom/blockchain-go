package tcp

import (
	"github.com/defaziom/blockchain-go/block"
	"github.com/defaziom/blockchain-go/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net"
	"testing"
	"time"
)

type MockTcpDialer struct {
	mock.Mock
	NetDialer
}

func (m *MockTcpDialer) Dial(address string) (net.Conn, error) {
	a := m.Called()
	return a.Get(0).(net.Conn), a.Error(1)
}

type MockPeer struct {
	mock.Mock
	Peer
}

func (m *MockPeer) SendResponseBlockChainMsg(blocks []*block.Block) error {
	a := m.Called()
	return a.Error(0)
}

func TestGetPeers(t *testing.T) {
	testConnList := []*database.PeerConnInfo{{
		Ip:   "1.1.1.1",
		Port: 42,
	}, {
		Ip:   "2.2.2.2",
		Port: 99,
	}}

	mConn := &MockConn{}
	mConn.On("Read").Return()
	mTcpDialer := &MockTcpDialer{}
	mTcpDialer.On("Dial").Return(mConn, nil)

	peers, _ := GetPeers(testConnList, mTcpDialer)
	assert.Len(t, peers, len(testConnList))
	mTcpDialer.AssertExpectations(t)
	for _, p := range peers {
		peerCon := p.(*PeerConn)
		assert.NotNil(t, peerCon.Conn)
	}
}

func TestBroadCastBlockToPeers(t *testing.T) {
	testBlock := &block.Block{
		Timestamp:     time.Time{},
		Data:          "test",
		PrevBlockHash: "",
		BlockHash:     "",
		Index:         0,
		Nonce:         0,
		Difficulty:    0,
	}
	mockPeer := &MockPeer{}
	mockPeer.On("SendResponseBlockChainMsg").Return(nil)
	peers := []Peer{mockPeer}
	c := make(chan Peer, 1)

	BroadCastBlockToPeers(testBlock, peers, c)

	mockPeer.AssertExpectations(t)
	actualPeer := <-c
	assert.Equal(t, mockPeer, actualPeer)
}
