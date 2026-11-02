package protocol

import (
	"fmt"
	"strings"
)

const (
	Protocol = "CLIPIK"
	Version  = "0.1.0"

	supportedMajor = "0.1."
)

func SupportsVersion(version string) bool {
	return strings.HasPrefix(version, supportedMajor)
}

func CheckVersion(version string) error {
	if !SupportsVersion(version) {
		return fmt.Errorf("unsuppored version: %q (want major %s)", version, supportedMajor)
	}

	return nil
}
