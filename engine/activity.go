package engine

import (
	"sync"
	"time"
)

// ActivityEvent is a single observable moment in a running simulation.
// Emitted by agents as they act, vote, or trigger fractures — streamed live to the UI.
type ActivityEvent struct {
	SimulationID string  `json:"simulation_id"`
	Round        int     `json:"round"`
	AgentID      string  `json:"agent_id"`
	AgentName    string  `json:"agent_name"`
	Archetype    string  `json:"archetype"`    // "disruptor" | "conformist" | "system"
	ActionType   string  `json:"action_type"`  // "react" | "propose" | "fracture" | "council"
	Snippet      string  `json:"snippet"`      // first 220 runes of the agent's output
	TokensUsed   int     `json:"tokens_used"`
	TotalTokens  int     `json:"total_tokens"` // running cumulative for this simulation
	Tension      float64 `json:"tension"`
	RuleID       string  `json:"rule_id,omitempty"`
	Accepted     *bool   `json:"accepted,omitempty"` // only set for "fracture" events
	Ts           int64   `json:"ts"`                 // unix millis
}

// ActivityBus is an in-memory pub/sub bus that fans out ActivityEvents to all
// SSE subscribers for a given simulation ID. Non-blocking: slow subscribers
// receive dropped events rather than stalling the simulation goroutine.
type ActivityBus struct {
	mu   sync.RWMutex
	subs map[string][]chan ActivityEvent
}

// GlobalActivityBus is the process-wide singleton used by all simulations.
var GlobalActivityBus = &ActivityBus{
	subs: make(map[string][]chan ActivityEvent),
}

// Subscribe creates and registers a buffered channel for the given simulation.
// The caller must call Unsubscribe when the SSE connection closes.
func (b *ActivityBus) Subscribe(simID string) chan ActivityEvent {
	ch := make(chan ActivityEvent, 256)
	b.mu.Lock()
	b.subs[simID] = append(b.subs[simID], ch)
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes and closes the channel registered by Subscribe.
func (b *ActivityBus) Unsubscribe(simID string, ch chan ActivityEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subs := b.subs[simID]
	for i, s := range subs {
		if s == ch {
			b.subs[simID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			return
		}
	}
}

// Emit broadcasts ev to all current subscribers. Drops the event if a
// subscriber's buffer is full to avoid blocking the simulation goroutine.
func (b *ActivityBus) Emit(simID string, ev ActivityEvent) {
	ev.Ts = time.Now().UnixMilli()
	b.mu.RLock()
	subs := b.subs[simID]
	b.mu.RUnlock()
	for _, ch := range subs {
		select {
		case ch <- ev:
		default:
		}
	}
}

// activitySnippet returns at most n runes of text, appending … if truncated.
func activitySnippet(text string, n int) string {
	runes := []rune(text)
	if len(runes) <= n {
		return text
	}
	return string(runes[:n]) + "…"
}

// firstRuleKey returns the first key from a tension delta map, or "".
func firstRuleKey(m map[string]float64) string {
	for k := range m {
		return k
	}
	return ""
}
