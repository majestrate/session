package cryptography

func Keygen() *KeyPair {
	kp := new(KeyPair)
	kp.Regen()
	return kp
}
