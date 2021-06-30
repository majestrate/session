package cryptography

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"
	"io"
	"io/fs"
	"io/ioutil"
)

var ErrBadSeedSize = errors.New("bad seed size")
var ErrDecryptError = errors.New("failed to decrypt")

type KeyPair struct {
	publicKey [32]byte
	secretKey [32]byte
}

func (keys *KeyPair) Regen() {
	io.ReadFull(rand.Reader, keys.secretKey[:])
	curve25519.ScalarBaseMult(&keys.publicKey, &keys.secretKey)
}

func (keys *KeyPair) SessionID() string {
	return fmt.Sprintf("05%s", hex.EncodeToString(keys.publicKey[:]))
}

func (keys *KeyPair) SaveFile(fname string) error {
	return ioutil.WriteFile(fname, keys.secretKey[:], fs.FileMode(0400))
}

func (keys *KeyPair) LoadFile(fname string) error {
	data, err := ioutil.ReadFile(fname)
	if err == nil && len(data) == 32 {
		copy(keys.secretKey[:], data)
		curve25519.ScalarBaseMult(&keys.publicKey, &keys.secretKey)
		return nil
	}
	if err == nil {
		err = ErrBadSeedSize
	}
	return err
}

func (keys *KeyPair) DecryptSessionMessage(data []byte) ([]byte, error) {
	out := make([]byte, 0)
	msg, ok := box.OpenAnonymous(out, data[:], &keys.publicKey, &keys.secretKey)
	if !ok {
		return nil, ErrDecryptError
	}
	return msg, nil
}
