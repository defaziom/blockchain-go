package task

import (
	"github.com/defaziom/blockchain-go/block"
	"github.com/defaziom/blockchain-go/blockchain"
	"github.com/defaziom/blockchain-go/tcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockJob struct {
	mock.Mock
	Job
	tasksLeft int
}

func (m *MockJob) GetNextTask() (Task, error) {
	a := m.Called()
	if m.tasksLeft > 0 {
		m.tasksLeft--
		return a.Get(0).(Task), a.Error(1)
	} else {
		return nil, nil
	}
}

type MockTask struct {
	mock.Mock
}

func (m *MockTask) Execute() error {
	a := m.Called()
	return a.Error(0)
}

type MockPeer struct {
	mock.Mock
	tcp.Peer
}

func (m *MockPeer) SendResponseBlockChainMsg(blocks []*block.Block) error {
	a := m.Called(blocks)
	return a.Error(0)
}

func (m *MockPeer) SendQueryAllMsg() error {
	a := m.Called()
	return a.Error(0)
}

func (m *MockPeer) SendAckMsg() error {
	a := m.Called()
	return a.Error(0)
}

func (m *MockPeer) IsClosed() bool {
	a := m.Called()
	return a.Get(0).(bool)
}

func (m *MockPeer) ReceiveMsg() (*tcp.PeerMsg, error) {
	a := m.Called()
	return a.Get(0).(*tcp.PeerMsg), a.Error(1)
}

func (m *MockPeer) ClosePeer() error {
	_ = m.Called()
	return nil
}

type MockBlockChain struct {
	mock.Mock
	blockchain.BlockChain
}

func (m *MockBlockChain) GetBlocks() *blockchain.SafeDoublyLinkedBlockList {
	a := m.Called()
	return a.Get(0).(*blockchain.SafeDoublyLinkedBlockList)
}

func (m *MockBlockChain) GetLatestBlock() *block.Block {
	a := m.Called()
	return a.Get(0).(*block.Block)
}

func (m *MockBlockChain) AddBlock(b *block.Block) error {
	a := m.Called(b)
	return a.Error(0)
}

func (m *MockBlockChain) ReplaceChain(bc blockchain.BlockChain) {
	_ = m.Called(bc)
	return
}

func TestPeerJobExecutor_Start(t *testing.T) {
	mTask := &MockTask{}
	mTask.On("Execute").Return(nil).Times(5)
	mJob := &MockJob{tasksLeft: 5}
	mJob.On("GetNextTask").Return(mTask, nil).Times(6)

	testJobExecutor := &PeerJobExecutor{Job: mJob}

	_ = testJobExecutor.Start()

	mTask.AssertExpectations(t)
	mJob.AssertExpectations(t)
}

func TestPeerJob_GetNextTask(t *testing.T) {
	testBlock := &block.Block{Data: "test"}
	mPeer := &MockPeer{}
	mIsClosed := mPeer.On("IsClosed").Return(false)
	mBc := &MockBlockChain{}
	mBc.On("GetBlocks").Return(&blockchain.SafeDoublyLinkedBlockList{Value: testBlock})
	mBc.On("GetLatestBlock").Return(testBlock)

	peerJob := &PeerJob{
		Peer:       mPeer,
		BlockChain: mBc,
	}

	mReceiveMsg := mPeer.On("ReceiveMsg").Return(&tcp.PeerMsg{Type: tcp.ACK}, nil)
	task, _ := peerJob.GetNextTask()
	_ = task.(*Ack)

	mReceiveMsg.Unset()
	mPeer.On("ReceiveMsg").Return(&tcp.PeerMsg{Type: tcp.QUERY_ALL}, nil)
	task, _ = peerJob.GetNextTask()
	_ = task.(*QueryAll)

	mReceiveMsg.Unset()
	mPeer.On("ReceiveMsg").Return(&tcp.PeerMsg{Type: tcp.QUERY_LATEST}, nil)
	task, _ = peerJob.GetNextTask()
	_ = task.(*QueryLatest)

	mReceiveMsg.Unset()
	mPeer.On("ReceiveMsg").Return(&tcp.PeerMsg{Type: tcp.RESPONSE_BLOCKCHAIN}, nil)
	task, _ = peerJob.GetNextTask()
	_ = task.(*ResponseBlockChain)

	mIsClosed.Unset()
	mPeer.On("IsClosed").Return(true)
	task, _ = peerJob.GetNextTask()
	assert.Nil(t, task)
}

func TestAck_Execute(t *testing.T) {
	mPeer := &MockPeer{}
	mPeer.On("ClosePeer").Return(nil)

	ackTask := &Ack{Peer: mPeer}
	_ = ackTask.Execute()
	mPeer.AssertExpectations(t)
}

func TestQueryLatest_Execute(t *testing.T) {
	testBlock := &block.Block{Data: "test"}

	mPeer := &MockPeer{}
	mPeer.On("SendResponseBlockChainMsg", []*block.Block{testBlock}).Return(nil)

	queryLatestTask := &QueryLatest{
		Block:       testBlock,
		PeerMsgTask: &PeerMsgTask{Peer: mPeer},
	}
	_ = queryLatestTask.Execute()
	mPeer.AssertExpectations(t)
}

