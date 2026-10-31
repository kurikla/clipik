package crypto

import (
	"crypto/ecdh"
	"crypto/hkdf"
	"crypto/sha256"
	"fmt"
)

const SessionInfo = "clipik"

func DeriveSessionKey(
	ephPriv *ecdh.PrivateKey,
	peerPub []byte,
	version string,
	clientPub, serverPub, clientID, serverID []byte,
) ([]byte, error) {
	pub, err := ecdh.X25519().NewPublicKey(peerPub)
	if err != nil {
		return nil, fmt.Errorf("parse peer pub: %w", err)
	}

	shared, err := ephPriv.ECDH(pub)
	if err != nil {
		return nil, fmt.Errorf("ecdh: %w", err)
	}

	salt := HandshakeSalt(version, clientPub, serverPub, clientID, serverID)
	key, err := hkdf.Key(sha256.New, shared, salt, SessionInfo, 32)
	if err != nil {
		return nil, fmt.Errorf("hkdf: %w", err)
	}

	return key, nil
}
