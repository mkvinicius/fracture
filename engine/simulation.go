package engine

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
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
	// Optional: council support. If CouncilLLM is nil, councils are skipped.
	CouncilLLM LLMCaller
	// Optional: if set, the top-N highest-power agents vote via LLM reasoning
	// instead of the heuristic. Produces more realistic fracture outcomes.
	VotingLLM LLMCaller
	Mode       ModeConfig
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
	Round         int          `json:"round"`
	ProposedBy    string       `json:"proposed_by"`
	Proposal      RuleProposal `json:"proposal"`
	Accepted      bool         `json:"accepted"`
	VoteBreakdown []VoteRecord `json:"vote_breakdown"`
	// Confidence is a 0.0-1.0 measure of how certain this fracture outcome is.
	// Derived from vote share, early convergence boost, and Shannon entropy penalty.
	Confidence    float64  `json:"confidence"`
	EvidenceTrail []string `json:"evidence_trail,omitempty"`
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

// AgentProposalAccuracy tracks how often each disruptor's fracture proposals were accepted.
// Persisted by the handler layer to calibrate future simulations.
type AgentProposalAccuracy struct {
	AgentID   string  `json:"agent_id"`
	AgentName string  `json:"agent_name"`
	Proposed  int     `json:"proposed"`
	Accepted  int     `json:"accepted"`
	Rate      float64 `json:"rate"` // accepted / proposed, 0 if proposed == 0
}

// SimulationResult is the final output after all rounds complete.
type SimulationResult struct {
	SimulationID     string                  `json:"simulation_id"`
	Question         string                  `json:"question"`
	Rounds           []RoundResult           `json:"rounds"`
	FractureEvents   []FractureEvent         `json:"fracture_events"`
	FinalWorld       WorldSnapshot           `json:"final_world"`
	ProbableFuture   string                  `json:"probable_future"`
	TensionMap       map[string]float64      `json:"tension_map"`
	RuptureScenarios []RuptureScenario       `json:"rupture_scenarios"`
	Coalitions       []Coalition             `json:"coalitions"`
	ActionPlaybook   *ActionPlaybook         `json:"action_playbook,omitempty"`
	ProposalAccuracy []AgentProposalAccuracy `json:"proposal_accuracy,omitempty"`
	TotalTokens      int                     `json:"total_tokens"`
	DurationMs       int64                   `json:"duration_ms"`
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
	cfg           SimulationConfig
	voter         *Voter
	results       []RoundResult
	events        []FractureEvent
	tokens        int
	atomicTokens  int64 // updated atomically per-agent for live feed
	startAt       time.Time
	proposalStats map[string]*agentProposalStat // agentID -> stats
}

type agentProposalStat struct {
	name     string
	proposed int
	accepted int
}

// NewSimulation creates a ready-to-run simulation.
func NewSimulation(cfg SimulationConfig) *Simulation {
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
	}
	if cfg.MaxRounds == 0 {
		cfg.MaxRounds = cfg.Mode.MaxRounds
		if cfg.MaxRounds == 0 {
			cfg.MaxRounds = DefaultConfigForMode(ModeStandard).MaxRounds
		}
	}
	if cfg.Mode.CouncilInterval == 0 {
		cfg.Mode.CouncilInterval = 5
	}
	return &Simulation{
		cfg:           cfg,
		voter:         NewVoter(cfg.Agents, cfg.VotingLLM),
		proposalStats: make(map[string]*agentProposalStat),
	}
}

