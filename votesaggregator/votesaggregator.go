package votesaggregator

import (
	"github.com/aragon/zkmultisig-node/db"
	"github.com/aragon/zkmultisig-node/types"
)

// VotesAggregator receives the votes and aggregates them to generate a zkProof
type VotesAggregator struct {
	db *db.SQLite
}

// New returns a VotesAggregator with the given SQLite db
func New(sqlite *db.SQLite) (*VotesAggregator, error) {
	return &VotesAggregator{db: sqlite}, nil
}

// AddVote adds to the VotesAggregator's db the given vote for the given
// CensusRoot
func (va *VotesAggregator) AddVote(processID uint64, votePackage types.VotePackage) error {
	// get the process from the db. It's assumed that if the processID
	// exists in the db, it exists in the SmartContract
	process, err := va.db.ReadProcessByProcessID(processID)
	if err != nil {
		return err
	}
	// check signature (babyjubjub) and MerkleProof
	if err := votePackage.Verify(process.CensusRoot); err != nil {
		return err
	}

	// store VotePackage in the SQL DB for the given CensusRoot
	return va.db.StoreVotePackage(processID, votePackage)
}
