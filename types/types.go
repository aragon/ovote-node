package types

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/poseidon"
	"github.com/vocdoni/arbo"
)

var (
	hashLen int = arbo.HashFunctionPoseidon.Len()
)

// CensusProof contains the proof of a PublicKey in the Census Tree
type CensusProof struct {
	Index       uint64             `json:"index"`
	PublicKey   *babyjub.PublicKey `json:"publicKey"`
	MerkleProof []byte             `json:"merkleProof"`
}

// VotePackage represents the vote sent by the User
type VotePackage struct {
	Signature   babyjub.SignatureComp
	CensusProof CensusProof
	Vote        []byte
}

// Process represents a voting process
type Process struct {
	// ID is determined by the SmartContract, is unique for each Process
	ID uint64
	// CensusRoot is determined by the SmartContract, the same CensusRoot
	// can be reused by different Processes
	CensusRoot       []byte
	EthBlockNum      uint64
	InsertedDatetime time.Time
}

func (vp *VotePackage) verifySignature() error {
	voteBI := arbo.BytesToBigInt(vp.Vote)
	sigUncompressed, err := vp.Signature.Decompress()
	if err != nil {
		return err
	}
	v := vp.CensusProof.PublicKey.VerifyPoseidon(
		voteBI, sigUncompressed)
	if !v {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}
func (vp *VotePackage) verifyMerkleProof(root []byte) error {
	indexBytes := Uint64ToIndex(vp.CensusProof.Index)
	pubKHashBytes, err := HashPubKBytes(vp.CensusProof.PublicKey)
	if err != nil {
		return err
	}
	v, err := arbo.CheckProof(arbo.HashFunctionPoseidon, indexBytes,
		pubKHashBytes, root, vp.CensusProof.MerkleProof)
	if err != nil {
		return err
	}
	if !v {
		return fmt.Errorf("merkleproof verification failed")
	}
	return nil
}

// Verify checks the signature and merkleproof of the VotePackage
func (vp *VotePackage) Verify(root []byte) error {
	if err := vp.verifySignature(); err != nil {
		return err
	}
	if err := vp.verifyMerkleProof(root); err != nil {
		return err
	}
	return nil
}

// Uint64ToIndex returns the bytes representation of the given uint64 that will
// be used as a leaf index in the MerkleTree
func Uint64ToIndex(u uint64) []byte {
	indexBytes := arbo.BigIntToBytes(32, big.NewInt(int64(int(u)))) //nolint:gomnd
	return indexBytes
}

// HashPubKBytes returns the bytes representation of the Poseidon hash of the
// given PublicKey, that will be used as a leaf value in the MerkleTree
func HashPubKBytes(pubK *babyjub.PublicKey) ([]byte, error) {
	pubKHash, err := poseidon.Hash([]*big.Int{pubK.X, pubK.Y})
	if err != nil {
		return nil, err
	}
	return arbo.BigIntToBytes(hashLen, pubKHash), nil
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
