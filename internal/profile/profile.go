// Package profile resolves named git config profiles to on-disk paths.
package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// FilePrefix is prepended to the profile name under $HOME.
	// Example: personal -> ~/.gitconfig.personal
	FilePrefix = ".gitconfig."
)

var namePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// ErrInvalidName is returned when a profile name fails validation.
type ErrInvalidName struct {
	Name string
}

func (e ErrInvalidName) Error() string {
	return fmt.Sprintf("invalid profile name %q: use only letters, digits, '.', '_' or '-'", e.Name)
}

// ErrNotFound is returned when the profile config file does not exist.
type ErrNotFound struct {
	Name string
	Path string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("profile %q not found: expected config at %s", e.Name, e.Path)
}

// Valid reports whether name is a safe profile identifier.
func Valid(name string) bool {
	return name != "" && namePattern.MatchString(name)
}

// PathFor returns the absolute path for ~/.gitconfig.{name}.
func PathFor(name string) (string, error) {
	name = strings.TrimSpace(name)
	if !Valid(name) {
		return "", ErrInvalidName{Name: name}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, FilePrefix+name), nil
}

// Resolve validates the profile and ensures its config file exists.
func Resolve(name string) (string, error) {
	path, err := PathFor(name)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNotFound{Name: name, Path: path}
		}
		return "", fmt.Errorf("stat profile config %s: %w", path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("profile config %s is a directory, expected a file", path)
	}

	return path, nil
}
