package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/git-comment/pkg/models"
)

type Client struct {
	host                 string
	model                string
	temp                 float64
	maxOptions           int
	useConventional      bool
	client               *http.Client
}

func NewClient(host, model string, temperature float64, maxOptions int, useConventional bool) *Client {
	return &Client{
		host:            host,
		model:           model,
		temp:            temperature,
		maxOptions:      maxOptions,
		useConventional: useConventional,
		client:          &http.Client{},
	}
}

func (c *Client) buildSystemPrompt() string {
	prompt := `You are an expert software engineer responsible for writing Git commit messages.

Analyze the provided Git diff and identify the primary intent of the changes.

Do not explain the diff.
Do not include numbering or markdown.
Output only one commit message per line.
Prefer clarity over verbosity.
Each message must be on its own line, maximum 72 characters, no trailing period.`

	if c.useConventional {
		prompt += "\nFollow Conventional Commits format. eg: type(scope): description"
	}

	prompt += fmt.Sprintf("\nReturn up to %d concise commit messages.", c.maxOptions)

	return prompt
}

func (c *Client) CheckAvailability() error {
	resp, err := c.client.Get(c.host + "/api/tags")
	if err != nil {
		return fmt.Errorf("unable to connect to Ollama at %s: %w", c.host, err)
	}
	defer resp.Body.Close()
	return nil
}

func (c *Client) ListModels() ([]string, error) {
	resp, err := c.client.Get(c.host + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var modelsResp models.ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, err
	}

	var modelNames []string
	for _, m := range modelsResp.Models {
		modelNames = append(modelNames, m.Name)
	}
	return modelNames, nil
}

func (c *Client) ValidateModel() error {
	models, err := c.ListModels()
	if err != nil {
		return err
	}

	for _, m := range models {
		if m == c.model || strings.HasPrefix(m, c.model+":") {
			return nil
		}
	}
	return fmt.Errorf("model %s not found", c.model)
}

func (c *Client) GenerateCommitMessages(diff string) ([]string, error) {
	maxChars := 32000
	if len(diff) > maxChars {
		diff = diff[:maxChars] + "\n\n[Diff truncated]"
	}

	request := models.OllamaRequest{
		Model:  c.model,
		Stream: false,
		Options: &models.Options{
			Temperature: c.temp,
		},
		Messages: []models.Message{
			{Role: "system", Content: c.buildSystemPrompt()},
			{Role: "user", Content: fmt.Sprintf("Analyze this Git diff and generate commit messages:\n\n%s", diff)},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.host+"/api/chat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ollamaResp models.OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, err
	}

	messages := c.parseResponse(ollamaResp.Message.Content)
	return messages, nil
}

func (c *Client) FormatCommitMessage(rawMessage string, diff string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert software engineer. A developer wrote this commit message:

"%s"

Based on the Git diff below, rewrite the message to be clear, concise, and follow Conventional Commits format (type(scope): description). Output only the reformatted commit message, nothing else.

Git diff:
%s`, rawMessage, diff)

	request := models.OllamaRequest{
		Model:  c.model,
		Stream: false,
		Options: &models.Options{
			Temperature: c.temp,
		},
		Messages: []models.Message{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Post(c.host+"/api/chat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ollamaResp models.OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", err
	}

	return strings.TrimSpace(ollamaResp.Message.Content), nil
}

func (c *Client) parseResponse(content string) []string {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	var messages []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(line, "1.")
		line = strings.TrimPrefix(line, "2.")
		line = strings.TrimPrefix(line, "3.")
		line = strings.TrimPrefix(line, "4.")
		line = strings.TrimPrefix(line, "5.")
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		line = strings.TrimSpace(line)
		line = strings.TrimRight(line, ".")

		if len(line) > 0 && len(line) <= 72 {
			messages = append(messages, line)
		}
	}

	if len(messages) > c.maxOptions {
		messages = messages[:c.maxOptions]
	}
	return messages
}
