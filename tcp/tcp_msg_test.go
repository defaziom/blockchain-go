package tcp

import (
	"bytes"
	"encoding/json"
	"github.com/defaziom/blockchain-go/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net"
	"strings"
	"testing"
)

type MockConn struct {
	mock.Mock
	net.Conn
	DataToBeRead *bytes.Buffer
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	_ = m.Called()
	return m.DataToBeRead.Read(b)
}

func TestReadData(t *testing.T) {
	expectedData := "test\n"

	mockConn := &MockConn{DataToBeRead: bytes.NewBufferString(expectedData)}
	mockConn.On("Read").Return()

	actualData, _ := ReadData(mockConn)
	assert.Equal(t, expectedData, string(actualData))

	expectedLongData := strings.Repeat("A", 2048) + "\n"
	mockConn.DataToBeRead = bytes.NewBufferString(expectedLongData)
	actualLongData, _ := ReadData(mockConn)
	assert.Equal(t, expectedLongData, string(actualLongData))
}

func TestPeerConn_ReceiveMsg(t *testing.T) {
	data, _ := json.Marshal([]*block.Block{})
	testMsg := &PeerMsg{
		Type: QUERY_LATEST,
		Data: data,
	}
	testData, _ := json.Marshal(testMsg)
	mockConn := &MockConn{DataToBeRead: bytes.NewBuffer(append(testData, byte('\n')))}
	mockConn.On("Read").Return()

	testPeerConn := PeerConn{
		Conn: mockConn,
	}

	actualMsg, err := testPeerConn.ReceiveMsg()
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	assert.Equal(t, *testMsg, *actualMsg)
}
