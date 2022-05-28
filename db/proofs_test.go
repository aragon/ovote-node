package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestProof(t *testing.T) {
	c := qt.New(t)

	database, err := sql.Open("sqlite3", filepath.Join(c.TempDir(), "testdb.sqlite3"))
	c.Assert(err, qt.IsNil)

	sqlite := NewSQLite(database)

	err = sqlite.Migrate()
	c.Assert(err, qt.IsNil)

	// prepare the votes
	processID := uint64(123)
	censusRoot := []byte("censusRoot")
	censusSize := uint64(100)
	ethBlockNum := uint64(10)
	resPubStartBlock := uint64(20)
	resPubWindow := uint64(20)
	minParticipation := uint8(20)
	minPositiveVotes := uint8(60)
	typ := uint8(1)

	err = sqlite.StoreProcess(processID, censusRoot, censusSize,
		ethBlockNum, resPubStartBlock, resPubWindow, minParticipation,
		minPositiveVotes, typ)
	c.Assert(err, qt.IsNil)

	proofs, err := sqlite.GetProofsByProcessID(processID)
	c.Assert(err, qt.IsNil)
	c.Assert(len(proofs), qt.Equals, 0)

	// expect no error, despite the ProofID is not stored yet
	err = sqlite.AddProofToProofID(processID, 42, []byte("testproof"))
	c.Assert(err, qt.IsNil)

	proofs, err = sqlite.GetProofsByProcessID(processID)
	c.Assert(err, qt.IsNil)
	c.Assert(len(proofs), qt.Equals, 0)

	err = sqlite.StoreProofID(processID, 42)
	c.Assert(err, qt.IsNil)

	proofs, err = sqlite.GetProofsByProcessID(processID)
	c.Assert(err, qt.IsNil)
	c.Assert(len(proofs), qt.Equals, 1)
	c.Assert(proofs[0].Proof, qt.DeepEquals, []byte{})

	err = sqlite.AddProofToProofID(processID, 42, []byte("testproof"))
	c.Assert(err, qt.IsNil)

	proofs, err = sqlite.GetProofsByProcessID(processID)
	c.Assert(err, qt.IsNil)
	c.Assert(len(proofs), qt.Equals, 1)
	c.Assert(proofs[0].Proof, qt.DeepEquals, []byte("testproof"))
}
