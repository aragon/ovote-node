package types

import (
	"math/big"
	"time"
)

// ProofInDB contains the proof data from an entry in the db
type ProofInDB struct {
	ProofID          uint64
	Proof            []byte
	PublicInputs     []byte
	InsertedDatetime time.Time
	ProcessID        uint64
}

// Proof represents a Groth16 zkSNARK proof
type Proof struct {
	A        [3]*big.Int    `json:"pi_a"`
	B        [3][2]*big.Int `json:"pi_b"`
	C        [3]*big.Int    `json:"pi_c"`
	Protocol string         `json:"protocol"`
}
