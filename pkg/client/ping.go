package client

import (
	"context"
	"errors"
	"time"

	"github.com/coder/websocket"
	"github.com/kurikla/clipik/internal/crypto"
	"github.com/kurikla/clipik/pkg/dto"
)

func (c *Client) runPingLoop(ctx context.Context, sc *serverConn) {
	t := time.NewTicker(c.cfg.PingInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-t.C:
			if _, err := c.pingOnce(ctx, sc); err != nil {
				c.logger.Warn("ping", "addr", sc.addr, "err", err)
				c.disconnect(sc.addr)
				return
			}
		}
	}

	return
}

func (c *Client) pingOnce(ctx context.Context, sc *serverConn) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	ping := dto.Ping{Ts: time.Now().UnixNano()}
	plain, err := sc.codec.Encode(ping)
	if err != nil {
		return 0, err
	}

	ciphertext, err := crypto.Encrypt(sc.aead, plain)
	if err != nil {
		return 0, err
	}

	c.logger.Info("ping to", "addr", sc.addr)
	if err := sc.pingWS.Write(ctx, websocket.MessageBinary, ciphertext); err != nil {
		return 0, err
	}

	_, data, err := sc.pingWS.Read(ctx)
	if err != nil {
		return 0, err
	}

	plaintext, err := crypto.Decrypt(sc.aead, data)
	if err != nil {
		return 0, err
	}

	var pong dto.Pong
	if err := sc.codec.Decode(plaintext, &pong); err != nil {
		return 0, err
	}

	c.logger.Info("pong from", "addr", sc.addr, "pong", pong)
	if pong.PingTs != ping.Ts {
		return 0, errors.New("pong mismatch")
	}

	return time.Since(time.Unix(0, ping.Ts)), nil
}
