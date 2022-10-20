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
}

type UnspentTxOut struct {
	TxOutId    string
	TxOutIndex int
	Address    string
	Amount     int
}

type UnspentTxOutList interface {
	Update(newTransactions []*TransactionIml)
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
	ContainsDuplicates(txs []*TransactionIml) bool
	ValidateTxForPool(t Transaction, p Pool) bool
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
	ProcessTransactions(transactions []*TransactionIml, blockIndex int) error
	UpdateUnspentTxOuts(transactions []*TransactionIml)
	ValidateBlockTransactions(transactions []*TransactionIml, blockIndex int) (valid bool, reason string)
	ValidateTransaction(t Transaction) (valid bool, reason string)
	GetTotalUnspentTxOutAmount(addr string) int
	GetUnspentTxOutsForAmount(amount int, address string) ([]*UnspentTxOut, int, error)
	GetUnspentTxOutList() *UnspentTxOutSlice
	UnspentTxOutToTxIn(unspentTxOuts []*UnspentTxOut) []*TxIn
	CreateTxOuts(sender string, recipient string, amount int, leftover int) []*TxOut
	CreateTransaction(txIns []*TxIn, txOuts []*TxOut, privateKey string) (*TransactionIml, error)
	Pool() *PoolSlice
}

type ServiceIml struct {
	UnspentTxOuts *UnspentTxOutSlice
	Validator
	PoolSlice *PoolSlice
}

type Pool interface {
	Replace(pool []Transaction)
	Update(list UnspentTxOutList)
	Add(tx Transaction)
	GetTxIns() map[string]*TxIn
	Contains(id string) bool
}

type PoolSlice []*TransactionIml

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
		return false, fmt.Sprintf("the TxOutIndex referred by the TxIn must match the block height"+
			": TxIn=%d"+
			" blockIndex=%d"+
			" TxId=%s", txIns[0].UnspentTxOut.TxOutIndex, blockIndex, t.GetId())
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

