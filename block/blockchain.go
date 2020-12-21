package block

import (
	"fmt"
	"github.com/boltdb/bolt"
)

type Blockchain struct {
	Tip []byte
	DB *bolt.DB
}

type BlockchainIterator struct {
	currentHash []byte
	DB *bolt.DB
}

func NewBlockChain() *Blockchain {
	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil); if err != nil {
		panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			fmt.Println("No blockchain found. Creating a new one...")
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket)); if err != nil {
				panic(err)
			}
			err = b.Put(genesis.Hash, genesis.Serialize()); if err != nil {
				fmt.Print("error putting in genesis block", err)
			}
			err = b.Put([]byte("l"), genesis.Hash); if err != nil {
				fmt.Println("error updating tail with genesis", err)
			}
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	}); if err != nil {
		panic(err)
	}

	return &Blockchain{
		Tip: tip,
		DB: db,
	}
}

func(bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	}); if err != nil {
		panic(err)
	}

	newBlock := NewBlock(data, lastHash)

	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize()); if err != nil {
			panic(err)
		}
		err = b.Put([]byte("l"), newBlock.Hash); if err != nil{
			panic(err)
		}
		bc.Tip = newBlock.Hash
		return nil
	}); if err != nil {
		panic(err)
	}
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{
		currentHash: bc.Tip,
		DB:          bc.DB,
	}
}

func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encBlock)
		return nil
	}); if err != nil {
		panic(err)
	}

	i.currentHash = block.PrevBlockHash
	return block
}