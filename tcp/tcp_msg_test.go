package tcp

import (
	"bytes"
	"encoding/json"
	"github.com/defaziom/blockchain-go/block"
	"github.com/stretchr/testify/assert"
	"net"
	"strings"
	"testing"
	"time"
)

type MockConn struct {
	DataToBeRead *bytes.Buffer
}

func (m MockConn) Read(b []byte) (n int, err error) {
	return m.DataToBeRead.Read(b)
}

func (m MockConn) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (m MockConn) Close() error {
	return nil
}

func (m MockConn) LocalAddr() net.Addr {
	return nil
}

func (m MockConn) RemoteAddr() net.Addr {
	return nil
}

func (m MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestReadData(t *testing.T) {
	expectedData := "test\n"

	mockConn := &MockConn{bytes.NewBufferString(expectedData)}

	msg := Peer{
		Msg:  nil,
		Conn: mockConn,
	}

	actualData, _ := ReadData(msg)
	assert.Equal(t, expectedData, string(actualData))

	expectedLongData := strings.Repeat("A", 2048) + "\n"
	mockConn.DataToBeRead = bytes.NewBufferString(expectedLongData)
	actualLongData, _ := ReadData(msg)
	assert.Equal(t, expectedLongData, string(actualLongData))
}

func TestConnMsg_LoadMsg(t *testing.T) {
	testMsg := &PeerMsg{
		Type: QUERY_LATEST,
		Data: []*block.Block{},
	}
	testData, _ := json.Marshal(testMsg)
	mockConn := &MockConn{bytes.NewBuffer(append(testData, byte('\n')))}

	testConnMsg := Peer{
		Msg:  &PeerMsg{},
		Conn: mockConn,
	}

	err := testConnMsg.LoadMsg()
	if err != nil {
		t.Logf(err.Error())
		t.Fail()
	}

	assert.Equal(t, *testMsg, *testConnMsg.Msg)
}
