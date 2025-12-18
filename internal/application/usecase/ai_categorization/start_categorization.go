// Package aicategorization contains AI categorization-related use cases.
package aicategorization

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strconv"
	"strings"
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

	// BatchDelay is the delay between batch requests to avoid rate limiting.
	// Gemini free tier allows 20 RPM, so we wait 5 seconds between batches.
	BatchDelay = 5 * time.Second

	// MaxRetries is the maximum number of retry attempts for rate-limited requests.
	MaxRetries = 5

	// DefaultRetryDelay is the fallback delay when we can't parse the API's suggested delay.
	DefaultRetryDelay = 45 * time.Second

	// RetryDelayBuffer is added to the API's suggested retry delay for safety margin.
	RetryDelayBuffer = 5 * time.Second

	// MaxRetryDelay caps the maximum wait time between retries.
	MaxRetryDelay = 120 * time.Second
)

// merchantPattern defines a pattern for extracting merchant base names from transaction descriptions.
type merchantPattern struct {
	pattern     *regexp.Regexp
	extractFunc func(matches []string, desc string) string
}

// merchantPatterns contains patterns for common merchant descriptions.
// Order matters - more specific patterns should come first.
var merchantPatterns = []merchantPattern{
	// Uber patterns: "UBER *TRIP", "UBER *EATS", etc.
	{
		pattern: regexp.MustCompile(`(?i)^UBER\s*\*`),
		extractFunc: func(_ []string, _ string) string {
			return "UBER"
		},
	},
	// iFood patterns
	{
		pattern: regexp.MustCompile(`(?i)IFOOD`),
		extractFunc: func(_ []string, _ string) string {
			return "IFOOD"
		},
	},
	// Rappi patterns
	{
		pattern: regexp.MustCompile(`(?i)RAPPI`),
		extractFunc: func(_ []string, _ string) string {
			return "RAPPI"
		},
	},
	// Netflix patterns
	{
		pattern: regexp.MustCompile(`(?i)NETFLIX`),
		extractFunc: func(_ []string, _ string) string {
			return "NETFLIX"
		},
	},
	// Spotify patterns
	{
		pattern: regexp.MustCompile(`(?i)SPOTIFY`),
		extractFunc: func(_ []string, _ string) string {
			return "SPOTIFY"
		},
	},
	// Amazon patterns
	{
		pattern: regexp.MustCompile(`(?i)AMAZON`),
		extractFunc: func(_ []string, _ string) string {
			return "AMAZON"
		},
	},
	// YouTube/Google patterns
	{
		pattern: regexp.MustCompile(`(?i)YOUTUBE`),
		extractFunc: func(_ []string, _ string) string {
			return "YOUTUBE"
		},
	},
	{
		pattern: regexp.MustCompile(`(?i)^GOOGLE\s*\*`),
		extractFunc: func(_ []string, _ string) string {
			return "GOOGLE"
		},
	},
	// MercadoLivre/MercadoPago patterns
	{
		pattern: regexp.MustCompile(`(?i)MERCADOLIVRE|MERCPAGO`),
		extractFunc: func(_ []string, _ string) string {
			return "MERCADOLIVRE"
		},
	},
	// PIX patterns: "PAG*" prefix
	{
		pattern: regexp.MustCompile(`(?i)^PAG\*`),
		extractFunc: func(_ []string, _ string) string {
			return "PAG*PIX"
		},
	},
	// PicPay patterns
	{
		pattern: regexp.MustCompile(`(?i)PICPAY`),
		extractFunc: func(_ []string, _ string) string {
			return "PICPAY"
		},
	},
	// Nubank patterns
	{
		pattern: regexp.MustCompile(`(?i)NUBANK`),
		extractFunc: func(_ []string, _ string) string {
			return "NUBANK"
		},
	},
	// PG * patterns: "PG *COMPANY"
	{
		pattern: regexp.MustCompile(`(?i)^PG\s*\*\s*(\w+)`),
		extractFunc: func(matches []string, _ string) string {
			if len(matches) > 1 {
				return strings.ToUpper(matches[1])
			}
			return "PG*"
		},
	},
	// Generic prefix patterns: extract first significant word
	{
		pattern: regexp.MustCompile(`(?i)^([A-Z]+(?:\s*[A-Z]+)?)`),
		extractFunc: func(matches []string, desc string) string {
			if len(matches) > 1 {
				// Return first word or two, cleaned up
				words := strings.Fields(matches[1])
				if len(words) > 0 {
					return strings.ToUpper(words[0])
				}
			}
			// Fallback: first word of description
			words := strings.Fields(desc)
			if len(words) > 0 {
				return strings.ToUpper(words[0])
			}
			return desc
		},
	},
}

