package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/aragon/zkmultisig-node/census"
	"github.com/aragon/zkmultisig-node/test"
	"github.com/aragon/zkmultisig-node/types"
	"github.com/dghubble/sling"
	qt "github.com/frankban/quicktest"
	"go.vocdoni.io/dvote/log"
)

var e2e = false

func init() {
	log.Init("debug", "stdout")
}

// TODO API Unit/E2E tests

func getAndPrintCensusInfo(c *qt.C, censusID uint64) {
	censusIDStr := strconv.Itoa(int(censusID))

	httpClient := &http.Client{}
	client := sling.New().Base("http://127.0.0.1:8080").Client(httpClient)
	req, err := client.New().Get("/census/" + censusIDStr).Request()
	c.Assert(err, qt.IsNil)
	res, err := httpClient.Do(req)
	c.Assert(err, qt.IsNil)
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)

	body, err := ioutil.ReadAll(res.Body)
	c.Assert(err, qt.IsNil)
	var ci census.Info
	err = json.Unmarshal(body, &ci)
	c.Assert(err, qt.IsNil)
	fmt.Printf("%#v\n", ci)
}

func TestPostNewCensus(t *testing.T) {
	if !e2e {
		t.Skip()
	}

	c := qt.New(t)

	nKeys := 100
	// generate the publicKeys
	log.Debugf("Generating %d PublicKeys", nKeys)
	keys := test.GenUserKeys(nKeys)
	reqData := newCensusReq{PublicKeys: keys.PublicKeys}

	httpClient := &http.Client{}
	client := sling.New().Base("http://127.0.0.1:8080").Client(httpClient)
	req, err := client.New().Post("/census").BodyJSON(reqData).Request()
	c.Assert(err, qt.IsNil)
	res, err := httpClient.Do(req)
	c.Assert(err, qt.IsNil)
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
}

func TestPostAddKeys(t *testing.T) {
	if !e2e {
		t.Skip()
	}

	c := qt.New(t)

	nKeys := 150
	// generate the publicKeys
	log.Debugf("Generating %d PublicKeys", nKeys)
	keys := test.GenUserKeys(nKeys)

	reqData := newCensusReq{PublicKeys: keys.PublicKeys[:100]}
	log.Debugf("%d PublicKeys created", nKeys)

	httpClient := &http.Client{}
	client := sling.New().Base("http://127.0.0.1:8080").Client(httpClient)
	req, err := client.New().Post("/census").BodyJSON(reqData).Request()
	c.Assert(err, qt.IsNil)
	res, err := httpClient.Do(req)
	c.Assert(err, qt.IsNil)
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)

	// get the censusID from the response
	body, err := ioutil.ReadAll(res.Body)
	c.Assert(err, qt.IsNil)
	var censusID uint64
	err = json.Unmarshal(body, &censusID)
	c.Assert(err, qt.IsNil)
	censusIDStr := strconv.Itoa(int(censusID))

	// Add the rest of the keys
	reqData = newCensusReq{PublicKeys: keys.PublicKeys[100:]}
	req, err = client.New().Post("/census/" + censusIDStr).BodyJSON(reqData).Request()
	c.Assert(err, qt.IsNil)
	res, err = httpClient.Do(req)
	c.Assert(err, qt.IsNil)
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
}

