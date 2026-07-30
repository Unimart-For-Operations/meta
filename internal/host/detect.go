// Package host provides host auto-detection for cmdr's Nix configuration.
//
// It scans cmdr/home/02-hosts/ for meta.nix files that match the current
// system username, returning the host directory name used as the flake target.
package host

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// stringFieldRegex extracts a string field from meta.nix, such as username = "cmdr";
var stringFieldRegex = regexp.MustCompile(`(?m)(?:^|[\s{;])([A-Za-z0-9_-]+)\s*=\s*"([^"]*)"`)

// listFieldRegex extracts a list field from meta.nix, such as capabilities = [ "baseline" ];
var listFieldRegex = regexp.MustCompile(`(?ms)(?:^|[\s{;])([A-Za-z0-9_-]+)\s*=\s*\[(.*?)\]\s*;`)

// listItemRegex extracts quoted list items from a Nix list body.
var listItemRegex = regexp.MustCompile(`"([^"]+)"`)

// Info holds the detected host configuration.
type Info struct {
	Name         string   // host directory name (e.g. "macbook", "studio")
	Platform     string   // platform subdirectory (e.g. "macos", "arch", "nixos")
	Username     string   // username from meta.nix
	Role         string   // semantic host role from meta.nix
	Capabilities []string // semantic capabilities from meta.nix
}

type metaFields struct {
	Strings map[string]string
	Lists   map[string][]string
}

// Detect scans the cmdr host directories for a meta.nix matching the current user.
// orgDir should be the root of the meta repo.
func Detect(orgDir string) (*Info, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}

	return DetectForUser(orgDir, currentUser.Username)
}

// detectPlatform detects the OS platform by reading /etc/os-release on Linux.
// Returns "macos" on Darwin, or one of "arch", "nixos", "ubuntu" on Linux,
// or "" if detection fails.
func detectPlatform() string {
	if runtime.GOOS == "darwin" {
		return "macos"
	}

	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return ""
	}

	id := ""
	idLike := ""

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "ID="):
			id = strings.Trim(strings.TrimPrefix(line, "ID="), "\"'")
		case strings.HasPrefix(line, "ID_LIKE="):
			idLike = strings.Trim(strings.TrimPrefix(line, "ID_LIKE="), "\"'")
		}
	}

	switch id {
	case "nixos", "ubuntu", "arch":
		return id
	}

	// Check ID_LIKE for derivatives (e.g., CachyOS has ID=cachyos, ID_LIKE=arch)
	for _, like := range strings.Fields(idLike) {
		if like == "arch" {
			return "arch"
		}
	}

	return ""
}

// findHostInPlatform searches for a matching username within a single platform
// directory. If multiple hosts match the username, prefers one whose directory
// name matches the system hostname.
func findHostInPlatform(hostsDir, platform, username string) *Info {
	hostname, _ := os.Hostname()
	var firstMatch *Info
	var hostnameMatch *Info

	platformDir := filepath.Join(hostsDir, platform)
	entries, err := os.ReadDir(platformDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		metaPath := filepath.Join(platformDir, entry.Name(), "meta.nix")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}

		meta := parseMetaFields(data)
		metaUsername := meta.Strings["username"]
		if metaUsername == "" || metaUsername != username {
			continue
		}

		info := &Info{
			Name:         entry.Name(),
			Platform:     platform,
			Username:     metaUsername,
			Role:         meta.Strings["role"],
			Capabilities: meta.Lists["capabilities"],
		}

		if firstMatch == nil {
			firstMatch = info
		}

		if hostname != "" && entry.Name() == hostname {
			hostnameMatch = info
		}
	}

	if hostnameMatch != nil {
		return hostnameMatch
	}
	return firstMatch
}

// DetectForUser scans for a host matching the given username.
func DetectForUser(orgDir, username string) (*Info, error) {
	hostsDir := filepath.Join(orgDir, "cmdr", "home", "02-hosts")

	// Detect the real OS platform and search there first.
	// This ensures NixOS machines match nixos/*, not arch/*.
	if platform := detectPlatform(); platform != "" {
		if info := findHostInPlatform(hostsDir, platform, username); info != nil {
			return info, nil
		}
	}

	// Fall back to old platform search order
	var searchDirs []string
	if runtime.GOOS == "darwin" {
		searchDirs = []string{"macos"}
	} else {
		searchDirs = []string{"arch", "nixos", "ubuntu"}
	}

	for _, platform := range searchDirs {
		if info := findHostInPlatform(hostsDir, platform, username); info != nil {
			return info, nil
		}
	}

	return nil, fmt.Errorf("no host config found for user %q in %s", username, hostsDir)
}

// ListHosts returns all available host configurations.
func ListHosts(orgDir string) ([]Info, error) {
	hostsDir := filepath.Join(orgDir, "cmdr", "home", "02-hosts")
	platforms := []string{"macos", "arch", "nixos", "ubuntu"}

	var hosts []Info
	for _, platform := range platforms {
		platformDir := filepath.Join(hostsDir, platform)
		entries, err := os.ReadDir(platformDir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			metaPath := filepath.Join(platformDir, entry.Name(), "meta.nix")
			data, err := os.ReadFile(metaPath)
			if err != nil {
				continue
			}

			meta := parseMetaFields(data)

			hosts = append(hosts, Info{
				Name:         entry.Name(),
				Platform:     platform,
				Username:     meta.Strings["username"],
				Role:         meta.Strings["role"],
				Capabilities: meta.Lists["capabilities"],
			})
		}
	}

	return hosts, nil
}

// GetHost returns a host configuration by host name.
func GetHost(orgDir, name string) (*Info, error) {
	hosts, err := ListHosts(orgDir)
	if err != nil {
		return nil, err
	}

	for _, h := range hosts {
		if h.Name == name {
			return &h, nil
		}
	}

	return nil, fmt.Errorf("host %q not found", name)
}

func parseMetaFields(data []byte) metaFields {
	fields := metaFields{
		Strings: map[string]string{},
		Lists:   map[string][]string{},
	}

	for _, match := range stringFieldRegex.FindAllSubmatch(data, -1) {
		if len(match) < 3 {
			continue
		}
		fields.Strings[string(match[1])] = strings.TrimSpace(string(match[2]))
	}

	for _, match := range listFieldRegex.FindAllSubmatch(data, -1) {
		if len(match) < 3 {
			continue
		}
		name := string(match[1])
		body := match[2]
		items := []string{}
		for _, item := range listItemRegex.FindAllSubmatch(body, -1) {
			if len(item) < 2 {
				continue
			}
			items = append(items, strings.TrimSpace(string(item[1])))
		}
		fields.Lists[name] = items
	}

	return fields
}
