// Acoin is a cryptocurrency with the eventual goal of being the currency for an entire decentralized internet.
//
// What exactly makes cryptocurrency so secure? How is this better than a government controlled currency? Can it be hacked?
//
// Crypto Explanation
// Cryptocurrency is made up of things called transactions, that are stored in things called blocks, that are arranged into something called a blockchain.
// Let's put aside transactions for a second and focus on what blocks are, and what a blockchain is.
//
// Blocks
// A crypto block consists of two parts. A header and a body. A header simply contains some meta data about the block, such as, the timestamp of when it was created,
// it's 'hash' (we'll discuss in a second), the previous blocks 'hash', a target for the hash (explained soon), and something called a nonce (also explained).
// Let's start with what exactly a hash is. A hash function is a special function that takes in some data of any size, and returns some string of letters and numbers,
// that somehow correlate to that data. It's impossible to 'unhash' a hash, and the cool thing is that if I hash the same data, it will return the exact hash each time.
// So what exactly do hashes do for a block? Well in order to make sure that there are a finite amount of blocks, and to make sure you can't just create blocks
// instantly, there has to be some sort of work done to create a hash. Here's where a target comes in. A hash in reality is just a large number formatted in something
// called base 16. A target is a predefined number that must be greater than the hash. Therefore if I hash some data and the hash returned is larger than the target,
// my hash is invalid. So what do I do now? Simply rehash the same data, but this time add in a number starting at the value of 1. Each time the hash is larger than
// the target, increase the number (called a nonce) by 1, and try again. Eventually, the hash should be less than the target, and voila you "mined" a block.
// So we understand how to create a hash, what a target is, and what the nonce is. Why do we need a hash though, and what does the hash of the previous block
// have to do with the new block? The hash is the reference to the block, and since it takes time to be found, it ensures validity to the block. The reference to
// the previous block's hash, is in order to create our 'blockchain'. Block C references block B which references block A. This way we can always go back x steps
// in the chain, and verify each block. This ensures each block is verifiable and valid. So that's the header of a block, simply some data that validates the block.
// The data part of the block is also the part of the block that gets hashed when finding a valid hash. The data is a list of things called transactions, and we'll get
// into that right now.
//
// Transactions
// Transactions are the 'data' part of the block and they make up the currency part of crypto. Each transaction stores things called inputs and outputs. This
// is about to get complicated so bear with me, and we'll try to take it slow. Outputs store our 'coins', the actual currency we use. Inputs are just references
// to outputs, and they store some data about which output they reference. When an input references an output, it creates a new output that stores the 'coins'
// that previously were stored in the referenced output. For example, block A has a transaction with an output. The output 'belongs' to Kevin, and it is storing 5
// coins. When Kevin decides to send Daniel 5 coins, a transaction is created. That transaction will now be inserted into the next mined block. Let's call that
// block B. The transaction on block B has an input that points to Block A's transaction output, where Kevin's 5 coins are stored. Once an output is referenced,
// it no longer can be used. So the transaction on block B creates a new output with 5 coins, and sets the owner to Daniel. Kevin can't access the coins on
// block A, since that output is referenced on block B. Daniel now has an open output with 5 coins that isn't referenced by any input, so he 'owns' 5 coins.
//
// Wallets and Transaction Signing
// So far, anyone who has the name Kevin can take the coins stored on block A, and send them around. In order to ensure each transaction is secured and legit,
// a 'wallet' is generated for kevin. A wallet is literally just 2 keys. A public key which becomes his 'address', and a private key which should be... private.
// Anything 'signed' with his private key, can be verified to be his, using his 'address' or public key. When someone wants to send the coins belonging to Kevin,
// he must sign the coins using Kevin's private key. If the private key is invalid, then the transaction is not processed. Daniel does not need to input his
// private key for the transfer, only the one who owns the coins must input his private key. This ensures privacy while at the same time makes transferring
// easy and simply.
//
// Proof of Work
// When people use the term proof of work, it simply means finding a hash for a block, that is less than the given target.

package block

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"time"
)

const (
	targetBits          = 24
	dbFile              = "blocks.db"
	blocksBucket        = "blockBucket"
	subsidy             = 10
	genesisCoinbaseData = "Genesis block for ACN"
	addressChecksumLen  = 4
	// version setting
	version    = byte(0x00)
	walletFile = "wallet.dat"
	walletChecksumLen = 4
)

// Block represents a single block withing a blockchain. A block contains headers, and the body (transactions). A block always references the previous block in a chain.
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

// NewGenesisBlock creates a new genesis block. The genesis block is the initial block created when a blockchain is created.
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

// NewBlock takes in a list of transactions, and the previous blocks hash, and creates a new block.
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
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

	err := encoder.Encode(b)
	if err != nil {
		fmt.Println("error serializing block", err)
	}

	return result.Bytes()
}

// Deserialize decodes a serialized block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	// decode d into block
	err := decoder.Decode(&block)
	if err != nil {
		fmt.Println("error decoding block", err)
		os.Exit(1)
	}

	return &block
}
