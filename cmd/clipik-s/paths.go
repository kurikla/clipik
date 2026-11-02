package main

import (
	"os"
	"path/filepath"
	"runtime"
)

func getConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDTA")
		if appData == "" {
			home, err := os.UserHomeDir()

			if err != nil {
				return ""
			}

			appData = home
		}

		return filepath.Join(appData, "Clipik")
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		return filepath.Join(home, "Library", "Application Support", "Clipik")
	default: // Linux, freebsd, openbsd, netbsd, ...
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		return filepath.Join(home, ".config", "clipik")
	}
}

func getBaseLibDir() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDTA")
		if appData == "" {
			home, err := os.UserHomeDir()

			if err != nil {
				return ""
			}

			appData = home
		}

		return filepath.Join(appData, "ClipikS")
	case "darwin":
		home, err := os.UserHomeDir()

		if err != nil {
			return ""
		}

		return filepath.Join(home, "Library", "Application Support", "ClipikS")
	default: // Linux, freebsd, openbsd, netbsd, ...
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		return filepath.Join(home, ".local", "clipik-s")
	}
}

func expandTilde(p string) string {
	if p == "" || p[0] != '~' {
		return p
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}

	if len(p) == 1 {
		return home
	}

	if p[1] != '/' && p[1] != '\\' {
		return p   // ~user — не поддерживаем
	}

	return filepath.Join(home, p[2:])
}
