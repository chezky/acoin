package cli

import (
	"github.com/chezky/acoin/block"
)

func (cli *CLI) createChain(address string) {
	_ = block.CreateBlockchain(address)
}

