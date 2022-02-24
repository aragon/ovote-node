package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// API allows external requests to the Node
type API struct {
	r *gin.Engine
}

// New returns a new API with the endpoints, without starting to listen
func New(censusBuilder, votesAggregator bool) (*API, error) {
	if !censusBuilder && !votesAggregator {
		return nil, fmt.Errorf("Can not create the API. At least" +
			" censusBuilder or votesAggregator should be active to start" +
			" the API. Use --help to see the list of available flags.")
	}

	a := API{}

	r := gin.Default()

	a.r = r

	return &a, nil
}

// Serve serves the API at the given port
func (a *API) Serve(port string) error {
	return a.r.Run(":" + port)
}
