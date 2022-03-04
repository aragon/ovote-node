package votesaggregator

import "github.com/aragon/zkmultisig-node/types"

// VotesAggregator receives the votes and aggregates them to generate a zkProof
type VotesAggregator struct {
}

// AddVote adds to the VotesAggregator's db the given vote for the given
// CensusID
func (va *VotesAggregator) AddVote(censusID uint64, vote types.VotePackage) error {
	// TODO get the CensusRoot for the given CensusID (including checking
	// if the CensusID exists)
	// TODO check signature (babyjubjub)
	// TODO check MerkleProof for the CensusID root
	// TODO store VotePackage in the SQL DB for the given CensusID
	return nil
}
