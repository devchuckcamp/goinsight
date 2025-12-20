package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chuckie/goinsight/internal/domain"
	"github.com/chuckie/goinsight/internal/jira"
	"github.com/chuckie/goinsight/internal/llm"
	"github.com/chuckie/goinsight/internal/repository"
	"github.com/chuckie/goinsight/internal/service"
	"github.com/go-chi/chi/v5"
)

// ServiceHandler is the refactored handler using the service layer
type ServiceHandler struct {
	feedbackService *service.FeedbackService
	jiraClient      *jira.Client
}

// NewServiceHandler creates a new service-based HTTP handler
func NewServiceHandler(feedbackService *service.FeedbackService, jiraClient *jira.Client) *ServiceHandler {
	return &ServiceHandler{
		feedbackService: feedbackService,
		jiraClient:      jiraClient,
	}
}

// HealthCheck returns the health status of the service
func (h *ServiceHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// Ask handles the main insight generation endpoint using the service layer
func (h *ServiceHandler) Ask(w http.ResponseWriter, r *http.Request) {
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

	// Measure execution time
	start := time.Now()

	// Use service layer to analyze feedback
	response, err := h.feedbackService.AnalyzeFeedback(r.Context(), req.Question)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Log execution time
	duration := time.Since(start)
	fmt.Printf("Ask endpoint completed in %v\n", duration)

	respondJSON(w, http.StatusOK, response)
}

// CreateJiraTickets handles converting insights into Jira tickets
func (h *ServiceHandler) CreateJiraTickets(w http.ResponseWriter, r *http.Request) {
	// Check if Jira is configured
	if h.jiraClient == nil {
		respondError(w, http.StatusServiceUnavailable,
			"Jira integration is not configured. Set JIRA_BASE_URL, JIRA_EMAIL, and JIRA_API_TOKEN environment variables.")
		return
	}

	// Parse request
	var req domain.JiraTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Convert to service request and create tickets
	serviceReq := service.JiraTicketRequest{
		Summary:         req.Summary,
		Recommendations: req.Recommendations,
		Actions:         req.Actions,
		Meta: service.JiraMetadata{
			ProjectKey:       req.Meta.ProjectKey,
			DefaultIssueType: req.Meta.DefaultIssueType,
			DefaultLabels:    req.Meta.DefaultLabels,
		},
	}

	result, err := h.feedbackService.CreateJiraTickets(r.Context(), serviceReq)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// GetAccountHealth retrieves ML predictions for a specific account
func (h *ServiceHandler) GetAccountHealth(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "id")
	if accountID == "" {
		respondError(w, http.StatusBadRequest, "Account ID is required")
		return
	}

	score, err := h.feedbackService.GetAccountRiskScore(r.Context(), accountID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get account health: %v", err))
		return
	}

	if score == nil {
		respondError(w, http.StatusNotFound, "Account not found")
		return
	}

	respondJSON(w, http.StatusOK, score)
}

// GetProductAreaPriorities retrieves ML predictions for product area priorities
func (h *ServiceHandler) GetProductAreaPriorities(w http.ResponseWriter, r *http.Request) {
	segment := r.URL.Query().Get("segment")

	impacts, err := h.feedbackService.GetProductAreaImpacts(r.Context(), segment)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get product area impacts: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, impacts)
}


// LegacyHandlerAdapter adapts the legacy Handler to work with the service layer
// This allows for gradual migration to the new architecture
type LegacyHandlerAdapter struct {
	*Handler
	feedbackService *service.FeedbackService
}

// NewLegacyHandlerAdapter creates an adapter for the legacy handler
func NewLegacyHandlerAdapter(
	handler *Handler,
	repo repository.FeedbackRepository,
	llmClient llm.Client,
	jiraClient *jira.Client,
) *LegacyHandlerAdapter {
	return &LegacyHandlerAdapter{
		Handler:         handler,
		feedbackService: service.NewFeedbackService(repo, llmClient, jiraClient),
	}
}
