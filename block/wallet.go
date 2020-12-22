package block

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"golang.org/x/crypto/ripemd160"
	"io/ioutil"
	"os"
)

type Wallet struct{
	PrivateKey ecdsa.PrivateKey
	PublicKey []byte
}

type Wallets struct {
	Wallets map[string]*Wallet
}

func NewWallet() *Wallet {
	private, public := newKeyPair()
	return &Wallet{
		PrivateKey: private,
		PublicKey:  public,
	}
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader); if err != nil {
		fmt.Println("error generating keypair", err)
		panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pubKey
}

func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)
	versionPayload := append([]byte{version}, pubKeyHash...)
	checkSum := checksum(versionPayload)

	fullPayload := append(versionPayload, checkSum...)
	address := Base58Encode(fullPayload)
	return address
}

func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:]); if err != nil {
		panic(err)
	}
	return RIPEMD160Hasher.Sum(nil)

}

func checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:addressChecksumLen]
}

func NewWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile()

	return &wallets, err
}

func (ws *Wallets) LoadFromFile() error {
	var wallets Wallets

	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := ioutil.ReadFile(walletFile); if err != nil {
		return err
	}

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets); if err != nil {
		return err
	}

	ws.Wallets = wallets.Wallets

	return nil
}

func (ws Wallets) SaveToFile() {
	var content bytes.Buffer

	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws); if err != nil {
		fmt.Println("error saving wallets to file", err)
	}
	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644); if err != nil {
		fmt.Println("error writing wallets to file", err)
	}
}

func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())

	ws.Wallets[address] = wallet
	return address
}

func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}