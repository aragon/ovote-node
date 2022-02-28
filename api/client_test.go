package api

import (
	"net/http"
	"testing"

	"github.com/dghubble/sling"
	qt "github.com/frankban/quicktest"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"go.vocdoni.io/dvote/log"
)

var e2e = false

func init() {
	log.Init("debug", "stdout")
}

// TODO API Unit/E2E tests

func TestGetCensus(t *testing.T) {
	if !e2e {
		t.Skip()
	}

	c := qt.New(t)

	httpClient := &http.Client{}
	client := sling.New().Base("http://127.0.0.1:8080").Client(httpClient)
	// client := sling.New().Base("http://127.0.0.1:8080")
	req, err := client.New().Get("/census/209").Request()
	c.Assert(err, qt.IsNil)
	_, err = httpClient.Do(req)
	c.Assert(err, qt.IsNil)
}

func TestPostNewCensus(t *testing.T) {
	if !e2e {
		t.Skip()
	}

	c := qt.New(t)

	nKeys := 100
	// generate the publicKeys
	log.Debugf("Generating %d PublicKeys", nKeys)
	var reqData newCensusReq
	for i := 0; i < nKeys; i++ {
		sk := babyjub.NewRandPrivKey()
		pubK := sk.Public()
		// pubKComp := sk.Public().Compress()
		// pubK := hex.EncodeToString(pubKComp[:])
		reqData.PublicKeys = append(reqData.PublicKeys, *pubK)
	}
	log.Debugf("%d PublicKeys created", nKeys)

	httpClient := &http.Client{}
	client := sling.New().Base("http://127.0.0.1:8080").Client(httpClient)
	req, err := client.New().Post("/census").BodyJSON(reqData).Request()
	c.Assert(err, qt.IsNil)
	_, err = httpClient.Do(req)
	c.Assert(err, qt.IsNil)
}
