package types

import (
	"encoding/hex"
	"fmt"

	"github.com/iden3/go-iden3-crypto/babyjub"
)

// CensusProof contains the proof of a PublicKey in the Census Tree
type CensusProof struct {
	Index       uint64
	PublicKey   babyjub.PublicKey
	MerkleProof []byte
}

// VotePackage represents the vote sent by the User
type VotePackage struct {
	Signature   babyjub.SignatureComp
	CensusProof CensusProof
	Vote        []byte
}

//
// // SignatureCompressedSize sets the size of the compressed Signature byte array
// const SignatureCompressedSize = 64
//
// // Signature contains a babyjubjub compressed Signature
// type Signature [SignatureCompressedSize]byte

// HexToPublicKey converts the given hex representation of a
// babyjub.PublicKeyComp, and returns the babyjub.PublicKey
func HexToPublicKey(h string) (*babyjub.PublicKey, error) {
	// pubKStr := c.Param("pubkey")
	// var pubK babyjub.PublicKey
	// err = json.Unmarshal([]byte(pubKStr), &pubK)
	// if err != nil {
	//         returnErr(c, err)
	//         return
	// }

	pubKCompBytes, err := hex.DecodeString(h)
	if err != nil {
		return nil, err
	}
	var pubKComp babyjub.PublicKeyComp
	if len(pubKComp[:]) != len(pubKCompBytes) {
		return nil, fmt.Errorf("unexpected pubK length: %d", len(pubKCompBytes))
	}
	copy(pubKComp[:], pubKCompBytes)
	pubK, err := pubKComp.Decompress()
	if err != nil {
		return nil, err
	}

	return pubK, nil
}
