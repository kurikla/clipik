package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"

	defaultLogStdout bool = false
)

var rootCmd = &cobra.Command{
	Use: "clipik-s",
	Short: "Clipik server",
	Long: `Server for clipboard synchronization`,
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfigFile(cmd)

		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if config.IdentityPath == "" {
			return errors.New("identity is empty")
		}

		if config.LogDir == "" {
			return errors.New("log_dir is empty")
		}

		if config.LogLevel == "" {
			return errors.New("log_level is  empty")
		}

		if !config.LogLevel.Valid() {
			return fmt.Errorf("log_level received %q (want: debug|info|warn|error)", config.LogLevel)
		}

		if config.Database == "" {
			return errors.New("database is  empty")
		}

		ctx := context.WithValue(cmd.Context(), configKey{}, config)
		cmd.SetContext(ctx)

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().String("config",    getConfigPath(),   "path to config file")
	rootCmd.PersistentFlags().String("identity",  getIdentityPath(), "path to identity file")
	rootCmd.PersistentFlags().String("log-level", "info",            "log level")
	rootCmd.PersistentFlags().String("log-dir",   getLogDir(),       "path to log directory")
	rootCmd.PersistentFlags().String("database",  getDatabasePath(), "path to sqlite3 database file")

	rootCmd.PersistentFlags().Bool("logstdout", defaultLogStdout, "enable logger stdout output")

}
