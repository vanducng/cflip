package providers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/vanducng/cflip/internal/config"
)

// GLMProvider implements the Provider interface for GLM by z.ai
type GLMProvider struct {
	config *config.Provider
}

// NewGLMProvider creates a new GLM provider
func NewGLMProvider() *GLMProvider {
	return &GLMProvider{
		config: &config.Provider{
			Name:        "glm",
			DisplayName: "GLM by z.ai",
			BaseURL:     "https://api.z.ai/api/anthropic",
			Models: map[string]string{
				"haiku":  "glm-4.5-air",
				"sonnet": "glm-4.6",
				"opus":   "glm-4.6",
			},
			AuthHeader: "authorization",
			EnvVars: map[string]string{
				"API_TIMEOUT_MS": "3000000",
			},
		},
	}
}

// Name returns the unique identifier for the provider
func (p *GLMProvider) Name() string {
	return p.config.Name
}

// DisplayName returns a human-readable name for the provider
func (p *GLMProvider) DisplayName() string {
	return p.config.DisplayName
}

// Description returns a brief description of the provider
func (p *GLMProvider) Description() string {
	return "GLM models from z.ai with Anthropic-compatible API"
}

// GetConfig returns the provider configuration
func (p *GLMProvider) GetConfig() *config.Provider {
	return p.config
}

// ValidateAPIKey validates if the provided API key is valid for GLM
func (p *GLMProvider) ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Basic format validation for z.ai API keys
	if !strings.HasPrefix(apiKey, "zai-") {
		return fmt.Errorf("invalid GLM API key format: should start with 'zai-'")
	}

	// Check minimum length
	if len(apiKey) < 40 {
		return fmt.Errorf("API key appears to be too short")
	}

	return nil
}

// GetModels returns the available models for this provider
func (p *GLMProvider) GetModels() map[string]string {
	return p.config.Models
}

// GetBaseURL returns the base URL for the provider's API
func (p *GLMProvider) GetBaseURL() string {
	return p.config.BaseURL
}

// RequiresSetup returns true if the provider requires additional setup
func (p *GLMProvider) RequiresSetup() bool {
	return true
}

// SetupInstructions returns setup instructions for the provider
func (p *GLMProvider) SetupInstructions() string {
	return `To use GLM by z.ai:

1. Visit https://platform.z.ai and register/login
2. Subscribe to the GLM Coding Plan
3. Generate an API key from your dashboard
4. Run: cflip switch glm
5. Enter your API key when prompted

Note: GLM Coding Plan is required for Claude Code integration.
Your API key will be securely stored in ~/.claude/settings.json`
}

// TestConnection makes a simple API call to verify the connection
func (p *GLMProvider) TestConnection(apiKey string) error {
	client := &http.Client{}
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", p.config.BaseURL+"/v1/messages", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(p.config.AuthHeader, "Bearer "+apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to GLM API: %w", err)
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
		return fmt.Errorf("access forbidden - ensure you have an active GLM Coding Plan")
	}

	return nil
}

// GetFeatureList returns special features available with GLM
func (p *GLMProvider) GetFeatureList() []string {
	return []string{
		"Code completion",
		"Repository Q&A",
		"Automated task management",
		"Vision MCP server",
		"Web Search MCP server",
	}
}