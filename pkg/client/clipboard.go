package client

import (
	"context"
	"time"

	"github.com/coder/websocket"
	"github.com/kurikla/clipik/internal/crypto"
	"github.com/kurikla/clipik/pkg/dto"
	"github.com/kurikla/clipik/pkg/model"
)

func (c *Client) SendClipboard(clip model.Clipboard) {
	c.mu.Lock()
	conns := make([]*serverConn, 0, len(c.connections))
	for _, sc := range c.connections {
		conns = append(conns, sc)
	}
	c.mu.Unlock()

	for _, sc := range conns {
		clipDto := dto.Clipboard{
			Type: dto.ClipboardTypeData,
			Data: &clip.Data,
			File: nil,
		}

		plaintext, err := sc.codec.Encode(clipDto)
		if err != nil {
			c.logger.Warn("encode", "addr", sc.addr, "err", err)
			continue
		}

		ciphertext, err := crypto.Encrypt(sc.aead, plaintext)
		if err != nil {
			c.logger.Warn("encrypt", "addr", sc.addr, "err", err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = sc.clipboardWS.Write(ctx, websocket.MessageBinary, ciphertext)
		cancel()

		if err != nil {
			c.logger.Warn("ws write", "addr", sc.addr, "err", err)
			c.disconnect(sc.addr)
		}
	}
}
