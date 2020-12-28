package block

import (
	"bytes"
	"encoding/gob"
	"os"
	"strconv"
)

// IntToHex takes in an int64 and returns a byte slice of that int64 formatted to base16
func IntToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 16))
}

// DBExists checks whether bolt has a db created. If yes that means there is a blockchain already created
func DBExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

func GobEncode(data interface{}) []byte{
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(&data); if err != nil {
		panic(err)
	}
	return buff.Bytes()
}
