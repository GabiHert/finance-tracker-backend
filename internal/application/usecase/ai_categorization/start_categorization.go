// Package aicategorization contains AI categorization-related use cases.
package aicategorization

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
	domainerror "github.com/finance-tracker/backend/internal/domain/error"
)

const (
	// BatchSize is the number of transactions to process per AI request.
	// Keeping this small (30-50) ensures Gemini can respond within timeout.
	BatchSize = 40

	// BatchTimeout is the timeout for processing a single batch.
	// Should be generous enough for AI to process BatchSize transactions.
	BatchTimeout = 45 * time.Second

	// MaxBatches is the maximum number of batches to process.
	// Prevents runaway processing (40 * 50 = 2000 transactions max).
	MaxBatches = 50
)

// splitIntoBatches divides transactions into batches of BatchSize.
func splitIntoBatches(transactions []*adapter.TransactionForAI) [][]*adapter.TransactionForAI {
	batches := make([][]*adapter.TransactionForAI, 0)

	for i := 0; i < len(transactions); i += BatchSize {
		end := i + BatchSize
		if end > len(transactions) {
			end = len(transactions)
		}
		batches = append(batches, transactions[i:end])
	}

	return batches
}

// StartCategorizationInput represents the input for starting AI categorization.
type StartCategorizationInput struct {
	UserID uuid.UUID
}

// StartCategorizationOutput represents the output of starting AI categorization.
type StartCategorizationOutput struct {
	JobID   string `json:"job_id"`
	Message string `json:"message"`
}

// StartCategorizationUseCase handles starting the AI categorization process.
type StartCategorizationUseCase struct {
	transactionRepo   adapter.TransactionRepository
	categoryRepo      adapter.CategoryRepository
	suggestionRepo    adapter.AISuggestionRepository
	aiService         adapter.AICategorizationService
	processingTracker ProcessingTracker
}

// NewStartCategorizationUseCase creates a new StartCategorizationUseCase instance.
func NewStartCategorizationUseCase(
	transactionRepo adapter.TransactionRepository,
	categoryRepo adapter.CategoryRepository,
	suggestionRepo adapter.AISuggestionRepository,
	aiService adapter.AICategorizationService,
	processingTracker ProcessingTracker,
) *StartCategorizationUseCase {
	return &StartCategorizationUseCase{
		transactionRepo:   transactionRepo,
		categoryRepo:      categoryRepo,
		suggestionRepo:    suggestionRepo,
		aiService:         aiService,
		processingTracker: processingTracker,
	}
}

// Execute starts the AI categorization process.
func (uc *StartCategorizationUseCase) Execute(ctx context.Context, input StartCategorizationInput) (*StartCategorizationOutput, error) {
	// Check if already processing
	if uc.processingTracker != nil && uc.processingTracker.IsProcessing(input.UserID) {
		return nil, domainerror.NewAISuggestionError(
			domainerror.ErrCodeAIAlreadyProcessing,
			"AI categorization is already in progress",
			domainerror.ErrAIAlreadyProcessing,
		)
	}

	// Clear any previous error before starting
	if uc.processingTracker != nil {
		uc.processingTracker.ClearError(input.UserID)
	}

	// Get uncategorized transactions
	uncategorizedTxs, err := uc.getUncategorizedTransactions(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get uncategorized transactions: %w", err)
	}

	if len(uncategorizedTxs) == 0 {
		return nil, domainerror.NewAISuggestionError(
			domainerror.ErrCodeAINoUncategorized,
			"No uncategorized transactions found",
			domainerror.ErrAINoUncategorized,
		)
	}

	// Generate job ID
	jobID := uuid.New().String()

	// Set processing state
	if uc.processingTracker != nil {
		uc.processingTracker.SetProcessing(input.UserID, jobID)
	}

	// Start async processing (in a goroutine for non-blocking response)
	go uc.processCategorizationAsync(context.Background(), input.UserID, uncategorizedTxs, jobID)

	return &StartCategorizationOutput{
		JobID:   jobID,
		Message: fmt.Sprintf("AI categorization started for %d uncategorized transactions", len(uncategorizedTxs)),
	}, nil
}

// getUncategorizedTransactions retrieves all uncategorized transactions for a user.
func (uc *StartCategorizationUseCase) getUncategorizedTransactions(ctx context.Context, userID uuid.UUID) ([]*entity.Transaction, error) {
	// Get all transactions for the user
	transactions, err := uc.transactionRepo.FindByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Filter uncategorized transactions (category_id is nil)
	uncategorized := make([]*entity.Transaction, 0)
	for _, tx := range transactions {
		if tx.CategoryID == nil {
			uncategorized = append(uncategorized, tx)
		}
	}

	return uncategorized, nil
}

