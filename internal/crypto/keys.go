package crypto

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"
)

func GenerateEphemeral() (*ecdh.PrivateKey, error) {
	return ecdh.X25519().GenerateKey(rand.Reader)
}

func LoadOrCreateIdentity(path string) (ed25519.PrivateKey, error) {
	if b, err := os.ReadFile(path); err == nil {
		if len(b) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("bad identity key size: %d", len(b))
		}

		return ed25519.PrivateKey(b), nil
	}

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, priv, 0o600); err != nil {
		return nil, fmt.Errorf("save identity: %w", err)
	}
	return priv, nil
}