func TestPostCloseCensus(t *testing.T) {
	if !e2e {
		t.Skip()
	}

	c := qt.New(t)

	nKeys := 100
	// generate the publicKeys
	log.Debugf("Generating %d PublicKeys", nKeys)
	keys := test.GenUserKeys(nKeys)
	reqData := newCensusReq{PublicKeys: keys.PublicKeys}

	httpClient := &http.Client{}
	client := sling.New().Base("http://127.0.0.1:8080").Client(httpClient)
	req, err := client.New().Post("/census").BodyJSON(reqData).Request()
	c.Assert(err, qt.IsNil)
	res, err := httpClient.Do(req)
	c.Assert(err, qt.IsNil)
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
	defer res.Body.Close() //nolint:errcheck

	// get the censusID from the response
	body, err := ioutil.ReadAll(res.Body)
	c.Assert(err, qt.IsNil)
	var censusID uint64
	err = json.Unmarshal(body, &censusID)
	c.Assert(err, qt.IsNil)
	censusIDStr := strconv.Itoa(int(censusID))

	time.Sleep(1 * time.Second)

	// close Census
	req, err = client.New().Post("/census/" + censusIDStr + "/close").BodyJSON(nil).Request()
	c.Assert(err, qt.IsNil)
	res, err = httpClient.Do(req)
	c.Assert(err, qt.IsNil)
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
	defer res.Body.Close() //nolint:errcheck

	body, err = ioutil.ReadAll(res.Body)
	c.Assert(err, qt.IsNil)
	var rootHex string
	err = json.Unmarshal(body, &rootHex)
	c.Assert(err, qt.IsNil)
	fmt.Println(rootHex)

	getAndPrintCensusInfo(c, censusID)
}

func TestGetProof(t *testing.T) {
	if !e2e {
		t.Skip()
	}

	c := qt.New(t)

	nKeys := 100
	// generate the publicKeys
	log.Debugf("Generating %d PublicKeys", nKeys)
	keys := test.GenUserKeys(nKeys)
	reqData := newCensusReq{PublicKeys: keys.PublicKeys}

	httpClient := &http.Client{}
	client := sling.New().Base("http://127.0.0.1:8080").Client(httpClient)
	req, err := client.New().Post("/census").BodyJSON(reqData).Request()
	c.Assert(err, qt.IsNil)
	res, err := httpClient.Do(req)
	c.Assert(err, qt.IsNil)
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
	defer res.Body.Close() //nolint:errcheck

	body, err := ioutil.ReadAll(res.Body)
	c.Assert(err, qt.IsNil)
	var censusID uint64
	err = json.Unmarshal(body, &censusID)
	c.Assert(err, qt.IsNil)
	censusIDStr := strconv.Itoa(int(censusID))

	time.Sleep(1 * time.Second)

	getAndPrintCensusInfo(c, censusID)

	// close Census
	req, err = client.New().Post("/census/" + censusIDStr + "/close").BodyJSON(nil).Request()
	c.Assert(err, qt.IsNil)
	res, err = httpClient.Do(req)
	c.Assert(err, qt.IsNil)
	c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
	defer res.Body.Close() //nolint:errcheck

	body, err = ioutil.ReadAll(res.Body)
	c.Assert(err, qt.IsNil)
	var rootHex string
	err = json.Unmarshal(body, &rootHex)
	c.Assert(err, qt.IsNil)
	fmt.Println(rootHex)
	root, err := hex.DecodeString(rootHex)
	c.Assert(err, qt.IsNil)

	getAndPrintCensusInfo(c, censusID)

	for i := 0; i < nKeys; i++ {
		pubKiComp := keys.PublicKeys[i].Compress()
		pubKiHex := hex.EncodeToString(pubKiComp[:])
		req, err = client.New().Get("/census/" + censusIDStr + "/merkleproof/" + pubKiHex).Request()
		c.Assert(err, qt.IsNil)
		res, err = httpClient.Do(req)
		c.Assert(err, qt.IsNil)
		c.Assert(res.StatusCode, qt.Equals, http.StatusOK)
		defer res.Body.Close() //nolint:errcheck

		body, err = ioutil.ReadAll(res.Body)
		c.Assert(err, qt.IsNil)
		var cp types.CensusProof
		err = json.Unmarshal(body, &cp)
		c.Assert(err, qt.IsNil)
		fmt.Printf("Index: %d, MerkleProof: %x\n", cp.Index, cp.MerkleProof)

		v, err := census.CheckProof(root, cp.MerkleProof, cp.Index, &keys.PublicKeys[i])
		c.Assert(err, qt.IsNil)
		c.Assert(v, qt.IsTrue)
	}
}
