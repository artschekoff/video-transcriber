package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ChunkDuration != 600 {
		t.Errorf("ChunkDuration: want 600, got %d", cfg.ChunkDuration)
	}
	if cfg.OpenAIAPIKey != "" {
		t.Errorf("OpenAIAPIKey: want empty, got %q", cfg.OpenAIAPIKey)
	}
}

func TestSaveAndLoad_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	origPath := configPath
	configPath = filepath.Join(dir, "settings.json")
	defer func() { configPath = origPath }()

	cfg := AppConfig{OpenAIAPIKey: "sk-test", ChunkDuration: 300}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.OpenAIAPIKey != "sk-test" {
		t.Errorf("OpenAIAPIKey: want sk-test, got %q", loaded.OpenAIAPIKey)
	}
	if loaded.ChunkDuration != 300 {
		t.Errorf("ChunkDuration: want 300, got %d", loaded.ChunkDuration)
	}
}

func TestLoad_ReturnsDefaultWhenNoFile(t *testing.T) {
	dir := t.TempDir()
	origPath := configPath
	configPath = filepath.Join(dir, "nonexistent.json")
	defer func() { configPath = origPath }()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ChunkDuration != 600 {
		t.Errorf("ChunkDuration: want 600, got %d", cfg.ChunkDuration)
	}
}
