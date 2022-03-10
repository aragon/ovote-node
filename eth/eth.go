package eth

// Client defines the interface that synchronizes with the Ethereum blockchain
// to obtain the processes data
type Client interface {
	// Sync scans the contract activity since the given fromBlock until the
	// current block, storing in the database all the updates on the
	// Processess
	Start(fromBlock uint64) error
}
