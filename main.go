package main

import (
	"flag"
	"fmt"
	"github.com/chezky/acoin/block"
	"os"
	"strconv"
)

type CLI struct{
	bc *block.Blockchain
}

func (cli *CLI) addBlock(data string) {
	cli.bc.AddBlock(data)
	fmt.Println("success")
}

func (cli *CLI) printChain() {
	itr := cli.bc.Iterator()

	for {
		blk := itr.Next()

		fmt.Printf("Prev Hash: %x\n", blk.PrevBlockHash)
		fmt.Printf("Data: %s\n", blk.Data)
		fmt.Printf("Hash %x\n", blk.Hash)
		prf := block.NewProofOfWork(blk)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(prf.Validate()))
		fmt.Println()

		if len(blk.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) Run() {
	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block Data")

	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:]); if err != nil {
			panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:]); if err != nil {
			panic(err)
		}
	default:
		os.Exit(1)
	}

	if addBlockCmd.Parsed(){
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func main() {
	bc := block.NewBlockChain()
	defer bc.DB.Close()
	cli := CLI{bc}
	cli.Run()
}