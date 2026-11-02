package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/kurikla/clipik/pkg/logger"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type configKey struct{}

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

func getDefaultInterfaces() ([]string, error) {
	ifaces, err := net.Interfaces()

	if err != nil {
		return nil, err
	}

	var out []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()

		if err != nil {
			continue
		}

		hasRoutable := false
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}

			if ipnet.IP.IsGlobalUnicast() && !ipnet.IP.IsLinkLocalUnicast() {
				hasRoutable = true
				break
			}
		}

		if !hasRoutable {
			continue
		}

		out = append(out, iface.Name)
	}

	return out, nil
}

// func defaultConfig(cmd *cobra.Command) (*Config, error) {
//	flags := cmd.Flags()
//	interfaces, err := getDefaultInterfaces()

//	if err != nil {
//		return nil, err
//	}

//	return &Config{
//		IdentityPath:	  flagOr(cmd, "identity",          flags.GetString, getIdentityPath),
//		LogDir:           flagOr(cmd, "log-dir",           flags.GetString, getLogDir),
//		Database:         flagOr(cmd, "database",          flags.GetString, getDatabasePath),
//		LogStdout:		  flagOr(cmd, "logstdout",		   flags.GetBool,   defaultLogStdout),
//		SizeLimit:        flagOr(cmd, "size-limit",        flags.GetInt64,  defaultSizeLimit),
//		HandshakeTimeout: flagOr(cmd, "handshake-timeout", flags.GetInt,    defaultHandshakeTimeout),
//		Port:             flagOr(cmd, "port",              flags.GetInt,    getDefaultPort),
//		TrimInterval:     flagOr(cmd, "trim-interval",     flags.GetInt,    defaultTrimInterval),
//		TrimLimit:        flagOr(cmd, "trim-limit",        flags.GetInt,    defaultTrimLimit),
//		FilesDir:		  flagOr(cmd, "files-dir",         flags.GetString, getFilesDir),

//		LogLevel:   logger.LogLevel(flagOr(cmd, "log-level", flags.GetString, string(logger.LogInfo))),
//		Interfaces: interfaces,
//	}, nil
// }


func saveConfigFile(cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(cfg.ConfigPath), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	b, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return os.WriteFile(cfg.ConfigPath, b, 0o644)
}

func loadConfigFile(cmd *cobra.Command) (*Config, error) {
	bindings := getBindings(cmd)

	path := (&binding[string]{
		field: "",
		name: "config",
		env: configEnv,
		get: cmd.Flags().GetString,
		fallback: getConfigPath,
	}).GetDefault()

	toSave := false

	cfg := &Config{}

	b, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := yaml.Unmarshal(b, cfg); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}

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

		toSave = true
	default:
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg.ConfigPath = path

	if toSave {
		saveConfigFile(cfg)
	}

	return cfg, nil
}
