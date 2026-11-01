package client

import (
	"bytes"
	"context"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/kurikla/clipik-s/internal/crypto"
	"github.com/kurikla/clipik-s/pkg/dto"
)

func (c *Client) handshake(ctx context.Context, addr string) (cipher.AEAD, dto.Format, string, error) {
	cEph, err := ecdh.X25519().GenerateKey(rand.Reader)

	if err != nil {
		return nil, "", "", err
	}

	cPub := cEph.PublicKey().Bytes()
	cID := c.identity.Public().(ed25519.PublicKey)

	cTrans := crypto.ClientTranscript(dto.Version, cPub, cID)
	cSig := ed25519.Sign(c.identity, cTrans)

	hs := dto.Handshake{
		Protocol: dto.Protocol,
		Version: dto.Version,
		Format: c.cfg.Format,
		ClientPub: cPub,
		ClientID: cID,
		ClientSig: cSig,
	}

	body, err := json.Marshal(hs)
	if err != nil {
		return nil, "", "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"http://"+addr+"/v1/handshake",
		bytes.NewReader(body),
	)

	if err != nil {
		return nil, "", "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)

	if err != nil {
		return nil, "", "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", "", fmt.Errorf("handshake status: %s", resp.Status)
	}

	var ack dto.HandshakeAck
	if err := json.NewDecoder(resp.Body).Decode(&ack); err != nil {
		return nil, "", "", err
	}

	if ack.Protocol != dto.Protocol {
		return nil, "", "", fmt.Errorf("wrong protocol: %q", ack.Protocol)
	}

	if !strings.HasPrefix(ack.Version, "1.") {
		return nil, "", "", fmt.Errorf("unsupported version: %q", ack.Version)
	}

	if ack.Format != c.cfg.Format {
		return nil, "", "", fmt.Errorf("format mismatch: %q", ack.Format)
	}

	if len(ack.ServerPub) != 32 {
		return nil, "", "", errors.New("bad server pub size")
	}

	if len(ack.ServerID) != ed25519.PublicKeySize {
		return nil, "", "", errors.New("bad server id size")
	}

	if len(ack.ServerSig) != ed25519.SignatureSize {
		return nil, "", "", errors.New("bad server sig size")
	}

	sTrans := crypto.ServerTranscript(ack.Version, cPub, ack.ServerPub, cID, ack.ServerID)

	if !ed25519.Verify(ack.ServerID, sTrans, ack.ServerSig) {
		return nil, "", "", errors.New("bad server signature")
	}

	sessionKey, err := crypto.DeriveSessionKey(
		cEph, ack.ServerPub, ack.Version,
		cPub, ack.ServerPub, cID, ack.ServerID,
	)

	if err != nil {
		return nil, "", "", err
	}

	aead, err := crypto.NewAEAD(sessionKey)
	if err != nil {
		return nil, "", "", err
	}

	return aead, ack.Format, ack.SessionToken, nil
}
