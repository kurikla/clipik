package server

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/kurikla/clipik/internal/crypto"
	"github.com/kurikla/clipik/internal/logger"
	"github.com/kurikla/clipik/pkg/dto"
	pkglog "github.com/kurikla/clipik/pkg/logger"
)

const (
	defaultHandshakeTimeout = 7 * time.Second
	defaultLogPrefix = "clipik"
	defaultSizeLimit int64 = 64 << 20
	defaultShutdownTimeout = 7 * time.Second
	defaultTrimInterval = 7 * time.Second
	defaultTrimLimit = 3000
)

type Config struct {
	Port             int
	Interfaces       []string
	FilesDir         string
	LogDir           string
	LogLevel         pkglog.LogLevel
	LogPrefix        string
	LogToStdout		 bool
	SizeLimit        int64
	HandshakeTimeout time.Duration
	IdentityPath     string
	Store			 Store
	Discoverer		 Discoverer
	TrimInterval	 time.Duration
	TrimLimit		 int
}

type Server struct {
	cfg      Config
	identity ed25519.PrivateKey
	logger	 *slog.Logger
	http	 *http.Server

	sessions       *sessionStore
	allowedFormats map[dto.Format]bool

	mdnsOn   bool

	gcCancel   context.CancelFunc
	gcDone     chan struct{}
	trimCancel context.CancelFunc
	trimDone   chan struct{}

	connsMu sync.Mutex
	conns   map[*websocket.Conn]struct{}

	logCloser io.Closer
}

