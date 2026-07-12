package config

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-comment/pkg/models"
	"gopkg.in/yaml.v3"
)

const (
	defaultHost            = "http://localhost:11434"
	defaultTemperature     = 0.2
	defaultMaxOptions      = 3
	defaultUseConventional = true
)

func Load() (*models.Config, error) {
	config := &models.Config{
		Host:                   defaultHost,
		Temperature:            defaultTemperature,
		MaxOptions:             defaultMaxOptions,
		UseConventionalCommits: defaultUseConventional,
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		config.Model = detectFirstModel(config.Host)
		return config, nil
	}

	configPath := filepath.Join(homeDir, ".git_comment.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			config.Model = detectFirstModel(config.Host)
			if writeErr := writeDefaultConfig(configPath, config); writeErr != nil {
				return config, nil
			}
			return config, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	if config.MaxOptions < 1 {
		config.MaxOptions = defaultMaxOptions
	}
	if config.Temperature <= 0 || config.Temperature > 2 {
		config.Temperature = defaultTemperature
	}

	return config, nil
}

func detectFirstModel(host string) string {
	resp, err := http.Get(host + "/api/tags")
	if err != nil {
		return "qwen2.5-coder"
	}
	defer resp.Body.Close()

	var modelsResp models.ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return "qwen2.5-coder"
	}

	if len(modelsResp.Models) > 0 {
		name := modelsResp.Models[0].Name
		name = strings.TrimSuffix(name, ":latest")
		return name
	}

	return "qwen2.5-coder"
}

func writeDefaultConfig(path string, config *models.Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
