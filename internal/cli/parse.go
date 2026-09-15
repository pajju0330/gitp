// Package cli parses gitp-specific flags from argv and leaves git args intact.
package cli

import (
	"fmt"
	"strings"
)

// Options are the flags owned by gitp (not forwarded to git).
type Options struct {
	// Profile is set when --profile / -p appears on the command line.
	Profile string
	// ProfileSet is true when the user explicitly passed --profile / -p.
	ProfileSet bool
	// Help requests gitp usage text.
	Help bool
	// Version requests gitp version output.
	Version bool
	// GitArgs are the remaining arguments forwarded to git.
	GitArgs []string
}

// Parse strips gitp flags from args and returns Options.
// Profile flags may appear anywhere; the last occurrence wins.
func Parse(args []string) (Options, error) {
	var opts Options
	out := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--help" || arg == "-h":
			opts.Help = true

		case arg == "--version" || arg == "-V":
			opts.Version = true

		case arg == "--profile" || arg == "-p":
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

		case strings.HasPrefix(arg, "--profile="):
			name := strings.TrimSpace(strings.TrimPrefix(arg, "--profile="))
			if name == "" {
				return Options{}, fmt.Errorf("--profile= requires a profile name")
			}
			opts.Profile = name
			opts.ProfileSet = true

		default:
			out = append(out, arg)
		}
	}

	opts.GitArgs = out
	return opts, nil
}

// Usage returns the human-readable help text for gitp.
func Usage() string {
	return `gitp — git CLI wrapper with named profiles

Usage:
  gitp [--profile|-p NAME] <git-command> [args...]
  gitp --help
  gitp --version

Options:
  -p, --profile NAME   Use ~/.gitconfig.NAME as the global git config
                       (overrides GIT_CONFIG_GLOBAL and the default profile)
  -h, --help           Show this help
  -V, --version        Show gitp version

Profile selection (highest wins):
  1. --profile / -p on the command line
  2. GITP_PROFILE environment variable
  3. default_profile in ~/.gitp/config
  4. git's normal global config (~/.gitconfig)

Examples:
  gitp --profile personal push -u origin main
  gitp -p work commit -m "fix"
  gitp status

Config (~/.gitp/config):
  default_profile = personal
`
}
