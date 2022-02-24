package censusbuilder

import (
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/aragon/zkmultisig-node/census"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"go.vocdoni.io/dvote/db"
	"go.vocdoni.io/dvote/db/pebbledb"
)

// CensusBuilder manages multiple Census MerkleTrees
type CensusBuilder struct {
	subDBsPath string
	db         db.Database

	// censuses contains the loaded census
	censuses map[uint64]*census.Census
}

// New loads the CensusBuilder
func New(database db.Database, subDBsPath string) (*CensusBuilder, error) {
	cb := &CensusBuilder{
		subDBsPath: subDBsPath,
		db:         database,
		censuses:   make(map[uint64]*census.Census),
	}

	wTx := cb.db.WriteTx()
	defer wTx.Discard()

	// if nextIndex is not set in the db, initialize it to 0
	_, err := cb.getNextCensusID(wTx)
	if err != nil {
		err = cb.setNextCensusID(wTx, 0)
		if err != nil {
			return nil, err
		}
	}

	// commit the db.WriteTx
	if err := wTx.Commit(); err != nil {
		return nil, err
	}

	return cb, nil
}

var dbKeyNextCensusID = []byte("nextCensusID")

func (cb *CensusBuilder) setNextCensusID(wTx db.WriteTx, nextCensusID uint64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(nextCensusID))
	if err := wTx.Set(dbKeyNextCensusID, b); err != nil {
		return err
	}
	return nil
}

func (cb *CensusBuilder) getNextCensusID(rTx db.ReadTx) (uint64, error) {
	b, err := rTx.Get(dbKeyNextCensusID)
	if err != nil {
		return 0, err
	}
	nextCensusID := binary.LittleEndian.Uint64(b)
	return nextCensusID, nil
}

// loadCensusIfNotYet will load the Census in memory if it is not loaded yet
func (cb *CensusBuilder) loadCensusIfNotYet(censusID uint64) error {
	if _, ok := cb.censuses[censusID]; !ok {
		// census not loaded, load it
		optsDB := db.Options{Path: cb.subDBsPath + "/" + strconv.Itoa(int(censusID))}
		database, err := pebbledb.New(optsDB)
		if err != nil {
			return err
		}
		optsCensus := census.Options{DB: database}
		c, err := census.New(optsCensus)
		if err != nil {
			return err
		}
		cb.censuses[censusID] = c
	}
	return nil
}

// NewCensus will create a new Census, if the Census already exists, will load it
func (cb *CensusBuilder) NewCensus(pubKs []babyjub.PublicKey) (uint64, error) {
	rTx := cb.db.ReadTx()
	defer rTx.Discard()
	nextCensusID, err := cb.getNextCensusID(rTx)
	if err != nil {
		return 0, err
	}
	// return the nextCensusID, while in background the Census will be
	// created
	err = cb.loadCensusIfNotYet(nextCensusID)
	if err != nil {
		return 0, err
	}

	err = cb.AddPublicKeys(nextCensusID, pubKs)
	if err != nil {
		return 0, err
	}

	// store nextCensusID+1 in the CensusBuilder.db
	wTx := cb.db.WriteTx()
	defer wTx.Discard()
	err = cb.setNextCensusID(wTx, nextCensusID+1)
	if err != nil {
		return 0, err
	}
	if err := wTx.Commit(); err != nil {
		return 0, err
	}

	return nextCensusID, nil
}

// CloseCensus closes the Census of the given censusID.
func (cb *CensusBuilder) CloseCensus(censusID uint64) error {
	err := cb.loadCensusIfNotYet(censusID)
	if err != nil {
		return err
	}
	return cb.censuses[censusID].Close()
}

// CensusRoot returns the Root of the Census if the Census is closed.
func (cb *CensusBuilder) CensusRoot(censusID uint64) ([]byte, error) {
	err := cb.loadCensusIfNotYet(censusID)
	if err != nil {
		return nil, err
	}
	root, err := cb.censuses[censusID].Root()
	if err != nil {
		return nil, fmt.Errorf("Can not get the CensusRoot, %s", err)
	}
	return root, nil
}

// AddPublicKeys adds the batch of given PublicKeys to the Census for the given
// censusID.
func (cb *CensusBuilder) AddPublicKeys(censusID uint64, pubKs []babyjub.PublicKey) error {
	err := cb.loadCensusIfNotYet(censusID)
	if err != nil {
		return err
	}
	invalids, err := cb.censuses[censusID].AddPublicKeys(pubKs)
	if err != nil {
		return err
	}
	if len(invalids) != 0 {
		fmt.Println("TODO handle invalids")
	}
	return nil
}
