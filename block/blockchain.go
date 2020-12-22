package block

import (
	"fmt"
	"github.com/boltdb/bolt"
	"os"
)

// Blockchain represents an entire blockchain. It stores the tip/tail/(hash of the last block) in a blockchain.
type Blockchain struct {
	Tip []byte
	DB  *bolt.DB
}

// BlockchainIterator stores the current hash of the block you are about to iterate over
type BlockchainIterator struct {
	currentHash []byte
	DB          *bolt.DB
}

// CreateBlockchain creates a blockchain. It first creates a genesis block, and signs the output with the address of the creator.
func CreateBlockchain(address string) *Blockchain {
	if DBExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			panic(err)
		}
		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			panic(err)
		}
		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			panic(err)
		}

		tip = genesis.Hash

		return nil
	})
	if err != nil {
		panic(err)
	}

	return &Blockchain{
		Tip: tip,
		DB:  db,
	}
}

// NewBlockChain doesn't create a blockchain, instead it identifies the tail of a previous blockchain, and that becomes the starting point of the new blockchain.
func NewBlockChain(address string) *Blockchain {
	if !DBExists() {
		fmt.Println("No existing blockchain found. Please create one first.")
		os.Exit(1)
	}

	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
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

	return &Blockchain{
		Tip: tip,
		DB:  db,
	}
}

// MineBlock takes in a list of transactions, finds the last hash of a blockchain, and creates a new block with the transactions and last hash.
// Then it updates the db and inserts the block, and updates the tail to be the hash of this new block.
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	var lastHash []byte

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			panic(err)
		}
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			panic(err)
		}
		bc.Tip = newBlock.Hash
		return nil
	})
	if err != nil {
		panic(err)
	}
}

// Iterator returns an iterator for a Blockchain
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{
		currentHash: bc.Tip,
		DB:          bc.DB,
	}
}

// Next is a method of BlockchainIterator that grabs the next block based on a hash. I.e, block A has a hash "abcd" and a lastHash of "wxyz". Iterator has
// currentHash as "abcd", it first finds the block with hash "abcd", which is block A, and then sets currentHash to the lastHash of block A, which is "wxyz".
// The next time iterator is callled, it searches for "wxyz", and stores that blocks lastHash in currentHash.
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encBlock)
		return nil
	})
	if err != nil {
		panic(err)
	}

	i.currentHash = block.PrevBlockHash
	return block
}
