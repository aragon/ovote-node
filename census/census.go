package census

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/poseidon"
	"github.com/vocdoni/arbo"
	"go.vocdoni.io/dvote/db"
)

var hashLen int = arbo.HashFunctionPoseidon.Len()
var maxLevels int = 64
var maxNLeafs uint64 = uint64(math.MaxUint64)
var maxKeyLen int = int(math.Ceil(float64(maxLevels) / float64(8))) //nolint:gomnd

// ErrCensusNotClosed is used when trying to do some action with the Census
// that needs the Census to be closed
var ErrCensusNotClosed = errors.New("Census not closed yet")

// ErrMaxNLeafsReached is used when trying to add a number of new publicKeys
// which would exceed the maximum number of keys in the census.
var ErrMaxNLeafsReached = fmt.Errorf("MaxNLeafs (%d) reached", maxNLeafs)

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

	// if nextIndex is not set in the db, initialize it to 0
	_, err = c.getNextIndex(wTx)
	if err != nil {
		err = c.setNextIndex(wTx, 0)
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

var dbKeyNextIndex = []byte("nextIndex")

func (c *Census) setNextIndex(wTx db.WriteTx, nextIndex uint64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(nextIndex))
	if err := wTx.Set(dbKeyNextIndex, b); err != nil {
		return err
	}
	return nil
}

func (c *Census) getNextIndex(rTx db.ReadTx) (uint64, error) {
	b, err := rTx.Get(dbKeyNextIndex)
	if err != nil {
		return 0, err
	}
	nextIndex := binary.LittleEndian.Uint64(b)
	return nextIndex, nil
}

// AddPublicKeys adds the batch of given PublicKeys, assigning incremental
// indexes to each one.
func (c *Census) AddPublicKeys(pubKs []babyjub.PublicKey) ([]arbo.Invalid, error) {
	wTx := c.db.WriteTx()
	defer wTx.Discard()

	nextIndex, err := c.getNextIndex(wTx)
	if err != nil {
		return nil, err
	}

	if nextIndex+uint64(len(pubKs)) > maxNLeafs {
		return nil, fmt.Errorf("%s, current index: %d, trying to add %d keys",
			ErrMaxNLeafsReached, nextIndex, len(pubKs))
	}
	var indexes [][]byte
	var pubKHashes [][]byte
	for i := 0; i < len(pubKs); i++ {
		// overflow in index should not be possible, as previously the
		// number of keys being added is already checked
		index := arbo.BigIntToBytes(maxKeyLen, big.NewInt(int64(int(nextIndex)+i))) //nolint:gomnd
		indexes = append(indexes[:], index)

		// store the mapping between PublicKey->Index
		pubKComp := pubKs[i].Compress()
		if err := wTx.Set(pubKComp[:], index); err != nil {
			return nil, err
		}

		pubKHash, err := poseidon.Hash([]*big.Int{pubKs[i].X, pubKs[i].Y})
		if err != nil {
			return nil, err
		}
		pubKHashes = append(pubKHashes, arbo.BigIntToBytes(hashLen, pubKHash))
	}

	invalids, err := c.tree.AddBatchWithTx(wTx, indexes, pubKHashes)
	if err != nil {
		return invalids, err
	}
	if len(invalids) != 0 {
		return invalids, fmt.Errorf("Can not add %d PublicKeys", len(invalids))
	}

	// TODO check overflow
	if err = c.setNextIndex(wTx, (nextIndex)+uint64(len(pubKs))); err != nil {
		return nil, err
	}

	// commit the db.WriteTx
	if err := wTx.Commit(); err != nil {
		return nil, err
	}

	return nil, nil
}
