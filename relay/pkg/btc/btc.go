package btc

import (
	"encoding/hex"
	"fmt"
	"math/big"
)

// Handle represents a handle to the Bitcoin chain.
type Handle interface {
	// GetHeaderByHeight returns the block header for the given block height.
	GetHeaderByHeight(height *big.Int) (*Header, error)

	// GetHeaderByDigest returns the block header for given digest (hash).
	GetHeaderByDigest(digest *Digest) (*Header, error)
}

// Header represents a Bitcoin block header.
type Header struct {
	// Hash is the hash of the block.
	Hash *Digest
	// Height is the height of the block in the blockchain.
	Height int64
	// PrevHash is the hash of the previous block.
	PrevHash *Digest
	// MerkleRoot is the hash of the root of the merkle tree of transactions in
	// the block.
	MerkleRoot *Digest
	// Raw is the serialized data of the block header; 80-bytes long.
	Raw []byte
}

func (h *Header) String() string {
	return fmt.Sprintf(
		"Hash: %s, Height: %d, PrevHash: %s, MerkleRoot: %s, Raw: %s",
		h.Hash,
		h.Height,
		h.PrevHash,
		h.MerkleRoot,
		hex.EncodeToString(h.Raw),
	)
}

// Config is a struct that contains the configuration needed to connect to a
// Bitcoin node.   This information will give access to a Bitcoin network.
type Config struct {
	URL      string
	Password string
	Username string
}
