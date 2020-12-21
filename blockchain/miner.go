package blockchain

import(
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"strconv"
)

const (
	targetBits = 24
)

type ProofOfWork struct {
	block *Block
	target *big.Int
}

func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 16))
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{
		b,
		target,
	}

	return pow
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var (
		hashInt big.Int
		hash [32]byte
	)
	maxNonce := math.MaxInt64
	nonce := 0

	//fmt.Printf("Mining the block containing \"%s\"\n", pow.block.)
	for nonce < maxNonce {
		data := pow.prepareDate(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce ++
		}
	}

	fmt.Printf("\n\n")

	return nonce, hash[:]

}

func (pow *ProofOfWork) prepareDate(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareDate(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1
}