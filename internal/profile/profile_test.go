package profile_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/prajwal/gitp/internal/profile"
)

func TestValid(t *testing.T) {
	cases := map[string]bool{
		"personal":      true,
		"work":          true,
		"client.acme":   true,
		"a_b-1":         true,
		"":              false,
		"../etc/passwd": false,
		"a/b":           false,
		"has space":     false,
	}
	for name, want := range cases {
		if got := profile.Valid(name); got != want {
			t.Errorf("Valid(%q)=%v want %v", name, got, want)
		}
	}
}

func TestResolve_OK(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".gitconfig.personal")
	if err := os.WriteFile(path, []byte("[user]\n\temail = a@b.c\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := profile.Resolve("personal")
	if err != nil {
		t.Fatal(err)
	}
	if got != path {
		t.Fatalf("got %q want %q", got, path)
	}
}

func TestResolve_NotFound(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	_, err := profile.Resolve("missing")
	var nf profile.ErrNotFound
	if !errors.As(err, &nf) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestResolve_InvalidName(t *testing.T) {
	_, err := profile.Resolve("../x")
	var inv profile.ErrInvalidName
	if !errors.As(err, &inv) {
		t.Fatalf("want ErrInvalidName, got %v", err)
	}
}

func TestListAndCreate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[user]\n\temail = d@e.f\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	path, err := profile.Create("personal", "default")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}

	names, err := profile.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "personal" {
		t.Fatalf("%v", names)
	}

	var buf bytes.Buffer
	if err := profile.Show("personal", &buf); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("d@e.f")) {
		t.Fatalf("%s", buf.String())
	}
}
