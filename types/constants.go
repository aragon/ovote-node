package types

import (
	"math"

	"github.com/vocdoni/arbo"
)

var (
	// MaxLevels indicates the maximum number of levels in the Census
	// MerkleTree
	MaxLevels int = 64
	// MaxNLeafs indicates the maximum number of leaves in the Census
	// MerkleTree
	MaxNLeafs uint64 = uint64(math.MaxUint64)
	// MaxKeyLen indicates the maximum key (index) length in the Census
	// MerkleTree
	MaxKeyLen int = int(math.Ceil(float64(MaxLevels) / float64(8))) //nolint:gomnd
	// EmptyRoot is a byte array of 0s, with the length of the hash
	// function output length used in the Census MerkleTree
	EmptyRoot = make([]byte, arbo.HashFunctionPoseidon.Len())
)
