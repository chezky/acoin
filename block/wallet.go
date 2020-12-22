package block

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"golang.org/x/crypto/ripemd160"
)

// Wallet represents a single wallet instance. A wallet contains a private key and a public key
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet returns a wallet struct, with a private and public key
func NewWallet() *Wallet {
	private, public := newKeyPair()
	return &Wallet{
		PrivateKey: private,
		PublicKey:  public,
	}
}

// NewKeyPair is responsible for getting a public and private key pair for a wallet upon its creation. If first creates a private key based on a elliptic.P256,
// which is any random number between 10^77. The public key is the x,y coordinates of the private key. Still unclear exactly what that means, but its a
// private-key specific public-key.
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Println("error generating keypair", err)
		panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pubKey
}

// GetAddress is responsible for getting an address for a wallet. It first runs the public key though a RIPEMD160 hash of a SHA256 hash of the public key.
// Then it appends the version to the head of the public key. Next, a checksum 4 bytes long is created, by double hashing the version+publicKey.
// The full payload is then created by appending the checksum to the end of the version+publicKey. Finally that value is base58 encoded and returned as the address.
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)
	versionPayload := append([]byte{version}, pubKeyHash...)
	checkSum := checksum(versionPayload)

	fullPayload := append(versionPayload, checkSum...)
	address := Base58Encode(fullPayload)
	return address
}

// HashPubKey takes in a public key slice. It first hashes with SHA256, and then hashes it again with RIPEMD160
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		panic(err)
	}
	return RIPEMD160Hasher.Sum(nil)

}

// checksum takes in a payload of bytes, hashes it, and then hashes that hash, and returns the first 4 bytes
func checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:addressChecksumLen]
}