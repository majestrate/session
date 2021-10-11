package cryptography

import (
	"bytes"
	"testing"
)

func TestSignEncrypt(t *testing.T) {
	sender := Keygen()
	recip := Keygen()

	message := "bepis"

	ct, err := sender.SignAndEncrypt(recip.Pubkey(), []byte(message))
	if err != nil {
		t.Fatalf("failed to sign and encrypt: %s", err.Error())
	}
	t.Logf("ct=%q", ct)
	msg, fromkey, err := recip.DecryptAndVerify(ct)
	if err != nil {
		t.Fatalf("cannot decrypt and verify: %s", err.Error())
	}
	if string(msg) != message {
		t.Fail()
	}
	senderkey := sender.Pubkey()
	if !bytes.Equal(fromkey, senderkey) {
		t.Fatalf("sender pubkey mismatch: %q != %q", fromkey, senderkey)
	}
}
