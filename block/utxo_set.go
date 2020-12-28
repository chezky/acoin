package block

import (
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"strings"
)

const (
	utxoBucket = "utxoBucket"
)

// UTXOSet is a struct that contains only a reference to a blockchain instance.
type UTXOSet struct {
	Blockchain *Blockchain
}

// Chainstate is the state of our current chain. It's important and a huge performance booster, as it allows us to simply keep a reference of UTXO's, instead
// of needing to loop over transaction in every block, just to find the UTXO's. Reindex function is used to run through the entire blockchain and find UTXO's.
// As you can imagine, this is a pretty memory intensive task the larger the chain gets, so we also have an Update function that is called every time an
// output is referenced or created.

// Reindex is used to reindex the chainstate. This is a pretty intensive task, so use wisely.
func (u UTXOSet) Reindex()  {
	db := u.Blockchain.DB
	bucketName := []byte(utxoBucket)

	// Delete the utxo bucket if it exists, and then recreate. Essentially empty out the bucket.
	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName); if err != nil {
			if !strings.Contains(err.Error(), "not found") {
				fmt.Println("error deleting utxoBucket", err)
			}
		}
		_, err = tx.CreateBucket(bucketName); if err != nil {
			fmt.Println("error creating utxo bucket", err)
		}
		return nil
	}); if err != nil {
		panic(err)
	}

	// Get a list of utxo's mapped to their transaction's ID
	UTXO := u.Blockchain.FindUTXO()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)


		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID); if err != nil {
				panic(err)
			}
			// Insert the serialized outputs, with their transaction's ID as the key
			err = b.Put(key, outs.Serialize())
		}
		return nil
	})
}

// FindSpendableOutputs runs through the utxo bucket, and checks if there are any UTXO's that are owned by the address requesting them.
func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.DB

	// check the db for outputs that belong to the address
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				// make sure the address owns them
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					// add the amount of coins to accumulated
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}
		return nil
	}); if err != nil {
		panic(err)
	}

	return accumulated, unspentOutputs
}

// FindUTXO is a method of UTXOSet, not to be confused with the Blockchain method of the same name. This FindUTXO is used to get the balance of an address.
// FindSpendableOutputs finds the first x amount of outputs that contain enough coins to satisfy the transfer. This function checks through every output.
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := *u.Blockchain.DB

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	}); if err != nil {
		panic(err)
	}

	return UTXOs
}

// Update is used to update the utxoBucket when there are newly referenced or created outputs. Pretty much every time a transaction is made, and also when a new block
// is added to the chain. Find the outputs on a TX that an input references, and then if they don't match the index in the input, that output is free.
func (u UTXOSet) Update(block *Block) {
	db := u.Blockchain.DB

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			// Skip coinbase transactions, as we don't care about their inputs.
			if !tx.IsCoinbase() {
				for _, vin := range tx.Vin {
					updatedOuts := TXOutputs{}
					// Get every output stored on the transaction that this input references. Don't worry, as we will make sure
					// to only store the ones that don't have the same index as vin.Vout.
					outsBytes := b.Get(vin.Txid)
					outs := DeserializeOutputs(outsBytes)

					// make sure to only store TX not referenced by inputs
					// TODO: um what if for example, the output on a transaction is referenced by a different input. It was just added to the bucket
					// since it isn't referenced by this input. Can it happen that even though it isn't referenced by this input, it still is referenced
					// elsewhere? Maybe I'm missing something. Needs to be checked out.
					for outIdx, out := range outs.Outputs {
						if outIdx != vin.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					// If there are no open outputs anymore on that transaction, delete the transaction.
					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(vin.Txid); if err != nil {
							fmt.Println("error deleting txid", vin.Txid, err)
						}
					// Otherwise, insert that output into the DB.
					} else {
						err := b.Put(vin.Txid, updatedOuts.Serialize()); if err != nil {
							fmt.Println("error inserting output", err)
						}
					}
				}
			}
			// Now is the part where we insert all the outputs on a new transaction. Applies to coinbase too, since we care about coinbase outputs.
			newOutputs := TXOutputs{}
			// for every output, store it in our struct
			for _, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}
			// Insert out slice of outputs into the db.
			err := b.Put(tx.ID, newOutputs.Serialize()); if err != nil {
				fmt.Println("error putting in serialized txID", tx.ID, err)
			}
		}
		return nil
	}); if err != nil {
		panic(err)
	}
}