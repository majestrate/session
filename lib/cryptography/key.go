package cryptography

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/jorrizza/ed2curve25519"
	"golang.org/x/crypto/nacl/box"
	"io/fs"
	"io/ioutil"
)

var ErrBadSeedSize = errors.New("bad seed size")
var ErrDecryptError = errors.New("failed to decrypt")

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

func (keys *KeyPair) DecryptSessionMessage(data []byte) ([]byte, error) {
	fmt.Printf("%q\n", string(data))
	var publicKey [32]byte
	var secretKey [32]byte
	sk := ed2curve25519.Ed25519PrivateKeyToCurve25519(keys.secretKey)
	pk := ed2curve25519.Ed25519PublicKeyToCurve25519(keys.publicKey)
	copy(secretKey[:], sk[:])
	copy(publicKey[:], pk[:])
	fmt.Printf("%q %q\n", sk[:], pk[:])
	out := make([]byte, 0)
	msg, ok := box.OpenAnonymous(out, data[:], &publicKey, &secretKey)
	if !ok {
		return nil, ErrDecryptError
	}
	return msg, nil
}
