package engine

import (
	"sync"
	"testing"
	"time"
)

// newBus creates a fresh ActivityBus for tests so they are isolated from
// GlobalActivityBus and from each other.
func newBus() *ActivityBus {
	return &ActivityBus{subs: make(map[string][]chan ActivityEvent)}
}

// receiveWithTimeout reads one event from ch or fails the test if nothing
// arrives within the deadline.
func receiveWithTimeout(t *testing.T, ch chan ActivityEvent, d time.Duration) ActivityEvent {
	t.Helper()
	select {
	case ev, ok := <-ch:
		if !ok {
			t.Fatal("channel was closed before an event arrived")
		}
		return ev
	case <-time.After(d):
		t.Fatalf("timed out waiting for event after %s", d)
		return ActivityEvent{}
	}
}

// assertClosed verifies that ch is closed (no event is pending).
func assertClosed(t *testing.T, ch chan ActivityEvent) {
	t.Helper()
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("expected channel to be closed, but received an event")
		}
		// channel closed as expected
	case <-time.After(50 * time.Millisecond):
		t.Fatal("expected channel to be closed, but it was still open after 50ms")
	}
}

// ── 1. Subscribe + Emit ───────────────────────────────────────────────────────

func TestSubscribeAndEmit(t *testing.T) {
	b := newBus()
	const simID = "sim-bus-001"

	ch := b.Subscribe(simID)
	defer b.Unsubscribe(simID, ch)

	b.Emit(simID, ActivityEvent{SimulationID: simID, Round: 1, AgentID: "agent-1"})

	ev := receiveWithTimeout(t, ch, time.Second)
	if ev.SimulationID != simID {
		t.Errorf("SimulationID = %q; want %q", ev.SimulationID, simID)
	}
	if ev.Round != 1 {
		t.Errorf("Round = %d; want 1", ev.Round)
	}
	if ev.AgentID != "agent-1" {
		t.Errorf("AgentID = %q; want %q", ev.AgentID, "agent-1")
	}
}

// ── 2. Multiple subscribers ───────────────────────────────────────────────────

func TestMultipleSubscribersReceiveSameEvent(t *testing.T) {
	b := newBus()
	const simID = "sim-bus-002"

	ch1 := b.Subscribe(simID)
	ch2 := b.Subscribe(simID)
	defer b.Unsubscribe(simID, ch1)
	defer b.Unsubscribe(simID, ch2)

	want := ActivityEvent{SimulationID: simID, Round: 5, AgentID: "agent-x", ActionType: "fracture"}
	b.Emit(simID, want)

	ev1 := receiveWithTimeout(t, ch1, time.Second)
	ev2 := receiveWithTimeout(t, ch2, time.Second)

	for _, ev := range []ActivityEvent{ev1, ev2} {
		if ev.Round != want.Round {
			t.Errorf("Round = %d; want %d", ev.Round, want.Round)
		}
		if ev.AgentID != want.AgentID {
			t.Errorf("AgentID = %q; want %q", ev.AgentID, want.AgentID)
		}
		if ev.ActionType != want.ActionType {
			t.Errorf("ActionType = %q; want %q", ev.ActionType, want.ActionType)
		}
	}
}

// ── 3. Unsubscribe ────────────────────────────────────────────────────────────

func TestUnsubscribeClosesChannelAndStopsDelivery(t *testing.T) {
	b := newBus()
	const simID = "sim-bus-003"

	ch := b.Subscribe(simID)
	b.Unsubscribe(simID, ch)

	// Channel should be closed immediately after Unsubscribe.
	assertClosed(t, ch)

	// Emitting after unsubscribe must not panic (drops silently).
	b.Emit(simID, ActivityEvent{SimulationID: simID})
}

// ── 4. Emit non-blocking when buffer is full ──────────────────────────────────

func TestEmitNonBlockingWhenBufferFull(t *testing.T) {
	b := newBus()
	const simID = "sim-bus-004"

	ch := b.Subscribe(simID)
	defer b.Unsubscribe(simID, ch)

	// Fill the 256-slot buffer completely.
	filler := ActivityEvent{SimulationID: simID, Round: 0}
	for i := 0; i < 256; i++ {
		b.Emit(simID, filler)
	}

	// The 257th emit must return immediately without blocking.
	done := make(chan struct{})
	go func() {
		b.Emit(simID, ActivityEvent{SimulationID: simID, Round: 257})
		close(done)
	}()

	select {
	case <-done:
		// non-blocking as expected
	case <-time.After(time.Second):
		t.Fatal("Emit blocked on a full channel")
	}
}

// ── 5. Emit sets Ts ───────────────────────────────────────────────────────────

