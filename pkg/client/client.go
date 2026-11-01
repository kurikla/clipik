package client

import (
	"context"
	"crypto/cipher"
	"crypto/ed25519"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/kurikla/clipik/internal/codec"
	"github.com/kurikla/clipik/internal/crypto"
	"github.com/kurikla/clipik/internal/logger"
	"github.com/kurikla/clipik/pkg/dto"
	pkglog "github.com/kurikla/clipik/pkg/logger"
	"github.com/kurikla/clipik/pkg/model"
)

const (
	defaultLogPrefix = "clipik-c"
	defaultSizeLimit int64 = 64 << 20
	defaultHandshakeTimeout = 7 * time.Second
	defaultPingInterval = 30 * time.Second
)

type Config struct {
	LogDir           string
	LogLevel         pkglog.LogLevel
	LogPrefix        string
	LogToStdout		 bool
	SizeLimit        int64
	HandshakeTimeout time.Duration
	IdentityPath	 string
	Store			 Store
	Discoverer		 Discoverer
	Format			 dto.Format
	PingInterval	 time.Duration
}

type Client struct {
	cfg      Config
	identity ed25519.PrivateKey
	logger   *slog.Logger
	http     *http.Client

	mu          sync.Mutex
	discovered  map[string]model.Service
	connections map[string]*serverConn

	stopCh   chan struct{}
	stopOnce sync.Once
}

type serverConn struct {
	addr        string
	token       string
	aead        cipher.AEAD
	codec       codec.Codec
	clipboardWS *websocket.Conn
	pingWS      *websocket.Conn
	cancel      context.CancelFunc
}

func New(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("client: config is nil")
	}

	if cfg.IdentityPath == "" {
		return nil, errors.New("client: IdentityPath is required")
	}

	if cfg.Store == nil {
		return nil, errors.New("client: Store is required")
	}

	if cfg.Discoverer == nil {
		return nil, errors.New("client: Discoverer is required")
	}

	if cfg.LogDir == "" {
		return nil, errors.New("client: LogDir is required")
	}

	if !cfg.LogLevel.Valid() {
		return nil, fmt.Errorf("client: invalid log level %q", cfg.LogLevel)
	}

	if cfg.LogPrefix == "" {
		cfg.LogPrefix = defaultLogPrefix
	}

	if cfg.SizeLimit == 0 {
		cfg.SizeLimit = defaultSizeLimit
	}

	if cfg.SizeLimit <= 0 {
		return nil, errors.New("client: SizeLimit must be positive")
	}

	if cfg.HandshakeTimeout == 0 {
		cfg.HandshakeTimeout = defaultHandshakeTimeout
	}

	if cfg.Format == "" {
		cfg.Format = dto.FormatMsgpack
	}

	if !cfg.Format.Valid() {
		return nil, errors.New("client: Format is invalid")
	}

	if cfg.PingInterval == 0 {
		cfg.PingInterval = defaultPingInterval
	}

	if cfg.PingInterval < 0 {
		return nil, errors.New("client: PingInterval must be positive")
	}

	lg, err := logger.New(&logger.Config{
		Dir: cfg.LogDir,
		Level: cfg.LogLevel,
		Prefix: cfg.LogPrefix,
		ToStdout: cfg.LogToStdout,
	})

	if err != nil {
		return nil, fmt.Errorf("logger: %w", err)
	}

	identity, err := crypto.LoadOrCreateIdentity(cfg.IdentityPath)

	if err != nil {
		return nil, fmt.Errorf("identity: %w", err)
	}

	lg.Info("client identity loaded",
		"path", cfg.IdentityPath,
		"id", fmt.Sprintf("%x", identity.Public().(ed25519.PublicKey)))

	return &Client{
		cfg: *cfg,
		identity: identity,
		logger: lg,
		http: &http.Client{Timeout: cfg.HandshakeTimeout},
		discovered: make(map[string]model.Service),
		connections: make(map[string]*serverConn),
		stopCh: make(chan struct{}),
	}, nil
}

func (c *Client) Start() error {
	c.logger.Info("client starting")
	obs := &Observer{client: c}
	return c.cfg.Discoverer.Start(obs)
}

func (c *Client) Stop() error {
	c.stopOnce.Do(func (){
		close(c.stopCh)

		if c.cfg.Discoverer != nil {
			_ = c.cfg.Discoverer.Stop()
		}

		c.mu.Lock()
		conns := c.connections
		c.connections = make(map[string]*serverConn)
		c.mu.Unlock()

		for _, sc := range conns {
			sc.cancel()
			sc.clipboardWS.Close(websocket.StatusNormalClosure, "")
			sc.pingWS.Close(websocket.StatusNormalClosure, "")
		}

		if c.cfg.Store != nil {
			if err := c.cfg.Store.Close(); err != nil {
				c.logger.Warn("store close", "err", err)
			}
		}
	})

	return nil
}

func (c *Client) Discovered() []model.Service {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make([]model.Service, 0, len(c.discovered))
	for _, s := range c.discovered {
		out = append(out, s)
	}

	return out
}

func (c *Client) Accept(service model.Service) {
	obs := &Observer{client: c}
	obs.OnAccept(service)
}

func (c *Client) Connected() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make([]string, 0, len(c.connections))
	for addr := range c.connections {
		out = append(out, addr)
	}

	return out
}

func (c *Client) History(limit int) ([]model.Clipboard, error) {
	return c.cfg.Store.List(limit)
}

func (c *Client) Get(id uuid.UUID) (*model.Clipboard, error) {
	return c.cfg.Store.Get(id)
}

func (c *Client) Latest() (*model.Clipboard, error){
	return c.cfg.Store.Latest()
}
