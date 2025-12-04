package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/chuckie/goinsight/internal/domain"
)

// Client handles communication with Jira Cloud REST API
type Client struct {
	baseURL    string
	email      string
	apiToken   string
	httpClient *http.Client
}

// NewClient creates a new Jira API client
func NewClient(baseURL, email, apiToken string) *Client {
	return &Client{
		baseURL:  baseURL,
		email:    email,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateIssue creates a single issue in Jira
func (c *Client) CreateIssue(spec domain.JiraTicketSpec) (*domain.JiraCreateResponse, error) {
	// Convert spec to Jira API format
	createReq := domain.JiraCreateRequest{
		Fields: domain.JiraIssueFields{
			Project: domain.JiraProject{
				Key: spec.ProjectKey,
			},
			Summary:     spec.Summary,
			Description: spec.Description,
			IssueType: domain.JiraIssueType{
				Name: spec.IssueType,
			},
			Labels: spec.Labels,
		},
	}

	// Add priority if specified
	if spec.Priority != "" {
		createReq.Fields.Priority = &domain.JiraPriority{
			Name: spec.Priority,
		}
	}

	// Add components if specified
	if len(spec.Components) > 0 {
		createReq.Fields.Components = make([]domain.JiraComponent, len(spec.Components))
		for i, comp := range spec.Components {
			createReq.Fields.Components[i] = domain.JiraComponent{Name: comp}
		}
	}

	// Marshal request body
	body, err := json.Marshal(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/rest/api/2/issue", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.email, c.apiToken)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jira API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var createResp domain.JiraCreateResponse
	if err := json.Unmarshal(respBody, &createResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &createResp, nil
}

// CreateIssues creates multiple issues in Jira
func (c *Client) CreateIssues(specs []domain.JiraTicketSpec) (*domain.JiraCreationResult, error) {
	result := &domain.JiraCreationResult{
		TicketSpecs:    specs,
		CreatedTickets: make([]domain.JiraCreateResponse, 0, len(specs)),
		Errors:         make([]string, 0),
	}

	for i, spec := range specs {
		created, err := c.CreateIssue(spec)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Ticket %d (%s): %v", i+1, spec.Summary, err))
			continue
		}
		result.CreatedTickets = append(result.CreatedTickets, *created)
	}

	return result, nil
}
