package cryptography

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	//	"golang.org/x/crypto/curve25519"
	"crypto/ed25519"
	"crypto/sha512"
	"github.com/majestrate/session/lib/cryptography/edwards25519"
	"golang.org/x/crypto/nacl/box"
	// "io"
	"io/fs"
	"io/ioutil"
)

var ErrBadSeedSize = errors.New("bad seed size")
var ErrDecryptError = errors.New("failed to decrypt")
var ErrEncryptError = errors.New("failed to encrypt")

type KeyPair struct {
	publicKey ed25519.PublicKey
	secretKey ed25519.PrivateKey
}

func (keys *KeyPair) edPubKey() []byte {
	return keys.publicKey[:]
}

func (keys *KeyPair) Pubkey() []byte {
	var X, Pub [32]byte
	copy(Pub[:], keys.publicKey)
	if edToCurve(&Pub, &X) {
		return X[:]
	}
	return nil
}

func (keys *KeyPair) Regen() {
	keys.publicKey, keys.secretKey, _ = ed25519.GenerateKey(rand.Reader)
}

func (keys *KeyPair) SessionID() string {
	return fmt.Sprintf("05%s", hex.EncodeToString(keys.Pubkey()))
}

func (keys *KeyPair) SaveFile(fname string) error {
	return ioutil.WriteFile(fname, keys.secretKey.Seed(), fs.FileMode(0400))
}

func (keys *KeyPair) LoadFile(fname string) error {
	data, err := ioutil.ReadFile(fname)
	if err == nil && len(data) == 32 {
		keys.secretKey = ed25519.NewKeyFromSeed(data)
		keys.publicKey = keys.secretKey.Public().(ed25519.PublicKey)
		return nil
	}
	if err == nil {
		err = ErrBadSeedSize
	}
	return err
}

func (keys *KeyPair) decryptOuterMessage(data []byte) ([]byte, error) {
	out := make([]byte, 0)
	var x, X, pub, priv [32]byte
	copy(pub[:], keys.publicKey)
	copy(priv[:], keys.secretKey)
	if edToCurve(&pub, &X) && edPrivToCurvePriv(&priv, &x) {

		msg, ok := box.OpenAnonymous(out, data[:], &X, &x)
		if !ok {
			return nil, ErrDecryptError
		}
		return msg, nil
	}
	return nil, fmt.Errorf("failed to compute our curve25519 keys")
}

func (keys *KeyPair) encryptOuterMessage(data []byte, toXKey *[32]byte) ([]byte, error) {
	out := make([]byte, 0)
	return box.SealAnonymous(out, data[:], toXKey, rand.Reader)
}

func edToCurve(ed *[32]byte, curve *[32]byte) bool {
	var A edwards25519.ExtendedGroupElement
	var x, oneMinusY edwards25519.FieldElement
	if !A.FromBytes(ed) {
		return false
	}
	edwards25519.FeOne(&oneMinusY)
	edwards25519.FeSub(&oneMinusY, &oneMinusY, &A.Y)
	edwards25519.FeOne(&x)
	edwards25519.FeAdd(&x, &x, &A.Y)
	edwards25519.FeInvert(&oneMinusY, &oneMinusY)
	edwards25519.FeMul(&x, &x, &oneMinusY)
	edwards25519.FeToBytes(curve, &x)
	return true
}

func edPrivToCurvePriv(ed *[32]byte, curve *[32]byte) bool {
	h := sha512.Sum512((*ed)[:])
	h[31] &= 127
	h[31] |= 64
	copy((*curve)[:], h[:32])
	return true
}

func (keys *KeyPair) SignAndEncrypt(recipX, data []byte) ([]byte, error) {

	var usEdKey [32]byte
	var themXKey [32]byte

	copy(themXKey[:], recipX)
	copy(usEdKey[:], keys.publicKey)

	data = addPadding(data)

	var body []byte
	body = append(body, data...)
	body = append(body, usEdKey[:]...)
	body = append(body, themXKey[:]...)

	sig := ed25519.Sign(keys.secretKey, body)
	var plain []byte
	plain = append(plain, data...)
	plain = append(plain, usEdKey[:]...)
	plain = append(plain, sig...)

	return keys.encryptOuterMessage(plain, &themXKey)
}

/// DecryptAndVerify takes a raw message and decrypts the outer message, verifies the inner message's signature and then returns the plaintext and the sender's pubkey
func (keys *KeyPair) DecryptAndVerify(data []byte) ([]byte, []byte, error) {
	plain, err := keys.decryptOuterMessage(data)
	if err != nil {
		return nil, nil, err
	}
	var themEdKey [32]byte
	var usEdKey [32]byte
	var themXKey [32]byte
	var usXKey [32]byte
	var sig [64]byte

	copy(sig[:], plain[len(plain)-64:])
	copy(themEdKey[:], plain[len(plain)-(32+64):len(plain)-64])
	copy(usEdKey[:], keys.publicKey)

	if !edToCurve(&themEdKey, &themXKey) {
		return nil, nil, fmt.Errorf("failed to convert ed25519 key to curve25519 key")
	}

	if !edToCurve(&usEdKey, &usXKey) {
		return nil, nil, fmt.Errorf("failed to convert ed25519 key to curve25519 key")
	}

	var body []byte
	msg := plain[:len(plain)-(32+64)]
	body = append(body, msg...)
	body = append(body, (themEdKey[:])...)
	body = append(body, (usXKey[:])...)
	if !ed25519.Verify(ed25519.PublicKey(themEdKey[:]), body, sig[:]) {
		return nil, nil, fmt.Errorf("failed to verify signature from %s, ed=%s sig=%s, data=%s, plain=%s", hex.EncodeToString(themXKey[:]), hex.EncodeToString(themEdKey[:]), hex.EncodeToString(sig[:]), hex.EncodeToString(body), hex.EncodeToString(plain))
	}
	return delPadding(msg), themXKey[:], nil
}
