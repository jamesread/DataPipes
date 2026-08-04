package config

import (
	"os"
	"path"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestReloadConfigFallsBackToSample(t *testing.T) {
	t.Setenv("DATAPIPES_CONFIG", "")
	t.Setenv("DATACLEANER_CONFIG", "")
	t.Setenv("DATA_CLEANER_CONFIG", "")
	t.Setenv("HOME", t.TempDir())

	cfg := ReloadConfig()
	loadedPath, err := LoadStatus()

	if err != nil {
		t.Fatalf("expected no load error, got %v", err)
	}
	if loadedPath != sampleConfigPathLabel {
		t.Fatalf("config path = %q, want %q", loadedPath, sampleConfigPathLabel)
	}
	if cfg == nil || cfg.Connections["sample_csv"] == nil {
		t.Fatal("expected sample_csv connection from built-in sample")
	}
	job := cfg.Jobs["sample"]
	if job == nil {
		t.Fatal("expected sample job from built-in sample")
	}
	if job.Extract != "sample_csv" || job.Load != "download_csv" {
		t.Fatalf("sample job = %+v", job)
	}
	if cfg.Network == nil || cfg.Network.BindProxy != ":8080" {
		t.Fatalf("network = %+v", cfg.Network)
	}
}

func TestReloadConfigUsesHomeFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("DATAPIPES_CONFIG", "")
	t.Setenv("DATACLEANER_CONFIG", "")
	t.Setenv("DATA_CLEANER_CONFIG", "")
	t.Setenv("HOME", home)

	content := []byte("network:\n  bindproxy: \":9090\"\n")
	if err := os.WriteFile(path.Join(home, ".datapipes-config.yaml"), content, 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := ReloadConfig()
	loadedPath, err := LoadStatus()
	wantPath := path.Join(home, ".datapipes-config.yaml")

	if err != nil {
		t.Fatalf("expected no load error, got %v", err)
	}
	if loadedPath != wantPath {
		t.Fatalf("config path = %q, want %q", loadedPath, wantPath)
	}
	if cfg.Network == nil || cfg.Network.BindProxy != ":9090" {
		t.Fatalf("network = %+v", cfg.Network)
	}
}

func TestSampleConfigValidates(t *testing.T) {
	cfg := newDefaultConfig()
	if err := yaml.UnmarshalStrict(sampleConfigYAML, &cfg); err != nil {
		t.Fatalf("sample config parse: %v", err)
	}
	if errs := validateConfig(cfg); len(errs) > 0 {
		t.Fatalf("sample config validation: %v", errs)
	}
}
