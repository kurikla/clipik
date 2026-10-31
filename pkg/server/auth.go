package server

import (
	"net/http"
	"strings"
)

func (s *Server) sessionFromRequest(r *http.Request) (*session, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, false
	}

	tok := strings.TrimPrefix(auth, "Bearer ")
	if tok == "" {
		return nil, false
	}

	return s.sessions.get(tok)
}
