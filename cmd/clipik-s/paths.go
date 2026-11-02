package main

import (
	"os"
	"path/filepath"
	"runtime"
)

func getConfigPath() string {
	envPath := os.Getenv(configEnv)
	if envPath != "" {
		return envPath
	}

	config_dir := getConfigDir()
	return filepath.Join(config_dir, "server.yaml")
}

func getIdentityPath() string {
	envPath := os.Getenv(identityEnv)
	if envPath != "" {
		return envPath
	}

	base_dir := getBaseLibDir()
	return filepath.Join(base_dir, "clipik.identity")
}

func getLogDir() string {
	envPath := os.Getenv(logDirEnv)
	if envPath != "" {
		return envPath
	}

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

		return filepath.Join(appData, "ClipikS", "Logs")
	case "darwin":
		home, err := os.UserHomeDir()

		if err != nil {
			return ""
		}

		return filepath.Join(home, "Library", "Application Support", "ClipikS", "Logs")
	default: // Linux, freebsd, openbsd, netbsd, ...
		home, err := os.UserHomeDir()

		if err != nil {
			return ""
		}

		return filepath.Join(home, ".local", "clipik-s", "logs")
	}
}

func getDatabasePath() string {
	envPath := os.Getenv(databaseEnv)
	if envPath != "" {
		return envPath
	}

	base_dir := getBaseLibDir()
	return filepath.Join(base_dir, "database.db")
}

func getFilesDir() string {
	envPath := os.Getenv(logDirEnv)
	if envPath != "" {
		return envPath
	}

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

		return filepath.Join(appData, "ClipikS", "Files")
	case "darwin":
		home, err := os.UserHomeDir()

		if err != nil {
			return ""
		}

		return filepath.Join(home, "Library", "Application Support", "ClipikS", "Files")
	default: // Linux, freebsd, openbsd, netbsd, ...
		home, err := os.UserHomeDir()

		if err != nil {
			return ""
		}

		return filepath.Join(home, ".local", "clipik-s", "files")
	}
}

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
