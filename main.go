package main

import (
	"flag"
	"fmt"
	"github.com/chezky/acoin/block"
	"os"
	"strconv"
)

type CLI struct{}

func(cli *CLI) createChain(address string) {
	bc := block.CreateBlockchain(address)
	itr := bc.Iterator()

	for {
		blk := itr.Next()

		fmt.Printf("Prev Hash: %x\n", blk.PrevBlockHash)
		fmt.Printf("TX count: %d\n", len(blk.Transactions))
		fmt.Printf("Hash %x\n", blk.Hash)
		prf := block.NewProofOfWork(blk)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(prf.Validate()))
		fmt.Println()

		if len(blk.PrevBlockHash) == 0 {
			break
		}

	}
}

	//func (cli *CLI) printChain() {
	//	itr := cli.bc.Iterator()
	//
	//	for {
	//		blk := itr.Next()
	//
	//		fmt.Printf("Prev Hash: %x\n", blk.PrevBlockHash)
	//		fmt.Printf("Data: %s\n", blk.Data)
	//		fmt.Printf("Hash %x\n", blk.Hash)
	//		prf := block.NewProofOfWork(blk)
	//		fmt.Printf("PoW: %s\n", strconv.FormatBool(prf.Validate()))
	//		fmt.Println()
	//
	//		if len(blk.PrevBlockHash) == 0 {
	//			break
	//		}
	//	}
	//}

func(cli *CLI) Run() {
	createChainCmd := flag.NewFlagSet("createchain", flag.ExitOnError)
	//printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	createChainAddress := createChainCmd.String("address", "", "Address to which initial chain should belong to")

	switch os.Args[1] {
	case "createchain":
		err := createChainCmd.Parse(os.Args[2:]); if err != nil {
			panic(err)
		}
	//case "printchain":
	//	err := printChainCmd.Parse(os.Args[2:]); if err != nil {
	//		panic(err)
	//	}
	default:
		os.Exit(1)
	}

	if createChainCmd.Parsed() {
		if *createChainAddress == "" {
			createChainCmd.Usage()
			os.Exit(1)
		}
		cli.createChain(*createChainAddress)
	}

	//if printChainCmd.Parsed() {
	//	cli.printChain()
	//}
}

func main() {
	cli := CLI{}
	cli.Run()
}