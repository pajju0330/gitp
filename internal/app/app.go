// Package app orchestrates flag parsing, profile resolution, and git execution.
package app

import (
	"fmt"
	"io"
	"os"

	"github.com/prajwal/gitp/internal/cli"
	"github.com/prajwal/gitp/internal/config"
	"github.com/prajwal/gitp/internal/git"
	"github.com/prajwal/gitp/internal/profile"
)

// Version is set at link time via -ldflags.
var Version = "dev"

// Dependencies can be replaced in tests.
type Dependencies struct {
	LoadConfig func() (config.Settings, error)
	Resolve    func(name string) (string, error)
	Environ    func() []string
	RunGit     func(args []string, env []string) error
	Stdout     io.Writer
	Stderr     io.Writer
}

func defaults() Dependencies {
	return Dependencies{
		LoadConfig: config.Load,
		Resolve:    profile.Resolve,
		Environ:    git.Environ,
		RunGit:     git.Run,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
	}
}

// Run is the main entry for gitp. args should be os.Args[1:].
func Run(args []string) int {
	return RunWith(defaults(), args)
}

// RunWith is like Run but accepts injectable dependencies (for tests).
func RunWith(deps Dependencies, args []string) int {
	opts, err := cli.Parse(args)
	if err != nil {
		fmt.Fprintf(deps.Stderr, "gitp: %v\n", err)
		return 2
	}

	if opts.Help {
		fmt.Fprint(deps.Stdout, cli.Usage())
		return 0
	}

	if opts.Version {
		fmt.Fprintf(deps.Stdout, "gitp %s\n", Version)
		return 0
	}

	profileName, err := selectProfile(opts, deps)
	if err != nil {
		fmt.Fprintf(deps.Stderr, "gitp: %v\n", err)
		return 1
	}

	env := deps.Environ()
	if profileName != "" {
		path, err := deps.Resolve(profileName)
		if err != nil {
			fmt.Fprintf(deps.Stderr, "gitp: %v\n", err)
			return 1
		}
		env = git.WithGlobalConfig(env, path)
	}

	if err := deps.RunGit(opts.GitArgs, env); err != nil {
		fmt.Fprintf(deps.Stderr, "gitp: %v\n", err)
		return 1
	}
	// syscall.Exec never returns on success.
	return 0
}

// selectProfile applies precedence:
// 1. explicit --profile / -p
// 2. GITP_PROFILE / ~/.gitp/config default_profile
// 3. empty (native git global config)
func selectProfile(opts cli.Options, deps Dependencies) (string, error) {
	if opts.ProfileSet {
		return opts.Profile, nil
	}

	settings, err := deps.LoadConfig()
	if err != nil {
		return "", err
	}
	return settings.DefaultProfile, nil
}
