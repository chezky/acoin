package main

import (
	_ "encoding/gob"
	"flag"
	"fmt"
	"github.com/chezky/acoin/blockchain"
	"os"
	"strconv"
)

type CLI struct {}

func (cli *CLI) Run() {
	creatBCCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	createBCAddress := creatBCCmd.String("address", "", "The address you want the value of")

	switch os.Args[1] {
	case "createblockchain":
		err := creatBCCmd.Parse(os.Args[2:]); if err != nil {
			panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:]); if err != nil{
			fmt.Println("error printchain", err)
		}
	default:
		os.Exit(1)
	}

	if creatBCCmd.Parsed() {
		if *createBCAddress == "" {
			creatBCCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBCAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func (cli *CLI) createBlockchain(address string) {
	bc := blockchain.CreateBlockChain(address)
	bc.Db.Close()
	fmt.Println("Done!")
}

func (cli *CLI) printChain() {
	bc := blockchain.CreateBlockChain("")
	defer bc.Db.Close()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("Prev hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Transactions count %d\n", len(block.Transactions))
		fmt.Printf("Hash %x\n", block.Hash)
		pow := blockchain.NewProofOfWork(block)
		fmt.Printf("PoW: %s", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func main() {
	cli := CLI{}
	cli.Run()
}