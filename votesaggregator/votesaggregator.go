package votesaggregator

import (
	"fmt"
	"time"

	"github.com/aragon/zkmultisig-node/db"
	"github.com/aragon/zkmultisig-node/types"
	"go.vocdoni.io/dvote/log"
)

const syncSleepTime = 6

// VotesAggregator receives the votes and aggregates them to generate a zkProof
type VotesAggregator struct {
	db *db.SQLite
}

// New returns a VotesAggregator with the given SQLite db
func New(sqlite *db.SQLite) (*VotesAggregator, error) {
	return &VotesAggregator{db: sqlite}, nil
}

// SyncProcesses actively checks if there are any processes closed, to trigger
// the generation of the zkInputs & zkProof of them. This method is designed to
// be called in a goroutine
func (va *VotesAggregator) SyncProcesses() {
	for {
		processes, err := va.db.ReadProcessesByStatus(types.ProcessStatusClosed)
		if err != nil {
			log.Error(err)
		}
		if len(processes) > 0 {
			process := processes[0]
			// WIP
			_ = process
			// generate zkInputs for the process
			// wait until prover is ready
			// send the zkInputs to the prover
			// zkProof, err := va.proverClient.ComputeProof(zkInputs)
		}

		time.Sleep(syncSleepTime * time.Second)
	}
}

// ProcessInfo returns info about the Process
func (va *VotesAggregator) ProcessInfo(processID uint64) (*types.Process, error) {
	// TODO add count of votes in the process
	return va.db.ReadProcessByID(processID)
}

// AddVote adds to the VotesAggregator's db the given vote for the given
// CensusRoot
func (va *VotesAggregator) AddVote(processID uint64, votePackage types.VotePackage) error {
	// get the process from the db. It's assumed that if the processID
	// exists in the db, it exists in the SmartContract
	process, err := va.db.ReadProcessByID(processID)
	if err != nil {
		return err
	}
	if process.Status != types.ProcessStatusOn {
		return fmt.Errorf("process EthEndBlockNum (%d) reached, votes"+
			" can not be added", process.EthEndBlockNum)
	}

	// check signature (babyjubjub) and MerkleProof
	if err := votePackage.Verify(process.CensusRoot); err != nil {
		return err
	}

	// store VotePackage in the SQL DB for the given CensusRoot
	return va.db.StoreVotePackage(processID, votePackage)
}
