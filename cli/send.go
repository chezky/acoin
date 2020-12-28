package cli

import (
	"fmt"
	"github.com/chezky/acoin/block"
	"os"
)

func (cli *CLI) send(from, to string, amount int) {

	if !block.ValidateAddress(from) {
		fmt.Println("The sender address is invalid")
		os.Exit(1)
	}

	if !block.ValidateAddress(to) {
		fmt.Println("The receiver address is invalid")
		os.Exit(1)
	}

	bc := block.NewBlockChain(from)
	defer bc.DB.Close()

	UTXOSet := block.UTXOSet{Blockchain: bc}

	tx := block.NewUTXOTransaction(from, to, amount, &UTXOSet)
	cbTx := block.NewCoinbaseTX(from, "")
	txs := []*block.Transaction{cbTx, tx}

	newBlock := bc.MineBlock(txs)
	UTXOSet.Update(newBlock)
	fmt.Println("Success!")
}
