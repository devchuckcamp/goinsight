package mocks

import (
	"context"
	"errors"

	"github.com/chuckie/goinsight/internal/domain"
	"github.com/chuckie/goinsight/internal/repository"
)

// MockFeedbackRepository is a mock implementation of FeedbackRepository for testing
type MockFeedbackRepository struct {
	// Configure return values
	QueryFeedbackResult          []map[string]any
	QueryFeedbackErr             error
	GetAccountRiskScoreResult    *domain.AccountRiskScore
	GetAccountRiskScoreErr       error
	GetRecentNegativeCountResult int
	GetRecentNegativeCountErr    error
	GetProductAreaImpactsResult  []map[string]any
	GetProductAreaImpactsErr     error
	GetFeedbackCountResult       int
	GetFeedbackCountErr          error
	GetFeedbackEnrichedResult    []domain.FeedbackEnriched
	GetFeedbackEnrichedErr       error

	// Track calls for assertion
	QueryFeedbackCalled          bool
	QueryFeedbackCallCount       int
	LastQueryFeedbackQuery       string
	GetAccountRiskScoreCalled    bool
	GetAccountRiskScoreCallCount int
	LastAccountRiskScoreID       string
	GetRecentNegativeCountCalled bool
	GetRecentNegativeCountCount  int
	LastRecentNegativeAccountID  string
	GetProductAreaImpactsCalled  bool
	GetProductAreaImpactsCount   int
	LastProductAreaSegment       string
	GetFeedbackCountCalled       bool
	GetFeedbackCountCount        int
	GetFeedbackEnrichedCalled    bool
	EnrichedCallCount            int
}

// NewMockFeedbackRepository creates a new mock repository
func NewMockFeedbackRepository() *MockFeedbackRepository {
	return &MockFeedbackRepository{
		QueryFeedbackResult:        []map[string]any{},
		GetAccountRiskScoreResult:  nil,
		GetRecentNegativeCountResult: 0,
		GetProductAreaImpactsResult: []map[string]any{},
		GetFeedbackCountResult:      0,
	}
}

// QueryFeedback implements FeedbackRepository.QueryFeedback
func (m *MockFeedbackRepository) QueryFeedback(ctx context.Context, query string) ([]map[string]any, error) {
	m.QueryFeedbackCalled = true
	m.QueryFeedbackCallCount++
	m.LastQueryFeedbackQuery = query
	return m.QueryFeedbackResult, m.QueryFeedbackErr
}

// GetAccountRiskScore implements FeedbackRepository.GetAccountRiskScore
func (m *MockFeedbackRepository) GetAccountRiskScore(ctx context.Context, accountID string) (*domain.AccountRiskScore, error) {
	m.GetAccountRiskScoreCalled = true
	m.GetAccountRiskScoreCallCount++
	m.LastAccountRiskScoreID = accountID
	return m.GetAccountRiskScoreResult, m.GetAccountRiskScoreErr
}

// GetRecentNegativeFeedbackCount implements FeedbackRepository.GetRecentNegativeFeedbackCount
func (m *MockFeedbackRepository) GetRecentNegativeFeedbackCount(ctx context.Context, accountID string) (int, error) {
	m.GetRecentNegativeCountCalled = true
	m.GetRecentNegativeCountCount++
	m.LastRecentNegativeAccountID = accountID
	return m.GetRecentNegativeCountResult, m.GetRecentNegativeCountErr
}

// GetProductAreaImpacts implements FeedbackRepository.GetProductAreaImpacts
func (m *MockFeedbackRepository) GetProductAreaImpacts(ctx context.Context, segment string) ([]map[string]any, error) {
	m.GetProductAreaImpactsCalled = true
	m.GetProductAreaImpactsCount++
	m.LastProductAreaSegment = segment
	return m.GetProductAreaImpactsResult, m.GetProductAreaImpactsErr
}

// GetFeedbackEnrichedCount implements FeedbackRepository.GetFeedbackEnrichedCount
func (m *MockFeedbackRepository) GetFeedbackEnrichedCount(ctx context.Context) (int, error) {
	m.GetFeedbackCountCalled = true
	m.GetFeedbackCountCount++
	return m.GetFeedbackCountResult, m.GetFeedbackCountErr
}

// SetupForSuccess configures the mock to return successful results
func (m *MockFeedbackRepository) SetupForSuccess() {
	m.QueryFeedbackErr = nil
	m.GetAccountRiskScoreErr = nil
	m.GetRecentNegativeCountErr = nil
	m.GetProductAreaImpactsErr = nil
	m.GetFeedbackCountErr = nil
}

