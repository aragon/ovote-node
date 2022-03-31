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

	chainID := uint64(3)
	processID := uint64(123)
	va, err := New(sqlite, chainID)
	c.Assert(err, qt.IsNil)

	// prepare the census
	keys := test.GenUserKeys(10)
	testCensus := test.GenCensus(c, keys)
	err = testCensus.Census.Close()
	c.Assert(err, qt.IsNil)

	censusRoot, err := testCensus.Census.Root()
	c.Assert(err, qt.IsNil)
	censusSize := uint64(len(keys.PublicKeys))
	votes := test.GenVotes(c, testCensus, chainID, processID)

	// store a process for the test
	ethBlockNum := uint64(10)
	ethEndBlockNum := uint64(20)
	resultsPublishingWindow := uint64(20)
	minParticipation := uint8(20)
	minPositiveVotes := uint8(60)
	err = sqlite.StoreProcess(processID, censusRoot, censusSize,
		ethBlockNum, ethEndBlockNum, resultsPublishingWindow, minParticipation,
		minPositiveVotes)
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
