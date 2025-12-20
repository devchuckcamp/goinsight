package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/chuckie/goinsight/internal/domain"
	"github.com/chuckie/goinsight/internal/jira"
	"github.com/chuckie/goinsight/internal/llm"
	"github.com/chuckie/goinsight/internal/repository"
)

// FeedbackService orchestrates business logic for feedback analysis
type FeedbackService struct {
	repo       repository.FeedbackRepository
	llmClient  llm.Client
	jiraClient *jira.Client
}

// NewFeedbackService creates a new feedback service
func NewFeedbackService(
	repo repository.FeedbackRepository,
	llmClient llm.Client,
	jiraClient *jira.Client,
) *FeedbackService {
	return &FeedbackService{
		repo:       repo,
		llmClient:  llmClient,
		jiraClient: jiraClient,
	}
}

// QueryRequest represents a question to analyze
type QueryRequest struct {
	Question string
}

// AnalyzeFeedback orchestrates the full workflow: SQL generation, execution, and insight generation
func (s *FeedbackService) AnalyzeFeedback(ctx context.Context, question string) (*domain.AskResponse, error) {
	// Validate input
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}

	// Step 1: Generate SQL from the question
	sqlQuery, err := s.llmClient.GenerateSQL(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SQL: %w", err)
	}

	// Step 2: Validate SQL for safety
	if err := s.validateSQL(sqlQuery); err != nil {
		return nil, err
	}

	// Step 3: Execute the SQL query
	queryResults, err := s.repo.QueryFeedback(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	// Step 4: Generate insights from the results
	insightJSON, err := s.llmClient.GenerateInsight(ctx, question, queryResults)
	if err != nil {
		return nil, fmt.Errorf("failed to generate insights: %w", err)
	}

	// Step 5: Parse the insight JSON
	var insightResult domain.InsightResult
	if err := json.Unmarshal([]byte(insightJSON), &insightResult); err != nil {
		return nil, fmt.Errorf("failed to parse insight response: %w", err)
	}

	// Step 6: Build the response
	dataPreview := queryResults
	if len(dataPreview) > 10 {
		dataPreview = dataPreview[:10]
	}

	response := &domain.AskResponse{
		Question:        question,
		SQL:             sqlQuery,
		DataPreview:     dataPreview,
		Summary:         insightResult.Summary,
		Recommendations: insightResult.Recommendations,
		Actions:         insightResult.Actions,
	}

	return response, nil
}

// validateSQL performs safety checks on the generated SQL query
func (s *FeedbackService) validateSQL(sqlQuery string) error {
	normalizedSQL := strings.ToUpper(strings.TrimSpace(sqlQuery))

	// Ensure it's only a SELECT
	if !strings.HasPrefix(normalizedSQL, "SELECT") {
		return fmt.Errorf(
			"unable to generate a valid data query. "+
				"This API analyzes customer feedback data. "+
				"Please ask questions about feedback, such as: "+
				"'What are the most common billing issues?' or "+
				"'Show me negative feedback from enterprise customers.'",
		)
	}

	// Check for dangerous SQL statement keywords as standalone words
	dangerous := []string{
		"\\bDROP\\b", "\\bDELETE\\b", "\\bINSERT\\b", "\\bUPDATE\\b",
		"\\bALTER\\b", "\\bCREATE\\s+TABLE\\b", "\\bCREATE\\s+INDEX\\b",
		"\\bTRUNCATE\\b", "\\bEXEC\\b", "\\bEXECUTE\\b",
	}

	for _, pattern := range dangerous {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("internal error compiling SQL validation pattern: %w", err)
		}

		if re.MatchString(normalizedSQL) {
			return fmt.Errorf("generated query contains forbidden SQL statement")
		}
	}

	return nil
}

// GetAccountRiskScore retrieves ML predictions for a specific account
func (s *FeedbackService) GetAccountRiskScore(ctx context.Context, accountID string) (*domain.AccountRiskScore, error) {
	return s.repo.GetAccountRiskScore(ctx, accountID)
}