// SetupForFailure configures the mock to return errors
func (m *MockFeedbackRepository) SetupForFailure(err error) {
	m.QueryFeedbackErr = err
	m.GetAccountRiskScoreErr = err
	m.GetRecentNegativeCountErr = err
	m.GetProductAreaImpactsErr = err
	m.GetFeedbackCountErr = err
}

// SetupForDatabaseError configures the mock to return database errors
func (m *MockFeedbackRepository) SetupForDatabaseError() {
	dbErr := errors.New("database error")
	m.SetupForFailure(dbErr)
}

// SetupForTimeout configures the mock to return timeout errors
func (m *MockFeedbackRepository) SetupForTimeout() {
	timeoutErr := context.DeadlineExceeded
	m.SetupForFailure(timeoutErr)
}

// SetQueryFeedbackResult sets the return value for QueryFeedback
func (m *MockFeedbackRepository) SetQueryFeedbackResult(result []map[string]any) {
	m.QueryFeedbackResult = result
}

// SetAccountRiskScoreResult sets the return value for GetAccountRiskScore
func (m *MockFeedbackRepository) SetAccountRiskScoreResult(result *domain.AccountRiskScore) {
	m.GetAccountRiskScoreResult = result
}

// SetRecentNegativeCountResult sets the return value for GetRecentNegativeFeedbackCount
func (m *MockFeedbackRepository) SetRecentNegativeCountResult(count int) {
	m.GetRecentNegativeCountResult = count
}

// SetProductAreaImpactsResult sets the return value for GetProductAreaImpacts
func (m *MockFeedbackRepository) SetProductAreaImpactsResult(result []map[string]any) {
	m.GetProductAreaImpactsResult = result
}

// SetFeedbackCountResult sets the return value for GetFeedbackEnrichedCount
func (m *MockFeedbackRepository) SetFeedbackCountResult(count int) {
	m.GetFeedbackCountResult = count
}

// SetGetFeedbackEnrichedResult sets the return value for GetFeedbackEnriched
func (m *MockFeedbackRepository) SetGetFeedbackEnrichedResult(result []domain.FeedbackEnriched) {
	m.GetFeedbackEnrichedResult = result
}

// SetGetFeedbackEnrichedError sets the error return value for GetFeedbackEnriched
func (m *MockFeedbackRepository) SetGetFeedbackEnrichedError(err error) {
	m.GetFeedbackEnrichedErr = err
}

// SetQueryFeedbackError sets the error return value for QueryFeedback
func (m *MockFeedbackRepository) SetQueryFeedbackError(err error) {
	m.QueryFeedbackErr = err
}

// GetFeedbackEnriched returns all enriched feedback (mock implementation)
func (m *MockFeedbackRepository) GetFeedbackEnriched(ctx context.Context) ([]domain.FeedbackEnriched, error) {
	m.GetFeedbackEnrichedCalled = true
	m.EnrichedCallCount++
	return m.GetFeedbackEnrichedResult, m.GetFeedbackEnrichedErr
}

// ResetCallCounts resets all call tracking
func (m *MockFeedbackRepository) ResetCallCounts() {
	m.QueryFeedbackCalled = false
	m.QueryFeedbackCallCount = 0
	m.GetAccountRiskScoreCalled = false
	m.GetAccountRiskScoreCallCount = 0
	m.GetRecentNegativeCountCalled = false
	m.GetRecentNegativeCountCount = 0
	m.GetProductAreaImpactsCalled = false
	m.GetProductAreaImpactsCount = 0
	m.GetFeedbackCountCalled = false
	m.GetFeedbackCountCount = 0
}

// MockTransaction is a mock implementation of Transaction for testing
type MockTransaction struct {
	CommitErr  error
	RollbackErr error
	Repository repository.FeedbackRepository

	CommitCalled    bool
	RollbackCalled  bool
	GetRepoCalled   bool
}

// NewMockTransaction creates a new mock transaction
func NewMockTransaction(repo repository.FeedbackRepository) *MockTransaction {
	return &MockTransaction{
		Repository: repo,
	}
}

// Commit implements Transaction.Commit
func (m *MockTransaction) Commit() error {
	m.CommitCalled = true
	return m.CommitErr
}

// Rollback implements Transaction.Rollback
func (m *MockTransaction) Rollback() error {
	m.RollbackCalled = true
	return m.RollbackErr
}

// GetRepository implements Transaction.GetRepository
func (m *MockTransaction) GetRepository() repository.FeedbackRepository {
	m.GetRepoCalled = true
	return m.Repository
}

// Reset resets call tracking
func (m *MockTransaction) Reset() {
	m.CommitCalled = false
	m.RollbackCalled = false
	m.GetRepoCalled = false
}
