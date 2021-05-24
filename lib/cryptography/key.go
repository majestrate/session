package cryptography

import (
	"fmt"
	"encoding/hex"
	"crypto/ed25519"
)

type KeyPair struct {
	publicKey ed25519.PublicKey
	secretKey ed25519.PrivateKey
}


func (keys *KeyPair) SessionID() string {
	return fmt.Sprintf("05%s", hex.EncodeToString(keys.publicKey[:]))
}
