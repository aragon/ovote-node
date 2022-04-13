package eth

import (
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vocdoni/arbo"
)

const (
	// eventNewProcessLen defines the length of an event log of newProcess
	eventNewProcessLen = 288 // = 32*9
)

// ClientInterf defines the interface that synchronizes with the Ethereum
// blockchain to obtain the processes data
type ClientInterf interface {
	// Sync scans the contract activity since the given fromBlock until the
	// current block, storing in the database all the updates on the
	// Processess
	Start(fromBlock uint64) error
}

// eventNewProcess contains the data received from an event log of newProcess
type eventNewProcess struct {
	Creator          common.Address
	ProcessID        uint64
	TxHash           [32]byte
	CensusRoot       [32]byte
	CensusSize       uint64
	ResPubStartBlock uint64
	ResPubWindow     uint64
	MinParticipation uint8
	MinPositiveVotes uint8
}

func parseEventNewProcess(d []byte) (*eventNewProcess, error) {
	if len(d) != eventNewProcessLen {
		return nil, fmt.Errorf("newProcess event log should be of length %d, current: %d",
			eventNewProcessLen, len(d))
	}
	var e eventNewProcess

	// contract event:
	// event EventProcessCreated(address creator, uint256 id,uint256
	// transactionHash,  uint256 censusRoot, uint64 censusSize, uint64
	// resPubStartBlock, uint64 resPubWindow, uint8 minParticipation, uint8
	// minPositiveVotes);

	creatorBytes := d[:32]
	e.Creator = common.BytesToAddress(creatorBytes[12:32])

	// WARNING for the moment is uint256 but probably will change to uint64
	// idBytes := new(big.Int).SetBytes(idBytes)
	idBytes := d[64-8 : 64] // uint64
	e.ProcessID = binary.BigEndian.Uint64(idBytes)

	copy(e.TxHash[:], d[64:96])

	// note that here Ethereum returns the CensusRoot in big endian
	copy(e.CensusRoot[:], arbo.SwapEndianness(d[96:128]))

	censusSizeBytes := d[160-8 : 160] // uint64
	e.CensusSize = binary.BigEndian.Uint64(censusSizeBytes)

	resPubStartBlockBytes := d[192-8 : 192] // uint64
	e.ResPubStartBlock = binary.BigEndian.Uint64(resPubStartBlockBytes)

	resPubWindowBytes := d[224-8 : 224] // uint64
	e.ResPubWindow = binary.BigEndian.Uint64(resPubWindowBytes)

	e.MinParticipation = uint8(d[255])
	e.MinPositiveVotes = uint8(d[287])
	return &e, nil
}
