package main

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	flag "github.com/spf13/pflag"
	"go.vocdoni.io/dvote/db"
	"go.vocdoni.io/dvote/db/pebbledb"
	"go.vocdoni.io/dvote/log"
)

var port, dir string

type api struct {
	r *gin.Engine
	sync.Mutex

	lastID int
	db     db.Database
}

func main() {
	flag.StringVarP(&port, "port", "p", "9000", "network port for the HTTP API")
	flag.StringVarP(&dir, "dir", "d", "~/.proverserver", "db & files directory")
	flag.Parse()

	opts := db.Options{Path: dir}
	database, err := pebbledb.New(opts)
	if err != nil {
		log.Fatal(err)
	}

	a := api{}
	a.db = database
	a.r = gin.Default()
	a.lastID = 0

	a.r.GET("/status", a.getStatus)
	a.r.POST("/proof", a.genProof)
	a.r.GET("/proof/:id", a.getProof)

	err = a.r.Run(":" + port)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *api) getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (a *api) genProof(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "TODO",
	})
}

func (a *api) getProof(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "TODO",
	})
}
