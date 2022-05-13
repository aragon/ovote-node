package main

import (
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"go.vocdoni.io/dvote/db"
	"go.vocdoni.io/dvote/db/pebbledb"
)

func TestPostNewCensusHandler(t *testing.T) {
	t.Skip()
	// TODO WIP

	c := qt.New(t)

	opts := db.Options{Path: c.TempDir()}
	database, err := pebbledb.New(opts)
	c.Assert(err, qt.IsNil)

	a := api{
		db:     database,
		lastID: 0,
	}

	c.Assert(a.isBusy(), qt.IsFalse)
	go a.busyFor(3 * time.Second)
	time.Sleep(1 * time.Second)
	c.Assert(a.isBusy(), qt.IsTrue)
	time.Sleep(2 * time.Second)
}
