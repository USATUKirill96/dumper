package app

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"dumper/config/env"
)

// Config represents the main application configuration
type Config struct {
	Environments *env.Config
	DumpsDir     string
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string, dumpsDir string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var environments env.Config
	if err := yaml.Unmarshal(data, &environments); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	if len(environments.Environments) == 0 {
		return nil, fmt.Errorf("no environments found in config")
	}

	// Create dumps directory if it doesn't exist
	if err := os.MkdirAll(dumpsDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating dumps directory: %w", err)
	}

	return &Config{
		Environments: &environments,
		DumpsDir:     dumpsDir,
	}, nil
}

// GetEnvironment returns an environment by name
func (c *Config) GetEnvironment(name string) *env.Environment {
	return c.Environments.GetEnvironmentByName(name)
}

// GetEnvironments returns all environments
func (c *Config) GetEnvironments() []env.Environment {
	return c.Environments.Environments
}
