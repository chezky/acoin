package cli

import (
	"fmt"
	"github.com/chezky/acoin/block"
	"strconv"
)

func (cli *CLI) printChain() {
	bc := block.NewBlockChain("")
	itr := bc.Iterator()

	for {
		blk := itr.Next()

		fmt.Printf("Prev Hash: %x\n", blk.PrevBlockHash)
		fmt.Printf("Block height: %d\n", blk.Height)
		fmt.Printf("TX count: %d\n", len(blk.Transactions))
		for idx, tx := range blk.Transactions {
			fmt.Printf("TX #: %x/%x\n", idx+1, len(blk.Transactions))
			fmt.Printf("TX ID: %x\n", tx.ID)
			fmt.Printf("# of inputs: %d\n", len(tx.Vin))
			for idxIn, in := range tx.Vin {
				fmt.Printf("Input %d/%d\n", idxIn+1, len(tx.Vin))
				fmt.Printf("Input ID: %x\n", in.Txid)
			}
			fmt.Printf("# of outputs %d\n", len(tx.Vout))
			for idxOut, out := range tx.Vout {
				fmt.Printf("output # %d/%d\n", idxOut+1, len(tx.Vout))
				fmt.Printf("output value: %d\n", out.Value)
				fmt.Printf("output pubKeyHash: %s\n", out.PubKeyHash)
			}
		}
		fmt.Printf("Hash %x\n", blk.Hash)
		prf := block.NewProofOfWork(blk)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(prf.Validate()))
		fmt.Println()

		if len(blk.PrevBlockHash) == 0 {
			break
		}
	}
}
