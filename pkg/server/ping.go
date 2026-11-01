package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket"

	"github.com/kurikla/clipik/pkg/dto"
	"github.com/kurikla/clipik/internal/crypto"
	"github.com/kurikla/clipik/internal/codec"
)

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.sessionFromRequest(r)

	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close(websocket. StatusNormalClosure, "")

	s.trackConn(conn)
	defer s.untrackConn(conn)

	cdc, err := codec.New(sess.format)

	s.logger.Info("ping client connected", "client_id", fmt.Sprintf("%x", sess.clientID))

	for {
		typ, data, err := conn.Read(r.Context())
		if err != nil {
			return
		}

		if typ != websocket.MessageBinary {
			continue
		}

		plaintext, err := crypto.Decrypt(sess.aead, data)
		if err != nil {
			return
		}

		var p dto.Ping
		if err := cdc.Decode(plaintext, &p); err != nil {
			return
		}

		pong := dto.Pong{PingTs: p.Ts, PongTs: time.Now().UnixNano()}
		plain, _ := cdc.Encode(pong)
		ciphertext, _ := crypto.Encrypt(sess.aead, plain)
		conn.Write(r.Context(), websocket.MessageBinary, ciphertext)
	}

	return
}
