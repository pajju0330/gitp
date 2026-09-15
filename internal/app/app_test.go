package app_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/prajwal/gitp/internal/app"
	"github.com/prajwal/gitp/internal/config"
	"github.com/prajwal/gitp/internal/git"
)

func TestRun_HelpAndVersion(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	deps := app.Dependencies{
		Stdout: stdout,
		Stderr: stderr,
		LoadConfig: func() (config.Settings, error) {
			return config.Settings{}, nil
		},
		Resolve: func(string) (string, error) { return "", nil },
		Environ: func() []string { return nil },
		RunGit:  func([]string, []string) error { return errors.New("should not run") },
	}

	code := app.RunWith(deps, []string{"--help"})
	if code != 0 || !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("help: code=%d out=%q", code, stdout.String())
	}

	stdout.Reset()
	code = app.RunWith(deps, []string{"--version"})
	if code != 0 || !strings.Contains(stdout.String(), "gitp") {
		t.Fatalf("version: code=%d out=%q", code, stdout.String())
	}
}

func TestRun_ExplicitProfileOverridesDefaultAndEnv(t *testing.T) {
	var gotArgs []string
	var gotEnv []string

	deps := app.Dependencies{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LoadConfig: func() (config.Settings, error) {
			return config.Settings{DefaultProfile: "default"}, nil
		},
		Resolve: func(name string) (string, error) {
			if name != "explicit" {
				t.Fatalf("resolve called with %q", name)
			}
			return "/tmp/.gitconfig.explicit", nil
		},
		Environ: func() []string {
			return []string{"GIT_CONFIG_GLOBAL=/old", "PATH=/bin"}
		},
		RunGit: func(args []string, env []string) error {
			gotArgs = append([]string{}, args...)
			gotEnv = append([]string{}, env...)
			return nil
		},
	}

	code := app.RunWith(deps, []string{"--profile", "explicit", "push", "-u"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if len(gotArgs) != 2 || gotArgs[0] != "push" || gotArgs[1] != "-u" {
		t.Fatalf("args=%v", gotArgs)
	}

	global := valueOf(gotEnv, git.EnvConfigGlobal)
	if global != "/tmp/.gitconfig.explicit" {
		t.Fatalf("GIT_CONFIG_GLOBAL=%q", global)
	}
	if countKey(gotEnv, git.EnvConfigGlobal) != 1 {
		t.Fatalf("duplicate GIT_CONFIG_GLOBAL in %v", gotEnv)
	}
}

func TestRun_UsesDefaultProfile(t *testing.T) {
	var resolved string
	deps := app.Dependencies{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LoadConfig: func() (config.Settings, error) {
			return config.Settings{DefaultProfile: "personal"}, nil
		},
		Resolve: func(name string) (string, error) {
			resolved = name
			return "/home/.gitconfig.personal", nil
		},
		Environ: func() []string { return nil },
		RunGit:  func([]string, []string) error { return nil },
	}

	code := app.RunWith(deps, []string{"status"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if resolved != "personal" {
		t.Fatalf("resolved=%q", resolved)
	}
}

func TestRun_NoProfilePassthrough(t *testing.T) {
	var gotEnv []string
	deps := app.Dependencies{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LoadConfig: func() (config.Settings, error) {
			return config.Settings{}, nil
		},
		Resolve: func(string) (string, error) {
			t.Fatal("Resolve should not be called")
			return "", nil
		},
		Environ: func() []string {
			return []string{"HOME=/tmp", "GIT_CONFIG_GLOBAL=/keep"}
		},
		RunGit: func(_ []string, env []string) error {
			gotEnv = env
			return nil
		},
	}

	code := app.RunWith(deps, []string{"status"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if valueOf(gotEnv, git.EnvConfigGlobal) != "/keep" {
		t.Fatalf("should leave existing GIT_CONFIG_GLOBAL alone when no profile: %v", gotEnv)
	}
}

func TestRun_MissingProfileFile(t *testing.T) {
	stderr := &bytes.Buffer{}
	deps := app.Dependencies{
		Stdout: &bytes.Buffer{},
		Stderr: stderr,
		LoadConfig: func() (config.Settings, error) {
			return config.Settings{}, nil
		},
		Resolve: func(string) (string, error) {
			return "", errors.New("profile \"x\" not found")
		},
		Environ: func() []string { return nil },
		RunGit:  func([]string, []string) error { return errors.New("should not run") },
	}

	code := app.RunWith(deps, []string{"-p", "x", "status"})
	if code != 1 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stderr.String(), "not found") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func valueOf(env []string, key string) string {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return strings.TrimPrefix(e, prefix)
		}
	}
	return ""
}

func countKey(env []string, key string) int {
	prefix := key + "="
	n := 0
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			n++
		}
	}
	return n
}
