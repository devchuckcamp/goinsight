package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/chuckie/goinsight/internal/db"
	"github.com/chuckie/goinsight/internal/domain"
	"github.com/chuckie/goinsight/internal/jira"
	"github.com/chuckie/goinsight/internal/llm"
	"github.com/go-chi/chi/v5"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	dbClient   *db.Client
	llmClient  llm.Client
	jiraClient *jira.Client
}

// NewHandler creates a new HTTP handler
func NewHandler(dbClient *db.Client, llmClient llm.Client, jiraClient *jira.Client) *Handler {
	return &Handler{
		dbClient:   dbClient,
		llmClient:  llmClient,
		jiraClient: jiraClient,
	}
}

// HealthCheck returns the health status of the service
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := h.dbClient.HealthCheck(); err != nil {
		respondJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "unhealthy",
			"error":  "database connection failed",
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// Ask handles the main insight generation endpoint
func (h *Handler) Ask(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req domain.AskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Question == "" {
		respondError(w, http.StatusBadRequest, "Question is required")
		return
	}

	// Step 1: Generate SQL from the question
	sqlQuery, err := h.llmClient.GenerateSQL(r.Context(), req.Question)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate SQL: %v", err))
		return
	}

	// Basic SQL safety check - ensure it's only a SELECT
	normalizedSQL := strings.ToUpper(strings.TrimSpace(sqlQuery))
	if !strings.HasPrefix(normalizedSQL, "SELECT") {
		respondError(w, http.StatusBadRequest, 
			"Unable to generate a valid data query from your question. "+
			"This API analyzes customer feedback data. "+
			"Please ask questions about feedback, such as: "+
			"'What are the most common billing issues?' or "+
			"'Show me negative feedback from enterprise customers.'")
		return
	}

	// Check for dangerous SQL statement keywords as standalone words
	// Use word boundaries to avoid false positives (e.g., "created_at" contains "CREATE")
	dangerous := []string{
		"\\bDROP\\b", "\\bDELETE\\b", "\\bINSERT\\b", "\\bUPDATE\\b",
		"\\bALTER\\b", "\\bCREATE\\s+TABLE\\b", "\\bCREATE\\s+INDEX\\b",
		"\\bTRUNCATE\\b", "\\bEXEC\\b", "\\bEXECUTE\\b",
	}
	for _, pattern := range dangerous {
		// Use regex to match word boundaries
		matched, _ := regexp.MatchString(pattern, normalizedSQL)
		if matched {
			respondError(w, http.StatusBadRequest, "Generated query contains forbidden SQL statement")
			return
		}
	}

	// Step 2: Execute the SQL query
	queryResults, err := h.dbClient.ExecuteQuery(sqlQuery)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Query execution failed: %v. SQL was: %s", err, sqlQuery))
		return
	}

	// Log the columns returned for debugging
	if len(queryResults) > 0 {
		var cols []string
		for col := range queryResults[0] {
			cols = append(cols, col)
		}
		fmt.Printf("DEBUG: Query returned %d rows with columns: %v\n", len(queryResults), cols)
	} else {
		fmt.Printf("DEBUG: Query returned 0 rows. SQL was: %s\n", sqlQuery)
	}

	// Step 3: Generate insights from the results
	insightJSON, err := h.llmClient.GenerateInsight(r.Context(), req.Question, queryResults)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate insights: %v", err))
		return
	}

	// Parse the insight JSON
	var insightResult domain.InsightResult
	if err := json.Unmarshal([]byte(insightJSON), &insightResult); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to parse insight response: %v", err))
		return
	}

	// Step 4: Build the response
	// Limit data preview to first 10 rows for brevity
	dataPreview := queryResults
	if len(dataPreview) > 10 {
		dataPreview = dataPreview[:10]
	}

	response := domain.AskResponse{
		Question:        req.Question,
		SQL:             sqlQuery,
		DataPreview:     dataPreview,
		Summary:         insightResult.Summary,
		Recommendations: insightResult.Recommendations,
		Actions:         insightResult.Actions,
	}

	respondJSON(w, http.StatusOK, response)
}

