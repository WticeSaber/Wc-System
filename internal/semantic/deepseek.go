package semantic

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"wc-predictor/internal/models"
)

const (
	deepSeekBaseURL   = "https://api.deepseek.com/v1"
	deepSeekModel     = "deepseek-chat"
	deepSeekMaxTokens = 150
	deepSeekTemp      = 0.2  // low temperature for deterministic JSON output
	callTimeout       = 12 * time.Second
	maxModifier       = 0.2
)

// modifierResponse is the expected JSON structure returned by DeepSeek.
type modifierResponse struct {
	Modifier  float64 `json:"modifier"`
	Reasoning string  `json:"reasoning"`
}

// DeepSeekClient wraps the OpenAI-compatible DeepSeek API client.
type DeepSeekClient struct {
	client *openai.Client
}

// NewDeepSeekClient creates a DeepSeekClient using the DEEPSEEK_API_KEY env variable.
// Returns nil if the key is not set (caller should treat this as "disabled").
func NewDeepSeekClient() *DeepSeekClient {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil
	}

	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = deepSeekBaseURL

	return &DeepSeekClient{client: openai.NewClientWithConfig(cfg)}
}

// GetMomentumModifier calls DeepSeek to obtain a float momentum adjustment
// coefficient in the range [-0.2, +0.2].
//
// Returns (0.0, error) on any failure. The caller should log the error and
// continue the calculation unmodified.
func (d *DeepSeekClient) GetMomentumModifier(
	ctx context.Context,
	homeTeam, awayTeam string,
	homeForm, awayForm []models.MatchResult,
) (float64, error) {
	if d == nil {
		return 0, fmt.Errorf("semantic: DeepSeek client not initialized (DEEPSEEK_API_KEY not set)")
	}

	prompt, err := BuildPrompt(homeTeam, awayTeam, homeForm, awayForm)
	if err != nil {
		return 0, err
	}

	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()

	resp, err := d.client.CreateChatCompletion(callCtx, openai.ChatCompletionRequest{
		Model: deepSeekModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a precise football analytics assistant. Always respond with valid JSON only.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   deepSeekMaxTokens,
		Temperature: deepSeekTemp,
	})
	if err != nil {
		return 0, fmt.Errorf("semantic: DeepSeek API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return 0, fmt.Errorf("semantic: DeepSeek returned no choices")
	}

	rawContent := resp.Choices[0].Message.Content

	var result modifierResponse
	if err := json.Unmarshal([]byte(rawContent), &result); err != nil {
		return 0, fmt.Errorf("semantic: failed to parse DeepSeek JSON response %q: %w", rawContent, err)
	}

	// Clamp to the declared range; log a warning if clamping was needed.
	if math.Abs(result.Modifier) > maxModifier {
		log.Printf("[semantic] WARNING: DeepSeek returned modifier=%.4f, clamping to ±%.1f", result.Modifier, maxModifier)
		if result.Modifier > 0 {
			result.Modifier = maxModifier
		} else {
			result.Modifier = -maxModifier
		}
	}

	log.Printf("[semantic] DeepSeek modifier=%.4f, reasoning=%q", result.Modifier, result.Reasoning)
	return result.Modifier, nil
}

// AsSemanticModifierFunc wraps the client into the SemanticModifierFunc type
// expected by the engine calculator. Safe to call with a nil client.
func (d *DeepSeekClient) AsSemanticModifierFunc() func(
	ctx context.Context,
	homeCanonical, awayCanonical string,
	homeForm, awayForm []models.MatchResult,
) (float64, error) {
	return func(ctx context.Context, homeCanonical, awayCanonical string, homeForm, awayForm []models.MatchResult) (float64, error) {
		return d.GetMomentumModifier(ctx, homeCanonical, awayCanonical, homeForm, awayForm)
	}
}
