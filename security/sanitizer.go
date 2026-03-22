package security

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

// ErrInjectionDetected is returned when prompt injection is detected.
var ErrInjectionDetected = errors.New("potential prompt injection detected in input")

// injectionPatterns are regex patterns that match common injection attempts.
var injectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)ignore (all |your )?(previous |prior |above )?instructions`),
	regexp.MustCompile(`(?i)disregard (all |your )?(previous |prior |above )?instructions`),
	regexp.MustCompile(`(?i)forget (everything|all|your instructions)`),
	regexp.MustCompile(`(?i)you are now`),
	regexp.MustCompile(`(?i)new (system |)prompt:`),
	regexp.MustCompile(`(?i)\[system\]`),
	regexp.MustCompile(`(?i)<\|im_start\|>`),
	regexp.MustCompile(`(?i)###\s*(instruction|system|prompt)`),
	regexp.MustCompile(`(?i)act as (if you are|a )(different|new|another)`),
	regexp.MustCompile(`(?i)reveal (your|the) (system |)prompt`),
	regexp.MustCompile(`(?i)print (your|the) (system |)prompt`),
	regexp.MustCompile(`(?i)what (are|were) your instructions`),
}

// sensitiveDataPatterns detect accidental leakage of secrets in outputs.
var sensitiveDataPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(sk-|sk-proj-)[a-zA-Z0-9\-_]{20,}`),              // OpenAI keys
	regexp.MustCompile(`(?i)anthropic[_-]?api[_-]?key\s*[:=]\s*\S+`),         // Anthropic keys
	regexp.MustCompile(`(?i)(password|passwd|secret|token)\s*[:=]\s*\S{8,}`), // Generic secrets
	regexp.MustCompile(`(?i)[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`), // Emails (optional redact)
}

// LLMSanitizer is the interface for the LLM-based injection detector.
// Uses a fast, cheap model (e.g., Claude Haiku) isolated from simulation context.
type LLMSanitizer interface {
	CheckInjection(ctx context.Context, input string) (bool, error) // true = safe
}

// Sanitizer validates external inputs before they reach simulation agents.
type Sanitizer struct {
	llm LLMSanitizer // optional — if nil, only regex checks are applied
}

// NewSanitizer creates a Sanitizer. Pass nil for llm to use regex-only mode.
func NewSanitizer(llm LLMSanitizer) *Sanitizer {
	return &Sanitizer{llm: llm}
}

// Sanitize validates an input string. Returns the cleaned string or an error.
func (s *Sanitizer) Sanitize(ctx context.Context, input string) (string, error) {
	if input == "" {
		return "", nil
	}

	// Step 1: Regex check (fast, no LLM cost)
	for _, pattern := range injectionPatterns {
		if pattern.MatchString(input) {
			return "", ErrInjectionDetected
		}
	}

	// Step 2: LLM check (deeper, catches sophisticated injections)
	if s.llm != nil {
		safe, err := s.llm.CheckInjection(ctx, input)
		if err != nil {
			// On LLM error, fail open (log but don't block) — availability > security here
			// In production, consider failing closed depending on risk tolerance
		} else if !safe {
			return "", ErrInjectionDetected
		}
	}

	return input, nil
}

// RedactOutput removes sensitive data from agent outputs before display.
func RedactOutput(output string) string {
	result := output
	for _, pattern := range sensitiveDataPatterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			// Keep first 4 chars for debugging, redact the rest
			if len(match) > 8 {
				return match[:4] + strings.Repeat("*", len(match)-4)
			}
			return "****"
		})
	}
	return result
}

// SanitizeBatch sanitizes a slice of inputs, returning only safe ones.
// Unsafe inputs are replaced with an empty string and logged.
func (s *Sanitizer) SanitizeBatch(ctx context.Context, inputs []string) ([]string, []error) {
	results := make([]string, len(inputs))
	errs := make([]error, len(inputs))
	for i, input := range inputs {
		clean, err := s.Sanitize(ctx, input)
		results[i] = clean
		errs[i] = err
	}
	return results, errs
}
