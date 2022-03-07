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

	// prepare the votes
	keys := test.GenUserKeys(10)
	testCensus := test.GenCensus(c, keys)
	err = testCensus.Census.Close()
	c.Assert(err, qt.IsNil)

	censusRoot, err := testCensus.Census.Root()
	c.Assert(err, qt.IsNil)
	votes := test.GenVotes(c, testCensus)

	for i := 0; i < len(votes); i++ {
		err = va.AddVote(censusRoot, votes[i])
		c.Assert(err, qt.IsNil)
	}

	// try to store a vote with already stored index
	err = va.AddVote(censusRoot, votes[0])
	c.Assert(err, qt.Not(qt.IsNil))
	c.Assert(err.Error(), qt.Equals, "UNIQUE constraint failed: votepackages.indx")

	// TODO try to store invalid votes/signatures/merkleproofs
}
