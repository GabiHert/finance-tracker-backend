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

	sb.WriteString(`Voce e um especialista em categorizacao de transacoes financeiras. Sua tarefa e analisar transacoes sem categoria e sugerir categorias apropriadas.

IMPORTANTE - IDIOMA:
- Todas as respostas devem ser em Portugues Brasileiro
- Nomes de categorias DEVEM ser em Portugues, EXCETO para termos comumente usados em ingles no Brasil:
  * Pet Shop, Delivery, Drive Thru, Shopping, Fast Food, Streaming, Fitness, E-commerce, Marketplace
  * Nomes de apps/servicos: Uber, iFood, Rappi, Netflix, Spotify, etc.
- Para outras categorias, use Portugues Brasileiro natural:
  * Supermercado, Restaurante, Transporte, Saude, Educacao, Lazer, Moradia, Vestuario
  * Servicos, Assinaturas, Viagem, Alimentacao, Combustivel, Farmacia, Padaria, Banco

Para cada transacao, voce deve:
1. Identificar um padrao (palavra-chave) para corresponder transacoes similares
2. Sugerir uma categoria existente ou propor uma nova
3. Identificar o tipo de correspondencia: "exact", "startsWith", ou "contains"

REGRAS IMPORTANTES:
- Prefira categorias existentes quando correspondem bem
- Para novas categorias, sugira nome (em Portugues, exceto termos comuns em ingles), icone (da lista abaixo), e cor hex
- A palavra-chave deve ser especifica para evitar falsos positivos, mas geral para capturar transacoes similares
- Use "contains" para parciais, "startsWith" para prefixo, "exact" para exatas
- Agrupe transacoes similares pelo padrao

ICONES DISPONIVEIS (use APENAS estes nomes exatos):
Financeiro: wallet, credit-card, bank, receipt, coins, piggy-bank, chart-line, dollar-sign
Alimentacao: utensils, coffee, pizza, apple, wine
Transporte: car, bus, plane, train, bike, gas-pump
Casa: home, bed, sofa, lamp, wrench
Entretenimento: music, film, gamepad, tv, ticket
Saude: heart, medical, pill, dumbbell
Educacao: book, graduation-cap, pencil
Compras: shopping-bag, shopping-cart, tag, gift, percent
Utilidades: bolt, wifi, phone, droplet, flame
Outros: briefcase, globe, star

SUGESTOES DE ICONES POR TIPO DE CATEGORIA:
- Supermercado: shopping-cart
- Restaurante/Alimentacao: utensils
- Pet Shop: heart
- Farmacia: medical
- Transporte/Uber: car
- Combustivel/Posto: gas-pump
- Streaming/Assinaturas: tv
- Delivery/iFood: utensils
- Shopping: shopping-bag
- Academia/Fitness: dumbbell
- Educacao: book
- Banco/Taxas: bank
- Moradia/Aluguel: home
- Lazer: gamepad
- Viagem: plane
- Servicos: briefcase

CATEGORIAS EXISTENTES:
`)

	if len(request.ExistingCategories) > 0 {
		for _, cat := range request.ExistingCategories {
			sb.WriteString(fmt.Sprintf("- ID: %s, Name: %s, Type: %s, Icon: %s\n",
				cat.ID, cat.Name, cat.Type, cat.Icon))
		}
	} else {
		sb.WriteString("(Nenhuma categoria existente)\n")
	}

	sb.WriteString("\nTRANSACOES PARA CATEGORIZAR:\n")
	for _, tx := range request.Transactions {
		sb.WriteString(fmt.Sprintf("- ID: %s, Description: \"%s\", Amount: %s, Date: %s, Type: %s\n",
			tx.ID, tx.Description, tx.Amount, tx.Date, tx.Type))
	}

	sb.WriteString(`

Responda com um array JSON de sugestoes. Cada sugestao deve ter:
{
  "transaction_id": "uuid da transacao principal",
  "suggested_category_id": "uuid da categoria existente ou null",
  "suggested_category_new": { "name": "string em Portugues", "icon": "string da lista de icones", "color": "#XXXXXX" } ou null,
  "match_type": "contains" | "startsWith" | "exact",
  "match_keyword": "palavra-chave/padrao para correspondencia",
  "affected_transaction_ids": ["uuids de outras transacoes que correspondem ao padrao"],
  "confidence": 0.0-1.0,
  "reasoning": "breve explicacao em Portugues"
}

Agrupe transacoes similares. Se multiplas transacoes correspondem ao mesmo padrao, inclua uma sugestao com todos os IDs afetados.

IMPORTANTE: Use APENAS icones da lista fornecida acima. Nao invente nomes de icones.

FORMATO DE RESPOSTA: Retorne apenas o array JSON, sem texto adicional.
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
