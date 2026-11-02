package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
)

var (
	defaultHandshakeTimeout int = 7
	defaultSizeLimit int64 = 64 << 20
	defaultTrimInterval int = 30
	defaultTrimLimit int = 3000
)

func getDefaultConfigPath() string {
	envPath := os.Getenv(configEnv)
	if envPath != "" {
		return envPath
	}

	config_dir := getConfigDir()
	return filepath.Join(config_dir, "server.yaml")
}

func getDefaultIdentityPath() string {
	envPath := os.Getenv(identityEnv)
	if envPath != "" {
		return envPath
	}

	base_dir := getBaseLibDir()
	return filepath.Join(base_dir, "clipik.identity")
}

func getDefaultLogDir() string {
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

func getDefaultDatabasePath() string {
	envPath := os.Getenv(databaseEnv)
	if envPath != "" {
		return envPath
	}

	base_dir := getBaseLibDir()
	return filepath.Join(base_dir, "database.db")
}

func getDefaultFilesDir() string {
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

func getDefaultPort() int {
	ln, err := net.Listen("tcp", ":0")

	if err != nil {
		fmt.Println(os.Stderr, err)
		return 8765
	}

	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
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
