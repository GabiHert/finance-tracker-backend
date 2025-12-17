// Package adapters provides implementations for external service integrations.
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"google.golang.org/api/option"

	"github.com/finance-tracker/backend/internal/application/adapter"
	"github.com/finance-tracker/backend/internal/domain/entity"
)

// GeminiService implements the AICategorizationService using Google Gemini.
type GeminiService struct {
	apiKey    string
	modelName string
	client    *genai.Client
}

// NewGeminiService creates a new Gemini service instance.
func NewGeminiService(apiKey string) *GeminiService {
	return &GeminiService{
		apiKey:    apiKey,
		modelName: "gemini-2.5-flash-lite",
	}
}

// IsAvailable checks if the Gemini service is available and properly configured.
func (s *GeminiService) IsAvailable() bool {
	return s.apiKey != ""
}

// Categorize analyzes transactions and returns categorization suggestions.
func (s *GeminiService) Categorize(ctx context.Context, request *adapter.AICategorizationRequest) ([]*adapter.AICategorizationResult, error) {
	if !s.IsAvailable() {
		return nil, fmt.Errorf("gemini service is not configured")
	}

	// Create client
	client, err := genai.NewClient(ctx, option.WithAPIKey(s.apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}
	defer client.Close()

	// Get the model
	model := client.GenerativeModel(s.modelName)

	// Configure model for JSON output
	model.SetTemperature(0.3)
	model.ResponseMIMEType = "application/json"

	// Build the prompt
	prompt := s.buildPrompt(request)

	// Generate response
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Parse response
	results, err := s.parseResponse(resp, request)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return results, nil
}

// buildPrompt creates the prompt for Gemini.
func (s *GeminiService) buildPrompt(request *adapter.AICategorizationRequest) string {
	var sb strings.Builder

	sb.WriteString(`You are a financial transaction categorization expert. Your task is to analyze uncategorized transactions and suggest appropriate categories.

For each transaction, you should:
1. Identify a pattern (keyword) that can be used to match similar transactions
2. Suggest either an existing category or propose a new one
3. Identify the match type: "exact", "startsWith", or "contains"

IMPORTANT RULES:
- Prefer existing categories when they match well
- For new categories, suggest a name, icon (from common icon libraries like lucide), and a hex color
- The match keyword should be specific enough to avoid false positives but general enough to catch similar transactions
- Use "contains" for partial matches, "startsWith" for prefix matches, "exact" for exact matches
- Group similar transactions together by their match keyword

EXISTING CATEGORIES:
`)

	if len(request.ExistingCategories) > 0 {
		for _, cat := range request.ExistingCategories {
			sb.WriteString(fmt.Sprintf("- ID: %s, Name: %s, Type: %s, Icon: %s\n",
				cat.ID, cat.Name, cat.Type, cat.Icon))
		}
	} else {
		sb.WriteString("(No existing categories)\n")
	}

	sb.WriteString("\nTRANSACTIONS TO CATEGORIZE:\n")
	for _, tx := range request.Transactions {
		sb.WriteString(fmt.Sprintf("- ID: %s, Description: \"%s\", Amount: %s, Date: %s, Type: %s\n",
			tx.ID, tx.Description, tx.Amount, tx.Date, tx.Type))
	}

	sb.WriteString(`

Respond with a JSON array of suggestions. Each suggestion should have:
{
  "transaction_id": "uuid of the primary transaction",
  "suggested_category_id": "uuid of existing category or null",
  "suggested_category_new": { "name": "string", "icon": "string", "color": "#XXXXXX" } or null,
  "match_type": "contains" | "startsWith" | "exact",
  "match_keyword": "the keyword/pattern to match",
  "affected_transaction_ids": ["uuids of other transactions that would match this pattern"],
  "confidence": 0.0-1.0,
  "reasoning": "brief explanation"
}

Group similar transactions together. If multiple transactions would match the same pattern, include one suggestion with all affected IDs.

RESPONSE FORMAT: Return only the JSON array, no additional text.
`)

	return sb.String()
}

// geminiSuggestion represents the raw response from Gemini.
type geminiSuggestion struct {
	TransactionID         string             `json:"transaction_id"`
	SuggestedCategoryID   *string            `json:"suggested_category_id"`
	SuggestedCategoryNew  *geminiNewCategory `json:"suggested_category_new"`
	MatchType             string             `json:"match_type"`
	MatchKeyword          string             `json:"match_keyword"`
	AffectedTransactionIDs []string          `json:"affected_transaction_ids"`
	Confidence            float64            `json:"confidence"`
	Reasoning             string             `json:"reasoning"`
}

type geminiNewCategory struct {
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

// parseResponse parses the Gemini response into AICategorizationResults.
func (s *GeminiService) parseResponse(resp *genai.GenerateContentResponse, request *adapter.AICategorizationRequest) ([]*adapter.AICategorizationResult, error) {
	if resp == nil || len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("empty response from gemini")
	}

	// Get the text content from the response
	var textContent string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			textContent = string(text)
			break
		}
	}

	if textContent == "" {
		return nil, fmt.Errorf("no text content in response")
	}

	// Clean the response (remove markdown code blocks if present)
	textContent = strings.TrimPrefix(textContent, "```json")
	textContent = strings.TrimPrefix(textContent, "```")
	textContent = strings.TrimSuffix(textContent, "```")
	textContent = strings.TrimSpace(textContent)

	// Parse JSON
	var suggestions []geminiSuggestion
	if err := json.Unmarshal([]byte(textContent), &suggestions); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w, content: %s", err, textContent)
	}

	// Convert to results
	results := make([]*adapter.AICategorizationResult, 0, len(suggestions))
	for _, s := range suggestions {
		result := &adapter.AICategorizationResult{
			MatchType:   entity.MatchType(s.MatchType),
			MatchKeyword: s.MatchKeyword,
			Confidence:  s.Confidence,
			Reasoning:   s.Reasoning,
		}

		// Parse transaction ID
		txID, err := uuid.Parse(s.TransactionID)
		if err != nil {
			continue // Skip invalid IDs
		}
		result.TransactionID = txID

		// Parse suggested category ID or new category
		if s.SuggestedCategoryID != nil && *s.SuggestedCategoryID != "" {
			catID, err := uuid.Parse(*s.SuggestedCategoryID)
			if err == nil {
				result.SuggestedCategoryID = &catID
			}
		} else if s.SuggestedCategoryNew != nil {
			result.SuggestedCategoryNew = &entity.SuggestedCategoryNew{
				Name:  s.SuggestedCategoryNew.Name,
				Icon:  s.SuggestedCategoryNew.Icon,
				Color: s.SuggestedCategoryNew.Color,
			}
		}

		// Parse affected transaction IDs
		result.AffectedTransactionIDs = make([]uuid.UUID, 0, len(s.AffectedTransactionIDs))
		for _, idStr := range s.AffectedTransactionIDs {
			if id, err := uuid.Parse(idStr); err == nil {
				result.AffectedTransactionIDs = append(result.AffectedTransactionIDs, id)
			}
		}

		// Validate match type
		switch result.MatchType {
		case entity.MatchTypeContains, entity.MatchTypeStartsWith, entity.MatchTypeExact:
			// Valid
		default:
			result.MatchType = entity.MatchTypeContains // Default to contains
		}

		results = append(results, result)
	}

	return results, nil
}
