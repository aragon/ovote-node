package api

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/aragon/zkmultisig-node/census"
	"github.com/aragon/zkmultisig-node/censusbuilder"
	"github.com/aragon/zkmultisig-node/db"
	"github.com/aragon/zkmultisig-node/test"
	"github.com/aragon/zkmultisig-node/types"
	"github.com/aragon/zkmultisig-node/votesaggregator"
	qt "github.com/frankban/quicktest"
	"github.com/gin-gonic/gin"
	"github.com/iden3/go-iden3-crypto/babyjub"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vocdoni/arbo"
	kvdb "go.vocdoni.io/dvote/db"
	"go.vocdoni.io/dvote/db/pebbledb"
	"go.vocdoni.io/dvote/log"
)

func newTestAPI(c *qt.C) (API, *db.SQLite) {
	// create the CensusBuilder
	opts := kvdb.Options{Path: c.TempDir()}
	database, err := pebbledb.New(opts)
	c.Assert(err, qt.IsNil)
	cb, err := censusbuilder.New(database, c.TempDir())
	c.Assert(err, qt.IsNil)
	r := gin.Default()

	// create the VotesAggregator
	sqlDB, err := sql.Open("sqlite3", filepath.Join(c.TempDir(), "testdb.sqlite3"))
	c.Assert(err, qt.IsNil)

	sqlite := db.NewSQLite(sqlDB)
	err = sqlite.Migrate()
	c.Assert(err, qt.IsNil)

	va, err := votesaggregator.New(sqlite)
	c.Assert(err, qt.IsNil)

	return API{r: r, cb: cb, va: va}, sqlite
}

func doPostNewCensus(c *qt.C, a API, pubKs []babyjub.PublicKey) uint64 {
	reqData := newCensusReq{PublicKeys: pubKs}
	jsonReqData, err := json.Marshal(reqData)
	c.Assert(err, qt.IsNil)

	req, err := http.NewRequest("POST", "/census", bytes.NewBuffer(jsonReqData))
	c.Assert(err, qt.IsNil)
	w := httptest.NewRecorder()
	a.r.ServeHTTP(w, req)
	c.Assert(w.Code, qt.Equals, http.StatusOK)

	// get the censusID from the response
	body, err := ioutil.ReadAll(w.Body)
	c.Assert(err, qt.IsNil)
	var censusID uint64
	err = json.Unmarshal(body, &censusID)
	c.Assert(err, qt.IsNil)
	return censusID
}

func doPostAddKeys(c *qt.C, a API, censusID uint64, pubKs []babyjub.PublicKey) {
	censusIDStr := strconv.Itoa(int(censusID))
	reqData := newCensusReq{PublicKeys: pubKs}
	jsonReqData, err := json.Marshal(reqData)
	c.Assert(err, qt.IsNil)
	req, err := http.NewRequest("POST", "/census/"+censusIDStr, bytes.NewBuffer(jsonReqData))
	c.Assert(err, qt.IsNil)
	w := httptest.NewRecorder()
	a.r.ServeHTTP(w, req)
	c.Assert(w.Code, qt.Equals, http.StatusOK)
}

func doPostCloseCensus(c *qt.C, a API, censusID uint64) []byte {
	censusIDStr := strconv.Itoa(int(censusID))
	req, err := http.NewRequest("POST", "/census/"+censusIDStr+"/close", nil)
	c.Assert(err, qt.IsNil)
	w := httptest.NewRecorder()
	a.r.ServeHTTP(w, req)
	c.Assert(w.Code, qt.Equals, http.StatusOK)

	body, err := ioutil.ReadAll(w.Body)
	c.Assert(err, qt.IsNil)
	var rootHex string
	err = json.Unmarshal(body, &rootHex)
	c.Assert(err, qt.IsNil)
	root, err := hex.DecodeString(rootHex)
	c.Assert(err, qt.IsNil)
	return root
}

func doGetProof(c *qt.C, a API, censusID uint64, pubK babyjub.PublicKey) types.CensusProof {
	censusIDStr := strconv.Itoa(int(censusID))
	pubKComp := pubK.Compress()
	pubKHex := hex.EncodeToString(pubKComp[:])

	req, err := http.NewRequest("GET", "/census/"+censusIDStr+"/merkleproof/"+pubKHex, nil)
	c.Assert(err, qt.IsNil)
	w := httptest.NewRecorder()
	a.r.ServeHTTP(w, req)
	c.Assert(w.Code, qt.Equals, http.StatusOK)

	body, err := ioutil.ReadAll(w.Body)
	c.Assert(err, qt.IsNil)
	var cp types.CensusProof
	err = json.Unmarshal(body, &cp)
	c.Assert(err, qt.IsNil)
	return cp
}

func doPostVote(c *qt.C, a API, processID uint64, vote types.VotePackage) {
	processIDStr := strconv.Itoa(int(processID))
	jsonReqData, err := json.Marshal(vote)
	c.Assert(err, qt.IsNil)
	fmt.Println("JSON", processIDStr, string(jsonReqData))
	req, err := http.NewRequest("POST", "/process/"+processIDStr, bytes.NewBuffer(jsonReqData))
	c.Assert(err, qt.IsNil)
	w := httptest.NewRecorder()
	a.r.ServeHTTP(w, req)
	c.Assert(w.Code, qt.Equals, http.StatusOK)
}