func TestEmitSetsTimestamp(t *testing.T) {
	b := newBus()
	const simID = "sim-bus-005"

	ch := b.Subscribe(simID)
	defer b.Unsubscribe(simID, ch)

	before := time.Now().UnixMilli()
	b.Emit(simID, ActivityEvent{SimulationID: simID})
	after := time.Now().UnixMilli()

	ev := receiveWithTimeout(t, ch, time.Second)
	if ev.Ts == 0 {
		t.Fatal("Ts was not set by Emit")
	}
	if ev.Ts < before || ev.Ts > after {
		t.Errorf("Ts %d is outside expected range [%d, %d]", ev.Ts, before, after)
	}
}

// ── 6. Different simIDs are isolated ─────────────────────────────────────────

func TestDifferentSimIDsAreIsolated(t *testing.T) {
	b := newBus()
	const simA = "sim-bus-006a"
	const simB = "sim-bus-006b"

	chA := b.Subscribe(simA)
	chB := b.Subscribe(simB)
	defer b.Unsubscribe(simA, chA)
	defer b.Unsubscribe(simB, chB)

	// Emit only to simA.
	b.Emit(simA, ActivityEvent{SimulationID: simA, Round: 42})

	// chA should receive the event.
	ev := receiveWithTimeout(t, chA, time.Second)
	if ev.Round != 42 {
		t.Errorf("chA Round = %d; want 42", ev.Round)
	}

	// chB must not receive anything.
	select {
	case unexpected := <-chB:
		t.Errorf("chB unexpectedly received event: %+v", unexpected)
	case <-time.After(50 * time.Millisecond):
		// correct: no event for a different simID
	}
}

// ── 7. activitySnippet — text shorter than n ──────────────────────────────────

func TestActivitySnippetShort(t *testing.T) {
	got := activitySnippet("hello", 10)
	if got != "hello" {
		t.Errorf("got %q; want %q", got, "hello")
	}
}

// ── 8. activitySnippet — text exactly n runes ────────────────────────────────

func TestActivitySnippetExact(t *testing.T) {
	got := activitySnippet("hello", 5)
	if got != "hello" {
		t.Errorf("got %q; want %q", got, "hello")
	}
}

// ── 9. activitySnippet — text longer than n ──────────────────────────────────

func TestActivitySnippetLong(t *testing.T) {
	got := activitySnippet("hello world", 5)
	const want = "hello…"
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

// ── 10. activitySnippet — unicode / multi-byte runes ─────────────────────────

func TestActivitySnippetUnicode(t *testing.T) {
	// Each Japanese character is 3 bytes in UTF-8, but 1 rune.
	// "日本語テスト" = 6 runes; truncating to 3 should give "日本語…"
	text := "日本語テスト"
	got := activitySnippet(text, 3)
	const want = "日本語…"
	if got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}

// ── 11. firstRuleKey — empty map ─────────────────────────────────────────────

func TestFirstRuleKeyEmptyMap(t *testing.T) {
	got := firstRuleKey(map[string]float64{})
	if got != "" {
		t.Errorf("got %q; want empty string", got)
	}
}

// ── 12. firstRuleKey — non-empty map ─────────────────────────────────────────

func TestFirstRuleKeyNonEmpty(t *testing.T) {
	m := map[string]float64{
		"rule-alpha": 0.5,
		"rule-beta":  1.2,
		"rule-gamma": 0.9,
	}
	got := firstRuleKey(m)
	if _, ok := m[got]; !ok {
		t.Errorf("firstRuleKey returned %q which is not a key in the map", got)
	}
	if got == "" {
		t.Error("firstRuleKey returned empty string for non-empty map")
	}
}

// ── Concurrent safety smoke test ─────────────────────────────────────────────

// TestEmitConcurrentlySafe verifies that multiple goroutines can call Emit
// simultaneously without data races (run with -race).
// Subscribers are registered before and unsubscribed after all emits are done
// to avoid the pre-existing race between Emit's RLock and Unsubscribe's close
// (which is a separate concern in the production code, not under test here).
func TestEmitConcurrentlySafe(t *testing.T) {
	b := newBus()
	const simID = "sim-bus-race"
	const workers = 8
	const emits = 50

	// Subscribe all channels before any emitting starts.
	channels := make([]chan ActivityEvent, workers)
	for i := 0; i < workers; i++ {
		channels[i] = b.Subscribe(simID)
	}

	// Emit concurrently from multiple goroutines.
	var wg sync.WaitGroup
	for g := 0; g < workers; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < emits; i++ {
				b.Emit(simID, ActivityEvent{SimulationID: simID, Round: id*emits + i})
			}
		}(g)
	}
	wg.Wait()

	// All emits are done; now unsubscribe safely.
	for _, ch := range channels {
		b.Unsubscribe(simID, ch)
	}
}
