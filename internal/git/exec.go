// Package git locates the system git binary and executes it.
package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const (
	// EnvConfigGlobal is the git environment variable that selects an
	// alternate global config file. When set by gitp, it overrides any
	// prior value (including one already present in the process environment).
	EnvConfigGlobal = "GIT_CONFIG_GLOBAL"
)

// Finder locates the git executable. Tests may replace this.
var Finder = exec.LookPath

// Run replaces the current process with git using the given args and env.
// env should already include any profile overrides.
// On success this function never returns.
func Run(args []string, env []string) error {
	bin, err := Finder("git")
	if err != nil {
		return fmt.Errorf("git not found in PATH: %w", err)
	}

	argv := append([]string{bin}, args...)
	return syscall.Exec(bin, argv, env)
}

// WithGlobalConfig returns a copy of baseEnv with GIT_CONFIG_GLOBAL set to
// path. Any existing GIT_CONFIG_GLOBAL entry is removed so the profile wins.
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

// Environ is a thin wrapper around os.Environ for testability.
func Environ() []string {
	return os.Environ()
}
