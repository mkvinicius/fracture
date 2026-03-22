package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ModelRole maps a function in the simulation to a specific model.
type ModelRole string

const (
	RoleConformist  ModelRole = "conformist"   // GPT-4o Mini or Gemini Flash
	RoleDisruptor   ModelRole = "disruptor"    // GPT-4o or Claude Sonnet
	RoleSynthesis   ModelRole = "synthesis"    // Claude Sonnet (final report)
	RoleSanitizer   ModelRole = "sanitizer"    // Claude Haiku (security)
	RoleCoherence   ModelRole = "coherence"    // Gemini Flash (fast checks)
)

// ModelConfig holds the API key and model name for a specific role.
type ModelConfig struct {
	Provider string `json:"provider"` // openai | anthropic | google | ollama
	Model    string `json:"model"`
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url,omitempty"` // for Ollama or custom endpoints
}

// RouterConfig maps each role to a ModelConfig.
type RouterConfig map[ModelRole]ModelConfig

// DefaultRouterConfig returns the recommended model configuration.
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		RoleConformist: {Provider: "openai", Model: "gpt-4o-mini"},
		RoleDisruptor:  {Provider: "openai", Model: "gpt-4o"},
		RoleSynthesis:  {Provider: "anthropic", Model: "claude-3-5-sonnet-20241022"},
		RoleSanitizer:  {Provider: "anthropic", Model: "claude-3-5-haiku-20241022"},
		RoleCoherence:  {Provider: "google", Model: "gemini-1.5-flash"},
	}
}

// Router selects the right model for each agent role and calls it.
type Router struct {
	cfg    RouterConfig
	client *http.Client
}

// NewRouter creates a Router with the given configuration.
func NewRouter(cfg RouterConfig) *Router {
	return &Router{
		cfg: cfg,
		client: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

// ForRole returns a caller bound to the model for the given role.
func (r *Router) ForRole(role ModelRole) *RoleCaller {
	cfg, ok := r.cfg[role]
	if !ok {
		// Fallback to conformist config
		cfg = r.cfg[RoleConformist]
	}
	return &RoleCaller{router: r, cfg: cfg, role: role}
}

// RoleCaller implements engine.LLMCaller for a specific model role.
type RoleCaller struct {
	router *Router
	cfg    ModelConfig
	role   ModelRole
}

// Call sends a chat completion request to the configured model.
// Returns (response text, tokens used, error).
func (rc *RoleCaller) Call(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (string, int, error) {
	switch rc.cfg.Provider {
	case "openai":
		return rc.callOpenAI(ctx, systemPrompt, userPrompt, maxTokens)
	case "anthropic":
		return rc.callAnthropic(ctx, systemPrompt, userPrompt, maxTokens)
	case "google":
		return rc.callGoogle(ctx, systemPrompt, userPrompt, maxTokens)
	case "ollama":
		return rc.callOllama(ctx, systemPrompt, userPrompt, maxTokens)
	default:
		return "", 0, fmt.Errorf("unknown provider: %s", rc.cfg.Provider)
	}
}

// ─── OpenAI ──────────────────────────────────────────────────────────────────

type openAIRequest struct {
	Model     string          `json:"model"`
	Messages  []openAIMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (rc *RoleCaller) callOpenAI(ctx context.Context, system, user string, maxTokens int) (string, int, error) {
	payload := openAIRequest{
		Model: rc.cfg.Model,
		Messages: []openAIMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		MaxTokens: maxTokens,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Authorization", "Bearer "+rc.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := rc.router.client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var result openAIResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", 0, fmt.Errorf("parse openai response: %w", err)
	}
	if result.Error != nil {
		return "", 0, fmt.Errorf("openai error: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", 0, fmt.Errorf("openai: no choices returned")
	}
	return result.Choices[0].Message.Content, result.Usage.TotalTokens, nil
}

// ─── Anthropic ───────────────────────────────────────────────────────────────

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (rc *RoleCaller) callAnthropic(ctx context.Context, system, user string, maxTokens int) (string, int, error) {
	payload := anthropicRequest{
		Model:     rc.cfg.Model,
		MaxTokens: maxTokens,
		System:    system,
		Messages:  []anthropicMessage{{Role: "user", Content: user}},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("x-api-key", rc.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := rc.router.client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var result anthropicResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", 0, fmt.Errorf("parse anthropic response: %w", err)
	}
	if result.Error != nil {
		return "", 0, fmt.Errorf("anthropic error: %s", result.Error.Message)
	}
	if len(result.Content) == 0 {
		return "", 0, fmt.Errorf("anthropic: no content returned")
	}
	total := result.Usage.InputTokens + result.Usage.OutputTokens
	return result.Content[0].Text, total, nil
}

// ─── Google Gemini ───────────────────────────────────────────────────────────

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		TotalTokenCount int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func (rc *RoleCaller) callGoogle(ctx context.Context, system, user string, maxTokens int) (string, int, error) {
	combined := system + "\n\n" + user
	payload := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: combined}}},
		},
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		rc.cfg.Model, rc.cfg.APIKey)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := rc.router.client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var result geminiResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", 0, fmt.Errorf("parse gemini response: %w", err)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", 0, fmt.Errorf("gemini: no content returned")
	}
	return result.Candidates[0].Content.Parts[0].Text, result.UsageMetadata.TotalTokenCount, nil
}

// ─── Ollama (local) ──────────────────────────────────────────────────────────

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (rc *RoleCaller) callOllama(ctx context.Context, system, user string, maxTokens int) (string, int, error) {
	baseURL := rc.cfg.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	payload := ollamaRequest{
		Model:  rc.cfg.Model,
		Prompt: system + "\n\n" + user,
		Stream: false,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := rc.router.client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var result ollamaResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", 0, fmt.Errorf("parse ollama response: %w", err)
	}
	// Ollama doesn't report token count in non-stream mode easily; estimate
	estimated := len(result.Response) / 4
	return result.Response, estimated, nil
}
