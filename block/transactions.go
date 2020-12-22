package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
)

// Transaction represents a single transaction
type Transaction struct {
	ID   []byte
	Vin  []TXInput  // list of inputs // inputs always reference an output
	Vout []TXOutput // list of outputs // outputs that are referenced are "taken", unreferenced outputs have a value that can be spent
}

// TXOutput represents a single transaction output. TXOutputs store "coins", and are locked by a public key. The key can only be unlocked by the coins owner.
// Outputs that are referenced are spent and can't be used. Unreferenced outputs are open to being sent/spent/transferred.
type TXOutput struct {
	Value      int // amount of "coins" on the outputs
	PubKeyHash []byte
}

// TXInput represents a single input. An input must always reference an output. The input contains an id of which transaction it references, and the index of
// which output within that referenced transaction. It also contains the address/signature of who spent the coins. An input means that coins were spent/send/transferred.
type TXInput struct {
	Txid      []byte
	Vout      int // index of the output it is referencing
	Signature []byte
	PubKey    []byte
}

// UsesKey checks if the inputs lock hash matches the hash of the public key. If yes then it belongs to that address.
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

// Lock takes in an address, decodes it, removes the version and checksum, and then assigns it to the outputs public key.
// Only this address can now unlock the output.
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	//[1: removes the version (first 1 byte, which is 2 numbers), and then :len(pubKeyHash)-4] removes the last checksum (last 4 bytes, 8 numbers long)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey checks if the output was locked with the public key it's being tested against.
func (out *TXOutput) IsLockedWithKey(pubKey []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKey) == 0
}

// IsCoinbase checks if it's the outputs from the genesis block
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// SetID sets a transaction id. It's a hash of an encoded transaction
func (tx *Transaction) SetID() {
	var (
		encoded bytes.Buffer
		hash    [32]byte
	)

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{
		Value:      value,
		PubKeyHash: nil,
	}
	txo.Lock([]byte(address))
	return txo
}

// NewCoinbaseTX creates a new transaction for the initial genesis block
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{
		Txid:      []byte{},
		Vout:      -1,
		Signature: nil,
		PubKey:    []byte(data),
	}

	txout := NewTXOutput(subsidy, to)

	tx := Transaction{
		ID:   nil,
		Vin:  []TXInput{txin},
		Vout: []TXOutput{*txout},
	}

	tx.SetID()
	return &tx
}

// HashTransactions joins together a slice of transaction ID's, and hashes them together. Used when preparing a blocks data
func (b *Block) HashTransactions() []byte {
	var (
		txHashes [][]byte
		txHash   [32]byte
	)

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}

// FindUnspentTransactions loops over every block in a blockchain. For every transaction, it checks for unspent transactions.
// For every Output in a transaction, check if that output was already "spent"/referenced by an input. The first/tail block will always be
// nil since no outputs have been referenced. Then check if any of those outputs belong to the address you requested. If they do, add the transaction.
// If not, for every input that belongs to the address, add it to a map. The next time the loops runs, when it looks for free outputs, if they are on that map
// they are definitely not free.
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()
	blockIdx := 0
	for {
		fmt.Println("block idx #", blockIdx)
		// iterate over every block
		block := bci.Next()
		blockIdx++

		// iterate over every transaction in a block
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				fmt.Println("SpentTXOs: ", spentTXOs)
				//	was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							fmt.Printf("Output #%d was already referenced\n", outIdx)
							continue Outputs
						}
					}
				}
				//if it wasn't spent yet, verify that is can be unlocked
				if out.IsLockedWithKey(pubKeyHash) {
					// only if it can be unlocked with that address, can it be counted as an unspentTX
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if !tx.IsCoinbase() {
				// spentTXOs starts off blank, since the tail block always has open outputs
				// if an input references an output, then that input is spent
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.Txid)
						fmt.Println("appending tx id: ", inTxID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	fmt.Println("length of unspentTX is: ", len(unspentTXs))
	return unspentTXs
}

// FindUTXOs checks and verifies all outputs in a transaction that contains unsent outputs, whether or not they belong to that public key
func (bc *Blockchain) FindUTXOs(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTXs {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

// New UTXOTransaction makes a transaction from address a to address b. It first gets a list of which outputs have address a's coins. If there aren't enough
// coins to satisfy the amount that address a would like to send, return an error that address a needs more coins. Otherwise, create an input that references
// the output that has address a's coins. Then create an output that has the amount being transferred, and lock it to address b.
// If there are too many coins on the output, i.e address a wants to send 5 and the output has 10, create another output locked to address a, and store the remainder of
// the coins on that new output. Finally, using the newly created input and output(s), return a transaction that can then be stored in a block.
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var (
		inputs  []TXInput
		outputs []TXOutput
	)

	wallets, err := NewWallets()
	if err != nil {
		panic(err)
	}
	wallet := wallets.GetWallet(from)
	pubKeyHash := HashPubKey(wallet.PublicKey)

	acc, validOutputs := bc.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		fmt.Println("ERROR: Not enough funds")
		// TODO: fix this
		os.Exit(1)
	}

	// build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			panic(err)
		}

		for _, out := range outs {
			input := TXInput{
				Txid:      txID,
				Vout:      out,
				Signature: nil,
				PubKey:    wallet.PublicKey,
			}
			inputs = append(inputs, input)
		}
	}

	// build a list of outputs
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from)) // if its not exactly the amount, create change. For example if its 30 and he only needed 25
	}

	tx := Transaction{
		ID:   nil,
		Vin:  inputs,
		Vout: outputs,
	}

	tx.SetID()
	return &tx
}

// FindSpendableOutputs gets a list of unspent transactions. It then loops over every transaction, and checks if any of the outputs match the signature.
// If they do, add the "coins"/value of the output to a variable. Once it finds enough coins, break the loop, and return a map of the outputs ID's that store the coins it found.
func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	return accumulated, unspentOutputs
}
