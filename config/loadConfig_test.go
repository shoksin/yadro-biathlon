package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	configContent := `{
		"laps": 2,
		"lapLen": 3650,
		"penaltyLen": 100,
		"firingLines": 1,
		"start": "09:35:00",
		"startDelta": "00:01:30"
	}`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if config.Laps != 2 {
		t.Errorf("Expected Laps=2, got %d", config.Laps)
	}
	if config.LapLen != 3650 {
		t.Errorf("Expected LapLen=3650, got %d", config.LapLen)
	}
	if config.PenaltyLen != 100 {
		t.Errorf("Expected PenaltyLen=100, got %d", config.PenaltyLen)
	}
	if config.Start != "09:35:00" {
		t.Errorf("Expected Start='09:35:00', got '%s'", config.Start)
	}
	if config.StartDelta != "00:01:30" {
		t.Errorf("Expected StartDelta='00:01:30', got '%s'", config.StartDelta)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("non_existent_file.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}
