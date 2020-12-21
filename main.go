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
	_ = block.CreateBlockchain(address)
}

func (cli *CLI) printChain() {
	bc := block.NewBlockChain("")
	itr := bc.Iterator()

	for {
		blk := itr.Next()

		fmt.Printf("Prev Hash: %x\n", blk.PrevBlockHash)
		fmt.Printf("TX count: %d\n", len(blk.Transactions))
		fmt.Printf("TX ID: %x\n", blk.Transactions[0].ID)
		fmt.Printf("Hash %x\n", blk.Hash)
		prf := block.NewProofOfWork(blk)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(prf.Validate()))
		fmt.Println()

		if len(blk.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) getBalance(address string) {
	bc := block.NewBlockChain(address)
	defer bc.DB.Close()
	balance := 0
	UTXOs := bc.FindUTXOs(address)
	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
	bc := block.NewBlockChain(from)
	defer bc.DB.Close()

	tx := block.NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*block.Transaction{tx})
	fmt.Println("Success!")
}

func(cli *CLI) Run() {
	createChainCmd := flag.NewFlagSet("createchain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	createChainAddress := createChainCmd.String("address", "", "Address to which initial chain should belong to")
	createBalanceAddress := getBalanceCmd.String("address", "", "Address to which balance you would like to check")
	createSendFrom := sendCmd.String("from", "", "Address to whom this money is coming from")
	createSendTo := sendCmd.String("to", "", "Address to whom this money is being sent to")
	createSendAmount := sendCmd.String("amount", "", "Amount of money being sent")


	switch os.Args[1] {
	case "createchain":
		err := createChainCmd.Parse(os.Args[2:]); if err != nil {
			panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:]); if err != nil {
			panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:]); if err != nil {
			panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:]); if err != nil {
			panic(err)
		}
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

	if getBalanceCmd.Parsed() {
		if *createBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*createBalanceAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *createSendAmount == "" || *createSendFrom == "" || *createSendTo == "" {
			sendCmd.Usage()
			os.Exit(1)
		}
		amt, err := strconv.Atoi(*createSendAmount); if err != nil {
			fmt.Println("Amount must be a number")
			os.Exit(1)
		}
		cli.send(*createSendFrom, *createSendTo, amt)
	}

}

func main() {
	cli := CLI{}
	cli.Run()
}