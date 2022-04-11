package types

import (
	"fmt"
	"math/big"

	"github.com/vocdoni/arbo"
)

// ZKCircuitMeta contains metadata related to the circuit configuration
type ZKCircuitMeta struct {
	NMaxVotes int
	NLevels   int
}

// ZKInputs contains the inputs used to generate the zkProof
type ZKInputs struct {
	Meta ZKCircuitMeta `json:"-"`

	// public inputs
	ChainID    *big.Int `json:"chainID"`
	ProcessID  *big.Int `json:"processID"`
	CensusRoot *big.Int `json:"censusRoot"`
	ReciptRoot *big.Int `json:"reciptRoot"`
	Result     *big.Int `json:"result"`

	/////
	// private inputs
	Vote []*big.Int `json:"vote"`
	// user's key related
	Index []*big.Int `json:"index"`
	PkX   []*big.Int `json:"pkX"`
	PkY   []*big.Int `json:"pkY"`
	// signatures
	S   []*big.Int `json:"s"`
	R8x []*big.Int `json:"r8x"`
	R8y []*big.Int `json:"r8y"`
	// census proofs
	Siblings       [][]*big.Int `json:"siblings"`
	ReciptSiblings [][]*big.Int `json:"reciptSiblings"`
}

// NewZKInputs returns an initialized ZKInputs struct
func NewZKInputs(nMaxVotes, nLevels int) *ZKInputs {
	z := &ZKInputs{}
	z.Meta.NMaxVotes = nMaxVotes
	z.Meta.NLevels = nLevels

	z.ChainID = big.NewInt(0)
	z.ProcessID = big.NewInt(0)
	z.CensusRoot = big.NewInt(0)
	z.ReciptRoot = big.NewInt(0)
	z.Result = big.NewInt(0)

	z.Vote = emptyBISlice(nMaxVotes)
	z.Index = emptyBISlice(nMaxVotes)
	z.PkX = emptyBISlice(nMaxVotes)
	z.PkY = emptyBISlice(nMaxVotes)
	z.S = emptyBISlice(nMaxVotes)
	z.R8x = emptyBISlice(nMaxVotes)
	z.R8y = emptyBISlice(nMaxVotes)
	z.Siblings = make([][]*big.Int, nMaxVotes)
	for i := 0; i < nMaxVotes; i++ {
		z.Siblings[i] = emptyBISlice(nLevels)
	}
	z.ReciptSiblings = make([][]*big.Int, nMaxVotes)
	for i := 0; i < nMaxVotes; i++ {
		z.ReciptSiblings[i] = emptyBISlice(nLevels)
	}

	return z
}

// emptyBISlice returns an bigint zeroes slice, of length n
func emptyBISlice(n int) []*big.Int {
	s := make([]*big.Int, n)
	for i := 0; i < len(s); i++ {
		s[i] = big.NewInt(0)
	}
	return s
}

// MerkleProofToZKInputsFormat prepares the given MerkleProof into the
// ZKInputs.Siblings format for the circuit
func (z *ZKInputs) MerkleProofToZKInputsFormat(p []byte) ([]*big.Int, error) {
	s, err := arbo.UnpackSiblings(arbo.HashFunctionPoseidon, p)
	if err != nil {
		return nil, err
	}
	if len(s) > z.Meta.NLevels {
		return nil, fmt.Errorf("Max nLevels: %d, number of siblings: %d", z.Meta.NLevels, len(s))
	}

	b := make([]*big.Int, len(s))
	for i := 0; i < len(s); i++ {
		b[i] = arbo.BytesToBigInt(s[i])
	}
	for i := len(b); i < z.Meta.NLevels; i++ {
		b = append(b, big.NewInt(0))
	}

	return b, nil
}