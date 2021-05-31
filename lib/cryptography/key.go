package cryptography

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
)

var ErrBadSeedSize = errors.New("bad seed size")

type KeyPair struct {
	publicKey ed25519.PublicKey
	secretKey ed25519.PrivateKey
}

func (keys *KeyPair) Regen() {
	keys.publicKey, keys.secretKey, _ = ed25519.GenerateKey(rand.Reader)
}

func (keys *KeyPair) SessionID() string {
	return fmt.Sprintf("05%s", hex.EncodeToString(keys.publicKey[:]))
}

func (keys *KeyPair) SaveFile(fname string) error {
	return ioutil.WriteFile(fname, keys.secretKey.Seed(), fs.FileMode(0400))
}

func (keys *KeyPair) LoadFile(fname string) error {
	data, err := ioutil.ReadFile(fname)
	if err == nil && len(data) == ed25519.SeedSize {
		keys.secretKey = ed25519.NewKeyFromSeed(data)
		keys.publicKey = keys.secretKey.Public().(ed25519.PublicKey)
		return nil
	}
	if err == nil {
		err = ErrBadSeedSize
	}
	return err
}
