package config

import (
	"os"
	"path/filepath"

	"github.com/git-comment/pkg/models"
	"gopkg.in/yaml.v3"
)

const (
	defaultModel           = "llama3.2"
	defaultHost            = "http://localhost:11434"
	defaultTemperature     = 0.2
	defaultMaxOptions      = 3
	defaultUseConventional = true
)

func Load() (*models.Config, error) {
	config := &models.Config{
		Model:                  defaultModel,
		Host:                   defaultHost,
		Temperature:            defaultTemperature,
		MaxOptions:             defaultMaxOptions,
		UseConventionalCommits: defaultUseConventional,
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config, nil
	}

	configPath := filepath.Join(homeDir, ".git_comment.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}
