package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
)

type Transaction struct {
	ID []byte
	Vin []TXInput
	Vout []TXOutput
}

type TXOutput struct {
	Value int
	ScriptPubKey string
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
		panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{
		Txid:      []byte{},
		Vout:      -1,
		ScriptSig: data,
	}
	
	txout := TXOutput{
		Value:        subsidy,
		ScriptPubKey: to,
	}

	tx := Transaction{
		ID:   nil,
		Vin:  []TXInput{txin},
		Vout: []TXOutput{txout},
	}

	tx.SetID()
	return &tx
}

func (b *Block) HashTransactions() []byte {
	var (
		txHashes [][]byte
		txHash [32]byte
	)

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}