// CreateJiraTickets handles converting insights into Jira tickets
func (h *Handler) CreateJiraTickets(w http.ResponseWriter, r *http.Request) {
	// Check if Jira is configured
	if h.jiraClient == nil {
		respondError(w, http.StatusServiceUnavailable, "Jira integration is not configured. Set JIRA_BASE_URL, JIRA_EMAIL, and JIRA_API_TOKEN environment variables.")
		return
	}

	// Parse request
	var req domain.JiraTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request has actions
	if len(req.Actions) == 0 {
		respondError(w, http.StatusBadRequest, "No actions provided to convert into tickets")
		return
	}

	// Set defaults for meta if not provided
	// Project key can come from request or will use the client's default
	if req.Meta.ProjectKey == "" {
		// Will be filled by Jira client from environment variable
		req.Meta.ProjectKey = ""
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

	// Step 1: Convert request to JSON for LLM prompt
	requestJSON, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to serialize request: %v", err))
		return
	}

	// Step 2: Use LLM to generate Jira ticket specifications
	prompt := llm.JiraTicketPrompt(string(requestJSON))
	ticketsJSON, err := h.llmClient.Generate(r.Context(), prompt)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate ticket specs: %v", err))
		return
	}

	// Strip markdown code fences if present (LLMs often wrap JSON in ```json ... ```)
	ticketsJSON = strings.TrimSpace(ticketsJSON)
	ticketsJSON = strings.TrimPrefix(ticketsJSON, "```json")
	ticketsJSON = strings.TrimPrefix(ticketsJSON, "```")
	ticketsJSON = strings.TrimSuffix(ticketsJSON, "```")
	ticketsJSON = strings.TrimSpace(ticketsJSON)

	// Parse the ticket specifications
	var ticketsResp domain.JiraTicketsResponse
	if err := json.Unmarshal([]byte(ticketsJSON), &ticketsResp); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to parse ticket specs: %v. Raw response: %s", err, ticketsJSON))
		return
	}

	if len(ticketsResp.Tickets) == 0 {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("LLM did not generate any ticket specifications. Raw response: %s", ticketsJSON))
		return
	}

	// Step 3: Create tickets in Jira
	result, err := h.jiraClient.CreateIssues(ticketsResp.Tickets)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create Jira tickets: %v", err))
		return
	}

	// Step 4: Return results
	respondJSON(w, http.StatusOK, result)
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{
		"error": message,
	})
}

// GetAccountHealth returns ML-based health and risk metrics for a specific account
// GET /api/accounts/{id}/health
func (h *Handler) GetAccountHealth(w http.ResponseWriter, r *http.Request) {
	// Extract account ID from URL path
	accountID := chi.URLParam(r, "id")
	if accountID == "" {
		respondError(w, http.StatusBadRequest, "Account ID is required")
		return
	}

	// Query account risk score
	row, err := h.dbClient.GetAccountRiskScore(accountID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to query account: %v", err))
		return
	}

	var risk domain.AccountRiskScore
	err = row.Scan(
		&risk.AccountID,
		&risk.ChurnProbability,
		&risk.HealthScore,
		&risk.RiskCategory,
		&risk.PredictedAt,
		&risk.ModelVersion,
	)
	if err != nil {
		respondError(w, http.StatusNotFound, fmt.Sprintf("Account not found: %s", accountID))
		return
	}

	// Get recent negative feedback count
	negativeCount, err := h.dbClient.GetRecentNegativeFeedbackCount(accountID)
	if err != nil {
		// Log error but don't fail the request
		negativeCount = 0
	}

	// Build response
	response := domain.AccountHealthResponse{
		AccountID:           risk.AccountID,
		ChurnProbability:    risk.ChurnProbability,
		HealthScore:         risk.HealthScore,
		RiskCategory:        risk.RiskCategory,
		RecentNegativeCount: negativeCount,
		PredictedAt:         risk.PredictedAt.Format("2006-01-02T15:04:05Z07:00"),
		ModelVersion:        risk.ModelVersion,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetProductAreaPriorities returns ML-based priority scores for product areas
// GET /api/priorities/product-areas?segment=enterprise
func (h *Handler) GetProductAreaPriorities(w http.ResponseWriter, r *http.Request) {
	// Get optional segment filter from query params
	segment := r.URL.Query().Get("segment")

	// Query product area impacts
	results, err := h.dbClient.GetProductAreaImpacts(segment)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to query priorities: %v", err))
		return
	}

	// Convert results to domain objects
	var impacts []domain.ProductAreaImpact
	for _, row := range results {
		impact := domain.ProductAreaImpact{
			ProductArea:       row["product_area"].(string),
			Segment:           row["segment"].(string),
			PriorityScore:     row["priority_score"].(float64),
			FeedbackCount:     int(row["feedback_count"].(int64)),
			AvgSentimentScore: row["avg_sentiment_score"].(float64),
			NegativeCount:     int(row["negative_count"].(int64)),
			CriticalCount:     int(row["critical_count"].(int64)),
			ModelVersion:      row["model_version"].(string),
		}
		
		// Parse timestamp
		if predictedAt, ok := row["predicted_at"].(string); ok {
			impact.PredictedAt, _ = time.Parse("2006-01-02T15:04:05Z07:00", predictedAt)
		}
		
		impacts = append(impacts, impact)
	}

	response := domain.ProductAreaPriorityResponse{
		ProductAreas: impacts,
	}

	respondJSON(w, http.StatusOK, response)
}
