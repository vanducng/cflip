package providers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/vanducng/cflip/internal/config"
)

// AnthropicProvider implements the Provider interface for Anthropic
type AnthropicProvider struct {
	config *config.Provider
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider() *AnthropicProvider {
	return &AnthropicProvider{
		config: &config.Provider{
			Name:        "anthropic",
			DisplayName: "Anthropic",
			BaseURL:     "https://api.anthropic.com",
			Models: map[string]string{
				"haiku":  "claude-3-5-haiku-20241022",
				"sonnet": "claude-3-5-sonnet-20241022",
				"opus":   "claude-3-opus-20240229",
			},
			AuthHeader: "x-api-key",
			EnvVars: map[string]string{
				"API_TIMEOUT_MS": "3000000",
			},
		},
	}
}

// Name returns the unique identifier for the provider
func (p *AnthropicProvider) Name() string {
	return p.config.Name
}

// DisplayName returns a human-readable name for the provider
func (p *AnthropicProvider) DisplayName() string {
	return p.config.DisplayName
}

// Description returns a brief description of the provider
func (p *AnthropicProvider) Description() string {
	return "Official Anthropic Claude API provider"
}

// GetConfig returns the provider configuration
func (p *AnthropicProvider) GetConfig() *config.Provider {
	return p.config
}

// ValidateAPIKey validates if the provided API key is valid for Anthropic
func (p *AnthropicProvider) ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Basic format validation for Anthropic API keys
	if !strings.HasPrefix(apiKey, "sk-ant-") {
		return fmt.Errorf("invalid Anthropic API key format: should start with 'sk-ant-'")
	}

	// Check minimum length
	if len(apiKey) < 50 {
		return fmt.Errorf("API key appears to be too short")
	}

	// TODO: Add actual API validation call if needed
	// For now, we'll do basic format validation
	return nil
}

// GetModels returns the available models for this provider
func (p *AnthropicProvider) GetModels() map[string]string {
	return p.config.Models
}

// GetBaseURL returns the base URL for the provider's API
func (p *AnthropicProvider) GetBaseURL() string {
	return p.config.BaseURL
}

// RequiresSetup returns true if the provider requires additional setup
func (p *AnthropicProvider) RequiresSetup() bool {
	return false
}

// SetupInstructions returns setup instructions for the provider
func (p *AnthropicProvider) SetupInstructions() string {
	return `To use Anthropic:

1. Get your API key from https://console.anthropic.com/
2. Ensure you have credits or a subscription
3. Run: cflip switch anthropic
4. Enter your API key when prompted

Your API key will be securely stored in ~/.claude/settings.json`
}

// TestConnection makes a simple API call to verify the connection
func (p *AnthropicProvider) TestConnection(apiKey string) error {
	client := &http.Client{}
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL+"/v1/messages", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(p.config.AuthHeader, apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Anthropic API: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	// Check for authentication errors
	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid API key")
	}

	if resp.StatusCode == 403 {
		return fmt.Errorf("access forbidden - check your API key and permissions")
	}

	return nil
}