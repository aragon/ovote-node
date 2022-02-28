package main

import (
	"os"
	"path/filepath"

	"github.com/aragon/zkmultisig-node/api"
	"github.com/aragon/zkmultisig-node/censusbuilder"
	"github.com/aragon/zkmultisig-node/votesaggregator"
	flag "github.com/spf13/pflag"
	"go.vocdoni.io/dvote/db"
	"go.vocdoni.io/dvote/db/pebbledb"
	"go.vocdoni.io/dvote/log"
)

// Config contains the main configuration parameters of the node
type Config struct {
	dir, logLevel, port            string
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
	flag.BoolVarP(&config.censusBuilder, "censusbuilder", "c", false, "CensusBuilder active")
	flag.BoolVarP(&config.votesAggregator, "votesaggregator", "v", false, "VotesAggregator active")

	flag.CommandLine.SortFlags = false
	flag.Parse()

	log.Init(config.logLevel, "stdout")

	log.Debugf("Config: %#v\n", config)

	var censusBuilder *censusbuilder.CensusBuilder
	var votesAggregator *votesaggregator.VotesAggregator
	if config.censusBuilder {
		opts := db.Options{Path: filepath.Join(config.dir, "censusbuilder")}
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
		log.Fatal("VotesAggregator not available yet")
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
