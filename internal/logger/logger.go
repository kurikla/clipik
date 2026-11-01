package logger

import (
	"io"
	"os"
	"fmt"
	"strings"
	"log/slog"
	"path/filepath"
	"compress/gzip"

	"github.com/kurikla/clipik/pkg/logger"
	"github.com/DeRuina/timberjack"
)

func New(logDir string, level logger.LogLevel, prefix string) (*slog.Logger, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("create log_dir: %w", err)
	}

	if prefix == "" {
		return nil, fmt.Errorf("log_prefix is required, received %q", prefix)
	}

	if err := archiveOldLogs(logDir, prefix); err != nil {
		return nil, fmt.Errorf("archive old logs: %w", err)
	}

	filename := filepath.Join(logDir, prefix+".log")

	tj := &timberjack.Logger{
		Filename:    filename,
		MaxSize:     100,
		MaxBackups:  0,
		MaxAge:      0,
		LocalTime:   true,
		Compression: "gzip",
		RotateAt:    []string{"00:00"},
		FileMode:    0o644,
	}

	sLevel := slogLevel(level)
	return slog.New(slog.NewTextHandler(tj, &slog.HandlerOptions{Level: sLevel})), nil
}

func slogLevel(l logger.LogLevel) slog.Level {
	switch l {
	case logger.LogDebug:
		return slog.LevelDebug
	case logger.LogInfo:
		return slog.LevelInfo
	case logger.LogWarn:
		return slog.LevelWarn
	case logger.LogError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func archiveOldLogs(logDir, prefix string) error {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".log") {
			continue
		}

		src := filepath.Join(logDir, name)
		dst := src + ".gz"

		if err := gzipFile(src, dst); err != nil {
			return fmt.Errorf("gzip %s: %w", name, err)
		}

		if err := os.Remove(src); err != nil {
			return fmt.Errorf("remove %s: %w", name, err)
		}
	}

	return nil
}

func gzipFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	gz := gzip.NewWriter(out)
	defer gz.Close()

	if _, err := io.Copy(gz, in); err != nil {
		return err
	}

	return nil
}
