package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// configPath is a var so tests can override it.
var configPath = filepath.Join(mustConfigDir(), "video-transcriber", "settings.json")

func mustConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return os.TempDir()
	}
	return dir
}

type AppConfig struct {
	OpenAIAPIKey  string `json:"openaiApiKey"`
	ChunkDuration int    `json:"chunkDuration"`
}

func DefaultConfig() AppConfig {
	return AppConfig{ChunkDuration: 600}
}

func Load() (AppConfig, error) {
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return DefaultConfig(), nil
	}
	if err != nil {
		return DefaultConfig(), err
	}
	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}
	return cfg, nil
}

func Save(cfg AppConfig) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0o600)
}
