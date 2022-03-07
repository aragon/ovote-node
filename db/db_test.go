package db

import (
	"database/sql"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/aragon/zkmultisig-node/types"
	qt "github.com/frankban/quicktest"
	"github.com/iden3/go-iden3-crypto/babyjub"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vocdoni/arbo"
)

func TestStoreAndReadVotes(t *testing.T) {
	c := qt.New(t)

	db, err := sql.Open("sqlite3", filepath.Join(c.TempDir(), "testdb.sqlite3"))
	c.Assert(err, qt.IsNil)

	sqlite := NewSQLite(db)

	err = sqlite.Migrate()
	c.Assert(err, qt.IsNil)

	// prepare the votes
	censusRoot := []byte("censusRoot")
	nVotes := 10

	var votesAdded []types.VotePackage
	for i := 0; i < nVotes; i++ {
		voteBytes := []byte("test")
		voteBI := arbo.BytesToBigInt(voteBytes)
		sk := babyjub.NewRandPrivKey()
		pubK := sk.Public()
		sigUncomp := sk.SignPoseidon(voteBI)
		sig := sigUncomp.Compress()
		vote := types.VotePackage{
			Signature: sig,
			CensusProof: types.CensusProof{
				Index:       uint64(i),
				PublicKey:   *pubK,
				MerkleProof: []byte("test" + strconv.Itoa(i)),
			},
			Vote: voteBytes,
		}
		votesAdded = append(votesAdded, vote)

		err = sqlite.StoreVotePackage(censusRoot, vote)
		c.Assert(err, qt.IsNil)
	}

	// try to store a vote with already stored index
	err = sqlite.StoreVotePackage(censusRoot, votesAdded[0])
	c.Assert(err, qt.Not(qt.IsNil))
	c.Assert(err.Error(), qt.Equals, "UNIQUE constraint failed: votepackages.indx")

	// read the stored votes
	votes, err := sqlite.ReadVotePackagesByCensusRoot(censusRoot)
	c.Assert(err, qt.IsNil)
	c.Assert(len(votes), qt.Equals, nVotes)
}
