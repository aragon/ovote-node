package votesaggregator

import (
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/aragon/zkmultisig-node/db"
	"github.com/aragon/zkmultisig-node/types"
	"github.com/vocdoni/arbo"
	"go.vocdoni.io/dvote/log"
)

const syncSleepTime = 6

// VotesAggregator receives the votes and aggregates them to generate a zkProof
type VotesAggregator struct {
	db      *db.SQLite
	chainID uint64 // determined by config
}

// New returns a VotesAggregator with the given SQLite db
func New(sqlite *db.SQLite, chainID uint64) (*VotesAggregator, error) {
	return &VotesAggregator{db: sqlite, chainID: chainID}, nil
}

// SyncProcesses actively checks if there are any processes closed, to trigger
// the generation of the zkInputs & zkProof of them. This method is designed to
// be called in a goroutine
func (va *VotesAggregator) SyncProcesses() {
	for {
		// if there are Frozen processes, generate their zkProofs
		processes, err := va.db.ReadProcessesByStatus(types.ProcessStatusFrozen)
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
			// update ProcessStatus to types.ProcessStatusFinished
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
	// for this initial version, only vote values with 0 or 1 are supported
	// TODO check vote value inside range

	// get the process from the db. It's assumed that if the processID
	// exists in the db, it exists in the SmartContract
	process, err := va.db.ReadProcessByID(processID)
	if err != nil {
		return err
	}
	if process.Status != types.ProcessStatusOn {
		return fmt.Errorf("process ResPubStartBlock (%d) reached,"+
			" votes can not be added", process.ResPubStartBlock)
	}

	// check signature (babyjubjub) and MerkleProof
	if err := votePackage.Verify(va.chainID, processID, process.CensusRoot); err != nil {
		return err
	}

	// store VotePackage in the SQL DB for the given CensusRoot
	return va.db.StoreVotePackage(processID, votePackage)
}

// GenerateZKInputs will generate the zkInputs for the given processID
func (va *VotesAggregator) GenerateZKInputs(processID uint64, nMaxVotes,
	nLevels /* tmp */ int) (*types.ZKInputs, error) {
	// TODO TMP, nMaxVotes & nLevels will be defined by the compiled circuits
	z := types.NewZKInputs(nMaxVotes, nLevels)

	z.ChainID = big.NewInt(int64(va.chainID))
	z.ProcessID = big.NewInt(int64(processID))
	process, err := va.db.ReadProcessByID(processID)
	if err != nil {
		return nil, err
	}
	z.CensusRoot = arbo.BytesToBigInt(process.CensusRoot)

	var receiptsKeys [][]byte
	var receiptsValues [][]byte

	// get db votes for the processID. It's assumed that the returned
	// votePackages are sorted by index
	votes, err := va.db.ReadVotePackagesByProcessID(processID)
	if err != nil {
		return nil, err
	}
	r := big.NewInt(0)
	for i := 0; i < len(votes); i++ {
		voteBI := arbo.BytesToBigInt(votes[i].Vote)
		if voteBI.Cmp(big.NewInt(1)) == 1 { // voteBI > 1:
			return nil, fmt.Errorf("invalid vote value") // TODO better error handling
		}
		r = new(big.Int).Add(r, new(big.Int).Mul(voteBI, votes[i].CensusProof.Weight))
		// TODO ensure that Weight does not overflow the field
		z.Vote[i] = voteBI
		z.Index[i] = big.NewInt(int64(votes[i].CensusProof.Index))

		z.PkX[i] = votes[i].CensusProof.PublicKey.X
		z.PkY[i] = votes[i].CensusProof.PublicKey.Y
		z.Weight[i] = votes[i].CensusProof.Weight
		sig, err := votes[i].Signature.Decompress()
		if err != nil {
			// TODO, probably instead of stopping the process, skip
			// that vote due wrong signature (having in mind, that
			// if the signature was wrong, should not be allowed to
			// be stored in the db
			return nil, err
		}
		z.S[i] = sig.S
		z.R8x[i] = sig.R8.X
		z.R8y[i] = sig.R8.Y
		z.Siblings[i], err = z.MerkleProofToZKInputsFormat(votes[i].CensusProof.MerkleProof)
		if err != nil {
			return nil, err
		}

		// prepare the receipt data with the index & pubK
		key := types.Uint64ToIndex(votes[i].CensusProof.Index)
		key = key[:int(math.Ceil(float64(nLevels)/8))] //nolint:gomnd
		receiptsKeys = append(receiptsKeys, key)
		pubKHashBytes, err := types.HashPubKBytes(
			votes[i].CensusProof.PublicKey,
			votes[i].CensusProof.Weight)
		if err != nil {
			return nil, err
		}
		receiptsValues = append(receiptsValues, pubKHashBytes[:])
	}
	z.Result = r
	z.NVotes = big.NewInt(int64(len(votes)))
	z.WithReceipts = big.NewInt(1)

	// compute the z.ReceiptsRoot & zk.ReceiptsSiblings
	err = z.ComputeReceipts(processID, receiptsKeys, receiptsValues)
	if err != nil {
		return nil, err
	}

	return z, nil
}
