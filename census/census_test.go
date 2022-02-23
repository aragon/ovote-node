package census

import (
	"math"
	"testing"

	qt "github.com/frankban/quicktest"
	"go.vocdoni.io/dvote/db"
	"go.vocdoni.io/dvote/db/pebbledb"
)

// NOTE: most of the methods of Census are wrappers over
// https://github.com/vocdoni/arbo.  The proper tests are in arbo's repo, here
// there are tests that check the specific code of the Census package.

func newTestDB(c *qt.C) db.Database {
	var database db.Database
	var err error
	opts := db.Options{Path: c.TempDir()}
	database, err = pebbledb.New(opts)
	c.Assert(err, qt.IsNil)
	return database
}

func newTestCensus(c *qt.C) *Census {
	database := newTestDB(c)
	opts := Options{database}
	census, err := NewCensus(opts)
	c.Assert(err, qt.IsNil)
	return census
}

func TestLastUsedIndex(t *testing.T) {
	c := qt.New(t)
	census := newTestCensus(c)

	// expect lastUsedIndex to be 0
	rTx := census.db.ReadTx()
	i, err := census.getLastUsedIndex(rTx)
	c.Assert(err, qt.IsNil)
	c.Assert(i, qt.Equals, uint64(0))
	rTx.Discard()

	// set the lastUsedIndex to 10
	wTx := census.db.WriteTx()
	err = census.setLastUsedIndex(wTx, 10)
	c.Assert(err, qt.IsNil)
	err = wTx.Commit()
	c.Assert(err, qt.IsNil)
	wTx.Discard()

	// expect lastUsedIndex to be 10
	rTx = census.db.ReadTx()
	i, err = census.getLastUsedIndex(rTx)
	c.Assert(err, qt.IsNil)
	c.Assert(i, qt.Equals, uint64(10))

	maxUint64 := uint64(math.MaxUint64)
	// set the lastUsedIndex to maxUint64
	wTx = census.db.WriteTx()
	err = census.setLastUsedIndex(wTx, maxUint64)
	c.Assert(err, qt.IsNil)
	err = wTx.Commit()
	c.Assert(err, qt.IsNil)
	wTx.Discard()

	// expect lastUsedIndex to be maxUint64
	rTx = census.db.ReadTx()
	i, err = census.getLastUsedIndex(rTx)
	c.Assert(err, qt.IsNil)
	c.Assert(i, qt.Equals, maxUint64)
}
