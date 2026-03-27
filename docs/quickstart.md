# FRACTURE — Quickstart

FRACTURE is a local strategic simulation tool. You ask a strategic question,
56 AI agents debate how market rules evolve over 20–50 rounds, and when
tension peaks a **FRACTURE POINT** occurs — a rule mutates and the agents
adapt. The result is a structured report: probable future, rupture scenarios,
and a concrete action playbook.

---

## Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| Go   | 1.22+   | `go version` |
| Node | 20+     | only needed for dashboard build |
| pnpm | 9+      | `npm i -g pnpm` |

At least one LLM API key is required for full simulation (see [API keys](#2-configure-api-keys)).
Without keys, FRACTURE runs in **heuristic mode** (deterministic, no LLM calls).

---

## 1. Install

```bash
git clone https://github.com/mkvinicius/fracture
cd fracture

# Tests only — no dashboard required
go test ./...

# Full binary (embeds the React dashboard)
make build        # builds dashboard + Go binary
./fracture        # starts on http://localhost:3000
```

Or with `go install`:

```bash
go install github.com/fracture/fracture@latest
fracture
```

---

## 2. Configure API keys

Open `http://localhost:3000` → **Settings** → paste your key(s).

Or via the API:

```bash
curl -X POST http://localhost:3000/api/v1/keys/validate \
  -H 'Content-Type: application/json' \
  -d '{"provider":"openai","key":"sk-..."}'
```

Supported providers:

| Provider   | Models used |
|------------|-------------|
| OpenAI     | `gpt-4o-mini` (conformist) + `gpt-4o` (disruptor) |
| Anthropic  | `claude-haiku-4-5` (conformist) + `claude-sonnet-4` (disruptor + synthesis) |
| Google     | `gemini-1.5-flash` (conformist + coherence) + `gemini-1.5-pro` (disruptor) |
| Ollama     | any local model (set `ollama_model` in config) |

Multiple providers can be active simultaneously — FRACTURE routes each agent
role to the most appropriate model.

Optional web-search keys (improve DeepSearch quality):

| Variable        | Purpose |
|-----------------|---------|
| `TAVILY_API_KEY`  | Primary web search |
| `SERPAPI_KEY`     | Fallback web search |

---

## 3. Run your first simulation

**Via the dashboard** — click **New Simulation**, fill in the form, click **Run**.

**Via the API:**

```bash
# Start simulation
SIM=$(curl -s -X POST http://localhost:3000/api/v1/simulations \
  -H 'Content-Type: application/json' \
  -d '{
    "question": "What happens if a major competitor launches a free tier?",
    "department": "market",
    "mode": "standard"
  }' | jq -r .id)

echo "Simulation ID: $SIM"

# Poll until done
watch -n2 "curl -s http://localhost:3000/api/v1/simulations/$SIM | jq '{status,current_round,fracture_count}'"

# Fetch the full report
curl -s http://localhost:3000/api/v1/simulations/$SIM/report | jq .
```

---

## 4. Understand the output

A completed simulation produces a `FullReport` with five sections:

### Probable Future
The most likely evolution if no rules break — narrative + 6/12/24-month timeline.

### Tension Map
Every active rule ranked by tension level (green → yellow → orange → red).
Red rules are fracture candidates.

### Rupture Scenarios
Specific rules that are likely to break, with:
- Probability estimate
- Who breaks it (agent archetype)
- How it happens
- Impact on your company

### Fracture Events
Every rule-mutation that actually occurred during the simulation, with the
vote breakdown that decided it.

### Action Playbook
Concrete recommendations derived from the simulation results.

---

## 5. Modes

| Mode       | Rounds | Ensemble runs | Use when |
|------------|--------|---------------|----------|
| `standard` | 20     | 1             | Quick exploration, daily use |
| `premium`  | 50     | 2             | High-stakes decisions, board prep |

Premium mode runs the simulation twice independently and merges results —
divergences between runs signal low-confidence areas.

---

## 6. Domains

Each simulation is anchored to one of seven rule domains:

| Domain        | What it models |
|---------------|---------------|
| `market`      | Competition, pricing, customer behaviour |
| `technology`  | Platform shifts, AI adoption, infrastructure |
| `regulation`  | Compliance, policy, legal risk |
| `behavior`    | Talent, culture, organisational dynamics |
| `culture`     | Social norms, media, public sentiment |
| `geopolitics` | Trade, sovereignty, macro risk |
| `finance`     | Capital markets, funding, cost of money |

---

## 7. Quick Pulse (no simulation needed)

For a rapid tension score without running a full simulation:

```bash
curl -s -X POST http://localhost:3000/api/v1/pulse \
  -H 'Content-Type: application/json' \
  -d '{"situation":"We are considering a 30% price increase next quarter","domain":"market"}' \
  | jq '{score,level,summary}'
```

Returns a 0–100 score, level (`low/medium/high/critical`), one-line summary,
and three specific risks — in seconds.

---

## 8. Submit feedback and improve accuracy

After a simulated scenario plays out in the real world, submit feedback.
FRACTURE uses it to recalibrate archetype accuracy weights so future
simulations for your company are more accurate.

```bash
curl -X POST http://localhost:3000/api/v1/simulations/$SIM/feedback \
  -H 'Content-Type: application/json' \
  -d '{
    "outcome": "accurate",
    "predicted_fracture": "Price war erupts within 3 months",
    "actual_outcome": "Competitor launched free tier; our churn rose 18%",
    "delta_score": 0.8
  }'
```

`delta_score` range: `-1.0` (completely wrong) → `0.0` (neutral) → `1.0` (exact).

---

## 9. Compare simulations

Run the same question with different modes or domains, then compare:

```bash
curl "http://localhost:3000/api/v1/simulations/compare?ids=$SIM_A,$SIM_B" \
  | jq '{common_fractures,divergent_fractures,confidence_delta}'
```

---

## 10. Export

```bash
# Markdown (for Notion, Confluence, etc.)
curl -O http://localhost:3000/api/v1/simulations/$SIM/export/markdown

# JSON (for downstream processing)
curl -O http://localhost:3000/api/v1/simulations/$SIM/export/json
```

---

## Development

```bash
make dev        # dashboard dev server + Go server in parallel (hot reload)
make test       # go test ./... -race
make lint       # golangci-lint
make coverage   # HTML coverage report → coverage.html
make fmt        # gofmt -w
make tidy       # go mod tidy
```

See [docs/api.md](api.md) for the full API reference.
