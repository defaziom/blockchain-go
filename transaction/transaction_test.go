package transaction

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockTransaction struct {
	mock.Mock
}

func (m *MockTransaction) CalcTransactionId() string {
	a := m.Called()
	return a.Get(0).(string)
}

func (m *MockTransaction) GetId() string {
	a := m.Called()
	return a.Get(0).(string)
}

func (m *MockTransaction) GetTxIns() []*TxIn {
	a := m.Called()
	return a.Get(0).([]*TxIn)
}

func (m *MockTransaction) GetTxOuts() []*TxOut {
	a := m.Called()
	return a.Get(0).([]*TxOut)
}

func (m *MockTransaction) GetTotalTxOutAmount() int {
	a := m.Called()
	return a.Get(0).(int)
}

func (m *MockTransaction) GetTotalTxInAmount() int {
	a := m.Called()
	return a.Get(0).(int)
}

type MockValidator struct {
	mock.Mock
	Validator
}

func (m *MockValidator) ValidateCoinbaseTx(t Transaction, blockIndex int) (valid bool, reason string) {
	a := m.Called(t, blockIndex)
	return a.Get(0).(bool), a.Get(1).(string)
	//TODO implement me
	panic("implement me")
}

func (m *MockValidator) ValidateTxAmount(t Transaction) bool {
	a := m.Called(t)
	return a.Get(0).(bool)
}

func (m *MockValidator) ValidateTxId(t Transaction) bool {
	a := m.Called(t)
	return a.Get(0).(bool)
}

func (m *MockValidator) ContainsDuplicates(txs []Transaction) bool {
	a := m.Called(txs)
	return a.Get(0).(bool)
}

func (m *MockValidator) ValidateSignedTx(txIn SignedTx, id string) (bool, string) {
	a := m.Called(txIn, id)
	return a.Get(0).(bool), a.Get(1).(string)
}

func TestTransactionIml_CalcTransactionId(t *testing.T) {
	txIns := []*TxIn{{
		UnspentTxOut: &UnspentTxOut{
			TxOutId:    "moustache",
			TxOutIndex: 42,
		},
		Signature: "signed by bob ross",
	}}
	txOuts := []*TxOut{{
		Address: "somewalletaddress",
		Amount:  99,
	}}
	transaction := &TransactionIml{
		Id:     "",
		TxIns:  txIns,
		TxOuts: txOuts,
	}

	id := transaction.CalcTransactionId()
	idBytes, err := hex.DecodeString(id)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(idBytes))
}

func TestTransactionIml_GetTotalTxOutAmount(t *testing.T) {
	tx := &TransactionIml{TxOuts: []*TxOut{{Amount: 42}, {Amount: 11}}}

	totalAmount := tx.GetTotalTxOutAmount()
	assert.Equal(t, 53, totalAmount)
}

func TestTransactionIml_GetTotalTxInAmount(t *testing.T) {
	tx := &TransactionIml{TxIns: []*TxIn{{UnspentTxOut: &UnspentTxOut{Amount: 42}},
		{UnspentTxOut: &UnspentTxOut{Amount: 11}}}}

	totalAmount := tx.GetTotalTxInAmount()
	assert.Equal(t, 53, totalAmount)
}

