package wallet

import (
	"fmt"
	"github.com/defaziom/blockchain-go/transaction"
)

type Wallet interface {
	SendToAddress(amount int, address string)
	GetBalance() int
}

type WalletService struct {
	transaction.Service
	PrivateKey string
	Address    string
}

func (w *WalletService) SendToAddress(amount int, address string) (*transaction.TransactionIml, error) {
	unspentTxOuts, leftover, err := w.Service.GetUnspentTxOutsForAmount(amount, w.Address)
	if err != nil {
		return nil, fmt.Errorf("error sending to address: %w", err)
	}
	txIns := w.Service.UnspentTxOutToTxIn(unspentTxOuts)
	txOuts := w.Service.CreateTxOuts(w.Address, address, amount, leftover)
	tx, err := w.Service.CreateTransaction(txIns, txOuts, w.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error sending to address: %w", err)
	}

	return tx, nil
}

func (w *WalletService) GetBalance() int {
	return w.Service.GetTotalUnspentTxOutAmount(w.Address)
}
