package cli

import (
	"github.com/chezky/acoin/block"
)

func (cli *CLI) createChain(address string) {
	bc := block.CreateBlockchain(address)
	defer bc.DB.Close()

	UTXOSet := block.UTXOSet{Blockchain: bc}
	UTXOSet.Reindex()
}

