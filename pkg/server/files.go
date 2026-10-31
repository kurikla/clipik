package server

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/kurikla/clipik-s/internal/crypto"
)

var fileIDRe = regexp.MustCompile(`^[a-zA-Z]0-9_-{1,64}$`)

func (s *Server) handleFileUpload(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.sessionFromRequest(r)

	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}

	id := r.Header.Get("X-File-ID")
	if !fileIDRe.MatchString(id) {
		http.Error(w, "invalid or missing X-File-ID", http.StatusBadRequest)
		return
	}

	limit := s.cfg.SizeLimit + int64(sess.aead.Overhead()) + 24
	r.Body = http.MaxBytesReader(w, r.Body, limit)

	encrypted, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read", http.StatusRequestEntityTooLarge)
		return
	}

	plaintext, err := crypto.Decrypt(sess.aead, encrypted)
	if err != nil {
		http.Error(w, "decrypt", http.StatusBadRequest)
		return
	}

	dstPath := filepath.Join(s.cfg.FilesDir, id)
	if err := os.WriteFile(dstPath, plaintext, 0o644); err != nil {
		http.Error(w, "write", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}
