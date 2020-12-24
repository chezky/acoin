package cli

import (
	"fmt"
	"github.com/chezky/acoin/block"
)

func (cli *CLI) getBalance(address string) {
	bc := block.NewBlockChain(address)
	defer bc.DB.Close()
	balance := 0
	pubKeyHash := block.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := bc.FindUTXOs(pubKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