// GetRecentNegativeFeedbackCount counts recent negative feedback for an account
func (s *FeedbackService) GetRecentNegativeFeedbackCount(ctx context.Context, accountID string) (int, error) {
	return s.repo.GetRecentNegativeFeedbackCount(ctx, accountID)
}

// GetProductAreaImpacts retrieves ML predictions for product area priorities
func (s *FeedbackService) GetProductAreaImpacts(ctx context.Context, segment string) ([]map[string]any, error) {
	return s.repo.GetProductAreaImpacts(ctx, segment)
}

// JiraTicketRequest wraps the action items and metadata for Jira ticket creation
type JiraTicketRequest struct {
	Summary         string
	Recommendations []string
	Actions         []domain.ActionItem
	Meta            JiraMetadata
}

// JiraMetadata holds Jira-specific configuration
type JiraMetadata struct {
	ProjectKey       string
	DefaultIssueType string
	DefaultLabels    []string
}

// CreateJiraTickets converts insight actions into Jira tickets
func (s *FeedbackService) CreateJiraTickets(ctx context.Context, req JiraTicketRequest) (*domain.JiraCreationResult, error) {
	// Validate Jira is configured
	if s.jiraClient == nil {
		return nil, fmt.Errorf("jira integration is not configured")
	}

	// Validate request
	if len(req.Actions) == 0 {
		return nil, fmt.Errorf("no actions provided to convert into tickets")
	}

	// Validate required Jira meta and set defaults
	if strings.TrimSpace(req.Meta.ProjectKey) == "" {
		return nil, fmt.Errorf("jira project key is required")
	}
	if req.Meta.DefaultIssueType == "" {
		req.Meta.DefaultIssueType = "Story"
	}
	if len(req.Meta.DefaultLabels) == 0 {
		req.Meta.DefaultLabels = []string{"feedback", "ai-insight"}
	}

	// Calculate magnitude for each action item
	for i := range req.Actions {
		req.Actions[i].Magnitude = domain.CalculateMagnitude(
			req.Actions[i],
			req.Summary,
			req.Recommendations,
		)
	}

	// Convert request to JSON for LLM prompt
	domainReq := domain.JiraTicketRequest{
		Summary:         req.Summary,
		Recommendations: req.Recommendations,
		Actions:         req.Actions,
		Meta: domain.JiraTicketMeta{
			ProjectKey:       req.Meta.ProjectKey,
			DefaultIssueType: req.Meta.DefaultIssueType,
			DefaultLabels:    req.Meta.DefaultLabels,
		},
	}

	requestJSON, err := json.MarshalIndent(domainReq, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}

	// Use LLM to generate Jira ticket specifications
	prompt := llm.JiraTicketPrompt(string(requestJSON))
	ticketsJSON, err := s.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ticket specs: %w", err)
	}

	// Strip markdown code fences if present
	ticketsJSON = strings.TrimSpace(ticketsJSON)
	ticketsJSON = strings.TrimPrefix(ticketsJSON, "```json")
	ticketsJSON = strings.TrimPrefix(ticketsJSON, "```")
	ticketsJSON = strings.TrimSuffix(ticketsJSON, "```")
	ticketsJSON = strings.TrimSpace(ticketsJSON)

	// Parse the ticket specifications
	var ticketsResp domain.JiraTicketsResponse
	if err := json.Unmarshal([]byte(ticketsJSON), &ticketsResp); err != nil {
		return nil, fmt.Errorf("failed to parse ticket specs: %w", err)
	}

	if len(ticketsResp.Tickets) == 0 {
		return nil, fmt.Errorf("LLM did not generate any ticket specifications")
	}

	// Create tickets in Jira
	result, err := s.jiraClient.CreateIssues(ticketsResp.Tickets)
	if err != nil {
		return nil, fmt.Errorf("failed to create jira tickets: %w", err)
	}

	return result, nil
}
