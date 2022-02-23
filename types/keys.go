package types

// SignatureCompressedSize sets the size of the compressed Signature byte array
const SignatureCompressedSize = 64

// Signature contains a babyjubjub compressed Signature
type Signature [SignatureCompressedSize]byte
