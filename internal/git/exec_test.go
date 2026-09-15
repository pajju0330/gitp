package git_test

import (
	"strings"
	"testing"

	"github.com/prajwal/gitp/internal/git"
)

func TestWithGlobalConfig_OverridesExisting(t *testing.T) {
	base := []string{
		"PATH=/usr/bin",
		"GIT_CONFIG_GLOBAL=/old/path",
		"HOME=/home/me",
	}
	got := git.WithGlobalConfig(base, "/new/profile")

	var global string
	count := 0
	for _, e := range got {
		if strings.HasPrefix(e, git.EnvConfigGlobal+"=") {
			count++
			global = e
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one GIT_CONFIG_GLOBAL, got %d in %v", count, got)
	}
	if global != "GIT_CONFIG_GLOBAL=/new/profile" {
		t.Fatalf("got %q", global)
	}
	if !contains(got, "PATH=/usr/bin") || !contains(got, "HOME=/home/me") {
		t.Fatalf("lost unrelated env: %v", got)
	}
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}