func New(cfg *Config) (*Server, error) {
	if cfg == nil {
		return nil, errors.New("server: config is nil")
	}

	if cfg.IdentityPath == "" {
		return nil, errors.New("server: IdentityPath is required")
	}

	if cfg.Store == nil {
		return nil, errors.New("server: Store is required")
	}

	if cfg.Discoverer == nil {
		return nil, errors.New("server: Discoverer is required")
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, fmt.Errorf("server: invalid port %d", cfg.Port)
	}

	if cfg.LogDir == "" {
		return nil, errors.New("server: LogDir is required")
	}

	if !cfg.LogLevel.Valid() {
		return nil, fmt.Errorf("server: invalid log level %q", cfg.LogLevel)
	}

	if cfg.LogPrefix == "" {
		cfg.LogPrefix = defaultLogPrefix
	}

	if cfg.SizeLimit == 0 {
		cfg.SizeLimit = defaultSizeLimit
	}

	if cfg.SizeLimit <= 0 {
		return nil, errors.New("server: SizeLimit must be positive")
	}

	if cfg.HandshakeTimeout == 0 {
		cfg.HandshakeTimeout = defaultHandshakeTimeout
	}

	if cfg.HandshakeTimeout < 0 {
		return nil, errors.New("server: HandshakeTimeout must be >= 0")
	}

	if cfg.TrimInterval == 0 {
		cfg.TrimInterval = defaultTrimInterval
	}

	if cfg.TrimLimit == 0 {
		cfg.TrimLimit = defaultTrimLimit
	}

	if cfg.TrimInterval < 0 {
		return nil, errors.New("server: TrimInterval must be >= 0")
	}

	if cfg.TrimLimit < 0 {
		return nil, errors.New("server: TrimLimit must be >= 0")
	}

	lg, closer, err := logger.New(&logger.Config{
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

	lg.Info("server identity loaded",
		"path", cfg.IdentityPath,
		"id", fmt.Sprintf("%x", identity.Public().(ed25519.PublicKey)))

	s := &Server{
		cfg:       *cfg,
		identity:  identity,
		logger:    lg,
		sessions:  newSessionStore(),
		conns:     make(map[*websocket.Conn]struct{}),
		logCloser: closer,
		allowedFormats: map[dto.Format]bool{
			dto.FormatJSON: true,
			dto.FormatMsgpack: true,
		},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/handshake", s.handleHandshakeHTTP)
	mux.HandleFunc("/v1/files/upload", s.handleFileUpload)

	mux.HandleFunc("/ws/v1/clipboard", s.handleClipboard)
	mux.HandleFunc("/ws/v1/ping", s.handlePing)

	s.http = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	return s, nil
}

func (s *Server) Start() error {
	s.logger.Info("Keep tcp", "addr", s.http.Addr)
	ln, err := net.Listen("tcp", s.http.Addr)

	if err != nil {
		return fmt.Errorf("listen %s: %w", s.http.Addr, err)
	}

	if err := s.startMDNS(); err != nil {
		ln.Close()
		return fmt.Errorf("mdns: %w", err)
	}

	s.logger.Info("start GC for sessions")
	gcCtx, gcCancel := context.WithCancel(context.Background())
	s.gcCancel = gcCancel
	s.gcDone = make(chan struct{})
	go s.runGC(gcCtx, time.Minute, 10*time.Minute)

	s.logger.Info("start Trim Store")
	trimCtx, trimCancel := context.WithCancel(context.Background())
	s.trimCancel = trimCancel
	s.trimDone = make(chan struct{})
	go s.runStoreTrim(trimCtx)

	s.logger.Info("server listening", "addr", s.http.Addr)
	if err := s.http.Serve(ln); err != nil && err != http.ErrServerClosed {
		gcCancel()
		<-s.gcDone
		if s.mdnsOn {
			_ = s.cfg.Discoverer.Unregister()
			s.mdnsOn = false
		}
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	return s.StopWithTimeout(defaultShutdownTimeout)
}

func (s *Server) StopWithTimeout(timeout time.Duration) error {
	s.logger.Info("server shutdown", "addr", s.http.Addr, "timeout", timeout)

	if s.mdnsOn {
		_ = s.cfg.Discoverer.Unregister()
		s.mdnsOn = false
	}

	s.closeAllConns("server shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := s.http.Shutdown(ctx)

	if s.gcCancel != nil {
		s.gcCancel()
	}

	if s.gcDone != nil {
		select {
		case <-s.gcDone:
		case <-ctx.Done():
			s.logger.Warn("gc shutdown timeout")
		}
	}

	if s.trimCancel != nil {
		s.trimCancel()
	}

	if s.trimDone != nil {
		select {
		case <-s.trimDone:
		case <-ctx.Done():
			s.logger.Warn("trim shutdown timeout")
		}
	}

	if s.cfg.Store != nil {
		if cerr := s.cfg.Store.Close(); cerr != nil {
			s.logger.Warn("store close", "err", cerr)
		}
	}

	s.logCloser.Close()

	return err
}

func (s *Server) startMDNS() error {
	if len(s.cfg.Interfaces) == 0 {
		s.logger.Info("mdns skipped: no interfaces configured")
		return nil
	}

	txt := []string{
		"v=1.0.0",
		fmt.Sprintf("port=%d", s.cfg.Port),
	}

	if err := s.cfg.Discoverer.Register(
		"clipik",
		"_clipik._tcp",
		s.cfg.Port,
		txt,
		s.cfg.Interfaces,
	); err != nil {
		return err
	}

	s.mdnsOn = true
	s.logger.Info("mdns registered",
		"service", "_clipik._tcp",
		"interfaces", s.cfg.Interfaces,
		"port", s.cfg.Port)

	return nil
}

func (s *Server) runStoreTrim(ctx context.Context) {
	defer close(s.trimDone)
	t := time.NewTicker(s.cfg.TrimInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("trim stopped")
			return
		case <-t.C:
			n, err := s.cfg.Store.Trim(s.cfg.TrimLimit)
			if err != nil {
				s.logger.Warn("store trim", "err", err)
			} else if n > 0 {
				s.logger.Info("store trimmed", "deleted", n, "limit", s.cfg.TrimLimit)
			}
		}
	}
}

func (s *Server) runGC(ctx context.Context, interval, maxIdle time.Duration) {
	defer close(s.gcDone)

	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("gc stopped")
			return
		case <-t.C:
			s.sessions.GC(maxIdle)
		}
	}
}
