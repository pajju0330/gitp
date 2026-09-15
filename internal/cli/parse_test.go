package cli_test

import (
	"testing"

	"github.com/prajwal/gitp/internal/cli"
)

func TestParse_ProfileLeading(t *testing.T) {
	opts, err := cli.Parse([]string{"--profile", "personal", "push", "-u", "origin", "main"})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.ProfileSet || opts.Profile != "personal" {
		t.Fatalf("profile: got set=%v name=%q", opts.ProfileSet, opts.Profile)
	}
	assertArgs(t, opts.GitArgs, "push", "-u", "origin", "main")
}

func TestParse_ProfileShortAndEquals(t *testing.T) {
	opts, err := cli.Parse([]string{"-p", "work", "status"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Profile != "work" {
		t.Fatalf("got %q", opts.Profile)
	}

	opts, err = cli.Parse([]string{"--profile=personal", "log"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Profile != "personal" {
		t.Fatalf("got %q", opts.Profile)
	}
}

func TestParse_ProfileAnywhereLastWins(t *testing.T) {
	opts, err := cli.Parse([]string{"push", "--profile", "a", "-u", "-p", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Profile != "b" {
		t.Fatalf("want last profile b, got %q", opts.Profile)
	}
	assertArgs(t, opts.GitArgs, "push", "-u")
}

func TestParse_NoProfile(t *testing.T) {
	opts, err := cli.Parse([]string{"status", "--short"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.ProfileSet {
		t.Fatal("expected ProfileSet=false")
	}
	assertArgs(t, opts.GitArgs, "status", "--short")
}

func TestParse_HelpVersion(t *testing.T) {
	opts, err := cli.Parse([]string{"--help"})
	if err != nil || !opts.Help {
		t.Fatalf("help: opts=%+v err=%v", opts, err)
	}
	opts, err = cli.Parse([]string{"-V"})
	if err != nil || !opts.Version {
		t.Fatalf("version: opts=%+v err=%v", opts, err)
	}
}

func TestParse_MissingProfileValue(t *testing.T) {
	_, err := cli.Parse([]string{"--profile"})
	if err == nil {
		t.Fatal("expected error")
	}
	_, err = cli.Parse([]string{"-p", "--help"})
	if err == nil {
		t.Fatal("expected error for flag-looking profile name")
	}
}

func TestParse_Verbose(t *testing.T) {
	opts, err := cli.Parse([]string{"-v", "status"})
	if err != nil || !opts.Verbose {
		t.Fatalf("%+v %v", opts, err)
	}
	assertArgs(t, opts.GitArgs, "status")
}

func TestParse_GitVerbosePassthrough(t *testing.T) {
	opts, err := cli.Parse([]string{"commit", "-v"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Verbose {
		t.Fatal("trailing -v must go to git, not gitp")
	}
	assertArgs(t, opts.GitArgs, "commit", "-v")
}

func TestParse_BuiltinPreserved(t *testing.T) {
	opts, err := cli.Parse([]string{"--profile", "p", "whoami"})
	if err != nil {
		t.Fatal(err)
	}
	if cli.BuiltinName(opts) != "whoami" {
		t.Fatalf("builtin=%q", cli.BuiltinName(opts))
	}
}

func assertArgs(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("args len: got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("args[%d]: got %q want %q", i, got[i], want[i])
		}
	}
}
