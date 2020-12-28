package block

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
)

// Transaction represents a single transaction
type Transaction struct {
	ID   []byte // ID of the transaction. It's how its identified.
	Vin  []TXInput  // list of inputs // inputs always reference an output
	Vout []TXOutput // list of outputs // outputs that are referenced are "taken", unreferenced outputs have a value that can be spent
}

// IsCoinbase checks if it's the outputs from the genesis block
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
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

	tx.ID = tx.Hash()
	return &tx
}

// Serialize serializes an entire transaction. Used when inserting a transaction into a boltDB bucket, or when hashing.
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx); if err != nil {
		panic(err)
	}
	return encoded.Bytes()
}

// Hash hashes a transaction, to be used for setting a transaction's ID.
func(tx *Transaction) Hash() []byte {
	var hash [32]byte

	// When serializing a transaction, remove the transaction ID. Used to make a trimmed copy for signing
	// see if u can avoid this step, by removing the pointer
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// FindUnspentTransactions loops over every block in a blockchain. For every transaction, it checks for unspent transactions.
// For every Output in a transaction, check if that output was already "spent"/referenced by an input. The first/tail block will always be
// nil since no outputs have been referenced. Then add the transaction to the UTXOs map if unreferenced. The next time the loops runs, when it looks for
// free outputs, if they are on that map they are definitely not free.
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	UTXOs := make(map[string]TXOutputs)
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
				// Was the output spent?
				// If that transaction id is in spentTXOs, verify that the spent output idx isn't that of the outputs index.
				// i.e: transaction B has two outputs [0,1], spentTXOs stored that transaction B has spentOutIdx (tx B's output index) 0 as spent. If the index (outIdx)
				// is 0, that means it's referencing the spent output 0, so skip it. It will find the next spentOutIdx as unspent since it's not referenced in spentTXOs.
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							fmt.Printf("Output #%d was already referenced\n", outIdx)
							continue Outputs
						}
					}
				}
				// if the output is not referenced, add it to UTXOs
				outs := UTXOs[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXOs[txID] = outs
			}
			// if tx is coinbase, skip this since it has no inputs that reference outputs
			if !tx.IsCoinbase() {
				// spentTXOs starts off blank, since the tail block always has open outputs
				// if an input references an output, then that input is spent
				// the reason why we don't just add every referenced output to spentTXOs, is that only the ones that can be unlocked are relevant, and
				// will be detected when searching for outputs. If we didn't list those as taken, we would think they were spendable.
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					fmt.Println("appending tx id: ", inTxID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return UTXOs
}

// New UTXOTransaction makes a transaction from address a to address b. It first gets a list of which outputs have address a's coins. If there aren't enough
// coins to satisfy the amount that address a would like to send, return an error that address a needs more coins. Otherwise, create an input that references
// the output that has address a's coins. Then create an output that has the amount being transferred, and lock it to address b.
// If there are too many coins on the output, i.e address a wants to send 5 and the output has 10, create another output locked to address a, and store the remainder of
// the coins on that new output. Finally, using the newly created input and output(s), return a transaction that can then be stored in a block.
func NewUTXOTransaction(from, to string, amount int, UTXOSet *UTXOSet) *Transaction {
	var (
		inputs  []TXInput
		outputs []TXOutput
	)

	wallets, err := NewWallets()
	if err != nil {
		panic(err)
	}
	wallet := wallets.GetWallet(from)
	// HashPubKey returns only the public key hashed with RIPEMD160 and SHA256, it doesn't add the version or checksum
	pubKeyHash := HashPubKey(wallet.PublicKey)

	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)

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
	// notice that when we transfer the amount, we lock it to the address it's being sent to. When the output is the remainder of the value sent, it's locked to the sender.
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from)) // if its not exactly the amount, create change. For example if its 30 and he only needed 25
	}

	tx := Transaction{
		ID:   nil,
		Vin:  inputs,
		Vout: outputs,
	}

	tx.ID = tx.Hash()
	UTXOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)
	return &tx
}

// Signing and Verifying is the necessary to ensure that the an open outputs cant just be spent by anyone. Without signing and needing to insert my private key,
// anyone can use my address to send themselves my coins. When we say sign, it means to create a hash of
// a transaction, and sign it using a private key. The hash doesn't contain all the data of the transaction. We remove the public key and signature before hashing.
// Instead we set the public key to the value of the outputs public key hash. Once we generate a hash, we can sign it with that hash and the senders private key.
// When we verify, we don't need the private key, we can check if the A. senders public key B. trimmed transaction hash C. signature, all match together.
// If any of those values aren't exactly what they are meant to be, then the verification will return false. I.e: the private key given was made up.
// During verify, we get the previous transaction outputs that any of our need-to-be-verified transaction references. Stored in those outputs are all the same public key
// that belongs to the one who signed the need-to-be-verified transaction.

// Sign takes in a private key and a list of previous transactions, and signs the transaction it was called with. The private key is used to do the signing,
// while the prevTXs is map holding transactions that contain outputs. Those outputs are the outputs that the transaction you are calling this method from has
// referenced.
func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	// CP transactions don't have real inputs and therefore are not signed
	if tx.IsCoinbase() {
		return
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTX := prevTXs[hex.EncodeToString(vin.Txid)]
		// just to double check that sig is nil
		txCopy.Vin[inID].Signature = nil
		// Set the public key to the value of the senders public key.
		// i.e: john sends money to dave, the public key here would be johns
		txCopy.Vin[inID].PubKey = prevTX.Vout[vin.Vout].PubKeyHash
		// Set the ID of the trimmed copy equal to the hashed trimmed copy.
		txCopy.ID = txCopy.Hash()
		// After the hash is created, remove the public key for safety
		txCopy.Vin[inID].PubKey = nil

		// r and s are key pairs that make up a signature
		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID); if err != nil {
			panic(err)
		}
		// concatenate them together to make a full signature
		signature := append(r.Bytes(),s.Bytes()...)

		tx.Vin[inID].Signature = signature
	}
}

// TrimmedCopy removes the Signature and PubKey from a transaction and sets them to nil. We don't need to sign the input keys, only the output keys
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	return Transaction{
		ID:   tx.ID,
		Vin:  inputs,
		Vout: tx.Vout,
	}
}

// Verify is used to verify a transactions signature is valid. Like Sign, prevTXs is a map of transactions that contain the outputs that THE transaction's
// inputs referenced.
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	// create a trimmed copy
	txCopy := tx.TrimmedCopy()
	// curve is the same curve used to generate the private key
	curve := elliptic.P256()

	for inID, vin := range tx.Vin {
		// Same steps as Sign, as we need to generate the exact same trimmed transaction hash that we used for signing.
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		// r and s will get the values of the signature. The signature is a key pair, r and s.
		r := big.Int{}
		s := big.Int{}
		// bear in mind this is the original transactions signature, as the trimmed copy doesn't contain a signature
		sigLen := len(vin.Signature)
		// r is the begging part of the signature, and it takes up half the byte length
		r.SetBytes(vin.Signature[:(sigLen/2)])
		// s is is the second half of the signature
		s.SetBytes(vin.Signature[(sigLen/2):])

		// x and y will become the public key. A public key is really two values that we previously concatenated into one
		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		// first half of the public key
		x.SetBytes(vin.PubKey[:(keyLen/2)])
		// second half of the public key
		y.SetBytes(vin.PubKey[(keyLen/2):])

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false
		}
	}
	return true
}