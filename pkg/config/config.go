package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type EnvironmentConfig struct {
	Name       string          `yaml:"name"`
	Proccesses []ProcessConfig `yaml:"processes"`
}

type ProcessConfig struct {
	Name    string   `yaml:"name"`
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
	WorkDir string   `yaml:"workDir"`
}

func LoadConfig(path string) (*EnvironmentConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var c EnvironmentConfig
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return &c, nil
}
