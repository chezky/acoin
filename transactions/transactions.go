package transactions

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
)

const (
	subsidy = 10
)

type Transaction struct {
	ID []byte
	Vin []TXInput
	Vout []TXOutput
}

type TXOutput struct {
	Value int
	ScriptPublicKey string
}

type TXInput struct {
	Txid []byte
	Vout int
	ScriptSig string
}

func (tx *Transaction) SetID() {
	var (
		encoded bytes.Buffer
		hash [32]byte
	)

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx); if err != nil {
		fmt.Println("error creating id", err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}
func NewCoinBaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to %s", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}
