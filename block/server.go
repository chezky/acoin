package block

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
)

const (
	protocol = "tcp"
	nodeVersion = 1
	commandLength = 12
)

var (
	nodeAddress string
	// hardcoded central node
	knownNodes      = []string{"localhost:3000"}
	miningAddress   string
	blocksInTransit [][]byte
	mempool = make(map[string]Transaction)
)

type Version struct {
	Version int
	BestHeight int
	AddrFrom string
}

type GetBlocks struct {
	AddrFrom string
}

type Inv struct {
	AddrFrom string
	Type string
	Items [][]byte
}

func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress); if err != nil {
		panic(err)
	}
	defer ln.Close()

	bc := NewBlockChain(nodeID)

	// if this address is not the central node
	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}

	for {
		conn, err := ln.Accept(); if err != nil {
			panic(err)
		}
		go handleConnection(conn, bc)
	}
}

func sendVersion(addr string, bc *Blockchain)  {
	bestHeight := bc.GetBestHeight()
	payload := GobEncode(Version{
		Version:    nodeVersion,
		BestHeight: bestHeight,
		AddrFrom:   nodeAddress,
	})

	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}

func handleInv(request []byte, bc *Blockchain) {
	var (
		buff bytes.Buffer
		payload Inv
	)

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload); if err != nil {
		fmt.Println("error decoding into payload for get blocks handler", err)
		return
	}

	fmt.Printf("Recieved inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		var newInTransit [][]byte

		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

func handleGetBlocks(request []byte, bc *Blockchain) {
	var (
		buff bytes.Buffer
		payload GetBlocks
	)

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload); if err != nil {
		fmt.Println("error decoding into payload for get blocks handler", err)
		return
	}

	blocks := bc.GetBlockHashes()
	sendInv(payload.AddrFrom, "block", blocks)
}

func handleVersion(request []byte, bc *Blockchain) {
	var (
		buff bytes.Buffer
		payload Version
	)

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload); if err != nil {
		fmt.Println("error decoding into payload for version handler", err)
		return
	}

	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if myBestHeight < foreignerBestHeight {
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.AddrFrom, bc)
	}

	if !nodeIsKnown(payload.AddrFrom) {
		knownNodes = append(knownNodes, payload.AddrFrom)
	}
}

func handleConnection(conn net.Conn, bc *Blockchain) {
	request, err := ioutil.ReadAll(conn); if err != nil {
		fmt.Println("error handling connection", err)
	}
	command := bytesToCommand(request[:commandLength])
	fmt.Printf("Recieved %s command\n", command)

	switch command {
	case "version":
		handleVersion(request, bc)
	default:
		fmt.Println("unknown command", command)
	}

	conn.Close()
}

func commandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}

func nodeIsKnown(addr string) bool {
	for _, node := range knownNodes {
		if node == addr {
			return true
		}
	}
	return false
}