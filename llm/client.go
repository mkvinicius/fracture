package llm

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync"
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
		RoleDisruptor:  {Provider: "anthropic", Model: "claude-haiku-4-5-20251001"},
		RoleSynthesis:  {Provider: "anthropic", Model: "claude-sonnet-4-20250514"},
		RoleSanitizer:  {Provider: "anthropic", Model: "claude-haiku-4-5-20251001"},
		RoleCoherence:  {Provider: "openai", Model: "gpt-4o-mini"},
	}
}

// cacheEntry holds a cached LLM response.
type cacheEntry struct {
	response string
	tokens   int
	expiry   time.Time
}

// responseCache is a thread-safe in-memory LLM response cache.
type responseCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
}

func newResponseCache() *responseCache {
	c := &responseCache{entries: make(map[string]cacheEntry)}
	go c.evictLoop()
	return c
}

func (c *responseCache) key(system, user string, maxTokens int) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s|%s|%d", system, user, maxTokens)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (c *responseCache) get(k string) (string, int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[k]
	if !ok || time.Now().After(e.expiry) {
		return "", 0, false
	}
	return e.response, e.tokens, true
}

func (c *responseCache) set(k, response string, tokens int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[k] = cacheEntry{
		response: response,
		tokens:   tokens,
		expiry:   time.Now().Add(10 * time.Minute),
	}
}

func (c *responseCache) evictLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, e := range c.entries {
			if now.After(e.expiry) {
				delete(c.entries, k)
			}
		}
		c.mu.Unlock()
	}
}

// Router selects the right model for each agent role and calls it.
// It includes a semaphore to cap concurrent API calls, a response cache,
// and retry logic with exponential backoff.
type Router struct {
	cfg       RouterConfig
	client    *http.Client
	semaphore chan struct{}
	cache     *responseCache
}

// NewRouter creates a Router with the given configuration.
func NewRouter(cfg RouterConfig) *Router {
	return &Router{
		cfg:       cfg,
		client:    &http.Client{Timeout: 90 * time.Second},
		semaphore: make(chan struct{}, 10), // max 10 concurrent API calls
		cache:     newResponseCache(),
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
// Includes semaphore throttling, response caching, and retry with exponential backoff.
func (rc *RoleCaller) Call(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (string, int, error) {
	// Check cache for deterministic roles
	if rc.role == RoleConformist || rc.role == RoleCoherence || rc.role == RoleSanitizer {
		k := rc.router.cache.key(systemPrompt, userPrompt, maxTokens)
		if resp, tokens, ok := rc.router.cache.get(k); ok {
			return resp, tokens, nil
		}
	}

	// Acquire semaphore slot (max 10 concurrent calls)
	select {
	case rc.router.semaphore <- struct{}{}:
		defer func() { <-rc.router.semaphore }()
	case <-ctx.Done():
		return "", 0, ctx.Err()
	}

	// Retry with exponential backoff (3 attempts)
	var (
		resp   string
		tokens int
		err    error
	)
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * 500 * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return "", 0, ctx.Err()
			}
		}
		switch rc.cfg.Provider {
		case "openai":
			resp, tokens, err = rc.callOpenAI(ctx, systemPrompt, userPrompt, maxTokens)
		case "anthropic":
			resp, tokens, err = rc.callAnthropic(ctx, systemPrompt, userPrompt, maxTokens)
		case "google":
			resp, tokens, err = rc.callGoogle(ctx, systemPrompt, userPrompt, maxTokens)
		case "ollama":
			resp, tokens, err = rc.callOllama(ctx, systemPrompt, userPrompt, maxTokens)
		default:
			return "", 0, fmt.Errorf("unknown provider: %s", rc.cfg.Provider)
		}
		if err == nil {
			break
		}
	}
	if err != nil {
		return "", 0, err
	}

	// Store in cache for eligible roles
	if rc.role == RoleConformist || rc.role == RoleCoherence || rc.role == RoleSanitizer {
		k := rc.router.cache.key(systemPrompt, userPrompt, maxTokens)
		rc.router.cache.set(k, resp, tokens)
	}
	return resp, tokens, nil
}

// ─── OpenAI ──────────────────────────────────────────────────────────────────

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
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
	// Conformist/coherence agents use lower temperature for speed and consistency
	temp := 0.7
	if rc.role == RoleConformist || rc.role == RoleCoherence {
		temp = 0.3
	}
	payload := openAIRequest{
		Model: rc.cfg.Model,
		Messages: []openAIMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		MaxTokens:   maxTokens,
		Temperature: temp,
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
