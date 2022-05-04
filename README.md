# zkmultisig-node [![GoDoc](https://godoc.org/github.com/aragon/zkmultisig-node?status.svg)](https://godoc.org/github.com/aragon/zkmultisig-node) [![Go Report Card](https://goreportcard.com/badge/github.com/aragon/zkmultisig-node)](https://goreportcard.com/report/github.com/aragon/zkmultisig-node) [![Test](https://github.com/aragon/zkmultisig-node/workflows/Test/badge.svg)](https://github.com/aragon/zkmultisig-node/actions?query=workflow%3ATest)

*Research project.*

This repo contains the zkMultisig node implementation, compatible with the [zkmultisig](https://github.com/aragon/zkmultisig) components. All code is in early stages.

## Usage
In the `cmd` dir:
```
> go run cmd.go --help
Usage of zkmultisig-node:
  -d, --dir string        storage data directory (default "~/.zkmultisig-node")
  -l, --logLevel string   log level (info, debug, warn, error) (default "info")
  -p, --port string       network port for the HTTP API (default "8080")
  -c, --censusbuilder     CensusBuilder active
  -v, --votesaggregator   VotesAggregator active
      --eth string        web3 provider url
      --addr string       zkMultisig contract address
      --block uint        Start scanning block (usually the block where the zkMultisig contract was deployed)
```

So for example, running the node as a CensusBuilder and VotesAggregator for the ChainID=1 would be:
```
go run cmd.go -c -v --chainid=1 \
--eth=wss://yourweb3url.com --addr=0xTheZKMultisigContractAddress --block=6678912
```


## Test
- Tests: `go test ./...` (need [go](https://go.dev/) installed)
- Linters: `golangci-lint run --timeout=5m -c .golangci.yml` (need [golangci-lint](https://golangci-lint.run/) installed)
