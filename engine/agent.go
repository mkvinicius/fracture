package engine

import (
	"context"
	"math/rand"
)

// AgentType distinguishes Conformists from Disruptors.
type AgentType string

const (
	AgentConformist AgentType = "conformist"
	AgentDisruptor  AgentType = "disruptor"
)

// AgentPermissions defines what an agent is allowed to do.
type AgentPermissions struct {
	CanProposeRule     bool
	CanAccessMemory    bool
	MaxTokensPerRound  int
	MaxRoundsActive    int
}

// ConformistPermissions — read world, react, no rule proposals.
var ConformistPermissions = AgentPermissions{
	CanProposeRule:    false,
	CanAccessMemory:   true,
	MaxTokensPerRound: 400,
	MaxRoundsActive:   999,
}

// DisruptorPermissions — can propose rule changes.
var DisruptorPermissions = AgentPermissions{
	CanProposeRule:    true,
	CanAccessMemory:   true,
	MaxTokensPerRound: 600,
	MaxRoundsActive:   999,
}

// Personality holds the character traits of an archetype.
type Personality struct {
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Traits      []string `json:"traits"`
	Goals       []string `json:"goals"`
	Biases      []string `json:"biases"`
	PowerWeight float64  `json:"power_weight"` // voting influence 0.0-1.0
}

// AgentAction is what an agent produces in a single round.
type AgentAction struct {
	AgentID            string        `json:"agent_id"`
	AgentType          AgentType     `json:"agent_type"`
	Text               string        `json:"text"`
	IsFractureProposal bool          `json:"is_fracture_proposal"`
	Proposal           *RuleProposal `json:"proposal,omitempty"`
	TensionDelta       map[string]float64 `json:"tension_delta"` // ruleID -> delta
	TokensUsed         int           `json:"tokens_used"`
}

// Agent is the interface every archetype must implement.
type Agent interface {
	ID() string
	Type() AgentType
	Permissions() AgentPermissions
	Personality() Personality
	// React produces an action given the current world state and round context.
	React(ctx context.Context, world *World, memory AgentMemory, round int, tension float64) (AgentAction, error)
}

// BaseAgent provides shared fields and helpers for all agent implementations.
type BaseAgent struct {
	id          string
	agentType   AgentType
	permissions AgentPermissions
	personality Personality
	llm         LLMCaller
}

func (b *BaseAgent) ID() string                    { return b.id }
func (b *BaseAgent) Type() AgentType               { return b.agentType }
func (b *BaseAgent) Permissions() AgentPermissions { return b.permissions }
func (b *BaseAgent) Personality() Personality      { return b.personality }

// FractureThreshold calculates the probability that a Disruptor fires a FRACTURE POINT.
// P = base_tension * (1 + dissatisfaction) * personality_factor
func FractureThreshold(tension float64, dissatisfaction float64, personalityFactor float64) float64 {
	p := tension * (1 + dissatisfaction) * personalityFactor
	if p > 0.95 {
		return 0.95
	}
	return p
}

// ShouldFireFracture returns true if a Disruptor should propose a rule change this round.
func ShouldFireFracture(threshold float64) bool {
	return rand.Float64() < threshold
}

// NewBaseAgent constructs a BaseAgent with all fields set.
func NewBaseAgent(id string, agentType AgentType, perms AgentPermissions, personality Personality) BaseAgent {
	return BaseAgent{
		id:          id,
		agentType:   agentType,
		permissions: perms,
		personality: personality,
	}
}

// LLMCaller is the interface the engine uses to call language models.
// Implemented in the llm package and injected at runtime.
type LLMCaller interface {
	Call(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (string, int, error)
}

// AgentMemory provides read access to an agent's historical context.
type AgentMemory interface {
	RecentActions(agentID string, n int) []AgentAction
	SimilarContexts(query string, n int) []string
}