func prepareVotes(c *qt.C, keys test.Keys, proofs []types.CensusProof) []types.VotePackage {
	var votes []types.VotePackage
	for i := 0; i < len(keys.PrivateKeys); i++ {
		voteBytes := []byte("test")
		voteBI := arbo.BytesToBigInt(voteBytes)
		sigUncomp := keys.PrivateKeys[i].SignPoseidon(voteBI)
		sig := sigUncomp.Compress()

		vote := types.VotePackage{
			Signature: sig,
			CensusProof: types.CensusProof{
				Index:       proofs[i].Index,
				PublicKey:   &keys.PublicKeys[i],
				MerkleProof: proofs[i].MerkleProof,
			},
			Vote: voteBytes,
		}
		votes = append(votes, vote)
	}
	return votes
}

func TestPostNewCensusHandler(t *testing.T) {
	c := qt.New(t)

	a, _ := newTestAPI(c)
	a.r.POST("/census", a.postNewCensus)

	nKeys := 100
	// generate the publicKeys
	log.Debugf("Generating %d PublicKeys", nKeys)
	keys := test.GenUserKeys(nKeys)
	// create a new census with the keys
	doPostNewCensus(c, a, keys.PublicKeys)
}

func TestPostAddKeysHandler(t *testing.T) {
	c := qt.New(t)

	a, _ := newTestAPI(c)
	a.r.POST("/census", a.postNewCensus)
	a.r.POST("/census/:censusid", a.postAddKeys)

	nKeys := 150
	// generate the publicKeys
	log.Debugf("Generating %d PublicKeys", nKeys)
	keys := test.GenUserKeys(nKeys)

	// create a new census with the first 100 keys
	censusID := doPostNewCensus(c, a, keys.PublicKeys[:100])

	// Add the rest of the keys
	doPostAddKeys(c, a, censusID, keys.PublicKeys[100:])
}

func TestPostCloseCensusHandler(t *testing.T) {
	c := qt.New(t)

	a, _ := newTestAPI(c)
	a.r.POST("/census", a.postNewCensus)
	a.r.POST("/census/:censusid", a.postAddKeys)
	a.r.POST("/census/:censusid/close", a.postCloseCensus)

	nKeys := 100
	// generate the publicKeys
	log.Debugf("Generating %d PublicKeys", nKeys)
	keys := test.GenUserKeys(nKeys)

	// create a new census with the first 100 keys
	censusID := doPostNewCensus(c, a, keys.PublicKeys)

	time.Sleep(1 * time.Second)

	_ = doPostCloseCensus(c, a, censusID)
}

func TestGetProofHandler(t *testing.T) {
	c := qt.New(t)

	a, _ := newTestAPI(c)
	a.r.POST("/census", a.postNewCensus)
	a.r.POST("/census/:censusid/close", a.postCloseCensus)
	a.r.GET("/census/:censusid/merkleproof/:pubkey", a.getMerkleProofHandler)

	nKeys := 100
	// generate the publicKeys
	log.Debugf("Generating %d PublicKeys", nKeys)
	keys := test.GenUserKeys(nKeys)

	// create a new census with the first 100 keys
	censusID := doPostNewCensus(c, a, keys.PublicKeys)

	time.Sleep(1 * time.Second)

	censusRoot := doPostCloseCensus(c, a, censusID)

	for i := 0; i < nKeys; i++ {
		cp := doGetProof(c, a, censusID, keys.PublicKeys[i])
		// fmt.Printf("Index: %d, MerkleProof: %x\n", cp.Index, cp.MerkleProof)

		v, err := census.CheckProof(censusRoot, cp.MerkleProof, cp.Index,
			&keys.PublicKeys[i])
		c.Assert(err, qt.IsNil)
		c.Assert(v, qt.IsTrue)
	}
}

func TestSendVotesHandler(t *testing.T) {
	c := qt.New(t)

	a, sqlite := newTestAPI(c)
	a.r.POST("/census", a.postNewCensus)
	a.r.POST("/census/:censusid/close", a.postCloseCensus)
	a.r.GET("/census/:censusid/merkleproof/:pubkey", a.getMerkleProofHandler)
	a.r.POST("/process/:processid", a.postVote)

	nKeys := 10
	// generate the publicKeys
	log.Debugf("Generating %d PublicKeys", nKeys)
	keys := test.GenUserKeys(nKeys)

	// create a new census with the first 100 keys
	censusID := doPostNewCensus(c, a, keys.PublicKeys)

	time.Sleep(1 * time.Second)

	censusRoot := doPostCloseCensus(c, a, censusID)

	var proofs []types.CensusProof
	for i := 0; i < nKeys; i++ {
		cp := doGetProof(c, a, censusID, keys.PublicKeys[i])
		v, err := census.CheckProof(censusRoot, cp.MerkleProof, cp.Index,
			&keys.PublicKeys[i])
		c.Assert(err, qt.IsNil)
		c.Assert(v, qt.IsTrue)
		proofs = append(proofs, cp)
	}

	// simulate SmartContract Process creation, by adding the CensusRoot in
	// the votesaggregator db
	processID := uint64(123)
	ethBlockNum := uint64(10)
	ethEndBlockNum := uint64(20)
	err := sqlite.StoreProcess(processID, censusRoot, ethBlockNum, ethEndBlockNum)
	c.Assert(err, qt.IsNil)

	// prepare the votes
	votes := prepareVotes(c, keys, proofs)
	// send the votes
	for i := 0; i < len(votes); i++ {
		doPostVote(c, a, processID, votes[i])
	}
}
