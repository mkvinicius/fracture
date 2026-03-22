package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SimulationConfig holds parameters for a single simulation run.
type SimulationConfig struct {
	ID         string
	CompanyID  string
	Question   string
	Department string
	MaxRounds  int
	Agents     []Agent
	World      *World
	Memory     AgentMemory
}

// RoundResult is streamed to the caller after each round completes.
type RoundResult struct {
	Round          int            `json:"round"`
	Actions        []AgentAction  `json:"actions"`
	FractureEvents []FractureEvent `json:"fracture_events"`
	Tension        float64        `json:"tension"`
	WorldSnapshot  WorldSnapshot  `json:"world_snapshot"`
	ElapsedMs      int64          `json:"elapsed_ms"`
}

// FractureEvent records a FRACTURE POINT activation and its outcome.
type FractureEvent struct {
	Round          int          `json:"round"`
	ProposedBy     string       `json:"proposed_by"`
	Proposal       RuleProposal `json:"proposal"`
	Accepted       bool         `json:"accepted"`
	VoteBreakdown  []VoteRecord `json:"vote_breakdown"`
}

// Coalition represents a group of agents aligned around a common interest.
type Coalition struct {
	Name        string   `json:"name"`
	AgentNames  []string `json:"agent_names"`
	SharedGoal  string   `json:"shared_goal"`
	Strength    float64  `json:"strength"` // 0.0-1.0
	IsDisruptive bool   `json:"is_disruptive"`
}

// ActionPlaybook is the strategic recommendation for the user.
type ActionPlaybook struct {
	Horizon90Days  []string `json:"horizon_90_days"`
	Horizon1Year   []string `json:"horizon_1_year"`
	Horizon3Years  []string `json:"horizon_3_years"`
	QuickWins      []string `json:"quick_wins"`
	CriticalRisks  []string `json:"critical_risks"`
}

// SimulationResult is the final output after all rounds complete.
type SimulationResult struct {
	SimulationID     string             `json:"simulation_id"`
	Question         string             `json:"question"`
	Rounds           []RoundResult      `json:"rounds"`
	FractureEvents   []FractureEvent    `json:"fracture_events"`
	FinalWorld       WorldSnapshot      `json:"final_world"`
	ProbableFuture   string             `json:"probable_future"`
	TensionMap       map[string]float64 `json:"tension_map"`
	RuptureScenarios []RuptureScenario  `json:"rupture_scenarios"`
	Coalitions       []Coalition        `json:"coalitions"`
	ActionPlaybook   *ActionPlaybook    `json:"action_playbook,omitempty"`
	TotalTokens      int                `json:"total_tokens"`
	DurationMs       int64              `json:"duration_ms"`
}

// RuptureScenario describes one possible future where a rule is broken.
type RuptureScenario struct {
	RuleID          string  `json:"rule_id"`
	RuleDescription string  `json:"rule_description"`
	Probability     float64 `json:"probability"`
	WhoBreaks       string  `json:"who_breaks"`
	HowItHappens    string  `json:"how_it_happens"`
	ImpactOnCompany string  `json:"impact_on_company"`
	HowToBeFirst    string  `json:"how_to_be_first"`
}

// Simulation orchestrates a full FRACTURE simulation run.
type Simulation struct {
	cfg     SimulationConfig
	voter   *Voter
	results []RoundResult
	events  []FractureEvent
	tokens  int
	startAt time.Time
}

// NewSimulation creates a ready-to-run simulation.
func NewSimulation(cfg SimulationConfig) *Simulation {
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
	}
	if cfg.MaxRounds == 0 {
		cfg.MaxRounds = 40
	}
	return &Simulation{
		cfg:   cfg,
		voter: NewVoter(cfg.Agents),
	}
}

// Run executes the simulation and streams RoundResults to the returned channel.
// The channel is closed when all rounds are done or ctx is cancelled.
func (s *Simulation) Run(ctx context.Context) <-chan RoundResult {
	out := make(chan RoundResult, 4)
	s.startAt = time.Now()

	go func() {
		defer close(out)

		for round := 1; round <= s.cfg.MaxRounds; round++ {
			select {
			case <-ctx.Done():
				return
			default:
			}

			roundStart := time.Now()
			tension := s.cfg.World.CalculateTension()

			// Run all agents in parallel
			actions := s.runAgentsParallel(ctx, round, tension)

			// Process FRACTURE POINT proposals
			var fractureEvents []FractureEvent
			for _, action := range actions {
				if action.IsFractureProposal && action.Proposal != nil {
					event := s.processFractureProposal(ctx, *action.Proposal, actions, round)
					fractureEvents = append(fractureEvents, event)
					s.events = append(s.events, event)
				}
			}

			// Apply tension deltas to world
			for _, action := range actions {
				for ruleID, delta := range action.TensionDelta {
					s.cfg.World.IncreaseTension(ruleID, delta)
				}
			}

			// Count tokens
			for _, a := range actions {
				s.tokens += a.TokensUsed
			}

			rr := RoundResult{
				Round:          round,
				Actions:        actions,
				FractureEvents: fractureEvents,
				Tension:        tension,
				WorldSnapshot:  s.cfg.World.Snapshot(round),
				ElapsedMs:      time.Since(roundStart).Milliseconds(),
			}
			s.results = append(s.results, rr)

			select {
			case out <- rr:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out
}

// runAgentsParallel runs all agents concurrently and collects their actions.
func (s *Simulation) runAgentsParallel(ctx context.Context, round int, tension float64) []AgentAction {
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		actions = make([]AgentAction, 0, len(s.cfg.Agents))
	)

	for _, agent := range s.cfg.Agents {
		wg.Add(1)
		go func(a Agent) {
			defer wg.Done()
			action, err := a.React(ctx, s.cfg.World, s.cfg.Memory, round, tension)
			if err != nil {
				action = AgentAction{
					AgentID:   a.ID(),
					AgentType: a.Type(),
					Text:      fmt.Sprintf("[error: %v]", err),
				}
			}
			mu.Lock()
			actions = append(actions, action)
			mu.Unlock()
		}(agent)
	}
	wg.Wait()
	return actions
}

// processFractureProposal runs the voting mechanism and applies the rule if accepted.
func (s *Simulation) processFractureProposal(
	ctx context.Context,
	proposal RuleProposal,
	actions []AgentAction,
	round int,
) FractureEvent {
	voteResult, breakdown := s.voter.Vote(ctx, proposal, actions)

	event := FractureEvent{
		Round:         round,
		ProposedBy:    proposal.ProposedByAgent,
		Proposal:      proposal,
		Accepted:      voteResult,
		VoteBreakdown: breakdown,
	}

	if voteResult {
		s.cfg.World.ApplyProposal(proposal)
	}
	return event
}

// Finalize collects the final result after Run() channel is drained.
func (s *Simulation) Finalize() SimulationResult {
	return SimulationResult{
		SimulationID:   s.cfg.ID,
		Question:       s.cfg.Question,
		Rounds:         s.results,
		FractureEvents: s.events,
		FinalWorld:     s.cfg.World.Snapshot(len(s.results)),
		TensionMap:     s.cfg.World.TensionMap,
		TotalTokens:    s.tokens,
		DurationMs:     time.Since(s.startAt).Milliseconds(),
	}
}
