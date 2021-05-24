package cryptography

import "crypto/ed25519"
import "crypto/rand"

func Keygen() *KeyPair {
	pk, sk, _ := ed25519.GenerateKey(rand.Reader)
	return &KeyPair{
		publicKey: pk,
		secretKey: sk,
	}
}