func TestTxValidator_ValidateCoinbaseTx(t *testing.T) {
	t.Run("ValidCoinbaseTx", func(t *testing.T) {
		bIndex := 42
		mTx := &MockTransaction{}
		mTx.On("CalcTransactionId").Return("moustache")
		mTx.On("GetId").Return("moustache")
		mTx.On("GetTxIns").Return([]*TxIn{{UnspentTxOut: &UnspentTxOut{TxOutIndex: bIndex}}})
		mTx.On("GetTxOuts").Return([]*TxOut{{Amount: COINBASE_AMOUNT}})

		v := &TxValidator{}
		valid, reason := v.ValidateCoinbaseTx(mTx, bIndex)
		mTx.AssertExpectations(t)
		assert.True(t, valid, reason)
	})

	t.Run("InvalidTxId", func(t *testing.T) {
		bIndex := 42
		mTx := &MockTransaction{}
		mTx.On("CalcTransactionId").Return("moustache")
		mTx.On("GetId").Return("nomoustache")

		v := &TxValidator{}
		valid, reason := v.ValidateCoinbaseTx(mTx, bIndex)
		mTx.AssertExpectations(t)
		assert.False(t, valid)
		assert.NotEmpty(t, reason)
	})

	t.Run("InvalidTxInLength", func(t *testing.T) {
		bIndex := 42
		mTx := &MockTransaction{}
		mTx.On("CalcTransactionId").Return("moustache")
		mTx.On("GetId").Return("moustache")
		mTx.On("GetTxIns").Return([]*TxIn{{}, {}})

		v := &TxValidator{}
		valid, reason := v.ValidateCoinbaseTx(mTx, bIndex)
		mTx.AssertExpectations(t)
		assert.False(t, valid)
		assert.NotEmpty(t, reason)
	})

	t.Run("InvalidTxInIndex", func(t *testing.T) {
		bIndex := 42
		mTx := &MockTransaction{}
		mTx.On("CalcTransactionId").Return("moustache")
		mTx.On("GetId").Return("moustache")
		mTx.On("GetTxIns").Return([]*TxIn{{UnspentTxOut: &UnspentTxOut{TxOutIndex: 99}}})

		v := &TxValidator{}
		valid, reason := v.ValidateCoinbaseTx(mTx, bIndex)
		mTx.AssertExpectations(t)
		assert.False(t, valid)
		assert.NotEmpty(t, reason)
	})

	t.Run("InvalidTxOutLength", func(t *testing.T) {
		bIndex := 42
		mTx := &MockTransaction{}
		mTx.On("CalcTransactionId").Return("moustache")
		mTx.On("GetId").Return("moustache")
		mTx.On("GetTxIns").Return([]*TxIn{{UnspentTxOut: &UnspentTxOut{TxOutIndex: bIndex}}})
		mTx.On("GetTxOuts").Return([]*TxOut{{}, {}})

		v := &TxValidator{}
		valid, reason := v.ValidateCoinbaseTx(mTx, bIndex)
		mTx.AssertExpectations(t)
		assert.False(t, valid)
		assert.NotEmpty(t, reason)
	})

	t.Run("InvalidTxOutAmount", func(t *testing.T) {
		bIndex := 42
		mTx := &MockTransaction{}
		mTx.On("CalcTransactionId").Return("moustache")
		mTx.On("GetId").Return("moustache")
		mTx.On("GetTxIns").Return([]*TxIn{{UnspentTxOut: &UnspentTxOut{TxOutIndex: bIndex}}})
		mTx.On("GetTxOuts").Return([]*TxOut{{Amount: 42}})

		v := &TxValidator{}
		valid, reason := v.ValidateCoinbaseTx(mTx, bIndex)
		mTx.AssertExpectations(t)
		assert.False(t, valid)
		assert.NotEmpty(t, reason)
	})
}

func TestTxValidator_ContainsDuplicates(t *testing.T) {
	uTxOut1 := &UnspentTxOut{TxOutId: "moustaches"}
	uTxOut2 := &UnspentTxOut{TxOutId: "are"}
	uTxOut3 := &UnspentTxOut{TxOutId: "fun"}

	tx1 := &TransactionIml{}
	tx2 := &TransactionIml{}

	v := &TxValidator{}

	t.Run("NoDuplicates", func(t *testing.T) {
		tx1.TxIns = []*TxIn{{UnspentTxOut: uTxOut1}, {UnspentTxOut: uTxOut2}}
		tx2.TxIns = []*TxIn{{UnspentTxOut: uTxOut3}}

		assert.False(t, v.ContainsDuplicates([]Transaction{tx1, tx2}))
	})

	t.Run("HasDuplicates", func(t *testing.T) {
		tx1.TxIns = []*TxIn{{UnspentTxOut: uTxOut1}, {UnspentTxOut: uTxOut2}}
		tx2.TxIns = []*TxIn{{UnspentTxOut: uTxOut3}, {UnspentTxOut: uTxOut1}}

		assert.True(t, v.ContainsDuplicates([]Transaction{tx1, tx2}))
	})
}

