// Package cli parses gitp-specific flags from argv and leaves git args intact.
package cli

import (
	"fmt"
	"strings"
)

// Built-in gitp commands (not forwarded to git).
var Builtins = map[string]bool{
	"whoami":     true,
	"doctor":     true,
	"profile":    true,
	"use":        true,
	"sync":       true,
	"ship":       true,
	"completion": true,
}

// Options are the flags owned by gitp (not forwarded to git).
type Options struct {
	Profile    string
	ProfileSet bool
	Verbose    bool
	Help       bool
	Version    bool
	GitArgs    []string
}

// Parse strips gitp flags from args and returns Options.
//
// --profile / -p may appear anywhere (last wins).
// -v / --verbose / -h / -V / --help / --version are only consumed before the
// first positional argument so git still receives e.g. `commit -v`.
func Parse(args []string) (Options, error) {
	var opts Options
	out := make([]string, 0, len(args))
	pastPositional := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Profile flags: allowed anywhere.
		if arg == "--profile" || arg == "-p" {
			if i+1 >= len(args) {
				return Options{}, fmt.Errorf("%s requires a profile name", arg)
			}
			i++
			name := strings.TrimSpace(args[i])
			if name == "" {
				return Options{}, fmt.Errorf("%s requires a non-empty profile name", arg)
			}
			if strings.HasPrefix(name, "-") {
				return Options{}, fmt.Errorf("invalid profile name %q", name)
			}
			opts.Profile = name
			opts.ProfileSet = true
			continue
		}
		if strings.HasPrefix(arg, "--profile=") {
			name := strings.TrimSpace(strings.TrimPrefix(arg, "--profile="))
			if name == "" {
				return Options{}, fmt.Errorf("--profile= requires a profile name")
			}
			opts.Profile = name
			opts.ProfileSet = true
			continue
		}

		if !pastPositional {
			switch arg {
			case "--help", "-h":
				opts.Help = true
				continue
			case "--version", "-V":
				opts.Version = true
				continue
			case "--verbose", "-v":
				opts.Verbose = true
				continue
			}
			if !strings.HasPrefix(arg, "-") {
				pastPositional = true
			}
		}

		out = append(out, arg)
	}

	opts.GitArgs = out
	return opts, nil
}

// BuiltinName returns the gitp builtin command if GitArgs starts with one.
func BuiltinName(opts Options) string {
	if len(opts.GitArgs) == 0 {
		return ""
	}
	if Builtins[opts.GitArgs[0]] {
		return opts.GitArgs[0]
	}
	return ""
}

// Usage returns the human-readable help text for gitp.
func Usage() string {
	return `gitp — git CLI wrapper with named profiles

Usage:
  gitp [flags] <git-command> [args...]
  gitp [flags] <builtin> [args...]

Flags (before the command):
  -p, --profile NAME   Use ~/.gitconfig.NAME (overrides everything else; may appear anywhere)
  -v, --verbose        Print which profile/config is active to stderr
  -h, --help           Show this help
  -V, --version        Show gitp version

Builtins:
  whoami               Show active profile identity
  doctor               Diagnose profile, SSH, and remote auth
  profile list|show|edit|create|use|init-ssh
  use [NAME|--clear]   Bind/clear profile for this repo (gitp.profile)
  sync                 fetch + rebase onto upstream
  ship                 push and open a PR (requires gh)
  completion bash|zsh|fish

Profile selection (highest wins):
  1. --profile / -p
  2. repo-local gitp.profile (gitp use NAME)
  3. GITP_PROFILE
  4. default_profile in ~/.gitp/config
  5. git's normal ~/.gitconfig

Examples:
  gitp --profile personal push -u origin main
  gitp profile use personal
  gitp use personal
  gitp -v status
`
}