// extractMerchantKey extracts a normalized merchant key from a transaction description.
// This key is used to group similar transactions together.
func extractMerchantKey(description string) string {
	desc := strings.ToUpper(strings.TrimSpace(description))
	if desc == "" {
		return ""
	}

	for _, mp := range merchantPatterns {
		matches := mp.pattern.FindStringSubmatch(desc)
		if matches != nil {
			return mp.extractFunc(matches, desc)
		}
	}

	// Fallback: return first word
	words := strings.Fields(desc)
	if len(words) > 0 {
		return words[0]
	}
	return desc
}

// transactionWithKey pairs a transaction with its extracted merchant key for sorting.
type transactionWithKey struct {
	transaction *adapter.TransactionForAI
	merchantKey string
}

// sortTransactionsByMerchant sorts transactions so similar ones (same merchant) are adjacent.
// This increases the chance that the AI will see related transactions in the same batch.
func sortTransactionsByMerchant(transactions []*adapter.TransactionForAI) []*adapter.TransactionForAI {
	if len(transactions) <= 1 {
		return transactions
	}

	// Extract merchant keys for all transactions
	txsWithKeys := make([]transactionWithKey, len(transactions))
	for i, tx := range transactions {
		txsWithKeys[i] = transactionWithKey{
			transaction: tx,
			merchantKey: extractMerchantKey(tx.Description),
		}
	}

	// Sort by merchant key to group similar transactions.
	// Using stable sort to preserve original order within each merchant group.
	sort.SliceStable(txsWithKeys, func(i, j int) bool {
		return txsWithKeys[i].merchantKey < txsWithKeys[j].merchantKey
	})

	// Extract sorted transactions
	sorted := make([]*adapter.TransactionForAI, len(transactions))
	for i, twk := range txsWithKeys {
		sorted[i] = twk.transaction
	}

	return sorted
}

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

// isRateLimitError checks if an error is a rate limit (429) error from the AI service.
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "quota") ||
		strings.Contains(errStr, "resource exhausted")
}

// retryDelayRegex matches "Please retry in Xs" or "retry in X seconds" patterns.
var retryDelayRegex = regexp.MustCompile(`retry in (\d+(?:\.\d+)?)\s*s`)

