package test

import (
	"github.com/aragon/zkmultisig-node/census"
	"github.com/aragon/zkmultisig-node/types"
	qt "github.com/frankban/quicktest"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/vocdoni/arbo"
	"go.vocdoni.io/dvote/db"
	"go.vocdoni.io/dvote/db/pebbledb"
)

// Keys contains the test PrivateKeys and test PublicKeys
type Keys struct {
	PrivateKeys []babyjub.PrivateKey
	PublicKeys  []babyjub.PublicKey
}

// Census contains the test Keys and census.Census
type Census struct {
	Keys   Keys
	Census *census.Census
}

// GenUserKeys returns n Keys
func GenUserKeys(nUsers int) Keys {
	var keys Keys
	for i := 0; i < nUsers; i++ {
		sk := babyjub.NewRandPrivKey()
		keys.PrivateKeys = append(keys.PrivateKeys, sk)
		keys.PublicKeys = append(keys.PublicKeys, *sk.Public())
	}
	return keys
}

// GenCensus returns a test Census containing the given Keys
func GenCensus(c *qt.C, keys Keys) *Census {
	optsDB := db.Options{Path: c.TempDir()}
	database, err := pebbledb.New(optsDB)
	c.Assert(err, qt.IsNil)

	optsCensus := census.Options{DB: database}
	cens, err := census.New(optsCensus)
	c.Assert(err, qt.IsNil)

	invalids, err := cens.AddPublicKeys(keys.PublicKeys)
	c.Assert(err, qt.IsNil)
	c.Assert(len(invalids), qt.Equals, 0)
	return &Census{Keys: keys, Census: cens}
}

// GenVotes generate the votes from the given Census
func GenVotes(c *qt.C, cens *Census) []types.VotePackage {
	var votes []types.VotePackage
	for i := 0; i < len(cens.Keys.PrivateKeys); i++ {
		voteBytes := []byte("test")
		voteBI := arbo.BytesToBigInt(voteBytes)
		sigUncomp := cens.Keys.PrivateKeys[i].SignPoseidon(voteBI)
		sig := sigUncomp.Compress()

		// get merkleproof
		index, proof, err := cens.Census.GetProof(&cens.Keys.PublicKeys[i])
		c.Assert(err, qt.IsNil)

		vote := types.VotePackage{
			Signature: sig,
			CensusProof: types.CensusProof{
				Index:       index,
				PublicKey:   &cens.Keys.PublicKeys[i],
				MerkleProof: proof,
			},
			Vote: voteBytes,
		}
		votes = append(votes, vote)
	}
	return votes
}
