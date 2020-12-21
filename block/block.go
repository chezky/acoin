package block

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
)

const (
	targetBits = 24
	dbFile = "blocks.db"
	blocksBucket = "blockBucket"
)

type Block struct {
	Timestamp int64
	Data []byte
	PrevBlockHash []byte
	Hash []byte
	Nonce int
}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp: time.Now().Unix(),
		Data: []byte(data),
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
