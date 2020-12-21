package block

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"time"
)

const (
	targetBits = 20
	dbFile = "blocks.db"
	blocksBucket = "blockBucket"
	subsidy = 10
	genesisCoinbaseData = "Genesis block for ACN"
)

type Block struct {
	Timestamp int64
	Transactions []*Transaction
	PrevBlockHash []byte
	Hash []byte
	Nonce int
}

func DBExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp: time.Now().Unix(),
		Transactions: transactions,
		PrevBlockHash: prevBlockHash,
		Hash: []byte{},
		Nonce: 0,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b); if err != nil {
		fmt.Println("error serializing block", err)
	}

	return result.Bytes()
}

func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	// decode d into block
	err := decoder.Decode(&block); if err != nil {
		fmt.Println("error decoding block", err)
	}

	return &block
}