func TestQueryAll_Execute(t *testing.T) {
	testBlocks := []*block.Block{{Data: "test"}}

	mPeer := &MockPeer{}
	mPeer.On("SendResponseBlockChainMsg", testBlocks).Return(nil)

	queryLatestTask := &QueryAll{
		Blocks:      testBlocks,
		PeerMsgTask: &PeerMsgTask{Peer: mPeer},
	}
	_ = queryLatestTask.Execute()

	mPeer.AssertExpectations(t)
}

func TestResponseBlockChain_Execute(t *testing.T) {

	responseBlockChain := &ResponseBlockChain{
		PeerMsgTask: &PeerMsgTask{
			Msg: &tcp.PeerMsg{},
		},
	}

	// Test zero chain size
	t.Run("Zero chain size", func(t *testing.T) {
		mPeer := &MockPeer{}
		mBlockChain := &MockBlockChain{}
		responseBlockChain.PeerMsgTask.Peer = mPeer
		responseBlockChain.BlockChain = mBlockChain
		responseBlockChain.PeerMsgTask.Msg.Data = []*block.Block{}

		mPeer.On("SendAckMsg").Return(nil)
		_ = responseBlockChain.Execute()
		mPeer.AssertExpectations(t)
	})

	// Test received chain is not longer than own chain
	t.Run("Received chain is not longer than own chain", func(t *testing.T) {
		receivedBlocks := []*block.Block{{Index: 0}, {Index: 1}}
		latestBlock := &block.Block{Index: 42}
		mPeer := &MockPeer{}
		mPeer.On("SendAckMsg").Return(nil)
		mBlockChain := &MockBlockChain{}
		mBlockChain.On("GetLatestBlock").Return(latestBlock)
		responseBlockChain.PeerMsgTask.Msg.Data = receivedBlocks
		responseBlockChain.PeerMsgTask.Peer = mPeer
		responseBlockChain.BlockChain = mBlockChain

		_ = responseBlockChain.Execute()

		mBlockChain.AssertExpectations(t)
		mPeer.AssertExpectations(t)
	})

	// Test block received is next block in chain
	t.Run("Block received is next block in chain", func(t *testing.T) {
		receivedBlocks := []*block.Block{{Index: 1, PrevBlockHash: "abc"}}
		latestBlock := &block.Block{Index: 0, BlockHash: "abc"}
		mPeer := &MockPeer{}
		mPeer.On("SendAckMsg").Return(nil)
		mBlockChain := &MockBlockChain{}
		mBlockChain.On("GetLatestBlock").Return(latestBlock)
		mBlockChain.On("AddBlock", receivedBlocks[0]).Return(nil)
		responseBlockChain.PeerMsgTask.Peer = mPeer
		responseBlockChain.BlockChain = mBlockChain
		responseBlockChain.PeerMsgTask.Msg.Data = receivedBlocks

		_ = responseBlockChain.Execute()

		mBlockChain.AssertExpectations(t)
		mPeer.AssertExpectations(t)
	})

	// Test if own chain is behind by more than one
	t.Run("Own chain is behind by more than one", func(t *testing.T) {
		receivedBlocks := []*block.Block{{Index: 2, PrevBlockHash: "abc"}}
		latestBlock := &block.Block{Index: 0, BlockHash: "asdf"}
		mPeer := &MockPeer{}
		mPeer.On("SendQueryAllMsg").Return(nil)
		mBlockChain := &MockBlockChain{}
		mBlockChain.On("GetLatestBlock").Return(latestBlock)
		responseBlockChain.PeerMsgTask.Peer = mPeer
		responseBlockChain.BlockChain = mBlockChain
		responseBlockChain.PeerMsgTask.Msg.Data = receivedBlocks

		_ = responseBlockChain.Execute()

		mBlockChain.AssertExpectations(t)
		mPeer.AssertExpectations(t)
		mPeer.AssertNotCalled(t, "SendAckMsg")
	})

	// Test if received chain is longer than own chain
	t.Run("Received chain is longer than own chain", func(t *testing.T) {
		receivedBlocks := []*block.Block{{Index: 0}, {Index: 1}, {Index: 2}}
		latestBlock := &block.Block{Index: 0, BlockHash: "asdf"}
		mPeer := &MockPeer{}
		mPeer.On("SendAckMsg").Return(nil)
		mBlockChain := &MockBlockChain{}
		mBlockChain.On("GetLatestBlock").Return(latestBlock)
		mBlockChain.On("ReplaceChain", mock.AnythingOfType("*blockchain.BlockChainIml")).Return(nil)
		responseBlockChain.PeerMsgTask.Peer = mPeer
		responseBlockChain.BlockChain = mBlockChain
		responseBlockChain.PeerMsgTask.Msg.Data = receivedBlocks

		_ = responseBlockChain.Execute()

		mBlockChain.AssertExpectations(t)
		mPeer.AssertExpectations(t)
	})
}
