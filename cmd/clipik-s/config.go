package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kurikla/clipik/pkg/logger"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type configKey struct{}

type FileConfig struct {
	Cfg Config `yaml:"server"`
}

type Config struct {
	ConfigPath       string          `yaml:"-"` // не сохраняется в файл
	IdentityPath	 string			 `yaml:"identity"`
	LogLevel         logger.LogLevel `yaml:"log_level"`
	LogDir           string          `yaml:"log_dir"`
	Database		 string			 `yaml:"database"`
	LogStdout		 bool			 `yaml:"logstdout"`
	SizeLimit        int64			 `yaml:"size_limit"`
	HandshakeTimeout int			 `yaml:"handshake_timeout"`
	Port             int             `yaml:"port"`
	TrimInterval     int             `yaml:"trim_interval"`
	TrimLimit        int             `yaml:"trim_limit"`
	FilesDir		 string			 `yaml:"files_dir"`
	Interfaces       []string        `yaml:"interfaces"`
}

func saveConfigFile(cfg *Config) error {
	fc := FileConfig{Cfg: *cfg}

	if err := os.MkdirAll(filepath.Dir(cfg.ConfigPath), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	b, err := yaml.Marshal(fc)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return os.WriteFile(cfg.ConfigPath, b, 0o644)
}

func loadConfigFile(cmd *cobra.Command) (*Config, error) {
	bindings := getBindings(cmd)

	path := (&binding[string]{
		field:    "",
		name:     "config",
		env:      configEnv,
		get:      cmd.Flags().GetString,
		fallback: getDefaultConfigPath,
	}).GetDefault()

	toSave := false

	cfg := &Config{}

	b, err := os.ReadFile(path)
	switch {
	case err == nil:
		var fc FileConfig

		if err := yaml.Unmarshal(b, &fc); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}

		cfg = &fc.Cfg

		flags := cmd.Flags()
		for _, b := range bindings {
			if err := b.Override(cfg, flags); err != nil {
				return nil, fmt.Errorf("override: %s: %w", b.Name(), err)
			}
		}
	case errors.Is(err, os.ErrNotExist):
		for _, b := range bindings {
			if err := b.Default(cfg); err != nil {
				return nil, fmt.Errorf("default %s: %w", b.Name(), err)
			}
		}

		interfaces, err := getDefaultInterfaces()

		if err != nil {
			return nil, err
		}

		cfg.Interfaces = interfaces

		toSave = true
	default:
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg.ConfigPath = expandTilde(path)

	if toSave {
		if err := saveConfigFile(cfg); err != nil {
			return nil, err
		}
	}

	cfg.IdentityPath = expandTilde(cfg.IdentityPath)
	cfg.LogDir       = expandTilde(cfg.LogDir)
	cfg.Database     = expandTilde(cfg.Database)
	cfg.FilesDir     = expandTilde(cfg.FilesDir)

	return cfg, nil
}
