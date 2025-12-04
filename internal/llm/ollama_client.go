package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaClient implements the Client interface for local Ollama
// Ollama is completely free and runs models locally
type OllamaClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewOllamaClient creates a new Ollama client
// baseURL is typically http://localhost:11434
// model examples: llama3, mixtral, codellama
// Install: https://ollama.ai
func NewOllamaClient(baseURL, model string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	// Model should be set by config, but provide fallback
	if model == "" {
		model = "llama3"
	}
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // Local inference can be slower
		},
	}
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// GenerateSQL implements the Client interface
func (c *OllamaClient) GenerateSQL(ctx context.Context, question string) (string, error) {
	systemPrompt := SQLGenerationPrompt()
	prompt := fmt.Sprintf("%s\n\nUser question: %s\n\nSQL query:", systemPrompt, question)

	reqBody := ollamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	response, err := c.makeRequest(ctx, reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to generate SQL: %w", err)
	}

	return response, nil
}

// GenerateInsight implements the Client interface
func (c *OllamaClient) GenerateInsight(ctx context.Context, question string, queryResults []map[string]any) (string, error) {
	prompt := InsightGenerationPrompt(question, queryResults)

	reqBody := ollamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	response, err := c.makeRequest(ctx, reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to generate insight: %w", err)
	}

	return response, nil
}

// Generate implements the Client interface - sends prompt directly to LLM
func (c *OllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := ollamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	response, err := c.makeRequest(ctx, reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return response, nil
}

func (c *OllamaClient) makeRequest(ctx context.Context, reqBody ollamaRequest) (string, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request (is Ollama running?): %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response ollamaResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.Response, nil
}
