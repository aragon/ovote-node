package main

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/aragon/zkmultisig-node/api"
	"github.com/aragon/zkmultisig-node/censusbuilder"
	"github.com/aragon/zkmultisig-node/db"
	"github.com/aragon/zkmultisig-node/votesaggregator"
	_ "github.com/mattn/go-sqlite3"
	flag "github.com/spf13/pflag"
	kvdb "go.vocdoni.io/dvote/db"
	"go.vocdoni.io/dvote/db/pebbledb"
	"go.vocdoni.io/dvote/log"
)

// Config contains the main configuration parameters of the node
type Config struct {
	dir, logLevel, port            string
	chainID                        uint64
	censusBuilder, votesAggregator bool
}

func main() {
	config := Config{}

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	flag.StringVarP(&config.dir, "dir", "d", filepath.Join(home, ".zkmultisig-node"),
		"storage data directory")
	flag.StringVarP(&config.logLevel, "logLevel", "l", "info", "log level (info, debug, warn, error)")
	flag.StringVarP(&config.port, "port", "p", "8080", "network port for the HTTP API")
	flag.Uint64Var(&config.chainID, "chainid", 42, "ChainID")
	flag.BoolVarP(&config.censusBuilder, "censusbuilder", "c", false, "CensusBuilder active")
	flag.BoolVarP(&config.votesAggregator, "votesaggregator", "v", false, "VotesAggregator active")
	// TODO add flag for configurable threshold of minimum census size (to prevent small censuses)
	// TODO add flag for eth-start-scanning-block (to prevent scanning since the genesis)

	flag.CommandLine.SortFlags = false
	flag.Parse()

	log.Init(config.logLevel, "stdout")

	log.Debugf("Config: %#v\n", config)

	var censusBuilder *censusbuilder.CensusBuilder
	var votesAggregator *votesaggregator.VotesAggregator
	if config.censusBuilder {
		opts := kvdb.Options{Path: filepath.Join(config.dir, "censusbuilder")}
		database, err := pebbledb.New(opts)
		if err != nil {
			log.Fatal(err)
		}

		censusBuilder, err = censusbuilder.New(database, filepath.Join(config.dir, "subsdb"))
		if err != nil {
			log.Fatal(err)
		}
	}
	if config.votesAggregator {
		sqlDB, err := sql.Open("sqlite3", filepath.Join(config.dir, "testdb.sqlite3"))
		if err != nil {
			log.Fatal(err)
		}
		sqlite := db.NewSQLite(sqlDB)
		votesAggregator, err = votesaggregator.New(sqlite, config.chainID)
		if err != nil {
			log.Fatal(err)
		}
	}

	a, err := api.New(censusBuilder, votesAggregator)
	if err != nil {
		log.Fatal(err)
	}
	err = a.Serve(config.port)
	if err != nil {
		log.Fatal(err)
	}
}
