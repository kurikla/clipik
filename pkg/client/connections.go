package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coder/websocket"
	"github.com/kurikla/clipik-s/internal/codec"
)

func (c *Client) connect(addr string) error {
	c.mu.Lock()

	if _, ok := c.connections[addr]; ok {
		c.mu.Unlock()
		return nil
	}

	c.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), c.cfg.HandshakeTimeout)
	defer cancel()

	aead, format, token, err := c.handshake(ctx, addr)
	if err != nil {
		return fmt.Errorf("handshake: %w", err)
	}

	cdc, err := codec.New(format)
	if err != nil {
		return fmt.Errorf("codec: %w", err)
	}

	clipWs, err := c.dialWS(ctx, addr, "/ws/v1/clipboard", token)
	if err != nil {
		return fmt.Errorf("clipboard ws: %w", err)
	}

	pingWS, err := c.dialWS(ctx, addr, "/ws/v1/ping", token)
	if err != nil {
		clipWs.Close(websocket.StatusNormalClosure, "")
		return fmt.Errorf("ping ws: %w", err)
	}

	connCtx, connCancel := context.WithCancel(context.Background())
	sc := &serverConn{
		addr: addr,
		token: token,
		aead: aead,
		codec: cdc,
		clipboardWS: clipWs,
		pingWS: pingWS,
		cancel: connCancel,
	}

	c.mu.Lock()
	c.connections[addr] = sc
	c.mu.Unlock()

	c.logger.Info("connected", "addr", addr)
	go c.runPingLoop(connCtx, sc)
	return nil
}

func (c *Client) disconnect(addr string) error {
	c.mu.Lock()
	sc, ok := c.connections[addr]
	if ok {
		delete(c.connections, addr)
	}
	c.mu.Unlock()

	if !ok {
		return fmt.Errorf("cannot delete connection", "addr", addr)
	}

	c.logger.Info("disconnecting", "addr", addr)
	sc.cancel()
	sc.clipboardWS.Close(websocket.StatusNormalClosure, "")
	sc.pingWS.Close(websocket.StatusNormalClosure, "")

	return nil
}

func (c *Client) dialWS(ctx context.Context, addr, path, token string) (*websocket.Conn, error) {
	hdrs := http.Header{}
	hdrs.Set("Authorization", "Bearer "+token)
	conn, _, err := websocket.Dial(
		ctx,
		"ws://"+addr+path,
		&websocket.DialOptions{
			HTTPHeader: hdrs,
		},
	)

	return conn, err
}
