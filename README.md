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
      --chainid uint      ChainID (default 42)
  -c, --censusbuilder     CensusBuilder active
  -v, --votesaggregator   VotesAggregator active
```

So for example, running the node as a CensusBuilder and VotesAggregator for the ChainID=1 would be:
```
go run cmd.go -c -v --chainid=1
```


## Test
- Tests: `go test ./...`
- Linters: `golangci-lint run --timeout=5m -c .golangci.yml`
