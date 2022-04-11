package types

import (
	"fmt"
	"math/big"
	"os"

	"github.com/vocdoni/arbo"
	kvdb "go.vocdoni.io/dvote/db"
	"go.vocdoni.io/dvote/db/pebbledb"
	"go.vocdoni.io/dvote/log"
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
	ChainID      *big.Int `json:"chainID"`
	ProcessID    *big.Int `json:"processID"`
	CensusRoot   *big.Int `json:"censusRoot"`
	ReceiptsRoot *big.Int `json:"receiptsRoot"`
	Result       *big.Int `json:"result"`

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
	Siblings         [][]*big.Int `json:"siblings"`
	ReceiptsSiblings [][]*big.Int `json:"receiptsSiblings"`
}

// NewZKInputs returns an initialized ZKInputs struct
func NewZKInputs(nMaxVotes, nLevels int) *ZKInputs {
	z := &ZKInputs{}
	z.Meta.NMaxVotes = nMaxVotes
	z.Meta.NLevels = nLevels

	z.ChainID = big.NewInt(0)
	z.ProcessID = big.NewInt(0)
	z.CensusRoot = big.NewInt(0)
	z.ReceiptsRoot = big.NewInt(0)
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
	z.ReceiptsSiblings = make([][]*big.Int, nMaxVotes)
	for i := 0; i < nMaxVotes; i++ {
		z.ReceiptsSiblings[i] = emptyBISlice(nLevels)
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

// ComputeReceipts builds a temp MerkleTree with all the given index &
// publicKeys (receiptsKeys & receiptsValues), to then compute the siblings of
// each recipt, adding the siblings & root of the receipts tree to
// ZKInputs.ReceiptsRoot & ZKInputs.ReceiptsSiblings.
func (z *ZKInputs) ComputeReceipts(receiptsKeys, receiptsValues [][]byte) error {
	// prepare receiptsTree
	dbPebble, err := pebbledb.New(kvdb.Options{Path: os.TempDir()})
	if err != nil {
		return err
	}
	receiptsTreeConfig := arbo.Config{
		Database:     dbPebble,
		MaxLevels:    z.Meta.NLevels,
		HashFunction: arbo.HashFunctionPoseidon,
	}
	receiptsTree, err := arbo.NewTree(receiptsTreeConfig)
	if err != nil {
		return err
	}

	// build the receiptsTree
	invalids, err := receiptsTree.AddBatch(receiptsKeys, receiptsValues)
	if err != nil {
		return err
	}
	if len(invalids) != 0 {
		return fmt.Errorf("Can not add %d PublicKeys to the receiptsTree", len(invalids))
	}

	// get the z.ReceiptsRoot
	receiptsRoot, err := receiptsTree.Root()
	if err != nil {
		return err
	}
	z.ReceiptsRoot = arbo.BytesToBigInt(receiptsRoot)

	// compute the z.ReceiptsSiblings
	for i := 0; i < len(receiptsKeys); i++ {
		_, _, receiptSiblings, existence, err := receiptsTree.GenProof(receiptsKeys[i])
		if err != nil {
			return err
		}
		if !existence {
			log.Error("should not happen")
			return fmt.Errorf("publicKey does not exist in the receiptsTree (%x)", receiptsValues[:])
		}
		z.ReceiptsSiblings[i], err = z.MerkleProofToZKInputsFormat(receiptSiblings)
		if err != nil {
			return err
		}
	}

	return nil
}
