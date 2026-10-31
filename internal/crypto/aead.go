package crypto

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

func NewAEAD(key []byte) (cipher.AEAD, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("bad key size: %d", len(key))
	}

	return chacha20poly1305.NewX(key)
}

func Encrypt(aead cipher.AEAD, plaintext []byte) ([]byte, error) {
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	out := make([]byte, len(nonce)+len(plaintext)+aead.Overhead())
	copy(out, nonce)
	aead.Seal(out[len(nonce):len(nonce)], nonce, plaintext, nil)
	return out, nil
}

func Decrypt(aead cipher.AEAD, data []byte) ([]byte, error) {
	if len(data) < aead.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short: %d", len(data))
	}

	nonce := data[:aead.NonceSize()]
	ct := data[aead.NonceSize():]
	return aead.Open(nil, nonce, ct, nil)
}
