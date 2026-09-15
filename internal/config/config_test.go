package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/prajwal/gitp/internal/config"
)

func TestLoad_FromEnv(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(config.EnvDefaultProfile, "from-env")

	// Even with a config file, env wins.
	dir := filepath.Join(home, config.DirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(dir, config.FileName)
	if err := os.WriteFile(cfg, []byte("default_profile = from-file\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if s.DefaultProfile != "from-env" {
		t.Fatalf("got %q", s.DefaultProfile)
	}
}

func TestLoad_FromFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(config.EnvDefaultProfile, "")

	dir := filepath.Join(home, config.DirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(dir, config.FileName)
	content := "# comment\n\ndefault_profile = personal\n; another\n"
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if s.DefaultProfile != "personal" {
		t.Fatalf("got %q", s.DefaultProfile)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(config.EnvDefaultProfile, "")

	s, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if s.DefaultProfile != "" {
		t.Fatalf("got %q", s.DefaultProfile)
	}
}
