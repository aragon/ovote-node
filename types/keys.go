package types

import "github.com/iden3/go-iden3-crypto/babyjub"

// PublicKey is an alias for the babyjubjub PublicKey
type PublicKey babyjub.PublicKey

// SignatureCompressedSize sets the size of the compressed Signature byte array
const SignatureCompressedSize = 64

// Signature contains a babyjubjub compressed Signature
type Signature [SignatureCompressedSize]byte
