package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
	_ "github.com/mattn/go-sqlite3"
)

func TestMetaTable(t *testing.T) {
	c := qt.New(t)

	db, err := sql.Open("sqlite3", filepath.Join(c.TempDir(), "testdb.sqlite3"))
	c.Assert(err, qt.IsNil)

	sqlite := NewSQLite(db)

	err = sqlite.Migrate()
	c.Assert(err, qt.IsNil)

	err = sqlite.InitMeta(42, 0)
	c.Assert(err, qt.IsNil)

	b, err := sqlite.GetLastSyncBlockNum()
	c.Assert(err, qt.IsNil)
	c.Assert(b, qt.Equals, uint64(0))

	err = sqlite.UpdateLastSyncBlockNum(1234)
	c.Assert(err, qt.IsNil)

	b, err = sqlite.GetLastSyncBlockNum()
	c.Assert(err, qt.IsNil)
	c.Assert(b, qt.Equals, uint64(1234))
}
