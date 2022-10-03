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
	"log"
	"math/big"
)

var COINBASE_AMOUNT = 50

// TxOut is a transaction output
type TxOut struct {
	Address string // ECDSA public key
	Amount  int
}

// TxIn is a transaction input
type TxIn struct {
	TxOutId    string
	TxOutIndex int
	Signature  string
}

type UnspentTxOut struct {
	TxOutId    string
	TxOutIndex int
	Address    string
	amount     int
}

type Transaction struct {
	Id     string
	TxIns  []*TxIn
	TxOuts []*TxOut
}

func (t *Transaction) CalcTransactionId() string {
	var txInContent string
	for _, txIn := range t.TxIns {
		txInContent = fmt.Sprint(txInContent, txIn.TxOutId, txIn.TxOutIndex)
	}

	var txOutContent string
	for _, txOut := range t.TxOuts {
		txOutContent = fmt.Sprint(txOutContent, txOut.Address, txOut.Amount)
	}

	hash := sha256.New()
	hash.Write([]byte(txInContent + txOutContent))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func ValidateTransaction(t *Transaction, unspentTxOuts []*UnspentTxOut) (valid bool, reason string) {
	if t.CalcTransactionId() != t.Id {
		return false, "invalid tx ID: " + t.Id
	}

	// Validate all TxIns
	for _, txIn := range t.TxIns {
		if txInValid, txInInvalidReason := ValidateTxIn(txIn, t, unspentTxOuts); !txInValid {
			return false, fmt.Sprintf("invalid TxIn: %s", txInInvalidReason)
		}
	}

	totalTxInAmount := 0
	for _, txIn := range t.TxIns {
		totalTxInAmount += GetTxInAmount(txIn, unspentTxOuts)
	}

	totalTxOutAmount := 0
	for _, txOut := range t.TxOuts {
		totalTxOutAmount += txOut.Amount
	}

	if totalTxOutAmount != totalTxInAmount {
		return false, "total TxOut amount does not equal total TxIn amount"
	}

	return true, ""
}

func ValidateCoinbaseTx(t *Transaction, blockIndex int) (valid bool, reason string) {
	if t.CalcTransactionId() != t.Id {
		return false, "invalid coinbase tx ID: " + t.Id
	}
	if len(t.TxIns) != 1 {
		return false, "only one TxIn must be specified in the coinbase transaction"
	}
	if t.TxIns[0].TxOutIndex != blockIndex {
		return false, "the TxOutIndex referred by the TxIn must match the block height"
	}
	if len(t.TxOuts) != 1 {
		return false, "only one TxOut must be specified in the coinbase transaction"
	}
	if t.TxOuts[0].Amount != COINBASE_AMOUNT {
		return false, "invalid coinbase amount"
	}
	return true, ""
}

func (txIn *TxIn) Sign(t *Transaction, privateKeyHex string, unspentTxOuts []*UnspentTxOut) error {
	referencedUnspentTxOut := FindUnspentTxOut(txIn.TxOutId, txIn.TxOutIndex, unspentTxOuts)
	if referencedUnspentTxOut == nil {
		return errors.New("cannot find referenced TxOut")
	}

	privateKey := &HexPrivateKey{}
	err := privateKey.UnMarshal(privateKeyHex)
	if err != nil {
		return err
	}
	var publicKey HexPublicKey
	publicKey = HexPublicKey(privateKey.PublicKey)
	publicKey.Marshal()

	if publicKey.Marshal() != referencedUnspentTxOut.Address {
		return errors.New("private key's public key does not match TxOut address")
	}

	bytesToSign, err := hex.DecodeString(t.Id)
	if err != nil {
		return fmt.Errorf("failed to decode transaction ID: %w", err)
	}

	sigBytes, err := ecdsa.SignASN1(rand.Reader, (*ecdsa.PrivateKey)(privateKey), bytesToSign)
	txIn.Signature = hex.EncodeToString(sigBytes)

	return nil
}

type MarshalableKey[T any] interface {
	Marshal() T
	UnMarshal(key T) error
}
type HexPublicKey ecdsa.PublicKey

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

type HexPrivateKey ecdsa.PrivateKey

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

func FindUnspentTxOut(id string, index int, unspentTxOuts []*UnspentTxOut) *UnspentTxOut {
	txOutIndex := slices.IndexFunc(unspentTxOuts, func(txOut *UnspentTxOut) bool {
		return txOut.TxOutId == id && txOut.TxOutIndex == index
	})

	if txOutIndex == -1 {
		return nil
	} else {
		return unspentTxOuts[txOutIndex]
	}

}

func GetTxInAmount(txIn *TxIn, unspentTxOuts []*UnspentTxOut) int {
	txOut := FindUnspentTxOut(txIn.TxOutId, txIn.TxOutIndex, unspentTxOuts)
	if txOut == nil {
		return 0
	} else {
		return txOut.amount
	}

}

func ValidateTxIn(txIn *TxIn, t *Transaction, unspentTxOuts []*UnspentTxOut) (valid bool, reason string) {
	referencedTxOut := FindUnspentTxOut(txIn.TxOutId, txIn.TxOutIndex, unspentTxOuts)
	if referencedTxOut == nil {
		return false, "referenced TxOut not found in unspent TxOuts"
	}

	pubKey := &HexPublicKey{}
	err := pubKey.UnMarshal(referencedTxOut.Address)
	if err != nil {
		return false, "Failed to unmarshal referenced TxOut address: " + referencedTxOut.Address
	}

	signature, err := hex.DecodeString(txIn.Signature)
	if err != nil {
		return false, "Failed to decode TxIn signature: " + txIn.Signature
	}
	hash, err := hex.DecodeString(t.Id)
	if err != nil {
		return false, "Failed to decode transaction ID: " + t.Id
	}
	valid = ecdsa.VerifyASN1((*ecdsa.PublicKey)(pubKey), hash, signature)
	if !valid {
		reason = "Invalid signature"
	}
	return valid, reason
}

func ValidateBlockTransactions(transactions []*Transaction, unspentTxOuts []*UnspentTxOut, blockIndex int) (valid bool, reason string) {
	coinbaseTx := transactions[0]
	if valid, reason = ValidateCoinbaseTx(coinbaseTx, blockIndex); !valid {
		return
	}

	// Check for duplicates
	txInMap := make(map[string]*TxIn)
	txInTotal := 0
	for _, t := range transactions {
		for _, txIn := range t.TxIns {
			txInMap[txIn.TxOutId] = txIn
			txInTotal++
		}
	}

	if len(txInMap) < txInTotal {
		return false, "Transaction list contains duplicate TxIns"
	}

	for _, t := range transactions[1:] {
		valid, reason = ValidateTransaction(t, unspentTxOuts)
		if !valid {
			return
		}
	}

	return true, ""
}

func CreateCoinbaseTransaction(addr string, blockIndex int) *Transaction {
	t := &Transaction{
		Id:     "",
		TxIns:  []*TxIn{{TxOutIndex: blockIndex}},
		TxOuts: []*TxOut{{Address: addr, Amount: COINBASE_AMOUNT}},
	}
	t.Id = t.CalcTransactionId()
	return t
}

func UpdateUnspentTxOuts(newTransactions []*Transaction, unspentTxOuts []*UnspentTxOut) []*UnspentTxOut {
	var newUnspentTxOuts []*UnspentTxOut
	for _, t := range newTransactions {
		for i, txOut := range t.TxOuts {
			newUnspentTxOuts = append(newUnspentTxOuts, &UnspentTxOut{
				TxOutId:    t.Id,
				TxOutIndex: i,
				Address:    txOut.Address,
				amount:     txOut.Amount,
			})
		}
	}

	var consumedTxOuts []*UnspentTxOut
	for _, t := range newTransactions {
		for _, txIn := range t.TxIns {
			consumedTxOuts = append(consumedTxOuts, &UnspentTxOut{
				TxOutId:    txIn.TxOutId,
				TxOutIndex: txIn.TxOutIndex,
			})
		}
	}

	var resultingUnspentTxOuts []*UnspentTxOut
	for _, txOut := range unspentTxOuts {
		if FindUnspentTxOut(txOut.TxOutId, txOut.TxOutIndex, consumedTxOuts) == nil {
			resultingUnspentTxOuts = append(resultingUnspentTxOuts, txOut)
		}
	}

	return append(resultingUnspentTxOuts, newUnspentTxOuts...)
}

func ProcessTransactions(transactions []*Transaction, unspentTxOuts []*UnspentTxOut, blockIndex int) []*UnspentTxOut {

	if valid, reason := ValidateBlockTransactions(transactions, unspentTxOuts, blockIndex); !valid {
		log.Printf("invalid block transactions: %s\n", reason)
		return unspentTxOuts
	}

	return UpdateUnspentTxOuts(transactions, unspentTxOuts)
}
