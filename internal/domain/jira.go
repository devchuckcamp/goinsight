package domain

import "strings"

// JiraTicketRequest represents the input to create Jira tickets from insights
type JiraTicketRequest struct {
	Question        string         `json:"question"`
	Summary         string         `json:"summary"`
	Recommendations []string       `json:"recommendations"`
	Actions         []ActionItem   `json:"actions"`
	Meta            JiraTicketMeta `json:"meta"`
}

// JiraTicketMeta contains Jira-specific configuration
type JiraTicketMeta struct {
	ProjectKey       string   `json:"project_key"`
	DefaultIssueType string   `json:"default_issue_type"`
	DefaultLabels    []string `json:"default_labels"`
}

// JiraTicketSpec represents a single Jira ticket specification
type JiraTicketSpec struct {
	ProjectKey  string   `json:"project_key"`
	IssueType   string   `json:"issue_type"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	Labels      []string `json:"labels"`
	Components  []string `json:"components"`
	EpicLink    *string  `json:"epic_link"`
}

// JiraTicketsResponse is returned by the LLM
type JiraTicketsResponse struct {
	Tickets []JiraTicketSpec `json:"tickets"`
}

// JiraCreateRequest is the actual Jira Cloud API request format
type JiraCreateRequest struct {
	Fields JiraIssueFields `json:"fields"`
}

// JiraIssueFields matches Jira Cloud REST API structure
type JiraIssueFields struct {
	Project     JiraProject       `json:"project"`
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	IssueType   JiraIssueType     `json:"issuetype"`
	Priority    *JiraPriority     `json:"priority,omitempty"`
	Labels      []string          `json:"labels,omitempty"`
	Components  []JiraComponent   `json:"components,omitempty"`
}

// JiraProject represents a Jira project reference
type JiraProject struct {
	Key string `json:"key"`
}

// JiraIssueType represents a Jira issue type
type JiraIssueType struct {
	Name string `json:"name"`
}

// JiraPriority represents a Jira priority
type JiraPriority struct {
	Name string `json:"name"`
}

// JiraComponent represents a Jira component
type JiraComponent struct {
	Name string `json:"name"`
}

// JiraCreateResponse is returned by Jira after creating an issue
type JiraCreateResponse struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Self string `json:"self"`
}

// JiraCreationResult tracks the result of creating multiple tickets
type JiraCreationResult struct {
	TicketSpecs    []JiraTicketSpec       `json:"ticket_specs"`
	CreatedTickets []JiraCreateResponse   `json:"created_tickets"`
	Errors         []string               `json:"errors,omitempty"`
}

// CalculateMagnitude computes a priority score (0-10) for an action item
// based on keyword analysis and text characteristics
func CalculateMagnitude(action ActionItem, summary string, recommendations []string) float64 {
	var score float64 = 5.0 // baseline: Medium priority

	// High urgency keywords in title or description
	urgentKeywords := []string{
		"critical", "urgent", "emergency", "blocker", "security", "data loss",
		"crash", "down", "outage", "breach", "vulnerability", "exploit",
	}
	for _, keyword := range urgentKeywords {
		if contains(action.Title, keyword) || contains(action.Description, keyword) {
			score += 2.5
			break
		}
	}

	// High impact keywords
	impactKeywords := []string{
		"revenue", "customer", "payment", "billing", "refund", "loss",
		"compliance", "legal", "audit", "performance", "scalability",
	}
	for _, keyword := range impactKeywords {
		if contains(action.Title, keyword) || contains(action.Description, keyword) {
			score += 1.5
			break
		}
	}

	// Investigation/analysis tasks (typically medium-high)
	investigateKeywords := []string{"investigate", "analyze", "research", "identify"}
	for _, keyword := range investigateKeywords {
		if contains(action.Title, keyword) {
			score += 1.0
			break
		}
	}

	// Documentation/update tasks (typically medium-low)
	docKeywords := []string{"documentation", "update", "review", "document"}
	for _, keyword := range docKeywords {
		if contains(action.Title, keyword) {
			score -= 1.0
			break
		}
	}

	// Check if action appears in multiple recommendations (higher priority)
	mentionCount := 0
	for _, rec := range recommendations {
		for _, keyword := range []string{action.Title} {
			if contains(rec, keyword) {
				mentionCount++
				break
			}
		}
	}
	if mentionCount > 0 {
		score += float64(mentionCount) * 0.5
	}

	// Clamp score between 0 and 10
	if score > 10 {
		score = 10
	}
	if score < 0 {
		score = 0
	}

	return score
}

// contains is a case-insensitive string contains check
func contains(text, substr string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(substr))
}