// processCategorizationAsync processes categorization in the background using batched requests.
func (uc *StartCategorizationUseCase) processCategorizationAsync(ctx context.Context, userID uuid.UUID, transactions []*entity.Transaction, jobID string) {
	startTime := time.Now()
	logger := slog.Default().With("jobID", jobID, "userID", userID.String(), "transactionCount", len(transactions))

	logger.Info("Starting AI categorization process")

	defer func() {
		if uc.processingTracker != nil {
			uc.processingTracker.ClearProcessing(userID)
		}
		logger.Info("AI categorization process completed", "duration", time.Since(startTime).String())
	}()

	// Get existing categories for the user
	categories, err := uc.categoryRepo.FindByOwner(ctx, entity.OwnerTypeUser, userID)
	if err != nil {
		logger.Error("Failed to get categories", "error", err.Error())
		uc.setProcessingError(userID, err)
		return
	}
	logger.Info("Loaded categories", "categoryCount", len(categories))

	// Convert transactions to AI format
	txsForAI := make([]*adapter.TransactionForAI, len(transactions))
	for i, tx := range transactions {
		txsForAI[i] = &adapter.TransactionForAI{
			ID:          tx.ID,
			Description: tx.Description,
			Amount:      tx.Amount.String(),
			Date:        tx.Date.Format("2006-01-02"),
			Type:        string(tx.Type),
		}
	}

	// Convert categories to AI format
	catsForAI := make([]*adapter.CategoryForAI, len(categories))
	for i, cat := range categories {
		catsForAI[i] = &adapter.CategoryForAI{
			ID:    cat.ID,
			Name:  cat.Name,
			Type:  string(cat.Type),
			Icon:  cat.Icon,
			Color: cat.Color,
		}
	}

	// Split transactions into batches
	batches := splitIntoBatches(txsForAI)
	totalBatches := len(batches)

	// Limit to MaxBatches
	if totalBatches > MaxBatches {
		logger.Warn("Transaction count exceeds maximum, processing first batches only",
			"totalTransactions", len(transactions),
			"maxProcessed", MaxBatches*BatchSize,
		)
		batches = batches[:MaxBatches]
		totalBatches = MaxBatches
	}

	logger.Info("Processing transactions in batches",
		"batchCount", totalBatches,
		"batchSize", BatchSize,
	)

	// Process each batch and collect results
	allResults := make([]*adapter.AICategorizationResult, 0)

	for batchNum, batch := range batches {
		batchLogger := logger.With("batch", batchNum+1, "totalBatches", totalBatches, "batchTransactions", len(batch))
		batchLogger.Info("Processing batch")

		// Create per-batch timeout context
		batchCtx, batchCancel := context.WithTimeout(ctx, BatchTimeout)

		request := &adapter.AICategorizationRequest{
			UserID:             userID,
			Transactions:       batch,
			ExistingCategories: catsForAI,
		}

		batchStartTime := time.Now()
		results, err := uc.aiService.Categorize(batchCtx, request)
		batchCancel()

		if err != nil {
			batchLogger.Error("Batch processing failed",
				"error", err.Error(),
				"duration", time.Since(batchStartTime).String(),
			)
			uc.setProcessingError(userID, err)
			return
		}

		batchLogger.Info("Batch completed",
			"resultCount", len(results),
			"duration", time.Since(batchStartTime).String(),
		)

		allResults = append(allResults, results...)
	}

	logger.Info("AI service completed all batches",
		"totalResults", len(allResults),
		"totalDuration", time.Since(startTime).String(),
	)

	// Convert results to suggestions and save
	suggestions := make([]*entity.AISuggestion, 0, len(allResults))
	for _, result := range allResults {
		var suggestion *entity.AISuggestion

		if result.SuggestedCategoryID != nil {
			suggestion = entity.NewAISuggestion(
				userID,
				result.TransactionID,
				*result.SuggestedCategoryID,
				result.MatchType,
				result.MatchKeyword,
				result.AffectedTransactionIDs,
			)
		} else if result.SuggestedCategoryNew != nil {
			suggestion = entity.NewAISuggestionWithNewCategory(
				userID,
				result.TransactionID,
				*result.SuggestedCategoryNew,
				result.MatchType,
				result.MatchKeyword,
				result.AffectedTransactionIDs,
			)
		}

		if suggestion != nil {
			suggestions = append(suggestions, suggestion)
		}
	}

	// Save suggestions in batch
	if len(suggestions) > 0 {
		if err := uc.suggestionRepo.CreateBatch(ctx, suggestions); err != nil {
			logger.Error("Failed to save suggestions", "error", err.Error(), "suggestionCount", len(suggestions))
			uc.setProcessingError(userID, err)
			return
		}
		logger.Info("Saved suggestions", "suggestionCount", len(suggestions))
	} else {
		logger.Info("No suggestions generated")
	}
}

// setProcessingError classifies and stores an error for the user.
func (uc *StartCategorizationUseCase) setProcessingError(userID uuid.UUID, err error) {
	if uc.processingTracker == nil {
		return
	}
	processingError := classifyError(err)
	uc.processingTracker.SetError(userID, processingError)
}