// Run executes the simulation and streams RoundResults to the returned channel.
// The channel is closed when all rounds are done or ctx is cancelled.
func (s *Simulation) Run(ctx context.Context) <-chan RoundResult {
	out := make(chan RoundResult, 4)
	s.startAt = time.Now()

	// Build councils once if a council LLM is configured
	var councils []Council
	if s.cfg.CouncilLLM != nil {
		councils = BuildCouncils(s.cfg.CouncilLLM)
	}

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

			// Run councils every CouncilInterval rounds (non-blocking on error)
			interval := s.cfg.Mode.CouncilInterval
			if interval <= 0 {
				interval = 5
			}
			if len(councils) > 0 && round%interval == 0 {
				RunAllCouncils(ctx, councils, s.cfg.World, round)
			}

			// Update social context for the next round: agents can see top signals
			s.cfg.World.SetPrevRoundInfluence(s.buildInfluenceSummary(actions))

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

			// Emit live activity event to any connected SSE subscribers.
			total := int(atomic.AddInt64(&s.atomicTokens, int64(action.TokensUsed)))
			archetype := "conformist"
			if a.Type() == AgentDisruptor {
				archetype = "disruptor"
			}
			actionType := "react"
			if action.IsFractureProposal {
				actionType = "propose"
			}
			GlobalActivityBus.Emit(s.cfg.ID, ActivityEvent{
				SimulationID: s.cfg.ID,
				Round:        round,
				AgentID:      a.ID(),
				AgentName:    a.Personality().Name,
				Archetype:    archetype,
				ActionType:   actionType,
				Snippet:      activitySnippet(action.Text, 220),
				TokensUsed:   action.TokensUsed,
				TotalTokens:  total,
				Tension:      tension,
				RuleID:       firstRuleKey(action.TensionDelta),
			})

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
		Confidence:    calculateFractureConfidence(breakdown, round, s.cfg.MaxRounds),
	}

	if voteResult {
		s.cfg.World.ApplyProposal(proposal)
	}

	// Emit fracture event to live activity feed.
	accepted := voteResult
	agentNameForFracture := proposal.ProposedByAgent
	for _, a := range s.cfg.Agents {
		if a.ID() == proposal.ProposedByAgent {
			agentNameForFracture = a.Personality().Name
			break
		}
	}
	total := int(atomic.LoadInt64(&s.atomicTokens))
	GlobalActivityBus.Emit(s.cfg.ID, ActivityEvent{
		SimulationID: s.cfg.ID,
		Round:        round,
		AgentID:      proposal.ProposedByAgent,
		AgentName:    agentNameForFracture,
		Archetype:    "disruptor",
		ActionType:   "fracture",
		Snippet:      activitySnippet(proposal.NewDescription, 220),
		TotalTokens:  total,
		Tension:      s.cfg.World.CalculateTension(),
		RuleID:       proposal.OriginalRuleID,
		Accepted:     &accepted,
	})

	// Track proposal accuracy per agent for post-simulation calibration.
	agentID := proposal.ProposedByAgent
	if agentID != "" {
		stat, ok := s.proposalStats[agentID]
		if !ok {
			// Find the agent name for reporting
			name := agentID
			for _, a := range s.cfg.Agents {
				if a.ID() == agentID {
					name = a.Personality().Name
					break
				}
			}
			stat = &agentProposalStat{name: name}
			s.proposalStats[agentID] = stat
		}
		stat.proposed++
		if voteResult {
			stat.accepted++
		}
	}

	return event
}

