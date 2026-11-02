package main

import "strings"

var (
	physicalPrefixes []string = []string{
		"eth", "enp", "eno", "ens", "enx", // Linux ethernet
		"wlan", "wlp", "wlx", "wwlan",     // Linux wifi
		"en",                              // macOS en0/en1
		"Ethernet", "Wi-Fi",               // Windows
	}

	virtualPrefixes []string = []string{
		"docker", "veth", "virbr", "br-", "vmnet", "vboxnet",
		"tun", "tap", "wg", "throne", "zt", "tailscale", "lo",
	}
)

func isPhysical(name string) bool {
	for _, p := range physicalPrefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

func isVirtual(name string) bool {
	for _, p := range virtualPrefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}
