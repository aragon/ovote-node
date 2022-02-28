package censusbuilder

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"go.vocdoni.io/dvote/db"
	"go.vocdoni.io/dvote/db/pebbledb"
)

func newTestDB(c *qt.C) db.Database {
	var database db.Database
	var err error
	opts := db.Options{Path: c.TempDir()}
	database, err = pebbledb.New(opts)
	c.Assert(err, qt.IsNil)
	return database
}

func TestNewCensus(t *testing.T) {
	c := qt.New(t)

	// create the CensusBuilder
	database := newTestDB(c)
	cb, err := New(database, c.TempDir())
	c.Assert(err, qt.IsNil)

	censusID1, err := cb.NewCensus()
	c.Assert(err, qt.IsNil)
	err = cb.CloseCensus(censusID1)
	c.Assert(err, qt.IsNil)

	c.Assert(censusID1, qt.Equals, uint64(0))

	_, err = cb.CensusRoot(censusID1)
	c.Assert(err, qt.IsNil)

	censusID2, err := cb.NewCensus()
	c.Assert(err, qt.IsNil)
	c.Assert(censusID1, qt.Equals, uint64(0))

	_, err = cb.CensusRoot(censusID2)
	c.Assert(err.Error(), qt.Equals, "Can not get the CensusRoot, Census not closed yet")
	_, err = cb.censuses[censusID2].IntermediateRoot()
	c.Assert(err, qt.IsNil)

	err = cb.CloseCensus(censusID2)
	c.Assert(err, qt.IsNil)
}

func TestAddPublicKeys(t *testing.T) {
	c := qt.New(t)

	nKeys := 100
	// generate the publicKeys
	var pubKs []babyjub.PublicKey
	for i := 0; i < nKeys; i++ {
		sk := babyjub.NewRandPrivKey()
		pubK := sk.Public()
		pubKs = append(pubKs, *pubK)
	}

	// create the CensusBuilder
	database := newTestDB(c)
	cb, err := New(database, c.TempDir())
	c.Assert(err, qt.IsNil)

	censusID1, err := cb.NewCensus()
	c.Assert(err, qt.IsNil)
	err = cb.AddPublicKeys(censusID1, pubKs)
	c.Assert(err, qt.IsNil)
	err = cb.CloseCensus(censusID1)
	c.Assert(err, qt.IsNil)

	root1, err := cb.CensusRoot(censusID1)
	c.Assert(err, qt.IsNil)

	// create a 2nd Census, with the same pubKs than the 1st one
	censusID2, err := cb.NewCensus()
	c.Assert(err, qt.IsNil)
	err = cb.AddPublicKeys(censusID2, pubKs)
	c.Assert(err, qt.IsNil)

	_, err = cb.CensusRoot(censusID2)
	c.Assert(err.Error(), qt.Equals, "Can not get the CensusRoot, Census not closed yet")
	root2, err := cb.censuses[censusID2].IntermediateRoot()
	c.Assert(err, qt.IsNil)

	// check that both roots are equal
	c.Assert(root2, qt.DeepEquals, root1)

	// create new pubKs
	pubKs = []babyjub.PublicKey{}
	for i := 0; i < nKeys; i++ {
		sk := babyjub.NewRandPrivKey()
		pubK := sk.Public()
		pubKs = append(pubKs, *pubK)
	}
	err = cb.AddPublicKeys(censusID2, pubKs)
	c.Assert(err, qt.IsNil)

	err = cb.CloseCensus(censusID2)
	c.Assert(err, qt.IsNil)

	root2, err = cb.CensusRoot(censusID2)
	c.Assert(err, qt.IsNil)
	c.Assert(root2, qt.Not(qt.DeepEquals), root1)
}
