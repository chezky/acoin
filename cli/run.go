package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type CLI struct{}

func (cli *CLI) Run() {
	createChainCmd := flag.NewFlagSet("createchain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)

	createChainAddress := createChainCmd.String("address", "", "Address to which initial chain should belong to")
	createBalanceAddress := getBalanceCmd.String("address", "", "Address to which balance you would like to check")
	createSendFrom := sendCmd.String("from", "", "Address to whom this money is coming from")
	createSendTo := sendCmd.String("to", "", "Address to whom this money is being sent to")
	createSendAmount := sendCmd.String("amount", "", "Amount of money being sent")

	switch os.Args[1] {
	case "createchain":
		err := createChainCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
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
		amt, err := strconv.Atoi(*createSendAmount)
		if err != nil {
			fmt.Println("Amount must be a number")
			os.Exit(1)
		}
		cli.send(*createSendFrom, *createSendTo, amt)
	}
	if createWalletCmd.Parsed() {
		cli.createWallet()
	}
}