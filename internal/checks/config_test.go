package checks

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadConfigYAML verifies loading a valid YAML config file.
func TestLoadConfigYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	yamlContent := `default_profile: internal-service`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.DefaultProfile != "internal-service" {
		t.Errorf("Expected default_profile 'internal-service', got %q", cfg.DefaultProfile)
	}
}

// TestLoadConfigJSON verifies loading a valid JSON config file.
func TestLoadConfigJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	jsonContent := `{"default_profile": "open-source"}`
	if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.DefaultProfile != "open-source" {
		t.Errorf("Expected default_profile 'open-source', got %q", cfg.DefaultProfile)
	}
}

// TestLoadConfigInvalidYAML verifies error handling for invalid YAML.
func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	invalidYAML := `default_profile: [invalid: yaml: syntax`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

// TestLoadConfigInvalidJSON verifies error handling for invalid JSON.
func TestLoadConfigInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	invalidJSON := `{"default_profile": invalid json}`
	if err := os.WriteFile(configPath, []byte(invalidJSON), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// TestLoadConfigMissingFile verifies error handling for missing file.
func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/config.yml")
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

// TestDiscoverConfigNoFile verifies that DiscoverConfig returns nil when no config exists.
func TestDiscoverConfigNoFile(t *testing.T) {
	// Change to a temp directory with no config files
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	cfg, err := DiscoverConfig()
	if err != nil {
		t.Errorf("DiscoverConfig returned error: %v", err)
	}
	if cfg != nil {
		t.Errorf("Expected nil config, got %v", cfg)
	}
}

// TestDiscoverConfigCurrentDir verifies discovery in current directory.
func TestDiscoverConfigCurrentDir(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create config in current directory
	yamlContent := `default_profile: application`
	if err := os.WriteFile(".gh-repo-health-report.yml", []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	cfg, err := DiscoverConfig()
	if err != nil {
		t.Fatalf("DiscoverConfig failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}
	if cfg.DefaultProfile != "application" {
		t.Errorf("Expected default_profile 'application', got %q", cfg.DefaultProfile)
	}
}

// TestDiscoverConfigPrecedence verifies that current directory takes precedence.
func TestDiscoverConfigPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create both YAML and JSON in current directory - YAML should win
	yamlContent := `default_profile: prototype`
	jsonContent := `{"default_profile": "archived"}`

	if err := os.WriteFile(".gh-repo-health-report.yml", []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create YAML config: %v", err)
	}
	if err := os.WriteFile(".gh-repo-health-report.json", []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to create JSON config: %v", err)
	}

	cfg, err := DiscoverConfig()
	if err != nil {
		t.Fatalf("DiscoverConfig failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}
	if cfg.DefaultProfile != "prototype" {
		t.Errorf("Expected YAML to take precedence, got %q", cfg.DefaultProfile)
	}
}
