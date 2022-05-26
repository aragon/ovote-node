package prover

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aragon/zkmultisig-node/types"
	qt "github.com/frankban/quicktest"
	"github.com/gin-gonic/gin"
)

func TestGenProof(t *testing.T) {
	c := qt.New(t)

	r := gin.Default()
	r.POST("/proof", mockGenProof)

	ts := httptest.NewServer(r)
	defer ts.Close()

	p := NewClient(ts.URL)
	zki := types.NewZKInputs(2, 2)
	pID, err := p.GenProof(zki)
	c.Assert(err, qt.IsNil)
	c.Assert(pID, qt.Equals, uint64(42))

	// now with handler that returns error
	r = gin.Default()
	r.POST("/proof", mockGenProofErr)
	ts = httptest.NewServer(r)
	defer ts.Close()

	p = NewClient(ts.URL)
	_, err = p.GenProof(zki)
	c.Assert(err, qt.Not(qt.IsNil))
	c.Assert(err.Error(), qt.Equals, "expected error msg")
}

func mockGenProof(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"id": 42,
	})
}

func mockGenProofErr(c *gin.Context) {
	c.JSON(http.StatusBadRequest, errorMsg{
		Message: "expected error msg",
	})
}
