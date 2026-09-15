// Package config loads gitp's own settings (distinct from git config profiles).
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DirName is the subdirectory under $HOME that stores gitp settings.
	DirName = ".gitp"
	// FileName is the settings file inside DirName.
	FileName = "config"
	// KeyDefaultProfile is the key for the default profile name.
	KeyDefaultProfile = "default_profile"
	// EnvDefaultProfile is an optional override for the default profile.
	EnvDefaultProfile = "GITP_PROFILE"
)

// Settings holds gitp application configuration.
type Settings struct {
	DefaultProfile string
}

// Path returns the absolute path to ~/.gitp/config.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, DirName, FileName), nil
}

// Load reads settings from the environment and ~/.gitp/config.
// Precedence for DefaultProfile: GITP_PROFILE > config file > empty.
func Load() (Settings, error) {
	var s Settings

	if v := strings.TrimSpace(os.Getenv(EnvDefaultProfile)); v != "" {
		s.DefaultProfile = v
		return s, nil
	}

	path, err := Path()
	if err != nil {
		return s, err
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return s, fmt.Errorf("open gitp config %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == KeyDefaultProfile {
			s.DefaultProfile = value
		}
	}
	if err := scanner.Err(); err != nil {
		return s, fmt.Errorf("read gitp config %s: %w", path, err)
	}

	return s, nil
}
