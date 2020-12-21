package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/boltdb/bolt"
	"os"
	"time"

	"github.com/chezky/acoin/transactions"
)

const (
	dbFile = "acoin.db"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
	blocksBucket = "blocksBucket"
)

type Block struct {
	Timestamp int64
	Transactions []*transactions.Transaction
	PrevBlockHash []byte
	Hash []byte
	Nonce int
}

type Blockchain struct {
	tip []byte
	Db  *bolt.DB
}

type Iterator struct {
	currentHash []byte
	db *bolt.DB
}

func dbExists() bool {
	if _, err  := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func NewGenesisBlock(coinbase *transactions.Transaction) *Block {
	return NewBlock([]*transactions.Transaction{coinbase}, []byte{})
}

func CreateBlockChain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil); if err != nil {
		panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := transactions.NewCoinBaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)
		b, err := tx.CreateBucket([]byte(blocksBucket)); if err != nil {
			panic(err)
		}
		err = b.Put(genesis.Hash, genesis.Serialize()); if err != nil {
			panic(err)
		}
		err = b.Put([]byte("l"), genesis.Hash); if err != nil {
			panic(err)
		}

		tip = genesis.Hash

		return nil
	}); if err != nil{
		panic(err)
	}

	return &Blockchain{tip, db}
}

func NewBlockchain() *Blockchain {
	if !dbExists() {
		fmt.Println("No existing blockchain, create one first")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil); if err != nil {
		panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		tip = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

func NewBlock(transactions []*transactions.Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		time.Now().Unix(),
		transactions,
		prevBlockHash,
		[]byte{},
		0,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash
	block.Nonce = nonce

	return block
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b); if err != nil {
		fmt.Println("error encoding block: ", err)
	}

	return result.Bytes()
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

func (bc *Blockchain) Iterator() *Iterator {
	bci := &Iterator{bc.tip, bc.Db}

	return bci
}

func (i *Iterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	}); if err != nil {
		fmt.Println("error getting next block", err)
	}

	i.currentHash = block.PrevBlockHash

	return block
}

func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block); if err != nil {
		fmt.Println("error decoding block", err)
	}

	return &block
}