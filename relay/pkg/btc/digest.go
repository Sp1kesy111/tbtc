package btc

import "encoding/hex"

// Digest represents a Bitcoin digest.
type Digest struct {
	littleEndianBytes [32]byte
}

// NewLittleEndianDigest creates a digest using given bytes assuming they
// are represented in the little-endian system.
func NewLittleEndianDigest(littleEndianBytes [32]byte) *Digest {
	return &Digest{littleEndianBytes}
}

// NewBigEndianDigest creates a digest using given bytes assuming they
// are represented in the big-endian system.
func NewBigEndianDigest(bigEndianBytes [32]byte) *Digest {
	return &Digest{reverse(bigEndianBytes)}
}

// LittleEndianBytes returns digest bytes in little-endian system.
func (d *Digest) LittleEndianBytes() [32]byte {
	return d.littleEndianBytes
}

// BigEndianBytes returns digest bytes in big-endian system.
func (d *Digest) BigEndianBytes() [32]byte {
	return reverse(d.littleEndianBytes)
}

// String represents the digest as the hexadecimal string of
// little-endian bytes.
func (d *Digest) String() string {
	return hex.EncodeToString(d.littleEndianBytes[:])
}

func reverse(bytes [32]byte) [32]byte {
	for i := len(bytes)/2 - 1; i >= 0; i-- {
		opposite := len(bytes) - 1 - i
		bytes[i], bytes[opposite] = bytes[opposite], bytes[i]
	}

	return bytes
}
