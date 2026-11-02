package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/kurikla/clipik/pkg/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use: "serve",
	Short: "Run the server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := cmd.Context().Value(configKey{}).(*Config)

		if cfg.SizeLimit <= 0 {
			return fmt.Errorf("size_limit must be positive integer, received %d", cfg.SizeLimit)
		}

		if cfg.HandshakeTimeout <= 0 {
			return fmt.Errorf("handshake_timeout must be positive integer, received %d", cfg.HandshakeTimeout)
		}

		if cfg.Port < 1 || cfg.Port > 65535 {
			return fmt.Errorf("--port invalid %d, out of range 1-65535", cfg.Port)
		}

		if cfg.TrimInterval <= 0 {
			return fmt.Errorf("trim_interval must be positive integer, received %d", cfg.TrimInterval)
		}

		if cfg.TrimLimit <= 0 {
			return fmt.Errorf("trim_limit must be positive integer, received %d", cfg.TrimLimit)
		}

		if cfg.FilesDir == "" {
			return errors.New("files_dir is empty")
		}

		store, err := server.NewSqliteStore(cfg.Database)

		if err != nil {
			return fmt.Errorf("creating sqlite3 store: %w", err)
		}

		srv, err := server.New(&server.Config{
			Port:             cfg.Port,
			Interfaces:       cfg.Interfaces,
			FilesDir:         cfg.FilesDir,
			LogDir:           cfg.LogDir,
			LogLevel:         cfg.LogLevel,
			LogPrefix:        "server",
			LogToStdout:      cfg.LogStdout,
			SizeLimit:        cfg.SizeLimit,
			HandshakeTimeout: time.Duration(cfg.HandshakeTimeout) * time.Second,
			IdentityPath:     cfg.IdentityPath,
			Store:			  store,
			Discoverer:       &server.ZeroconfDiscoverer{},
			TrimInterval:     time.Duration(cfg.TrimInterval) * time.Second,
			TrimLimit:		  cfg.TrimLimit,
		})

		if err != nil {
			return fmt.Errorf("creating server: %w", err)
		}

		errCh := make(chan error, 1)
		go func() {
			errCh <- srv.Start()
		}()

		select {
		case <-cmd.Context().Done():
			return srv.StopWithTimeout(7 * time.Second)
		case err := <-errCh:
			return err
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().Int64("size-limit", defaultSizeLimit, "size limit for clipboard")

	serveCmd.Flags().Int("handshake-timeout", defaultHandshakeTimeout, "timeout for wait to websocket handshake (seconds)")
	serveCmd.Flags().Int("port",              getDefaultPort(),        "port to listen on")
	serveCmd.Flags().Int("trim-interval",     defaultTrimInterval,     "interval trim database history")
	serveCmd.Flags().Int("trim-limit",        defaultTrimLimit,        "trim history limit")

	serveCmd.Flags().String("files-dir", getDefaultFilesDir(), "path to files directory")
}
