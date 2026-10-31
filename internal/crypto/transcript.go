package crypto

import "crypto/sha256"

const HandshakeContext = "clipik"

func ClientTranscript(version string, clientPub, clientID []byte) []byte {
	h := sha256.New()
	h.Write([]byte(HandshakeContext))
	h.Write([]byte("client"))
	h.Write([]byte(version))
	h.Write(clientPub)
	h.Write(clientID)
	return h.Sum(nil)
}

func ServerTranscript(version string, clientPub, serverPub, clientID, serverID []byte) []byte {
	h := sha256.New()
	h.Write([]byte(HandshakeContext))
	h.Write([]byte("server"))
	h.Write([]byte(version))
	h.Write(clientPub)
	h.Write(serverPub)
	h.Write(clientID)
	h.Write(serverID)
	return h.Sum(nil)
}

func HandshakeSalt(version string, clientPub, serverPub, clientID, serverID []byte) []byte {
	h := sha256.New()
	h.Write([]byte(HandshakeContext))
	h.Write([]byte("session"))
	h.Write([]byte(version))
	h.Write(clientPub)
	h.Write(serverPub)
	h.Write(clientID)
	h.Write(serverID)
	return h.Sum(nil)
}
