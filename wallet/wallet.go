package wallet

import (
	"fmt"
	"github.com/defaziom/blockchain-go/transaction"
)

type Wallet interface {
	SendToAddress(amount int, address string)
	GetWallet() *Info
}

type Service struct {
	TxService  transaction.Service
	PrivateKey string
	Address    string
}

type Info struct {
	Address string
	Balance int
}

func (w *Service) SendToAddress(amount int, address string) (*transaction.TransactionIml, error) {
	unspentTxOuts, leftover, err := w.TxService.GetUnspentTxOutsForAmount(amount, w.Address)
	if err != nil {
		return nil, fmt.Errorf("error sending to address: %w", err)
	}
	txIns := w.TxService.UnspentTxOutToTxIn(unspentTxOuts)
	txOuts := w.TxService.CreateTxOuts(w.Address, address, amount, leftover)
	tx, err := w.TxService.CreateTransaction(txIns, txOuts, w.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error sending to address: %w", err)
	}

	return tx, nil
}

func (w *Service) GetInfo() *Info {
	return &Info{
		Address: w.Address,
		Balance: w.TxService.GetTotalUnspentTxOutAmount(w.Address),
	}
}

func GetWalletFromPrivateKey(key string, ts transaction.Service) (*Service, error) {
	w := &Service{
		TxService:  ts,
		PrivateKey: key,
	}
	address, err := transaction.GetPublicKeyFromPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("could not generate wallet from private key: %w", err)
	}
	w.Address = address
	return w, nil
}
