package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		APIToken:      "test-token-123",
		DefaultOutput: "text",
	}
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.APIToken != "test-token-123" {
		t.Errorf("APIToken = %q, want %q", loaded.APIToken, "test-token-123")
	}
	if loaded.DefaultOutput != "text" {
		t.Errorf("DefaultOutput = %q, want %q", loaded.DefaultOutput, "text")
	}
}

func TestSaveSetsFilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{APIToken: "secret"}
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}

func TestLoadMissingFileReturnsError(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestEnvVarOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cfg := &Config{APIToken: "file-token", DefaultOutput: "text"}
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	t.Setenv("QWILR_API_TOKEN", "env-token")

	token := ResolveToken(path)
	if token != "env-token" {
		t.Errorf("ResolveToken = %q, want %q", token, "env-token")
	}
}

func TestResolveTokenFallsBackToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cfg := &Config{APIToken: "file-token", DefaultOutput: "text"}
	if err := Save(cfg, path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	os.Unsetenv("QWILR_API_TOKEN")

	token := ResolveToken(path)
	if token != "file-token" {
		t.Errorf("ResolveToken = %q, want %q", token, "file-token")
	}
}
