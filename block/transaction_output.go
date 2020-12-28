package block

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// TXOutput represents a single transaction output. TXOutputs store "coins", and are locked by a public key. The key can only be unlocked by the coins owner.
// Outputs that are referenced are spent and can't be used. Unreferenced outputs are open to being sent/spent/transferred.
type TXOutput struct {
	Value      int // amount of "coins" on the outputs
	PubKeyHash []byte // this hash is the public key hash of the guy who owns the output
}

// TXOutputs is an instance of a slice of transaction outputs.
type TXOutputs struct {
	Outputs []TXOutput // slice of TX Outputs.
}

// NewTXOutput takes in a value and address, and creates a new TX output. It locks the transaction output to the address inputted.
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{
		Value:      value,
		PubKeyHash: nil,
	}
	txo.Lock([]byte(address))
	return txo
}

// Lock takes in an address, decodes it, removes the version and checksum, and then assigns it to the outputs public key.
// Only this address can now unlock the output.
// address here is the base58 encoded version+hash+checksum, part of the functions logic is to decode, and remove the version and checksum
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	//[1: removes the version (first 1 byte, which is 2 numbers), and then :len(pubKeyHash)-4] removes the last checksum (last 4 bytes, 8 numbers long)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// Serialize takes in a slice of TXOutputs and serializes them.
func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs); if err != nil {
		fmt.Println("error serializing outputs", err)
	}
	return buff.Bytes()
}

// DeserializeOutputs takes in a byte slice of serialized outputs, and returns the decoded outputs as TXOutputs.
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs); if err != nil {
		fmt.Println("error decoding outputs", err)
	}
	return outputs
}