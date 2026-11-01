package server

import (
	"context"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"strings"
	"net/http"
	"encoding/json"

	"github.com/kurikla/clipik/internal/crypto"
	"github.com/kurikla/clipik/pkg/dto"
)

func (s *Server) handleHandshakeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, cancel := context.WithTimeout(r.Context(), s.cfg.HandshakeTimeout)
	defer cancel()

	var hs dto.Handshake
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&hs); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if hs.Protocol != dto.Protocol {
		http.Error(w, "wrong protocol", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(hs.Version, "1.") {
		http.Error(w, "unsupported version", http.StatusBadRequest)
		return
	}

	if !s.allowedFormats[hs.Format] {
		http.Error(w, "unsupported format", http.StatusBadRequest)
		return
	}

	if len(hs.ClientPub) != 32 {
		http.Error(w, "bad client_pub size", http.StatusBadRequest)
		return
	}

	if len(hs.ClientID) != ed25519.PublicKeySize {
		http.Error(w, "bad client_id size", http.StatusBadRequest)
		return
	}

	if len(hs.ClientSig) != ed25519.SignatureSize {
		http.Error(w, "bad client_sig size", http.StatusBadRequest)
		return
	}

	cTrans := crypto.ClientTranscript(hs.Version, hs.ClientPub, hs.ClientID)
	if !ed25519.Verify(hs.ClientID, cTrans, hs.ClientSig) {
		http.Error(w, "bad client signature", http.StatusUnauthorized)
		return
	}

	sEph, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	sPub := sEph.PublicKey().Bytes()
	sID := s.identity.Public().(ed25519.PublicKey)

	sTrans := crypto.ServerTranscript(hs.Version, hs.ClientPub, sPub, hs.ClientID, sID)
	sSig := ed25519.Sign(s.identity, sTrans)

	sessionKey, err := crypto.DeriveSessionKey(
		sEph,
		hs.ClientPub,
		hs.Version,
		hs.ClientPub,
		sPub,
		hs.ClientID,
		sID,
	)

	if err != nil {
		http.Error(w, "derive", http.StatusInternalServerError)
		return
	}

	sEph = nil

	aead, err := crypto.NewAEAD(sessionKey)
	if err != nil {
		http.Error(w, "aead", http.StatusInternalServerError)
		return
	}

	token, err := s.sessions.create(aead, hs.ClientID, hs.Format)
	if err != nil {
		http.Error(w, "session", http.StatusInternalServerError)
		return
	}

	ack := dto.HandshakeAck{
		Protocol:     dto.Protocol,
		Version:      dto.Version,
		Format:       hs.Format,
		ServerPub:    sPub,
		ServerID:     sID,
		ServerSig:    sSig,
		SessionToken: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ack)
	s.logger.Info("handshake ok",
		"client_id", fmt.Sprintf("%x", hs.ClientID),
		"format", hs.Format,
		"token_prefix", token[:8])
}
