package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/kurikla/clipik/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Binding interface {
	Name() string
	Default(cfg *Config) error
	Override(cfg *Config, fs *pflag.FlagSet) error
}

type binding[T any] struct {
	field    string
	name     string
	env		 string
	get      func(string) (T, error)
	fallback any
}

func getBindings(cmd *cobra.Command) []Binding {
	flags := cmd.Flags()

	getLogLevel := func(name string) (logger.LogLevel, error) {
		s, err := flags.GetString(name)

		if err != nil {
			return "", err
		}

		return logger.LogLevel(s), nil
	}

	return []Binding{
		&binding[int64]{"SizeLimit", "size-limit", "", flags.GetInt64, defaultSizeLimit},

		&binding[bool]{"LogStdout", "logstdout", "", flags.GetBool, defaultLogStdout},

		&binding[logger.LogLevel]{"LogLevel", "log-level", logLevelEnv, getLogLevel, logger.LogInfo},

		&binding[string]{"IdentityPath", "identity",  identityEnv, flags.GetString, getIdentityPath},
		&binding[string]{"LogDir",       "log-dir",   logDirEnv,   flags.GetString, getLogDir},
		&binding[string]{"Database",     "database",  databaseEnv, flags.GetString, getDatabasePath},
		&binding[string]{"FilesDir",		"files-dir", "",          flags.GetString, getFilesDir},

		&binding[int]{"HandshakeTimeout", "handshake-timeout", "", flags.GetInt, defaultHandshakeTimeout},
		&binding[int]{"Port",			 "port",              "", flags.GetInt, getDefaultPort},
		&binding[int]{"TrimInterval",	 "trim-interval",     "", flags.GetInt, defaultTrimInterval},
		&binding[int]{"TrimLimit",		 "trim-limit",        "", flags.GetInt, defaultTrimLimit},

		// LogLevel:   logger.LogLevel(flagOr(cmd, "log-level", flags.GetString, string(logger.LogInfo))),
		// Interfaces: interfaces,

	}
}

func (b *binding[T]) Name() string {
	return b.field
}

func (b *binding[T]) Default(cfg *Config) error {
	if err := setField(cfg, b.field, b.GetDefault()); err != nil {
		return err
	}

	return nil
}

func (b *binding[T]) Override(cfg *Config, fs *pflag.FlagSet) error {
	v, err := b.fieldOverride(cfg, fs)

	if err != nil {
		return err
	}

	err = setField(cfg, b.field, v)

	if err != nil {
		return err
	}

	return nil
}

func (b *binding[T]) GetDefault() T {
	if v, err := b.get(b.name); err == nil {
		return v
	}

	if fn, ok := b.fallback.(func() T); ok {
		return fn()
	}

	if v, ok := b.fallback.(T); ok {
		return v
	}

	var zero T
	return zero
}

func (b *binding[T]) fieldOverride(cfg *Config, fs *pflag.FlagSet) (T, error) {
	if fs.Changed(b.name) {
		if v, err := b.get(b.name); err == nil {
			return v, nil
		}

		return b.cfgVal(cfg)
	}

	if b.env == "" {
		return b.cfgVal(cfg)
	}

	s := os.Getenv(b.env)
	if s == "" {
		return b.cfgVal(cfg)
	}

	var zero T
	switch any(zero).(type) {
	case string:
		return any(s).(T), nil
	case logger.LogLevel:
		l := logger.LogLevel(s)
		return any(l).(T), nil
	case int:
		n, err := strconv.Atoi(s)

		if err != nil {
			return zero, err
		}

		return any(n).(T), nil
	case int64:
		n, err := strconv.ParseInt(s, 10, 64)

		if err != nil {
			return zero, err
		}

		return any(n).(T), nil
	case bool:
		bv, err := strconv.ParseBool(s)

		if err != nil {
			return zero, err
		}

		return any(bv).(T), nil
	}

	return zero, fmt.Errorf("unsupported env type %T", zero)
}

func (b *binding[T]) cfgVal(cfg *Config) (T, error) {
	raw, err := getField(cfg, b.field)

	if err != nil {
		var zero T
		return zero, err
	}

	v, ok := raw.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("field %q: cannot cast %T to %T", b.field, raw, zero)
	}

	return v, nil
}