// parseRetryDelay extracts the suggested retry delay from an error message.
// Returns the parsed delay plus a buffer, or a default delay if parsing fails.
func parseRetryDelay(err error, attempt int) time.Duration {
	if err == nil {
		return DefaultRetryDelay
	}

	errStr := strings.ToLower(err.Error())
	matches := retryDelayRegex.FindStringSubmatch(errStr)

	if len(matches) >= 2 {
		if seconds, parseErr := strconv.ParseFloat(matches[1], 64); parseErr == nil {
			delay := time.Duration(seconds)*time.Second + RetryDelayBuffer
			if delay > MaxRetryDelay {
				return MaxRetryDelay
			}
			return delay
		}
	}

	// Fallback: exponential backoff based on attempt number
	delay := DefaultRetryDelay * time.Duration(attempt+1)
	if delay > MaxRetryDelay {
		return MaxRetryDelay
	}
	return delay
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
			uc.processingTracker.ClearProgress(userID)
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

	// Sort transactions by merchant to group similar ones together.
	// This increases the chance that similar transactions (e.g., multiple UBER trips)
	// end up in the same batch, allowing the AI to properly group them.
	sortedTxs := sortTransactionsByMerchant(txsForAI)
	logger.Info("Sorted transactions by merchant for better grouping")

	// Split transactions into batches
	batches := splitIntoBatches(sortedTxs)
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

	// Initialize progress tracking
	if uc.processingTracker != nil {
		uc.processingTracker.SetProgress(userID, ProcessingProgress{
			ProcessedTransactions: 0,
			TotalTransactions:     len(transactions),
			CurrentBatch:          0,
			TotalBatches:          totalBatches,
		})
	}

	// Process each batch and save results incrementally
	totalSavedSuggestions := 0

	for batchNum, batch := range batches {
		batchLogger := logger.With("batch", batchNum+1, "totalBatches", totalBatches, "batchTransactions", len(batch))

		// Update progress before processing batch
		if uc.processingTracker != nil {
			processedSoFar := batchNum * BatchSize
			uc.processingTracker.SetProgress(userID, ProcessingProgress{
				ProcessedTransactions: processedSoFar,
				TotalTransactions:     len(transactions),
				CurrentBatch:          batchNum + 1,
				TotalBatches:          totalBatches,
			})
		}

		// Add delay between batches to avoid rate limiting (skip first batch)
		if batchNum > 0 {
			batchLogger.Info("Waiting between batches to avoid rate limits", "delay", BatchDelay.String())
			time.Sleep(BatchDelay)
		}

		batchLogger.Info("Processing batch")

		request := &adapter.AICategorizationRequest{
			UserID:             userID,
			Transactions:       batch,
			ExistingCategories: catsForAI,
		}

		// Retry loop for rate-limited requests
		var results []*adapter.AICategorizationResult
		var err error
		for attempt := 0; attempt <= MaxRetries; attempt++ {
			// Create per-batch timeout context
			batchCtx, batchCancel := context.WithTimeout(ctx, BatchTimeout)

			batchStartTime := time.Now()
			results, err = uc.aiService.Categorize(batchCtx, request)
			batchCancel()

			if err == nil {
				batchLogger.Info("Batch completed",
					"resultCount", len(results),
					"duration", time.Since(batchStartTime).String(),
					"attempt", attempt+1,
				)
				break
			}

			// Check if it's a rate limit error
			if isRateLimitError(err) && attempt < MaxRetries {
				retryDelay := parseRetryDelay(err, attempt)
				batchLogger.Warn("Rate limited, retrying after delay",
					"error", err.Error(),
					"attempt", attempt+1,
					"maxRetries", MaxRetries,
					"retryDelay", retryDelay.String(),
				)
				time.Sleep(retryDelay)
				continue
			}

			// Non-rate-limit error or max retries exceeded
			batchLogger.Error("Batch processing failed",
				"error", err.Error(),
				"duration", time.Since(batchStartTime).String(),
				"attempt", attempt+1,
			)
			// Set error with info about partial results
			uc.setProcessingErrorWithPartialCount(userID, err, totalSavedSuggestions)
			return
		}

		// Convert batch results to suggestions and save immediately
		batchSuggestions := uc.convertResultsToSuggestions(userID, results)
		if len(batchSuggestions) > 0 {
			if err := uc.suggestionRepo.CreateBatch(ctx, batchSuggestions); err != nil {
				batchLogger.Error("Failed to save batch suggestions", "error", err.Error(), "count", len(batchSuggestions))
				// Continue processing, don't fail entire job for save error
			} else {
				totalSavedSuggestions += len(batchSuggestions)
				batchLogger.Info("Saved batch suggestions", "count", len(batchSuggestions), "totalSaved", totalSavedSuggestions)
			}
		}

		// Update progress after batch completion
		if uc.processingTracker != nil {
			processedSoFar := (batchNum + 1) * BatchSize
			if processedSoFar > len(transactions) {
				processedSoFar = len(transactions)
			}
			uc.processingTracker.SetProgress(userID, ProcessingProgress{
				ProcessedTransactions: processedSoFar,
				TotalTransactions:     len(transactions),
				CurrentBatch:          batchNum + 1,
				TotalBatches:          totalBatches,
			})
		}
	}

	// All batches completed successfully - suggestions were saved per-batch
	logger.Info("AI categorization completed all batches",
		"totalSavedSuggestions", totalSavedSuggestions,
		"totalDuration", time.Since(startTime).String(),
	)
}

// convertResultsToSuggestions converts AI categorization results to AISuggestion entities.
func (uc *StartCategorizationUseCase) convertResultsToSuggestions(userID uuid.UUID, results []*adapter.AICategorizationResult) []*entity.AISuggestion {
	suggestions := make([]*entity.AISuggestion, 0, len(results))

	for _, result := range results {
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

	return suggestions
}

// setProcessingError classifies and stores an error for the user.
func (uc *StartCategorizationUseCase) setProcessingError(userID uuid.UUID, err error) {
	if uc.processingTracker == nil {
		return
	}
	processingError := classifyError(err)
	uc.processingTracker.SetError(userID, processingError)
}

// setProcessingErrorWithPartialCount classifies and stores an error with info about saved suggestions.
func (uc *StartCategorizationUseCase) setProcessingErrorWithPartialCount(userID uuid.UUID, err error, savedCount int) {
	if uc.processingTracker == nil {
		return
	}
	processingError := classifyError(err)

	// Append info about saved suggestions to the message if any were saved
	if savedCount > 0 {
		processingError.Message = fmt.Sprintf("%s %d sugestões foram salvas.", processingError.Message, savedCount)
	}

	uc.processingTracker.SetError(userID, processingError)
}