// buildInfluenceSummary creates a brief social context string from the top-3
// most powerful agents' actions. Used to inform agents about the social landscape.
func (s *Simulation) buildInfluenceSummary(actions []AgentAction) string {
	if len(actions) == 0 {
		return ""
	}

	// Build power-weight lookup
	powerByID := make(map[string]float64, len(s.cfg.Agents))
	nameByID := make(map[string]string, len(s.cfg.Agents))
	for _, a := range s.cfg.Agents {
		p := a.Personality()
		powerByID[a.ID()] = p.PowerWeight
		nameByID[a.ID()] = p.Name
	}

	// Sort actions by power weight descending
	sorted := make([]AgentAction, len(actions))
	copy(sorted, actions)
	sort.Slice(sorted, func(i, j int) bool {
		return powerByID[sorted[i].AgentID] > powerByID[sorted[j].AgentID]
	})

	const maxSignals = 4
	var sb strings.Builder
	sb.WriteString("Signals from the most influential players last round:\n")
	for i, a := range sorted {
		if i >= maxSignals || a.Text == "" {
			break
		}
		name := nameByID[a.AgentID]
		if name == "" {
			name = a.AgentID
		}
		sb.WriteString(fmt.Sprintf("- %s: %s\n", name, truncate(a.Text, 120)))
	}
	return sb.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

// ─── Calibration ─────────────────────────────────────────────────────────────

// AgentCalibration holds a calibrated accuracy weight for a single agent or archetype.
// Sourced from memory.ArchetypeCalibration and converted at the handler layer.
type AgentCalibration struct {
	AgentID        string
	AccuracyWeight float64 // 0.3 (less trusted) to 2.0 (highly trusted); 1.0 = neutral
}

// calibratedAgent wraps an Agent and overrides its PowerWeight based on calibration.
type calibratedAgent struct {
	Agent
	calibratedPersonality Personality
}

func (c *calibratedAgent) Personality() Personality { return c.calibratedPersonality }

// ApplyCalibration returns a new agent slice where each agent's PowerWeight has been
// scaled by its AccuracyWeight from calibration. Agents not present in calibrations
// are returned unchanged. PowerWeight is clamped to [0.1, 10.0].
func ApplyCalibration(agents []Agent, calibrations []AgentCalibration) []Agent {
	if len(calibrations) == 0 {
		return agents
	}

	// Build lookup: agentID → accuracy weight
	weightByID := make(map[string]float64, len(calibrations))
	for _, cal := range calibrations {
		weightByID[cal.AgentID] = cal.AccuracyWeight
	}

	result := make([]Agent, len(agents))
	for i, a := range agents {
		w, ok := weightByID[a.ID()]
		if !ok || w == 1.0 {
			result[i] = a
			continue
		}
		p := a.Personality()
		base := p.PowerWeight
		if base == 0 {
			base = 0.5
		}
		adjusted := base * w
		if adjusted < 0.1 {
			adjusted = 0.1
		}
		if adjusted > 10.0 {
			adjusted = 10.0
		}
		p.PowerWeight = adjusted
		result[i] = &calibratedAgent{Agent: a, calibratedPersonality: p}
	}
	return result
}

// calculateFractureConfidence produces a 0.0-1.0 confidence score for a fracture event.
// Formula:
//   - Base = weighted yes-share (or no-share if rejected)
//   - Early convergence boost: +0.10 if fracture fired in first 30% of rounds
//   - Shannon entropy penalty: subtract H(p) * 0.15 where H(p) = -p*log2(p)-(1-p)*log2(1-p)
//
// Result is clamped to [0.05, 0.95].
func calculateFractureConfidence(breakdown []VoteRecord, round, maxRounds int) float64 {
	if len(breakdown) == 0 {
		return 0.5
	}

	var totalWeight, yesWeight float64
	for _, v := range breakdown {
		totalWeight += v.Weight
		if v.Vote {
			yesWeight += v.Weight
		}
	}
	if totalWeight == 0 {
		return 0.5
	}

	p := yesWeight / totalWeight
	// Use the majority share as base (whichever side won)
	base := p
	if base < 0.5 {
		base = 1.0 - p
	}

	// Early convergence boost: fires in first 30% of rounds
	var boost float64
	if maxRounds > 0 && float64(round)/float64(maxRounds) <= 0.30 {
		boost = 0.10
	}

	// Shannon entropy penalty
	entropy := 0.0
	if p > 0 && p < 1 {
		entropy = -(p*(math.Log(p)/math.Log(2)) + (1-p)*(math.Log(1-p)/math.Log(2)))
	}
	penalty := entropy * 0.15

	conf := base + boost - penalty
	if conf < 0.05 {
		return 0.05
	}
	if conf > 0.95 {
		return 0.95
	}
	return conf
}

// Finalize collects the final result after Run() channel is drained.
// It takes a safe snapshot of the world so the returned TensionMap is an
// immutable copy, not a live reference to the mutable World map.
func (s *Simulation) Finalize() SimulationResult {
	finalWorld := s.cfg.World.Snapshot(len(s.results))

	// Build proposal accuracy slice for calibration
	var accuracy []AgentProposalAccuracy
	for id, stat := range s.proposalStats {
		rate := 0.0
		if stat.proposed > 0 {
			rate = float64(stat.accepted) / float64(stat.proposed)
		}
		accuracy = append(accuracy, AgentProposalAccuracy{
			AgentID:   id,
			AgentName: stat.name,
			Proposed:  stat.proposed,
			Accepted:  stat.accepted,
			Rate:      rate,
		})
	}

	return SimulationResult{
		SimulationID:     s.cfg.ID,
		Question:         s.cfg.Question,
		Rounds:           s.results,
		FractureEvents:   s.events,
		FinalWorld:       finalWorld,
		TensionMap:       finalWorld.TensionMap, // safe copy from Snapshot — not a live reference
		ProposalAccuracy: accuracy,
		TotalTokens:      s.tokens,
		DurationMs:       time.Since(s.startAt).Milliseconds(),
	}
}
