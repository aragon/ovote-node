package eth

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/aragon/zkmultisig-node/db"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/vocdoni/arbo"
	"go.vocdoni.io/dvote/log"
)

const (
	// eventNewProcessLen defines the length of an event log of newProcess
	eventNewProcessLen = 288 // = 32*9
	// eventResultPublishedLen defines the length of an event log of
	// resultPublished
	eventResultPublishedLen = 160 // = 32*5
	// eventProcessClosedLen defines the length of an event log of
	// processClosed
	eventProcessClosedLen = 96 // = 32*3
)

// ClientInterf defines the interface that synchronizes with the Ethereum
// blockchain to obtain the processes data
type ClientInterf interface {
	// Sync scans the contract activity since the given fromBlock until the
	// current block, storing in the database all the updates on the
	// Processess
	Start(fromBlock uint64) error
}

// Client implements the ClientInterf that reads data from the Ethereum
// blockchain
type Client struct {
	client       *ethclient.Client
	db           *db.SQLite
	contractAddr common.Address
}

// Options is used to pass the parameters to load a new Client
type Options struct {
	EthURL       string
	SQLite       *db.SQLite
	ContractAddr common.Address
}

// New loads a new Client
func New(opts Options) (*Client, error) {
	client, err := ethclient.Dial(opts.EthURL)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &Client{
		client:       client,
		db:           opts.SQLite,
		contractAddr: opts.ContractAddr,
	}, nil
}

// Sync synchronizes from the configured zkmultisig contract events from the
// given block to the current block height.
func (c *Client) Sync(startBlock uint64) error {
	header, err := c.client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Error(err)
		return err
	}
	currBlockNum := header.Number
	log.Debugf("Sync blocknum: %d", currBlockNum)

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(startBlock)),
		ToBlock:   currBlockNum,
		Addresses: []common.Address{
			c.contractAddr,
		},
	}
	logs, err := c.client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Error(err)
		return err
	}
	for i := 0; i < len(logs); i++ {
		err = c.processEventLog(logs[i])
		if err != nil {
			log.Error(err)
		}
	}
	return nil
}

func (c *Client) processEventLog(eventLog types.Log) error {
	// depending on eventLog.Data length, parse the different types of
	// event logs
	switch l := len(eventLog.Data); l {
	case eventNewProcessLen:
		e, err := parseEventNewProcess(eventLog.Data)
		if err != nil {
			return fmt.Errorf("blocknum: %d, error parsing event log"+
				" (newProcess): %x, err: %s",
				eventLog.BlockNumber, eventLog.Data, err)
		}
		log.Debugf("blocknum: %d, newProcess event received, %v",
			eventLog.BlockNumber, e)
		// store the process in the db
		err = c.db.StoreProcess(e.ProcessID, e.CensusRoot[:], e.CensusSize,
			eventLog.BlockNumber, e.ResPubStartBlock, e.ResPubWindow,
			e.MinParticipation, e.MinPositiveVotes)
		if err != nil {
			return fmt.Errorf("error parsing event log (newProcess): %x, err: %s",
				eventLog.Data, err)
		}
	case eventResultPublishedLen:
		e, err := parseEventResultPublished(eventLog.Data)
		if err != nil {
			return fmt.Errorf("blocknum: %d, error parsing event log"+
				" (resultPublished): %x, err: %s",
				eventLog.BlockNumber, eventLog.Data, err)
		}
		log.Debugf("blocknum: %d, resultPublished event received, %v",
			eventLog.BlockNumber, e)
	case eventProcessClosedLen:
		e, err := parseEventProcessClosed(eventLog.Data)
		if err != nil {
			return fmt.Errorf("blocknum: %d, error parsing event log"+
				" (processClosed): %x, err: %s",
				eventLog.BlockNumber, eventLog.Data, err)
		}
		log.Debugf("blocknum: %d, processclosed event received, %v",
			eventLog.BlockNumber, e)
	default:
		return fmt.Errorf("unrecognized event log with length %d", l)
	}

	return nil
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

// String implements the String interface for eventNewProcess
func (e *eventNewProcess) String() string {
	return fmt.Sprintf("eventNewProcess: Creator %s, ProcessID: %d, TxHash: %s,"+
		" CensusRoot: %s, CensusSize: %d, ResPubStartBlock: %d, ResPubWindow: %d,"+
		" MinParticipation: %d, MinPositiveVotes: %d",
		e.Creator, e.ProcessID, hex.EncodeToString(e.TxHash[:]),
		arbo.BytesToBigInt(e.CensusRoot[:]), e.CensusSize,
		e.ResPubStartBlock, e.ResPubWindow, e.MinParticipation,
		e.MinPositiveVotes)
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

type eventResultPublished struct {
	Publisher    common.Address
	ProcessID    uint64
	ReceiptsRoot [32]byte
	Result       uint64
	NVotes       uint64
}

// String implements the String interface for eventResultPublished
func (e *eventResultPublished) String() string {
	return fmt.Sprintf("eventResultPublished: Publisher %s, ProcessID: %d,"+
		" ReceiptsRoot: %s, Result: %d, NVotes: %d",
		e.Publisher, e.ProcessID,
		arbo.BytesToBigInt(e.ReceiptsRoot[:]), e.Result, e.NVotes)
}

func parseEventResultPublished(d []byte) (*eventResultPublished, error) {
	if len(d) != eventResultPublishedLen {
		return nil, fmt.Errorf("resultPublished event log should be of"+
			" length %d, current: %d", eventResultPublishedLen, len(d))
	}

	// event EventResultPublished(address publisher, uint256 id, uint256
	// receiptsRoot, uint64 result, uint64 nVotes);

	var e eventResultPublished

	publisherBytes := d[:32]
	e.Publisher = common.BytesToAddress(publisherBytes[12:32])

	idBytes := d[64-8 : 64] // uint64
	e.ProcessID = binary.BigEndian.Uint64(idBytes)

	// note that here Ethereum returns the CensusRoot in big endian
	copy(e.ReceiptsRoot[:], arbo.SwapEndianness(d[64:96]))

	result := d[128-8 : 128] // uint64
	e.Result = binary.BigEndian.Uint64(result)

	nVotes := d[160-8 : 160] // uint64
	e.NVotes = binary.BigEndian.Uint64(nVotes)

	return &e, nil
}

type eventProcessClosed struct {
	Caller    common.Address
	ProcessID uint64
	Success   bool
}

// String implements the String interface for eventProcessClosed
func (e *eventProcessClosed) String() string {
	return fmt.Sprintf("eventProcessClosed: Caller %s, ProcessID: %d, Success: %t",
		e.Caller, e.ProcessID, e.Success)
}

// event EventProcessClosed(address caller, uint256 id, bool success);
func parseEventProcessClosed(d []byte) (*eventProcessClosed, error) {
	if len(d) != eventProcessClosedLen {
		return nil, fmt.Errorf("processClosed event log should be of length %d, current: %d",
			eventProcessClosedLen, len(d))
	}

	var e eventProcessClosed

	creatorBytes := d[:32]
	e.Caller = common.BytesToAddress(creatorBytes[12:32])

	idBytes := d[64-8 : 64] // uint64
	e.ProcessID = binary.BigEndian.Uint64(idBytes)

	success := d[96-1 : 96] // uint64
	if success[0] == byte(1) {
		e.Success = true
	}

	return &e, nil
}
