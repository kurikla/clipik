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
		return filepath.Join(home, ".config", "clipik-s")
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
