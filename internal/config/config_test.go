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
	if s.DefaultProfile != "from-env" || s.EnvDefault != "from-env" || s.FileDefault != "from-file" {
		t.Fatalf("%+v", s)
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
	if s.DefaultProfile != "personal" || s.FileDefault != "personal" {
		t.Fatalf("%+v", s)
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

func TestSetDefaultProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(config.EnvDefaultProfile, "")

	if err := config.SetDefaultProfile("work"); err != nil {
		t.Fatal(err)
	}
	s, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if s.FileDefault != "work" {
		t.Fatalf("%+v", s)
	}

	if err := config.SetDefaultProfile(""); err != nil {
		t.Fatal(err)
	}
	s, err = config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if s.FileDefault != "" {
		t.Fatalf("%+v", s)
	}
}