func (v *TxValidator) ContainsDuplicates(txs []*TransactionIml) bool {
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

func (v *TxValidator) ValidateTxForPool(tx Transaction, p Pool) bool {

	pTxInMap := p.GetTxIns()
	for _, txIn := range tx.GetTxIns() {
		lookup := fmt.Sprint(txIn.UnspentTxOut.TxOutId, txIn.UnspentTxOut.TxOutIndex)
		_, exists := pTxInMap[lookup]
		if exists {
			return false
		}
	}

	return true
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

func (u *UnspentTxOutSlice) Update(txs []*TransactionIml) {
	// Collect new unspent transactions created by new transactions
	var newUnspentTxOuts []*UnspentTxOut
	for _, t := range txs {
		for i, txOut := range t.GetTxOuts() {
			newUnspentTxOuts = append(newUnspentTxOuts, &UnspentTxOut{
				TxOutId:    t.GetId(),
				TxOutIndex: i,
				Address:    txOut.Address,
				Amount:     txOut.Amount,
			})
		}
	}

	// Build a list of UnspentTxOuts consumed by new transactions
	var consumedTxOuts UnspentTxOutSlice
	for _, t := range txs {
		for _, txIn := range t.GetTxIns() {
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

func (s *ServiceIml) ValidateBlockTransactions(transactions []*TransactionIml, blockIndex int) (valid bool,
	reason string) {
	coinbaseTx := transactions[0]
	if valid, reason = s.Validator.ValidateCoinbaseTx(coinbaseTx, blockIndex); !valid {
		return false, fmt.Sprintf("CoinbaseTx validation failed: %s", reason)
	}

	if s.Validator.ContainsDuplicates(transactions) {
		return false, "Transaction list contains duplicate TxIns"
	}

	if len(transactions) > 1 {
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
	}

	return true, ""
}

func (s *ServiceIml) ValidateTransaction(t Transaction) (valid bool, reason string) {
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

func (s *ServiceIml) ProcessTransactions(transactions []*TransactionIml, blockIndex int) error {
	if valid, reason := s.ValidateBlockTransactions(transactions, blockIndex); !valid {
		return errors.New(fmt.Sprintf("invalid block transactions: %s\n", reason))
	}

	s.UnspentTxOuts.Update(transactions)

	return nil
}

func (s *ServiceIml) UpdateUnspentTxOuts(transactions []*TransactionIml) {
	s.UnspentTxOuts.Update(transactions)
}
func (s *ServiceIml) GetTotalUnspentTxOutAmount(addr string) int {
	total := 0
	for _, unspentTxOut := range *s.UnspentTxOuts {
		if addr == unspentTxOut.Address {
			total += unspentTxOut.Amount
		}
	}
	return total
}

func (s *ServiceIml) GetUnspentTxOutsForAmount(amount int, address string) ([]*UnspentTxOut, int, error) {
	totalUnspentTxOutAmount := 0
	var unspentTxOuts []*UnspentTxOut
	for _, unspentTxOut := range *s.UnspentTxOuts {
		if address == unspentTxOut.Address {
			totalUnspentTxOutAmount += unspentTxOut.Amount
			unspentTxOuts = append(unspentTxOuts, unspentTxOut)
			if totalUnspentTxOutAmount >= amount {
				return unspentTxOuts, totalUnspentTxOutAmount - amount, nil
			}
		}
	}
	return nil, 0, errors.New("Insufficient amount for address " + address)
}

func (s *ServiceIml) UnspentTxOutToTxIn(unspentTxOuts []*UnspentTxOut) []*TxIn {
	txIns := make([]*TxIn, len(unspentTxOuts))
	for i, unspentTxOut := range unspentTxOuts {
		txIns[i] = &TxIn{
			UnspentTxOut: unspentTxOut,
		}
	}

	return txIns
}

func (s *ServiceIml) CreateTxOuts(sender string, recipient string, amount int, leftover int) []*TxOut {
	txOuts := []*TxOut{{Address: recipient, Amount: amount}}

	if leftover > 0 {
		return append(txOuts, &TxOut{Address: sender, Amount: leftover})
	} else {
		return txOuts
	}
}

func (s *ServiceIml) CreateTransaction(txIns []*TxIn, txOuts []*TxOut, privateKey string) (*TransactionIml, error) {
	tx := &TransactionIml{
		TxIns:  txIns,
		TxOuts: txOuts,
	}
	tx.Id = tx.CalcTransactionId()

	for _, txIn := range tx.TxIns {
		err := txIn.Sign(tx.Id, privateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to sign TxIn for transaction: %w", err)
		}
	}

	return tx, nil
}

func (s *ServiceIml) GetUnspentTxOutList() *UnspentTxOutSlice {
	return s.UnspentTxOuts
}

func (s *ServiceIml) Pool() *PoolSlice {
	return s.PoolSlice
}

func (p *PoolSlice) Replace(pool []Transaction) {
	*p = make([]*TransactionIml, len(pool))
	for i, tx := range pool {
		(*p)[i] = tx.(*TransactionIml)
	}
}
func (p *PoolSlice) Update(unspentTxOuts UnspentTxOutList) {
	newPool := make([]*TransactionIml, 0)
	for _, tx := range *p {
		for _, txIn := range tx.TxIns {
			if unspentTxOuts.Find(txIn.UnspentTxOut.TxOutId, txIn.UnspentTxOut.TxOutIndex) != nil {
				newPool = append(newPool, tx)
				break
			}
		}
	}
	*p = newPool
}

func (p *PoolSlice) Add(tx Transaction) {
	*p = append(*p, tx.(*TransactionIml))
}

func (p *PoolSlice) GetTxIns() map[string]*TxIn {
	txInMap := make(map[string]*TxIn, len(*p))
	for _, tx := range *p {
		for _, txIn := range tx.TxIns {
			lookup := fmt.Sprint(txIn.UnspentTxOut.TxOutId, txIn.UnspentTxOut.TxOutIndex)
			txInMap[lookup] = txIn
		}
	}

	return txInMap
}

func (p *PoolSlice) Contains(id string) bool {
	for _, tx := range *p {
		if tx.Id == id {
			return true
		}
	}
	return false
}

func GeneratePrivateKey() (string, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", err
	}

	hexKey := &HexPrivateKey{PrivateKey: *key}

	return hexKey.Marshal(), nil
}

func GetPublicKeyFromPrivateKey(key string) (string, error) {
	privateKey := &HexPrivateKey{}
	err := privateKey.UnMarshal(key)
	if err != nil {
		return "", fmt.Errorf("failed to get public key: %w", err)
	}
	pubKey := &HexPublicKey{PublicKey: privateKey.PublicKey}
	return pubKey.Marshal(), nil
}
