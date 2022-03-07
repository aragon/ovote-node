package types

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/vocdoni/arbo"
	"go.vocdoni.io/dvote/db"
	"go.vocdoni.io/dvote/db/pebbledb"
)

func TestVerifyCensusProof(t *testing.T) {
	c := qt.New(t)

	optsDB := db.Options{Path: c.TempDir()}
	database, err := pebbledb.New(optsDB)
	c.Assert(err, qt.IsNil)

	arboConfig := arbo.Config{
		Database:     database,
		MaxLevels:    256,
		HashFunction: arbo.HashFunctionPoseidon,
		// ThresholdNLeafs: not specified, use the default
	}
	tree, err := arbo.NewTree(arboConfig)
	c.Assert(err, qt.IsNil)

	// create key
	sk := babyjub.NewRandPrivKey()
	pubK := sk.Public()

	index := uint64(1)
	indexBytes := Uint64ToIndex(index)
	value, err := HashPubKBytes(pubK)
	c.Assert(err, qt.IsNil)

	err = tree.Add(indexBytes, value)
	c.Assert(err, qt.IsNil)

	_, _, proof, _, err := tree.GenProof(indexBytes)
	c.Assert(err, qt.IsNil)

	vote := []byte("votetest")
	voteBI := arbo.BytesToBigInt(vote)
	sig := sk.SignPoseidon(voteBI)

	vp := VotePackage{
		Signature: sig.Compress(),
		CensusProof: CensusProof{
			Index:       index,
			PublicKey:   pubK,
			MerkleProof: proof,
		},
		Vote: vote,
	}

	root, err := tree.Root()
	c.Assert(err, qt.IsNil)

	c.Assert(vp.verifySignature(), qt.IsNil)
	c.Assert(vp.verifyMerkleProof(root), qt.IsNil)
	c.Assert(vp.Verify(root), qt.IsNil)

	vp.CensusProof.Index++
	c.Assert(vp.verifySignature(), qt.IsNil)
	c.Assert(vp.verifyMerkleProof(root), qt.Not(qt.IsNil))
	c.Assert(vp.Verify(root), qt.Not(qt.IsNil))
}
