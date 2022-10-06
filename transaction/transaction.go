package transaction

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"math/big"
)

var COINBASE_AMOUNT = 50

// TxOut is a transaction output
type TxOut struct {
	Address string // ECDSA public key
	Amount  int
}

type SignedTx interface {
	Sign(data string, privateKeyHex string) error
	Validate(data string) (valid bool, reason string)
}

// TxIn is a transaction input
type TxIn struct {
	UnspentTxOut *UnspentTxOut
	Signature    string
	SignedTx
}

type UnspentTxOut struct {
	TxOutId    string
	TxOutIndex int
	Address    string
	Amount     int
}

type UnspentTxOutList interface {
	Update(newTransactions []Transaction)
	Find(id string, index int) *UnspentTxOut
}

type UnspentTxOutSlice []*UnspentTxOut

type Transaction interface {
	CalcTransactionId() string
	GetId() string
	GetTxIns() []*TxIn
	GetTxOuts() []*TxOut
	GetTotalTxOutAmount() int
	GetTotalTxInAmount() int
}

type Validator interface {
	ValidateCoinbaseTx(t Transaction, blockIndex int) (valid bool, reason string)
	ValidateTxAmount(t Transaction) bool
	ValidateTxId(t Transaction) bool
	ValidateSignedTx(signedTx SignedTx, data string) (valid bool, reason string)
	ContainsDuplicates(txs []Transaction) bool
}

type TxValidator struct {
}

type TransactionIml struct {
	Id     string
	TxIns  []*TxIn
	TxOuts []*TxOut
}

type MarshalableKey[T any] interface {
	Marshal() T
	UnMarshal(key T) error
}
type HexPublicKey struct {
	ecdsa.PublicKey
	MarshalableKey[string]
}
type HexPrivateKey struct {
	ecdsa.PrivateKey
	MarshalableKey[string]
}

type Service interface {
	ProcessTransactions(transactions []Transaction, unspentTxOuts UnspentTxOutList, blockIndex int) error
	ValidateBlockTransactions(v Validator, transactions []Transaction, blockIndex int) (valid bool, reason string)
}

type ServiceIml struct {
	UnspentTxOuts *UnspentTxOutSlice
	Validator
}

