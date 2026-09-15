// Package app orchestrates flag parsing, profile resolution, and command dispatch.
package app

import (
	"fmt"
	"io"
	"os"

	"github.com/prajwal/gitp/internal/cli"
	"github.com/prajwal/gitp/internal/cmd"
	"github.com/prajwal/gitp/internal/config"
	"github.com/prajwal/gitp/internal/git"
	"github.com/prajwal/gitp/internal/profile"
)

// Version is set at link time via -ldflags.
var Version = "dev"

// Dependencies can be replaced in tests.
type Dependencies struct {
	LoadConfig  func() (config.Settings, error)
	Resolve     func(name string) (string, error)
	Environ     func() []string
	RunGit      func(args []string, env []string) error
	RepoProfile func(env []string) string
	Stdout      io.Writer
	Stderr      io.Writer
}

func defaults() Dependencies {
	return Dependencies{
		LoadConfig:  config.Load,
		Resolve:     profile.Resolve,
		Environ:     git.Environ,
		RunGit:      git.Run,
		RepoProfile: git.RepoProfile,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
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

	env := deps.Environ()
	name, source, err := selectProfile(opts, deps, env)
	if err != nil {
		fmt.Fprintf(deps.Stderr, "gitp: %v\n", err)
		return 1
	}

	var path string
	if name != "" {
		path, err = deps.Resolve(name)
		if err != nil {
			fmt.Fprintf(deps.Stderr, "gitp: %v\n", err)
			return 1
		}
		env = git.WithGlobalConfig(env, path)
	}

	if opts.Verbose {
		if name == "" {
			fmt.Fprintf(deps.Stderr, "gitp: profile=(none) source=git-default\n")
		} else {
			fmt.Fprintf(deps.Stderr, "gitp: profile=%s source=%s config=%s\n", name, source, path)
		}
	}

	c := cmd.Context{
		Stdout:      deps.Stdout,
		Stderr:      deps.Stderr,
		Env:         env,
		ProfileName: name,
		ProfilePath: path,
		Source:      source,
		Verbose:     opts.Verbose,
	}

	if builtin := cli.BuiltinName(opts); builtin != "" {
		return dispatchBuiltin(c, builtin, opts.GitArgs[1:])
	}

	if err := deps.RunGit(opts.GitArgs, env); err != nil {
		fmt.Fprintf(deps.Stderr, "gitp: %v\n", err)
		return 1
	}
	return 0
}

func dispatchBuiltin(c cmd.Context, name string, args []string) int {
	switch name {
	case "whoami":
		return cmd.Whoami(c)
	case "doctor":
		return cmd.Doctor(c)
	case "profile":
		return cmd.Profile(c, args)
	case "use":
		return cmd.Use(c, args)
	case "sync":
		return cmd.Sync(c)
	case "ship":
		return cmd.Ship(c, args)
	case "completion":
		return cmd.Completion(c, args)
	default:
		fmt.Fprintf(c.Stderr, "gitp: unknown command %q\n", name)
		return 2
	}
}

// selectProfile applies precedence:
// 1. --profile / -p
// 2. repo-local gitp.profile
// 3. GITP_PROFILE / ~/.gitp/config default_profile
// 4. empty
func selectProfile(opts cli.Options, deps Dependencies, env []string) (name, source string, err error) {
	if opts.ProfileSet {
		return opts.Profile, "flag", nil
	}

	if deps.RepoProfile != nil {
		if rp := deps.RepoProfile(env); rp != "" {
			return rp, "repo", nil
		}
	}

	settings, err := deps.LoadConfig()
	if err != nil {
		return "", "", err
	}
	if settings.EnvDefault != "" {
		return settings.EnvDefault, "env", nil
	}
	if settings.FileDefault != "" {
		return settings.FileDefault, "default", nil
	}
	// fallback for older Settings usage in tests that only set DefaultProfile
	if settings.DefaultProfile != "" {
		return settings.DefaultProfile, "default", nil
	}
	return "", "", nil
}
