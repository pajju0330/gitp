// Package config loads and writes gitp's own settings.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DirName             = ".gitp"
	FileName            = "config"
	KeyDefaultProfile   = "default_profile"
	EnvDefaultProfile   = "GITP_PROFILE"
	RepoConfigKey       = "gitp.profile"
)

// Settings holds gitp application configuration.
type Settings struct {
	// DefaultProfile is the effective default: GITP_PROFILE, else file.
	DefaultProfile string
	// FileDefault is default_profile from ~/.gitp/config only.
	FileDefault string
	// EnvDefault is GITP_PROFILE if set.
	EnvDefault string
}

// Path returns the absolute path to ~/.gitp/config.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, DirName, FileName), nil
}

// Dir returns ~/.gitp.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, DirName), nil
}

// Load reads settings from the environment and ~/.gitp/config.
func Load() (Settings, error) {
	var s Settings
	s.EnvDefault = strings.TrimSpace(os.Getenv(EnvDefaultProfile))

	path, err := Path()
	if err != nil {
		return s, err
	}

	fileDefault, err := readFileDefault(path)
	if err != nil {
		return s, err
	}
	s.FileDefault = fileDefault

	if s.EnvDefault != "" {
		s.DefaultProfile = s.EnvDefault
	} else {
		s.DefaultProfile = s.FileDefault
	}
	return s, nil
}

func readFileDefault(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("open gitp config %s: %w", path, err)
	}
	defer f.Close()

	var def string
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
		if strings.TrimSpace(key) == KeyDefaultProfile {
			def = strings.TrimSpace(value)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read gitp config %s: %w", path, err)
	}
	return def, nil
}

// SetDefaultProfile writes default_profile to ~/.gitp/config.
// Empty name clears the default.
func SetDefaultProfile(name string) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}

	path, err := Path()
	if err != nil {
		return err
	}

	existing, _ := readOtherKeys(path)
	var b strings.Builder
	b.WriteString("# gitp configuration\n")
	if name != "" {
		b.WriteString(KeyDefaultProfile + " = " + name + "\n")
	}
	for _, line := range existing {
		b.WriteString(line + "\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func readOtherKeys(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		raw := scanner.Text()
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		key, _, ok := strings.Cut(line, "=")
		if ok && strings.TrimSpace(key) == KeyDefaultProfile {
			continue
		}
		lines = append(lines, raw)
	}
	return lines, scanner.Err()
}
