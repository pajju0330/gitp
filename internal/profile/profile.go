// Package profile resolves and manages named git config profiles.
package profile

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	FilePrefix = ".gitconfig."
)

var namePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

type ErrInvalidName struct{ Name string }

func (e ErrInvalidName) Error() string {
	return fmt.Sprintf("invalid profile name %q: use only letters, digits, '.', '_' or '-'", e.Name)
}

type ErrNotFound struct {
	Name string
	Path string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("profile %q not found: expected config at %s", e.Name, e.Path)
}

func Valid(name string) bool {
	return name != "" && namePattern.MatchString(name)
}

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

// List returns sorted profile names from ~/.gitconfig.*.
func List() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}
	matches, err := filepath.Glob(filepath.Join(home, FilePrefix+"*"))
	if err != nil {
		return nil, err
	}
	var names []string
	for _, m := range matches {
		base := filepath.Base(m)
		name := strings.TrimPrefix(base, FilePrefix)
		if !Valid(name) {
			continue
		}
		info, err := os.Stat(m)
		if err != nil || info.IsDir() {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// Create writes a new profile file. If from is non-empty, copies that profile
// (or "default" for ~/.gitconfig). Fails if the target already exists.
func Create(name, from string) (string, error) {
	path, err := PathFor(name)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("profile %q already exists at %s", name, path)
	} else if !os.IsNotExist(err) {
		return "", err
	}

	var content []byte
	switch {
	case from == "" || from == "default":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		src := filepath.Join(home, ".gitconfig")
		content, err = os.ReadFile(src)
		if err != nil {
			if os.IsNotExist(err) {
				content = []byte("[user]\n\tname = \n\temail = \n")
			} else {
				return "", err
			}
		}
	default:
		src, err := Resolve(from)
		if err != nil {
			return "", err
		}
		content, err = os.ReadFile(src)
		if err != nil {
			return "", err
		}
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// Show writes the profile file contents to w.
func Show(name string, w io.Writer) error {
	path, err := Resolve(name)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// Edit opens the profile file in $EDITOR (or vi).
func Edit(name string) error {
	path, err := Resolve(name)
	if err != nil {
		return err
	}
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if editor == "" {
		editor = "vi"
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// DefaultSSHKeyPath returns ~/.ssh/id_ed25519_{name}.
func DefaultSSHKeyPath(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ssh", "id_ed25519_"+name), nil
}

// InitSSH ensures an ed25519 key exists for the profile and sets core.sshCommand.
// Returns the public key path and whether a new key was generated.
func InitSSH(name, email string) (pubPath string, generated bool, err error) {
	cfgPath, err := Resolve(name)
	if err != nil {
		return "", false, err
	}
	keyPath, err := DefaultSSHKeyPath(name)
	if err != nil {
		return "", false, err
	}
	pubPath = keyPath + ".pub"

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(keyPath), 0o700); err != nil {
			return "", false, err
		}
		if email == "" {
			email = name + "@local"
		}
		cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", keyPath, "-C", email, "-N", "")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", false, fmt.Errorf("ssh-keygen: %w", err)
		}
		generated = true
	} else if err != nil {
		return "", false, err
	}

	sshCmd := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes", keyPath)
	set := exec.Command("git", "config", "--file", cfgPath, "core.sshCommand", sshCmd)
	if out, err := set.CombinedOutput(); err != nil {
		return "", generated, fmt.Errorf("set core.sshCommand: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return pubPath, generated, nil
}
