package server

import (
	"fmt"
	"net/http"
	"unicode/utf8"

	"github.com/coder/websocket"
	"github.com/gabriel-vasile/mimetype"

	"github.com/kurikla/clipik-s/pkg/dto"
	"github.com/kurikla/clipik-s/pkg/model"
	"github.com/kurikla/clipik-s/internal/codec"
	"github.com/kurikla/clipik-s/internal/crypto"
)

func (s *Server) handleClipboard(w http.ResponseWriter, r *http.Request) {
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

	s.logger.Info("client connected", "client_id", fmt.Sprintf("%x", sess.clientID))

	cdc, err := codec.New(sess.format)

	if err != nil {
		return
	}

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
			s.logger.Warn("decrypt", "err", err)
			return
		}

		var clip dto.Clipboard
		if err := cdc.Decode(plaintext, &clip); err != nil {
			return
		}

		if clip.Type == dto.ClipboardTypeFile {
			conn.Close(websocket.StatusInternalError, fmt.Sprintf("unsupported type %q", clip.Type))
			return
		}

		if clip.Data == nil {
			conn.Close(websocket.StatusInvalidFramePayloadData, fmt.Sprintf("data cannot be null when type is %q", clip.Type))
			return
		}

		m := mimetype.Detect(*clip.Data).String()
		if m == "application/octet-stream" && utf8.Valid(*clip.Data) {
			m = "text/plain; charset=utf-8"
		}

		s.logger.Info("clipboard", "mime", m, "size", len(*clip.Data))
		saved, err := s.store.Save(model.Clipboard{
			Mime: m,
			Data: *clip.Data,
		})

		if err != nil {
			s.logger.Warn("store save", "err", err)
			continue
		}

		s.logger.Info("clipboard saved",
					  "id", saved.ID,
					  "mime", saved.Mime,
					  "size", len(saved.Data),
					  "created_at", saved.CreatedAt)
	}

	return
}
