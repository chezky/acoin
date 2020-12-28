package block

import (
	"crypto/sha256"
)

// MerkleTree is an instance of a MerkleTree.
type MerkleTree struct {
	RootNode *MerkleNode // RootNode is the top most node of a MerkleTree.
}

// MerkleNode is an instance of a MerkleNode. It can contain two other MerkleNode(s), each one corresponding to the left and right of this MerkleNode.
// It also contains some data, which is a sha256 hash.
type MerkleNode struct {
	Left *MerkleNode
	Right *MerkleNode
	Data []byte
}

// NewMerkleNode takes in a left and right node, and some data. The data is used when creating the initial leaf nodes, and it's a serialized transaction.
// If it's a lead node, i.e: left and right are nil, then hash the data and store is as nMode.Data.
// If it's not a leaf node, concatenate the left and right, which are just the two nodes below this new node that actually form this node, and then hash them and
// then that's your new node's data.
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	nMode := MerkleNode{}

	// If there are no MerkleNodes on either side, just store a hash of the data. Otherwise...
	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		nMode.Data = hash[:]
	// ... append the left and right's data to form a single concatenated data source, and then stored that data after it's hashed.
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		nMode.Data = hash[:]
	}
	// keep the left and right MerkleNodes in this new MerkleNode
	nMode.Left = left
	nMode.Right = right
	return &nMode
}

// NewMerkleTree creates a new merkle tree. The data passed is a slice of serialized transactions. Before anything, if the amount of tx is odd, duplicate
// the last one. The merkle tree afterwards creates a leaf node for every transaction, and those leaf nodes contain a hash of the transaction as their data.
// Next, we run a loop for half the amount of transactions.
func NewMerkleTree(data [][]byte) *MerkleTree {
	// nodes is a slice of MerkleNode
	var nodes []MerkleNode

	// If there are an odd number of transactions, double the last transaction
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	// for each transaction in the slice of serialized transactions
	for _, datum := range data {
		// create a new node, with left and right being nil
		node := NewMerkleNode(nil, nil, datum)
		// add the new node to the slide of nodes
		nodes = append(nodes, *node)
	}

	// loop for as many times as half the amount of transactions, as each time it runs, the amount of nodes halve.
	// Example, there are 4 transactions, which forms 4 lead nodes. Two more nodes e and f, are created concatenating the hash of two nodes each. one new node is a
	// concatenation of node a+b, while the other is a concatenation of node c,d. Another node g is created by concatenating e,f. g is the root node of the merkle tree.
	// In this example, i runs twice, while j runs a total of 8 times. 4 times per loop of i. First loop finishes with 2 nodes, and second loop finishes with one node.
	// The goal is to have this for loop always end with nodes being of length 1.
	for i:=0; i < len(data)/2; i++ {
		var newLevel []MerkleNode

		// run an even number of times
		for j:=0; j < len(nodes); j+=2 {
			// Create a new node, with no data, but have it contain two merkle nodes as left and right nodes.
			// Since merkle trees are always an even number of nodes, there will always be a left and right node.
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			// That newly creates node, containing the hash of two concatenated nodes, it part of the next level.
			newLevel = append(newLevel, *node)
		}
		// Set nodes equal to the new level of nodes, and then rerun loop i.
		nodes = newLevel
	}

	return &MerkleTree{
		// nodes should only contain one value by the end of the function
		RootNode: &nodes[0],
	}
}