package cryptography

import (
	"encoding/hex"
	"golang.org/x/crypto/blake2b"
)

func B2SumHex(data string) string {
	sum := blake2b.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}
