// Package aicategorization contains AI categorization-related use cases.
package aicategorization

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
)

// GetStatusInput represents the input for getting AI categorization status.
type GetStatusInput struct {
	UserID uuid.UUID
}

// GetStatusOutput represents the output of getting AI categorization status.
type GetStatusOutput struct {
	UncategorizedCount      int                 `json:"uncategorized_count"`
	IsProcessing            bool                `json:"is_processing"`
	PendingSuggestionsCount int                 `json:"pending_suggestions_count"`
	JobID                   string              `json:"job_id,omitempty"`
	HasError                bool                `json:"has_error"`
	Error                   *ProcessingError    `json:"error,omitempty"`
	Progress                *ProcessingProgress `json:"progress,omitempty"`
}

// GetStatusUseCase handles retrieving AI categorization status.
type GetStatusUseCase struct {
	transactionRepo   adapter.TransactionRepository
	suggestionRepo    adapter.AISuggestionRepository
	processingTracker ProcessingTracker
}

// ProcessingProgress tracks the progress of batch processing.
type ProcessingProgress struct {
	ProcessedTransactions int `json:"processed_transactions"`
	TotalTransactions     int `json:"total_transactions"`
	CurrentBatch          int `json:"current_batch"`
	TotalBatches          int `json:"total_batches"`
}

// ProcessingTracker interface for tracking processing state (will be implemented in-memory or Redis).
type ProcessingTracker interface {
	// Existing methods for processing state.
	IsProcessing(userID uuid.UUID) bool
	GetJobID(userID uuid.UUID) string
	SetProcessing(userID uuid.UUID, jobID string)
	ClearProcessing(userID uuid.UUID)

	// Error tracking methods.
	SetError(userID uuid.UUID, err *ProcessingError)
	GetError(userID uuid.UUID) *ProcessingError
	ClearError(userID uuid.UUID)
	HasError(userID uuid.UUID) bool

	// Progress tracking methods.
	SetProgress(userID uuid.UUID, progress ProcessingProgress)
	GetProgress(userID uuid.UUID) ProcessingProgress
	ClearProgress(userID uuid.UUID)
}

// NewGetStatusUseCase creates a new GetStatusUseCase instance.
func NewGetStatusUseCase(
	transactionRepo adapter.TransactionRepository,
	suggestionRepo adapter.AISuggestionRepository,
	processingTracker ProcessingTracker,
) *GetStatusUseCase {
	return &GetStatusUseCase{
		transactionRepo:   transactionRepo,
		suggestionRepo:    suggestionRepo,
		processingTracker: processingTracker,
	}
}

// Execute retrieves the AI categorization status for a user.
func (uc *GetStatusUseCase) Execute(ctx context.Context, input GetStatusInput) (*GetStatusOutput, error) {
	// Get uncategorized transaction count
	uncategorizedCount, err := uc.getUncategorizedCount(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	// Get pending suggestions count
	pendingCount, err := uc.suggestionRepo.GetPendingCount(ctx, input.UserID)
	if err != nil {
		return nil, err
	}

	// Check if processing is in progress
	isProcessing := false
	jobID := ""
	if uc.processingTracker != nil {
		isProcessing = uc.processingTracker.IsProcessing(input.UserID)
		jobID = uc.processingTracker.GetJobID(input.UserID)
	}

	// Check for processing error
	hasError := false
	var processingError *ProcessingError
	if uc.processingTracker != nil {
		hasError = uc.processingTracker.HasError(input.UserID)
		if hasError {
			processingError = uc.processingTracker.GetError(input.UserID)
		}
	}

	// Get progress if processing
	var progress *ProcessingProgress
	if isProcessing && uc.processingTracker != nil {
		p := uc.processingTracker.GetProgress(input.UserID)
		if p.TotalTransactions > 0 {
			progress = &p
		}
	}

	return &GetStatusOutput{
		UncategorizedCount:      uncategorizedCount,
		IsProcessing:            isProcessing,
		PendingSuggestionsCount: pendingCount,
		JobID:                   jobID,
		HasError:                hasError,
		Error:                   processingError,
		Progress:                progress,
	}, nil
}

// getUncategorizedCount retrieves the count of uncategorized transactions.
func (uc *GetStatusUseCase) getUncategorizedCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return uc.transactionRepo.CountUncategorizedByUser(ctx, userID)
}

// InMemoryProcessingTracker is a simple in-memory implementation of ProcessingTracker.
type InMemoryProcessingTracker struct {
	mu         sync.RWMutex
	processing map[uuid.UUID]string
	errors     map[uuid.UUID]*ProcessingError
	progress   map[uuid.UUID]ProcessingProgress
}

// NewInMemoryProcessingTracker creates a new in-memory processing tracker.
func NewInMemoryProcessingTracker() *InMemoryProcessingTracker {
	return &InMemoryProcessingTracker{
		processing: make(map[uuid.UUID]string),
		errors:     make(map[uuid.UUID]*ProcessingError),
		progress:   make(map[uuid.UUID]ProcessingProgress),
	}
}

// IsProcessing checks if a user is currently processing.
func (t *InMemoryProcessingTracker) IsProcessing(userID uuid.UUID) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, ok := t.processing[userID]
	return ok
}

// GetJobID gets the job ID for a user.
func (t *InMemoryProcessingTracker) GetJobID(userID uuid.UUID) string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.processing[userID]
}

// SetProcessing sets the processing state for a user.
func (t *InMemoryProcessingTracker) SetProcessing(userID uuid.UUID, jobID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.processing[userID] = jobID
}

// ClearProcessing clears the processing state for a user.
func (t *InMemoryProcessingTracker) ClearProcessing(userID uuid.UUID) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.processing, userID)
}

// SetError stores an error for a user.
func (t *InMemoryProcessingTracker) SetError(userID uuid.UUID, err *ProcessingError) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.errors[userID] = err
}

// GetError retrieves the error for a user.
func (t *InMemoryProcessingTracker) GetError(userID uuid.UUID) *ProcessingError {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.errors[userID]
}

// ClearError removes the error for a user.
func (t *InMemoryProcessingTracker) ClearError(userID uuid.UUID) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.errors, userID)
}

// HasError checks if a user has an error.
func (t *InMemoryProcessingTracker) HasError(userID uuid.UUID) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, ok := t.errors[userID]
	return ok
}

// SetProgress stores the progress for a user.
func (t *InMemoryProcessingTracker) SetProgress(userID uuid.UUID, progress ProcessingProgress) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.progress[userID] = progress
}

// GetProgress retrieves the progress for a user.
func (t *InMemoryProcessingTracker) GetProgress(userID uuid.UUID) ProcessingProgress {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.progress[userID]
}

// ClearProgress removes the progress for a user.
func (t *InMemoryProcessingTracker) ClearProgress(userID uuid.UUID) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.progress, userID)
}

// TransactionForCategorization represents minimal transaction data for AI processing.
type TransactionForCategorization struct {
	ID          uuid.UUID
	Description string
	Amount      string
	Date        string
	Type        entity.TransactionType
}
