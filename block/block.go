package block

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"time"
)

const (
	targetBits = 18
	dbFile = "blocks.db"
	blocksBucket = "blockBucket"
	subsidy = 10
	genesisCoinbaseData = "Genesis block for ACN"
	addressChecksumLen = 4
	// version setting
	version = byte(0x00)
	walletFile = "wallet.dat"
)

// Block represents a single block withing a blockchain. A block contains headers, and the body (transactions). A block always references the previous block in a chain.
type Block struct {
	Timestamp int64
	Transactions []*Transaction
	PrevBlockHash []byte
	Hash []byte
	Nonce int
}

// DBExists checks whether bolt has a db created. If yes that means there is a blockchain already created
func DBExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// NewGenesisBlock creates a new genesis block. The genesis block is the initial block created when a blockchain is created.
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

// NewBlock takes in a list of transactions, and the previous blocks hash, and creates a new block.
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

// Serialize serializes/encodes a block
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b); if err != nil {
		fmt.Println("error serializing block", err)
	}

	return result.Bytes()
}

// Deserialize decodes a serialized block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	// decode d into block
	err := decoder.Decode(&block); if err != nil {
		fmt.Println("error decoding block", err)
		os.Exit(1)
	}

	return &block
}
