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

func TestStoreProcess(t *testing.T) {
	c := qt.New(t)

	db, err := sql.Open("sqlite3", filepath.Join(c.TempDir(), "testdb.sqlite3"))
	c.Assert(err, qt.IsNil)

	sqlite := NewSQLite(db)

	err = sqlite.Migrate()
	c.Assert(err, qt.IsNil)

	// prepare the votes
	processID := uint64(123)
	censusRoot := []byte("censusRoot")
	ethBlockNum := uint64(10)

	err = sqlite.StoreProcess(processID, censusRoot, ethBlockNum)
	c.Assert(err, qt.IsNil)

	// try to store the same processID, expecting error
	err = sqlite.StoreProcess(processID, censusRoot, ethBlockNum)
	c.Assert(err, qt.Not(qt.IsNil))
	c.Assert(err.Error(), qt.Equals, "UNIQUE constraint failed: processes.processID")

	// try to store the a different processID, but the same censusRoot,
	// expecting no error
	err = sqlite.StoreProcess(processID+1, censusRoot, ethBlockNum)
	c.Assert(err, qt.IsNil)

	process, err := sqlite.ReadProcessByProcessID(processID)
	c.Assert(err, qt.IsNil)
	c.Assert(process.ID, qt.Equals, processID)
	c.Assert(process.CensusRoot, qt.DeepEquals, censusRoot)
	c.Assert(process.EthBlockNum, qt.Equals, ethBlockNum)

	// read the stored votes
	processes, err := sqlite.ReadProcesses()
	c.Assert(err, qt.IsNil)
	c.Assert(len(processes), qt.Equals, 2)
	c.Assert(processes[0].ID, qt.Equals, processID)
	c.Assert(processes[0].CensusRoot, qt.DeepEquals, censusRoot)
	c.Assert(processes[0].EthBlockNum, qt.Equals, ethBlockNum)
}

func TestStoreAndReadVotes(t *testing.T) {
	c := qt.New(t)

	db, err := sql.Open("sqlite3", filepath.Join(c.TempDir(), "testdb.sqlite3"))
	c.Assert(err, qt.IsNil)

	sqlite := NewSQLite(db)

	err = sqlite.Migrate()
	c.Assert(err, qt.IsNil)

	// store a processID in which the votes will be related
	processID := uint64(123)
	censusRoot := []byte("censusRoot")
	ethBlockNum := uint64(10)

	err = sqlite.StoreProcess(processID, censusRoot, ethBlockNum)
	c.Assert(err, qt.IsNil)

	// prepare the votes
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
				PublicKey:   pubK,
				MerkleProof: []byte("test" + strconv.Itoa(i)),
			},
			Vote: voteBytes,
		}
		votesAdded = append(votesAdded, vote)

		err = sqlite.StoreVotePackage(processID, vote)
		c.Assert(err, qt.IsNil)
	}

	// try to store a vote with already stored index
	err = sqlite.StoreVotePackage(processID, votesAdded[0])
	c.Assert(err, qt.Not(qt.IsNil))
	c.Assert(err.Error(), qt.Equals, "UNIQUE constraint failed: votepackages.indx")

	// read the stored votes
	votes, err := sqlite.ReadVotePackagesByProcessID(processID)
	c.Assert(err, qt.IsNil)
	c.Assert(len(votes), qt.Equals, nVotes)
}