func (t *TransactionIml) CalcTransactionId() string {
	var txInContent string
	for _, txIn := range t.TxIns {
		txInContent = fmt.Sprint(txInContent, txIn.UnspentTxOut.TxOutId, txIn.UnspentTxOut.TxOutIndex)
	}

	var txOutContent string
	for _, txOut := range t.TxOuts {
		txOutContent = fmt.Sprint(txOutContent, txOut.Address, txOut.Amount)
	}

	hash := sha256.New()
	hash.Write([]byte(txInContent + txOutContent))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (t *TransactionIml) GetTotalTxOutAmount() int {
	totalTxOutAmount := 0
	for _, txOut := range t.TxOuts {
		totalTxOutAmount += txOut.Amount
	}

	return totalTxOutAmount
}

func (t *TransactionIml) GetTotalTxInAmount() int {
	totalTxInAmount := 0
	for _, txIn := range t.TxIns {
		totalTxInAmount += txIn.UnspentTxOut.Amount
	}

	return totalTxInAmount
}

func (t *TransactionIml) GetTxIns() []*TxIn {
	return t.TxIns
}

func (t *TransactionIml) GetTxOuts() []*TxOut {
	return t.TxOuts
}

func (t *TransactionIml) GetId() string {
	return t.Id
}

func (v *TxValidator) ValidateTxId(t Transaction) bool {
	return t.GetId() == t.CalcTransactionId()
}

func (v *TxValidator) ValidateTxAmount(t Transaction) bool {

	totalTxInAmount := t.GetTotalTxInAmount()
	totalTxOutAmount := t.GetTotalTxOutAmount()

	return totalTxInAmount == totalTxOutAmount
}

func (v *TxValidator) ValidateCoinbaseTx(t Transaction, blockIndex int) (valid bool, reason string) {
	if t.CalcTransactionId() != t.GetId() {
		return false, "invalid coinbase tx ID: " + t.GetId()
	}
	txIns := t.GetTxIns()
	if len(txIns) != 1 {
		return false, "only one TxIn must be specified in the coinbase transaction"
	}
	if txIns[0].UnspentTxOut.TxOutIndex != blockIndex {
		return false, "the TxOutIndex referred by the TxIn must match the block height"
	}
	txOuts := t.GetTxOuts()
	if len(txOuts) != 1 {
		return false, "only one TxOut must be specified in the coinbase transaction"
	}
	if txOuts[0].Amount != COINBASE_AMOUNT {
		return false, "invalid coinbase Amount"
	}
	return true, ""
}

func (v *TxValidator) ValidateSignedTx(signedTx SignedTx, data string) (valid bool, reason string) {
	return signedTx.Validate(data)
}

func (v *TxValidator) ContainsDuplicates(txs []Transaction) bool {
	// Check for duplicates
	txInMap := make(map[string]*TxIn)
	txInTotal := 0
	for _, t := range txs {
		for _, txIn := range t.GetTxIns() {
			txInMap[txIn.UnspentTxOut.TxOutId] = txIn
			txInTotal++
		}
	}

	return len(txInMap) < txInTotal
}

func (txIn *TxIn) Sign(data string, privateKeyHex string) error {
	privateKey := &HexPrivateKey{}
	err := privateKey.UnMarshal(privateKeyHex)
	if err != nil {
		return err
	}
	publicKey := &HexPublicKey{PublicKey: privateKey.PublicKey}
	publicKey.Marshal()

	if publicKey.Marshal() != txIn.UnspentTxOut.Address {
		return errors.New("private key's public key does not match TxOut address")
	}

	bytesToSign, err := hex.DecodeString(data)
	if err != nil {
		return fmt.Errorf("failed to decode transaction ID: %w", err)
	}

	sigBytes, err := ecdsa.SignASN1(rand.Reader, &privateKey.PrivateKey, bytesToSign)
	txIn.Signature = hex.EncodeToString(sigBytes)

	return nil
}

func (txIn *TxIn) Validate(data string) (valid bool, reason string) {
	pubKey := &HexPublicKey{}
	err := pubKey.UnMarshal(txIn.UnspentTxOut.Address)
	if err != nil {
		return false, "Failed to unmarshal referenced TxOut address: " + txIn.UnspentTxOut.Address
	}

	signature, err := hex.DecodeString(txIn.Signature)
	if err != nil {
		return false, "Failed to decode TxIn signature: " + txIn.Signature
	}
	hash, err := hex.DecodeString(data)
	if err != nil {
		return false, "Failed to decode transaction ID: " + data
	}
	valid = ecdsa.VerifyASN1(&pubKey.PublicKey, hash, signature)
	if !valid {
		reason = "Invalid signature"
	}
	return valid, reason
}

func (u *UnspentTxOutSlice) Update(transactions []Transaction) {
	txs := make([]*TransactionIml, len(transactions))
	for i, v := range transactions {
		txs[i] = v.(*TransactionIml)
	}
	// Collect new unspent transactions created by new transactions
	var newUnspentTxOuts []*UnspentTxOut
	for _, t := range txs {
		for i, txOut := range t.TxOuts {
			newUnspentTxOuts = append(newUnspentTxOuts, &UnspentTxOut{
				TxOutId:    t.Id,
				TxOutIndex: i,
				Address:    txOut.Address,
				Amount:     txOut.Amount,
			})
		}
	}

	// Build a list of UnspentTxOuts consumed by new transactions
	var consumedTxOuts UnspentTxOutSlice
	for _, t := range txs {
		for _, txIn := range t.TxIns {
			consumedTxOuts = append(consumedTxOuts, &UnspentTxOut{
				TxOutId:    txIn.UnspentTxOut.TxOutId,
				TxOutIndex: txIn.UnspentTxOut.TxOutIndex,
			})
		}
	}

	// Filter out the consumed TxOuts from the list of all UnspentTxOuts
	var resultingUnspentTxOuts []*UnspentTxOut
	for _, txOut := range *u {
		if consumedTxOuts.Find(txOut.TxOutId, txOut.TxOutIndex) == nil {
			resultingUnspentTxOuts = append(resultingUnspentTxOuts, txOut)
		}
	}

	*u = append(resultingUnspentTxOuts, newUnspentTxOuts...)
}

func (u *UnspentTxOutSlice) Find(id string, index int) *UnspentTxOut {
	txOutIndex := slices.IndexFunc(*u, func(txOut *UnspentTxOut) bool {
		return txOut.TxOutId == id && txOut.TxOutIndex == index
	})

	if txOutIndex == -1 {
		return nil
	} else {
		return (*u)[txOutIndex]
	}

}

func (k *HexPublicKey) Marshal() string {
	return fmt.Sprintf("%x", elliptic.Marshal(elliptic.P256(), k.X, k.Y))
}

func (k *HexPublicKey) UnMarshal(key string) error {
	bytes, err := hex.DecodeString(key)
	if err != nil {
		return err
	}
	k.Curve = elliptic.P256()
	k.X, k.Y = elliptic.Unmarshal(elliptic.P256(), bytes)
	return nil
}

func (k *HexPrivateKey) Marshal() string {
	return hex.EncodeToString(k.D.Bytes())
}

func (k *HexPrivateKey) UnMarshal(key string) error {
	k.D, _ = new(big.Int).SetString(key, 16)
	if k.D == nil {
		return errors.New("failed to unmarshal hex string: " + key)
	}
	k.PublicKey.Curve = elliptic.P256()
	k.PublicKey.X, k.PublicKey.Y = k.PublicKey.Curve.ScalarBaseMult(k.D.Bytes())

	return nil
}

func (s *ServiceIml) ValidateBlockTransactions(transactions []Transaction, blockIndex int) (valid bool, reason string) {
	coinbaseTx := transactions[0]
	if valid, reason = s.Validator.ValidateCoinbaseTx(coinbaseTx, blockIndex); !valid {
		return
	}

	if s.Validator.ContainsDuplicates(transactions) {
		return false, "Transaction list contains duplicate TxIns"
	}

	for _, t := range transactions[1:] {

		// Validate ID
		if !s.Validator.ValidateTxId(t) {
			return false, "invalid tx ID: " + t.GetId()
		}

		// Validate all TxIns
		for _, txIn := range t.GetTxIns() {
			valid, reason = s.Validator.ValidateSignedTx(txIn, t.GetId())
			if !valid {
				return
			}
		}

		// Validate tx amounts
		if !s.Validator.ValidateTxAmount(t) {
			return false, "total TxOut Amount does not equal total TxIn Amount for TxId: " + t.GetId()
		}
	}

	return true, ""
}

func CreateCoinbaseTransaction(addr string, blockIndex int) *TransactionIml {
	t := &TransactionIml{
		Id:     "",
		TxIns:  []*TxIn{{UnspentTxOut: &UnspentTxOut{TxOutIndex: blockIndex}}},
		TxOuts: []*TxOut{{Address: addr, Amount: COINBASE_AMOUNT}},
	}
	t.Id = t.CalcTransactionId()
	return t
}

func (s *ServiceIml) ProcessTransactions(transactions []Transaction, blockIndex int) error {
	if valid, reason := s.ValidateBlockTransactions(transactions, blockIndex); !valid {
		return errors.New(fmt.Sprintf("invalid block transactions: %s\n", reason))
	}

	s.UnspentTxOuts.Update(transactions)
	return nil
}

func GeneratePrivateKey() (string, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", err
	}

	hexKey := &HexPrivateKey{PrivateKey: *key}

	return hexKey.Marshal(), nil
}