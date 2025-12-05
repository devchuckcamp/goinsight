package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Default LLM models for each provider
const (
	DefaultOpenAIModel = "gpt-4o-mini"
	DefaultGroqModel   = "llama-3.3-70b-versatile"
	DefaultOllamaModel = "llama3"
)

// Config holds all application configuration loaded from environment variables
type Config struct {
	// Database
	DatabaseURL string

	// LLM Provider
	OpenAIAPIKey string
	GroqAPIKey   string
	OllamaURL    string
	LLMModel     string
	LLMProvider  string

	// Server
	Port string
	Env  string

	// Jira
	JiraBaseURL    string
	JiraEmail      string
	JiraAPIToken   string
	JiraProjectKey string

	// Debug
	Debug bool
}

// Load reads configuration from environment variables
// It attempts to load from .env file first (for local development)
// but doesn't fail if the file doesn't exist (for production/Docker environments)
func Load() (*Config, error) {
	// Try to load .env file - ignore error if it doesn't exist
	// In production/Docker, env vars will be set directly
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		GroqAPIKey:   getEnv("GROQ_API_KEY", ""),
		OllamaURL:    getEnv("OLLAMA_URL", "http://localhost:11434"),
		LLMModel:     getEnv("LLM_MODEL", ""),
		LLMProvider:  getEnv("LLM_PROVIDER", "mock"),
		Port:         getEnv("PORT", "8080"),
		Env:          getEnv("ENV", "development"),
		JiraBaseURL:    getEnv("JIRA_BASE_URL", ""),
		JiraEmail:      getEnv("JIRA_EMAIL", ""),
		JiraAPIToken:   getEnv("JIRA_API_TOKEN", ""),
		JiraProjectKey: getEnv("JIRA_PROJECT_KEY", ""),
		Debug:          getEnvBool("DEBUG", false),
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	// Validate LLM configuration based on provider
	switch cfg.LLMProvider {
	case "openai":
		if cfg.OpenAIAPIKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is required when LLM_PROVIDER=openai")
		}
		if cfg.LLMModel == "" {
			cfg.LLMModel = DefaultOpenAIModel
		}
	case "groq":
		if cfg.GroqAPIKey == "" {
			return nil, fmt.Errorf("GROQ_API_KEY is required when LLM_PROVIDER=groq")
		}
		if cfg.LLMModel == "" {
			cfg.LLMModel = DefaultGroqModel
		}
	case "ollama":
		if cfg.LLMModel == "" {
			cfg.LLMModel = DefaultOllamaModel
		}
	case "mock":
		// No validation needed for mock
	default:
		return nil, fmt.Errorf("invalid LLM_PROVIDER: %s (must be: openai, groq, ollama, or mock)", cfg.LLMProvider)
	}

	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool retrieves a boolean environment variable
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		boolVal, err := strconv.ParseBool(value)
		if err == nil {
			return boolVal
		}
	}
	return defaultValue
}