func TestTxIn_SignAndValidate(t *testing.T) {
	privateKey, err := GeneratePrivateKey()
	assert.NoError(t, err)

	hexPrivateKey := &HexPrivateKey{}
	err = hexPrivateKey.UnMarshal(privateKey)
	assert.NoError(t, err)

	hexPublickey := &HexPublicKey{PublicKey: hexPrivateKey.PublicKey}
	publicKey := hexPublickey.Marshal()

	txIn := &TxIn{UnspentTxOut: &UnspentTxOut{Address: publicKey}}
	err = txIn.Sign(hex.EncodeToString([]byte("moustache")), privateKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, txIn.Signature)

	valid, reason := txIn.Validate(hex.EncodeToString([]byte("moustache")))
	assert.True(t, valid, reason)
	assert.Empty(t, reason)
}

func TestUnspentTxOutSlice_Update(t *testing.T) {
	var unspentTxOuts UnspentTxOutSlice
	unspentTxOuts = []*UnspentTxOut{{TxOutId: "one", TxOutIndex: 0}, {TxOutId: "two", TxOutIndex: 1},
		{TxOutId: "three", TxOutIndex: 2}}
	txs := []*TransactionIml{{
		Id: "moustache",
		TxIns: []*TxIn{
			{UnspentTxOut: &UnspentTxOut{TxOutId: "one", TxOutIndex: 0}},
			{UnspentTxOut: &UnspentTxOut{TxOutId: "three", TxOutIndex: 2}},
		},
		TxOuts: []*TxOut{
			{Address: "Bob Ross' address", Amount: 42},
		},
	}}

	var transactions []Transaction
	for _, v := range txs {
		transactions = append(transactions, v)
	}

	unspentTxOuts.Update(transactions)
	assert.Len(t, unspentTxOuts, 2)
	assert.Contains(t, unspentTxOuts, &UnspentTxOut{
		TxOutId:    "moustache",
		TxOutIndex: 0,
		Address:    "Bob Ross' address",
		Amount:     42,
	})
	assert.Contains(t, unspentTxOuts, &UnspentTxOut{TxOutId: "two", TxOutIndex: 1})
}

func TestUnspentTxOutSlice_Find(t *testing.T) {
	var unspentTxOuts UnspentTxOutSlice
	unspentTxOuts = []*UnspentTxOut{{TxOutId: "one", TxOutIndex: 0}, {TxOutId: "two", TxOutIndex: 1},
		{TxOutId: "three", TxOutIndex: 2}}

	found := unspentTxOuts.Find("two", 1)
	assert.Equal(t, &UnspentTxOut{TxOutId: "two", TxOutIndex: 1}, found)

	notFound := unspentTxOuts.Find("moustache", 42)
	assert.Nil(t, notFound)
}

func TestServiceIml_ValidateBlockTransactions(t *testing.T) {
	mV := &MockValidator{}
	mV.On("ValidateCoinbaseTx", mock.Anything, mock.Anything).Return(true, "")
	mV.On("ContainsDuplicates", mock.Anything).Return(false)
	mV.On("ValidateTxId", mock.Anything).Return(true)
	mV.On("ValidateSignedTx", mock.Anything, mock.Anything).Return(true, "")
	mV.On("ValidateTxAmount", mock.Anything).Return(true)

	txs := make([]Transaction, 2)
	txs[0] = &TransactionIml{}
	txs[1] = &TransactionIml{TxIns: []*TxIn{{}}}

	s := &ServiceIml{Validator: mV}

	valid, reason := s.ValidateBlockTransactions(txs, 42)
	mV.AssertExpectations(t)
	assert.True(t, valid, reason)
}

func TestCreateCoinbaseTransaction(t *testing.T) {
	cbTx := CreateCoinbaseTransaction(hex.EncodeToString([]byte("bob ross address")), 42)
	v := &TxValidator{}

	valid, reason := v.ValidateCoinbaseTx(cbTx, 42)
	assert.True(t, valid, reason)
}
