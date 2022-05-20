package api

import (
	"math/big"

	"github.com/iden3/go-iden3-crypto/babyjub"
)

type newCensusReq struct {
	// babyjub.PublicKey Unmarshaler takes care of parsing hex
	// representation of compressed PublicKeys
	PublicKeys []babyjub.PublicKey `json:"publicKeys"`
	Weights    []*big.Int          `json:"weights"`
}
