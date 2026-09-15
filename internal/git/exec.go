// Package git locates and runs the system git binary.
package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const (
	EnvConfigGlobal = "GIT_CONFIG_GLOBAL"
	RepoProfileKey  = "gitp.profile"
)

var Finder = exec.LookPath

// Run replaces the current process with git. On success this never returns.
func Run(args []string, env []string) error {
	bin, err := Finder("git")
	if err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
	}
	argv := append([]string{bin}, args...)
	return syscall.Exec(bin, argv, env)
}

// Output runs git as a subprocess and returns stdout (trimmed) and exit error.
func Output(args []string, env []string) (string, error) {
	bin, err := Finder("git")
	if err != nil {
		return "", fmt.Errorf("git not found in PATH: %w", err)
	}
	cmd := exec.Command(bin, args...)
	cmd.Env = env
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	out := strings.TrimSpace(stdout.String())
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return out, fmt.Errorf("%s", msg)
	}
	return out, nil
}

// RunInherit runs git inheriting stdio; returns exit code (0 on success).
func RunInherit(args []string, env []string) int {
	bin, err := Finder("git")
	if err != nil {
		fmt.Fprintf(os.Stderr, "gitp: git not found in PATH: %v\n", err)
		return 1
	}
	cmd := exec.Command(bin, args...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "gitp: %v\n", err)
		return 1
	}
	return 0
}

// WithGlobalConfig returns a copy of baseEnv with GIT_CONFIG_GLOBAL set to path.
func WithGlobalConfig(baseEnv []string, path string) []string {
	out := make([]string, 0, len(baseEnv)+1)
	prefix := EnvConfigGlobal + "="
	for _, e := range baseEnv {
		if strings.HasPrefix(e, prefix) {
			continue
		}
		out = append(out, e)
	}
	return append(out, prefix+path)
}

func Environ() []string {
	return os.Environ()
}

// RepoProfile returns the local gitp.profile value, or "" if unset / not a repo.
func RepoProfile(env []string) string {
	out, err := Output([]string{"config", "--local", "--get", RepoProfileKey}, env)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// ConfigGet returns a single config value under the given env (profile-aware).
func ConfigGet(key string, env []string) string {
	out, err := Output([]string{"config", "--get", key}, env)
	if err != nil {
		return ""
	}
	return out
}
