package types

// CensusProof contains the proof of a PublicKey in the Census Tree
type CensusProof struct {
	Index       uint64
	MerkleProof []byte
}

// Vote represents the vote sent by the User
type Vote struct {
	Signature   Signature
	CensusProof CensusProof
	Vote        []byte
}
