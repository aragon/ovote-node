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
func (va *VotesAggregator) AddVote(processID uint64, vote types.VotePackage) error {
	// TODO get the CensusRoot exists in the list of accepted CensusRoots
	// (comes from the SmartContract)
	// TODO check signature (babyjubjub)
	// TODO check MerkleProof for the CensusRoot

	// store VotePackage in the SQL DB for the given CensusRoot
	return va.db.StoreVotePackage(processID, vote)
}
