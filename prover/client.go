// Package prover implements the prover client to interact with the
// prover-server
package prover

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/aragon/zkmultisig-node/types"
)

// Proof represents a Groth16 zkSNARK proof
type Proof struct {
	// TODO maybe move this to types package
	A        [3]*big.Int    `json:"pi_a"`
	B        [3][2]*big.Int `json:"pi_b"`
	C        [3]*big.Int    `json:"pi_c"`
	Protocol string         `json:"protocol"`
}

// Client implements the prover http client, used to make requests to the
// prover server
type Client struct {
	url string
	c   *http.Client
}

// NewClient returns a new Client for the given proverURL
func NewClient(proverURL string) *Client {
	httpClient := &http.Client{}
	return &Client{
		url: proverURL,
		c:   httpClient,
	}
}

type errorMsg struct {
	Message string `json:"message"`
}

// GenProof sends the given ZKInputs to the prover-server to trigger the
// zkProof generation
func (c *Client) GenProof(zki *types.ZKInputs) (uint64, error) {
	jsonZKI, err := json.Marshal(zki)
	if err != nil {
		return 0, err
	}
	resp, err := c.c.Post(c.url+"/proof", "application/json", bytes.NewBuffer(jsonZKI))
	if err != nil {
		return 0, err
	}

	// resp.body.id contains the id to use to retrieve the proof later
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode == http.StatusBadRequest {
		var errMsg errorMsg
		if err = json.Unmarshal(body, &errMsg); err != nil {
			return 0, err
		}
		return 0, errors.New(errMsg.Message)
	}

	var m map[string]uint64
	err = json.Unmarshal(body, &m)
	if err != nil {
		return 0, err
	}

	return m["id"], nil
}
