package block

import (
	"bytes"
)

// TXInput represents a single input. An input must always reference an output. The input contains an id of which transaction it references, and the index of
// which output within that referenced transaction. It also contains the address/signature of who spent the coins. An input means that coins were spent/send/transferred.
type TXInput struct {
	Txid      []byte // the id of the transaction that contains the output this input is referencing
	Vout      int // index of the output it is referencing
	Signature []byte // signature is propagated when the transaction is signed, and it's a concatenated key pair generated using a private key and a trimmed transaction hash
	PubKey    []byte // the public key of the sender. i.e: the one who owns the output
}

// UsesKey checks if the inputs lock hash matches the hash of the public key. If yes then it belongs to that address.
// pubKeyHash is just the public key hashed, without any added version or checksum
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

// IsLockedWithKey checks if the output was locked with the public key it's being tested against.
// pubKey is the hashed public key without version or checksum
func (out *TXOutput) IsLockedWithKey(pubKey []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKey) == 0
}
