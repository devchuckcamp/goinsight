package llm

import "context"

// Client is the interface that all LLM providers must implement
type Client interface {
	// GenerateSQL takes a natural language question and generates a safe SQL SELECT query
	GenerateSQL(ctx context.Context, question string) (string, error)

	// GenerateInsight takes the original question and query results, returns analysis
	GenerateInsight(ctx context.Context, question string, queryResults []map[string]any) (string, error)

	// Generate sends a prompt directly to the LLM without any wrapping
	Generate(ctx context.Context, prompt string) (string, error)
}
