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

const systemPrompt = `You are an expert software engineer responsible for writing Git commit messages.

Analyze the provided Git diff and identify the primary intent of the changes
(feature, fix, refactor, chore, docs, style, test, perf, etc.).

Return exactly three distinct commit message options, each offering a
different valid framing of the same change (e.g. one focused on the "what",
one on the "why", one on the affected component/scope).

Follow the Conventional Commits format strictly: type(scope): description
- type: feat, fix, refactor, chore, docs, style, test, perf, or build
- scope: the affected module, file, or component (lowercase, no spaces)
- description: imperative mood, lowercase, no trailing period

Formatting rules:
- Output only the three messages, nothing else
- Number them 1, 2, 3, each on its own line, followed by the message on the next line
- No explanations, no markdown, no extra commentary
- Each commit message must be 72 characters or fewer
- Prefer clarity and specificity over generic phrasing (avoid vague verbs like "update" or "change" unless nothing more specific applies)

Output format example:
1.
feat(auth): persist user session using Supabase
2.
fix(camera): prevent null controller during initialization
3.
refactor(auth): connect login flow with backend session storage`

type Client struct {
	host   string
	model  string
	temp   float64
	client *http.Client
}

func NewClient(host, model string, temperature float64) *Client {
	return &Client{
		host:   host,
		model:  model,
		temp:   temperature,
		client: &http.Client{},
	}
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
			{Role: "system", Content: systemPrompt},
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

	messages := parseResponse(ollamaResp.Message.Content)
	return messages, nil
}

func parseResponse(content string) []string {
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
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimSpace(line)

		if len(line) > 0 && len(line) <= 72 && !strings.HasSuffix(line, ".") {
			messages = append(messages, line)
		}
	}

	if len(messages) > 3 {
		messages = messages[:3]
	}
	return messages
}
