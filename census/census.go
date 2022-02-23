package census

import (
	"encoding/binary"

	"github.com/vocdoni/arbo"
	"go.vocdoni.io/dvote/db"
)

var maxLevels int = 32

// Census contains the MerkleTree with the PublicKeys
type Census struct {
	tree     *arbo.Tree
	db       db.Database
	Editable bool
}

// Options is used to pass the parameters to load a new Census
type Options struct {
	// DB defines the database that will be used for the census
	DB db.Database
}

// NewCensus loads the census
func NewCensus(opts Options) (*Census, error) {
	arboConfig := arbo.Config{
		Database:     opts.DB,
		MaxLevels:    maxLevels,
		HashFunction: arbo.HashFunctionPoseidon,
		// ThresholdNLeafs: not specified, use the default
	}

	wTx := opts.DB.WriteTx()
	defer wTx.Discard()

	tree, err := arbo.NewTreeWithTx(wTx, arboConfig)
	if err != nil {
		return nil, err
	}

	c := &Census{
		tree:     tree,
		Editable: true,
		db:       opts.DB,
	}

	// if lastUsedIndex is not set in the db, initialize it to 0
	_, err = c.getLastUsedIndex(wTx)
	if err != nil {
		err = c.setLastUsedIndex(wTx, 0)
		if err != nil {
			return nil, err
		}
	}

	// commit the db.WriteTx
	if err := wTx.Commit(); err != nil {
		return nil, err
	}

	return c, nil
}

var dbKeyLastUsedIndex = []byte("lastUsedIndex")

func (c *Census) setLastUsedIndex(wTx db.WriteTx, lastUsedIndex uint64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(lastUsedIndex))
	if err := wTx.Set(dbKeyLastUsedIndex, b); err != nil {
		return err
	}
	return nil
}

func (c *Census) getLastUsedIndex(rTx db.ReadTx) (uint64, error) {
	b, err := rTx.Get(dbKeyLastUsedIndex)
	if err != nil {
		return 0, err
	}
	lastUsedIndex := binary.LittleEndian.Uint64(b)
	return lastUsedIndex, nil
}
