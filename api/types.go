package api

import "github.com/iden3/go-iden3-crypto/babyjub"

type newCensusReq struct {
	// babyjub.PublicKey Unmarshaler takes care of parsing hex
	// representation of compressed PublicKeys
	PublicKeys []babyjub.PublicKey `json:"publicKeys"`
}
