package server

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/kurikla/clipik-s/pkg/dto"
)

type session struct {
	aead      cipher.AEAD
	clientID  []byte
	format    dto.Format
	createdAt time.Time
	lastSeen  time.Time
}

type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*session
}

func newSessionStore() *sessionStore {
	return &sessionStore{sessions: make(map[string]*session)}
}

func (s *sessionStore) create(aead cipher.AEAD, clientID []byte, format dto.Format) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	tok := hex.EncodeToString(b)
	now := time.Now()
	s.mu.Lock()
	s.sessions[tok] = &session{
		aead: aead,
		clientID: clientID,
		format: format,
		createdAt: now,
		lastSeen: now,
	}
	s.mu.Unlock()
	return tok, nil
}

func (s *sessionStore) get(tok string) (*session, bool) {
	s.mu.RLock()
	sess, ok := s.sessions[tok]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	s.mu.Lock()
	sess.lastSeen = time.Now()
	s.mu.Unlock()
	return sess, true
}

func (s *sessionStore) delete(tok string) {
	s.mu.Lock()
	delete(s.sessions, tok)
	s.mu.Unlock()
}

func (s *sessionStore) GC(maxIdle time.Duration) {
	cutoff := time.Now().Add(-maxIdle)
	s.mu.Lock()
	for tok, sess := range s.sessions {
		if sess.lastSeen.Before(cutoff) {
			delete(s.sessions, tok)
		}
	}
	s.mu.Unlock()
}
