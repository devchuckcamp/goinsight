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

// GroqClient implements the Client interface for Groq API
// Groq offers free tier with fast inference
type GroqClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewGroqClient creates a new Groq client
// apiKey should be loaded from GROQ_API_KEY environment variable
// Get free API key at: https://console.groq.com
func NewGroqClient(apiKey, model string) *GroqClient {
	// Model should be set by config, but provide fallback
	if model == "" {
		model = "llama-3.3-70b-versatile"
	}
	return &GroqClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type groqRequest struct {
	Model    string        `json:"model"`
	Messages []groqMessage `json:"messages"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponse struct {
	Choices []struct {
		Message groqMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// GenerateSQL implements the Client interface
func (c *GroqClient) GenerateSQL(ctx context.Context, question string) (string, error) {
	systemPrompt := SQLGenerationPrompt()

	reqBody := groqRequest{
		Model: c.model,
		Messages: []groqMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: question},
		},
	}

	response, err := c.makeRequest(ctx, reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to generate SQL: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from Groq")
	}

	return response.Choices[0].Message.Content, nil
}

// GenerateInsight implements the Client interface
func (c *GroqClient) GenerateInsight(ctx context.Context, question string, queryResults []map[string]any) (string, error) {
	prompt := InsightGenerationPrompt(question, queryResults)

	reqBody := groqRequest{
		Model: c.model,
		Messages: []groqMessage{
			{Role: "user", Content: prompt},
		},
	}

	response, err := c.makeRequest(ctx, reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to generate insight: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from Groq")
	}

	return response.Choices[0].Message.Content, nil
}

// Generate implements the Client interface - sends prompt directly to LLM
func (c *GroqClient) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := groqRequest{
		Model: c.model,
		Messages: []groqMessage{
			{Role: "user", Content: prompt},
		},
	}

	response, err := c.makeRequest(ctx, reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from Groq")
	}

	return response.Choices[0].Message.Content, nil
}

func (c *GroqClient) makeRequest(ctx context.Context, reqBody groqRequest) (*groqResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Groq API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response groqResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("Groq API error: %s", response.Error.Message)
	}

	return &response, nil
}
