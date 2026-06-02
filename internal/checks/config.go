package checks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration file structure.
type Config struct {
	DefaultProfile string `yaml:"default_profile" json:"default_profile"`
}

// LoadConfig loads configuration from an explicit file path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	ext := filepath.Ext(path)
	var cfg Config

	switch ext {
	case ".yml", ".yaml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	return &cfg, nil
}

// DiscoverConfig searches for config files in standard locations and returns the first found.
func DiscoverConfig() (*Config, error) {
	// Search order: current directory, then home directory
	searchPaths := []string{
		".gh-repo-health-report.yml",
		".gh-repo-health-report.yaml",
		".gh-repo-health-report.json",
	}

	// Add home directory paths
	if home, err := os.UserHomeDir(); err == nil {
		searchPaths = append(searchPaths,
			filepath.Join(home, ".gh-repo-health-report.yml"),
			filepath.Join(home, ".gh-repo-health-report.yaml"),
			filepath.Join(home, ".gh-repo-health-report.json"),
		)
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return LoadConfig(path)
		}
	}

	// No config file found - not an error
	return nil, nil
}
