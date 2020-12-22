package block

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"strconv"
)

// ProofOfWork represent a single ProofOfWork instance.
type ProofOfWork struct {
	block *Block
	target *big.Int
}

// IntToHex takes in an int64 and returns a byte slice of that int64 formatted to base16
func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 16))
}

// NewProofOfWork takes in a block and returns a ProofOfWork. the PoW target is set to the largest number allowed. For example, if we only want to allow hashes smaller
// than 1000, only if the hash is 999 or less will it pass. The target is 256 - targetBits. 256 since the hashing algorithm returns maximum 256 bits. Which is 32 bytes. If
// targetBits is 24, then the maximum number of bits is 232 which is 29 bytes. Any hash smaller than 29 bytes will be accepted. The smaller targetBits is, the easier
// it is to mine a new block.
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	return &ProofOfWork{
		block: b,
		target: target,
	}
}

// PrepareData takes in nonce, and returns a byte slice. The nonce is an int, that when added to the blocks hash, returns a hash that meets the requirement of the target.
// The byte slice returned is a slice of a blocks previous hash, transactions, timestamp, targetBits, and nonce.
func (pow *ProofOfWork) PrepareData(nonce int) []byte {
	return bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
}

// Run is a method of ProofOfWork. Run is the method that actually does what we call "mine". It's an incredibly memory intensive task, since it runs, checks if true, and
// if not, keeps running until true. The maximum value it can run until, is until the largest number an int64 can be. Every time the hash is greater than the target, the
// for loop increases nonce by 1. Each time it runs it first gets a byte slice of all the data, by calling PrepareData, then it hashes it using sha256, then checks to see if
// it satisfies the target.
func (pow *ProofOfWork) Run() (int, []byte) {
	var (
		hashInt big.Int
		hash [32]byte
	)

	nonce := 0
	fmt.Printf("Mining block containing %d transactions\n", len(pow.block.Transactions))

	for nonce < math.MaxInt64 {
		data := pow.PrepareData(nonce)
		hash = sha256.Sum256(data)
		if nonce % 100000 == 99999 {
			fmt.Printf("\r%x", hash)
		}
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}

	fmt.Print("\n\n")

	return nonce, hash[:]
}

// Validate is a method that verifies that the hash of a block is actual less than its target.
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.PrepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1
}
