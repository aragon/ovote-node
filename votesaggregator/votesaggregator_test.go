package votesaggregator

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/aragon/zkmultisig-node/db"
	"github.com/aragon/zkmultisig-node/test"
	qt "github.com/frankban/quicktest"
	_ "github.com/mattn/go-sqlite3"
)

func TestStoreAndReadVotes(t *testing.T) {
	c := qt.New(t)

	sqlDB, err := sql.Open("sqlite3", filepath.Join(c.TempDir(), "testdb.sqlite3"))
	c.Assert(err, qt.IsNil)

	sqlite := db.NewSQLite(sqlDB)
	err = sqlite.Migrate()
	c.Assert(err, qt.IsNil)

	va, err := New(sqlite)
	c.Assert(err, qt.IsNil)

	// prepare the census
	keys := test.GenUserKeys(10)
	testCensus := test.GenCensus(c, keys)
	err = testCensus.Census.Close()
	c.Assert(err, qt.IsNil)

	censusRoot, err := testCensus.Census.Root()
	c.Assert(err, qt.IsNil)
	votes := test.GenVotes(c, testCensus)

	// store a process for the test
	processID := uint64(123)
	ethBlockNum := uint64(10)
	err = sqlite.StoreProcess(processID, censusRoot, ethBlockNum)
	c.Assert(err, qt.IsNil)

	for i := 0; i < len(votes); i++ {
		err = va.AddVote(processID, votes[i])
		c.Assert(err, qt.IsNil)
	}

	// try to store a vote with already stored index
	err = va.AddVote(processID, votes[0])
	c.Assert(err, qt.Not(qt.IsNil))
	c.Assert(err.Error(), qt.Equals, "UNIQUE constraint failed: votepackages.indx")

	// try to store invalid merkleproofs
	votes[0].CensusProof.Index = 11
	err = va.AddVote(processID, votes[0])
	c.Assert(err.Error(), qt.Equals, "merkleproof verification failed")

	// try to store invalid merkleproofs
	votes[0].Vote = []byte("invalidvotecontent")
	err = va.AddVote(processID, votes[0])
	c.Assert(err.Error(), qt.Equals, "signature verification failed")
}
