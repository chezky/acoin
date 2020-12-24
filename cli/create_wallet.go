package cli

import (
	"fmt"
	"github.com/chezky/acoin/block"
)

func (cli *CLI) createWallet() {
	wallets, err := block.NewWallets()
	if err != nil {
		fmt.Println("error creating wallet", err)
		//os.Exit(1)
	}
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address is %s\n", address)
}
