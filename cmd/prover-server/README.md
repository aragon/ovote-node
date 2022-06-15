# prover-server

`prover-server` is a wrapper over [rapidsnark](https://github.com/iden3/rapidsnark), to provide an API REST to generate the proofs.

## Usage
Build the binary: `go build`

Usage:
```
Usage of prover-server:
  -d, --dir string    db & files directory (default "~/.proverserver")
  -p, --port string   network port for the HTTP API (default "9000")
```
