package api

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aragon/zkmultisig-node/censusbuilder"
	"github.com/aragon/zkmultisig-node/votesaggregator"
	"github.com/gin-gonic/gin"
	"go.vocdoni.io/dvote/log"
)

// API allows external requests to the Node
type API struct {
	r  *gin.Engine
	cb *censusbuilder.CensusBuilder
}

// New returns a new API with the endpoints, without starting to listen
func New(censusBuilder *censusbuilder.CensusBuilder,
	votesAggregator *votesaggregator.VotesAggregator) (*API, error) {
	if censusBuilder == nil && votesAggregator == nil {
		return nil, fmt.Errorf("Can not create the API. At least" +
			" censusBuilder or votesAggregator should be active to start" +
			" the API. Use --help to see the list of available flags.")
	}

	a := API{}

	r := gin.Default()

	if censusBuilder != nil {
		a.cb = censusBuilder

		// r.GET("/census", a.getCensus) // TODO
		r.POST("/census", a.postNewCensus)
		r.GET("/census/:censusid", a.getCensus)
		r.POST("/census/:censusid", a.postAddKeys)
		r.POST("/census/:censusid/close", a.postCloseCensus)
		r.GET("/census/:censusid/merkleproof/:pubkey", a.getMerkleProofHandler)
	}

	a.r = r

	return &a, nil
}

// Serve serves the API at the given port
func (a *API) Serve(port string) error {
	return a.r.Run(":" + port)
}

type errorMsg struct {
	Message string `json:"message"`
}

func returnErr(c *gin.Context, err error) {
	log.Warnw("HTTP API Bad request error", "err", err)
	c.JSON(http.StatusBadRequest, errorMsg{
		Message: err.Error(),
	})
}

func (a *API) postNewCensus(c *gin.Context) {
	var d newCensusReq
	err := c.ShouldBindJSON(&d)
	if err != nil {
		returnErr(c, err)
		return
	}

	censusID, err := a.cb.NewCensus()
	if err != nil {
		returnErr(c, err)
		return
	}

	go a.cb.AddPublicKeysAndStoreError(censusID, d.PublicKeys)

	c.JSON(http.StatusOK, censusID)
}

func (a *API) postAddKeys(c *gin.Context) {
	censusIDStr := c.Param("censusid")
	censusIDInt, err := strconv.Atoi(censusIDStr)
	if err != nil {
		returnErr(c, err)
		return
	}
	censusID := uint64(censusIDInt)

	var d newCensusReq
	err = c.ShouldBindJSON(&d)
	if err != nil {
		returnErr(c, err)
		return
	}

	go a.cb.AddPublicKeysAndStoreError(censusID, d.PublicKeys)

	c.JSON(http.StatusOK, censusID)
}

func (a *API) postCloseCensus(c *gin.Context) {
	censusIDStr := c.Param("censusid")
	censusIDInt, err := strconv.Atoi(censusIDStr)
	if err != nil {
		returnErr(c, err)
		return
	}
	censusID := uint64(censusIDInt)

	if err = a.cb.CloseCensus(censusID); err != nil {
		returnErr(c, err)
		return
	}
	root, err := a.cb.CensusRoot(censusID)
	if err != nil {
		returnErr(c, err)
		return
	}
	log.Debugf("[CensusID=%d] closed. Root: %x", censusID, root)
	c.JSON(http.StatusOK, hex.EncodeToString(root))
}

func (a *API) getCensus(c *gin.Context) {
	censusIDStr := c.Param("censusid")
	censusID, err := strconv.Atoi(censusIDStr)
	if err != nil {
		returnErr(c, err)
		return
	}
	censusInfo, err := a.cb.CensusInfo(uint64(censusID))
	if err != nil {
		returnErr(c, err)
		return
	}
	c.JSON(http.StatusOK, censusInfo)
}

func (a *API) getMerkleProofHandler(c *gin.Context) {
	// censusID := c.Param("censusID")
	// pubKey := c.Param("pubkey")

	// TODO check if census is closed
	// TODO get MerkleProof
}
