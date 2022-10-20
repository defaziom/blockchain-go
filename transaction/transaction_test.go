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

type MockPool struct {
	mock.Mock
	Pool
}

func (m *MockPool) GetTxIns() map[string]*TxIn {
	a := m.Called()
	return a.Get(0).(map[string]*TxIn)
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

		assert.False(t, v.ContainsDuplicates([]*TransactionIml{tx1, tx2}))
	})

	t.Run("HasDuplicates", func(t *testing.T) {
		tx1.TxIns = []*TxIn{{UnspentTxOut: uTxOut1}, {UnspentTxOut: uTxOut2}}
		tx2.TxIns = []*TxIn{{UnspentTxOut: uTxOut3}, {UnspentTxOut: uTxOut1}}

		assert.True(t, v.ContainsDuplicates([]*TransactionIml{tx1, tx2}))
	})
}

func TestTxValidator_ValidateTxInsForPool(t *testing.T) {

	t.Run("Valid", func(t *testing.T) {
		txInMap := map[string]*TxIn{"moustache1": {}, "bobross3": {}}
		txIns := []*TxIn{{UnspentTxOut: &UnspentTxOut{TxOutId: "moustachewax", TxOutIndex: 1}}}
		tx := &TransactionIml{TxIns: txIns}
		mP := &MockPool{}
		mP.On("GetTxIns").Return(txInMap)

		v := &TxValidator{}
		valid := v.ValidateTxForPool(tx, mP)

		mP.AssertExpectations(t)
		assert.True(t, valid)
	})

	t.Run("Invalid", func(t *testing.T) {
		txInMap := map[string]*TxIn{"moustache1": {}, "bobross3": {}}
		txIns := []*TxIn{{UnspentTxOut: &UnspentTxOut{TxOutId: "moustache", TxOutIndex: 1}}}
		tx := &TransactionIml{TxIns: txIns}
		mP := &MockPool{}
		mP.On("GetTxIns").Return(txInMap)

		v := &TxValidator{}
		valid := v.ValidateTxForPool(tx, mP)

		mP.AssertExpectations(t)
		assert.False(t, valid)
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

	unspentTxOuts.Update(txs)
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

	txs := make([]*TransactionIml, 2)
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

func TestServiceIml_GetTotalUnspentTxOutAmount(t *testing.T) {
	unspentTxOuts := UnspentTxOutSlice([]*UnspentTxOut{
		{Address: "Mr Bob Ross", Amount: 42},
		{Address: "Mr Bob Ross", Amount: 24},
		{Address: "Mr Moustache", Amount: 64},
	})
	s := &ServiceIml{UnspentTxOuts: &unspentTxOuts}
	total := s.GetTotalUnspentTxOutAmount("Mr Bob Ross")
	assert.Equal(t, 66, total)
}

func TestServiceIml_GetUnspentTxOutsForAmount(t *testing.T) {

	bobRossTxOut1 := &UnspentTxOut{Address: "Mr Bob Ross", Amount: 42}
	bobRossTxOut2 := &UnspentTxOut{Address: "Mr Bob Ross", Amount: 24}

	unspentTxOuts := UnspentTxOutSlice([]*UnspentTxOut{
		bobRossTxOut1, bobRossTxOut2,
		{Address: "Mr Moustache", Amount: 64},
	})
	s := &ServiceIml{UnspentTxOuts: &unspentTxOuts}

	t.Run("NoLeftover", func(t *testing.T) {
		txOuts, rem, err := s.GetUnspentTxOutsForAmount(66, "Mr Bob Ross")
		assert.NoError(t, err)
		assert.Equal(t, txOuts, []*UnspentTxOut{bobRossTxOut1, bobRossTxOut2})
		assert.Equal(t, 0, rem)
	})

	t.Run("Leftover", func(t *testing.T) {
		txOuts, rem, err := s.GetUnspentTxOutsForAmount(60, "Mr Bob Ross")
		assert.NoError(t, err)
		assert.Equal(t, txOuts, []*UnspentTxOut{bobRossTxOut1, bobRossTxOut2})
		assert.Equal(t, 6, rem)
	})

	t.Run("InsufficientAmount", func(t *testing.T) {
		_, _, err := s.GetUnspentTxOutsForAmount(100, "Mr Bob Ross")
		assert.Error(t, err)
	})
}

func TestServiceIml_UnspentTxOutToTxIn(t *testing.T) {
	unspentTxOuts := UnspentTxOutSlice([]*UnspentTxOut{
		{Address: "Mr Bob Ross", Amount: 42},
		{Address: "Mr Bob Ross", Amount: 24},
		{Address: "Mr Moustache", Amount: 64},
	})
	s := &ServiceIml{}
	txIns := s.UnspentTxOutToTxIn(unspentTxOuts)

	assert.Len(t, txIns, len(unspentTxOuts))
	for _, v := range txIns {
		assert.Contains(t, unspentTxOuts, v.UnspentTxOut)
	}
}

func TestServiceIml_CreateTxOuts(t *testing.T) {
	s := &ServiceIml{}

	t.Run("NoLeftovers", func(t *testing.T) {
		expectedSender := "Mr Bob Ross"
		expectedRecipient := "Mr Moustache"
		txOuts := s.CreateTxOuts(expectedSender, expectedRecipient, 42, 0)
		assert.Len(t, txOuts, 1)
		assert.Equal(t, expectedRecipient, txOuts[0].Address)
	})

	t.Run("Leftovers", func(t *testing.T) {
		expectedSender := "Mr Bob Ross"
		expectedRecipient := "Mr Moustache"
		expectedAmount := 42
		expectedLeftover := 99
		txOuts := s.CreateTxOuts(expectedSender, expectedRecipient, expectedAmount, expectedLeftover)
		assert.Len(t, txOuts, 2)
		assert.Contains(t, txOuts, &TxOut{Address: expectedRecipient, Amount: expectedAmount})
		assert.Contains(t, txOuts, &TxOut{Address: expectedSender, Amount: expectedLeftover})
	})
}

func TestServiceIml_CreateTransaction(t *testing.T) {
	key, _ := GeneratePrivateKey()
	addr, _ := GetPublicKeyFromPrivateKey(key)
	txIns := []*TxIn{
		{UnspentTxOut: &UnspentTxOut{TxOutId: "one", TxOutIndex: 0, Address: addr}},
		{UnspentTxOut: &UnspentTxOut{TxOutId: "two", TxOutIndex: 1, Address: addr}},
	}
	txOuts := []*TxOut{{Address: "Mr Bob Ross", Amount: 42}, {Address: "Mr Moustache", Amount: 99}}
	s := &ServiceIml{}

	tx, err := s.CreateTransaction(txIns, txOuts, key)
	assert.NoError(t, err)
	assert.NotEmpty(t, tx.Id)
	assert.Equal(t, txIns, tx.TxIns)
	assert.Equal(t, txOuts, tx.TxOuts)

	for _, v := range txIns {
		assert.NotEmpty(t, v.Signature)
	}
}

func TestPoolSlice_Add(t *testing.T) {
	p := PoolSlice([]*TransactionIml{{Id: "Bob Ross tx"}})
	p.Add(&TransactionIml{Id: "Moustache tx"})

	assert.Len(t, p, 2)
	assert.Equal(t, &TransactionIml{Id: "Moustache tx"}, p[1])
}

func TestPoolSlice_Replace(t *testing.T) {
	p := PoolSlice([]*TransactionIml{{Id: "Bob Ross tx"}, {Id: "Moustache tx"}})
	replacementPool := []Transaction{&TransactionIml{Id: "Other Bob Ross tx"}, &TransactionIml{Id: "Other Moustache tx"}}
	p.Replace(replacementPool)

	assert.Len(t, p, 2)
	assert.Equal(t, []*TransactionIml{{Id: "Other Bob Ross tx"}, {Id: "Other Moustache tx"}}, []*TransactionIml(p))
}

func TestPoolSlice_Update(t *testing.T) {
	tx1 := &TransactionIml{
		TxIns: []*TxIn{
			{UnspentTxOut: &UnspentTxOut{TxOutId: "1", TxOutIndex: 0}},
			{UnspentTxOut: &UnspentTxOut{TxOutId: "2", TxOutIndex: 1}},
		},
		Id: "one",
	}
	tx2 := &TransactionIml{
		TxIns: []*TxIn{
			{UnspentTxOut: &UnspentTxOut{TxOutId: "3", TxOutIndex: 2}},
			{UnspentTxOut: &UnspentTxOut{TxOutId: "4", TxOutIndex: 3}},
		},
		Id: "two",
	}
	p := PoolSlice([]*TransactionIml{tx1, tx2})

	unspentTxOuts := UnspentTxOutSlice([]*UnspentTxOut{
		{TxOutId: "4", TxOutIndex: 3},
		{TxOutId: "3", TxOutIndex: 2},
	})
	p.Update(&unspentTxOuts)
	assert.Len(t, p, 1)
	assert.Contains(t, p, tx2)
}

func TestPoolSlice_GetTxIns(t *testing.T) {
	expectedTxIns := make([]*TxIn, 4)
	for i := 0; i < 4; i++ {
		expectedTxIns[i] = &TxIn{UnspentTxOut: &UnspentTxOut{
			TxOutId: "moustache", TxOutIndex: i,
		}}
	}
	tx1 := &TransactionIml{
		TxIns: expectedTxIns[:2],
		Id:    "one",
	}
	tx2 := &TransactionIml{
		TxIns: expectedTxIns[2:],
		Id:    "two",
	}
	p := PoolSlice([]*TransactionIml{tx1, tx2})

	txIns := p.GetTxIns()

	v, ok := txIns["moustache0"]
	assert.True(t, ok)
	assert.NotNil(t, v)
	v, ok = txIns["moustache1"]
	assert.True(t, ok)
	assert.NotNil(t, v)
	v, ok = txIns["moustache2"]
	assert.True(t, ok)
	assert.NotNil(t, v)
	v, ok = txIns["moustache3"]
	assert.True(t, ok)
	assert.NotNil(t, v)
}
