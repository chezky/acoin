package block

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
)

// Wallets is an instance of multiple Wallet('s). More specifically it is a map of a wallets address to a Wallet struct.
// i.e : [197QdQzchU4aMF3pTryySADwCsSC6cpj4A:[*Wallet]]
type Wallets struct {
	Wallets map[string]*Wallet
}

// NewWallets loads in all the wallets stored in the wallet.dat file
func NewWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile()

	return &wallets, err
}

// LoadFromFile checks if wallet.dat exists, and reads in the contents. The wallets are gob encoded, so it first decodes them.
func (ws *Wallets) LoadFromFile() error {
	var wallets Wallets

	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		return err
	}

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		return err
	}

	ws.Wallets = wallets.Wallets

	return nil
}

// SaveToFile saves a map of wallets to wallet.dat file. It firsts encodes them and then saves them to file.
func (ws Wallets) SaveToFile() {
	var content bytes.Buffer

	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		fmt.Println("error saving wallets to file", err)
	}
	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		fmt.Println("error writing wallets to file", err)
	}
}

// CreateWallet creates a new wallet, gets its address using GetAddress, and then sets the wallet to be included in the the Wallets struct,
// and finally returns the address
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())

	ws.Wallets[address] = wallet
	return address
}

// GetWallet gets a specific wallet within a map of wallets. It takes in the wallet address, and returns the wallet
func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}
