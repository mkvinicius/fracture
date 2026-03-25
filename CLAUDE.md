# CLAUDE.md — FRACTURE v2.5.0

This file provides context for AI assistants working on this codebase.

## Project Overview

FRACTURE is a local strategic simulation tool. A user asks a strategic question, DeepSearch gathers real-world context, and 56 AI agents simulate how market rules evolve over 20–50 rounds. When tension exceeds a threshold, a FRACTURE POINT occurs and a rule mutates.

**Version**: 2.5.0
**Go module**: `github.com/fracture/fracture`
**Main binary**: `main.go` (embeds `dashboard/dist`)

## Repository Structure

```
main.go              Entry point: logger, DB, router, HTTP server
api/handler.go       All REST handlers + simulation runner goroutine
engine/
  engine.go          RunSimulation(): rounds loop, tension, voting
  agents.go          56 agents (37 conformist, 19 disruptor), Personality
  world.go           World struct: Rules map, TensionMap, Evidence string
  world_domains.go   DefaultWorldForDomain(), 7 domain rule sets
  report.go          FullReport: narrative, fractures, playbook, ensemble
  export.go          ReportToMarkdown()
  compare.go         CompareReports(): common/divergent fractures, deltas
deepsearch/
  agent.go           Agent.Research(): web search + LLM synthesis loop
  domain_research.go DomainResearcher: semaphore, cache, ResearchDomains()
db/
  db.go              SQLite helpers: jobs, simulations, domain contexts
  schema.sql         Tables: config, simulations, simulation_jobs, feedback,
                     audit_log, agent_memory, archetype_calibration,
                     domain_contexts, rag_documents, domain_research_state
  migrations.go      Idempotent versioned migrations (schema_version table)
  rounds.go          SaveRound(), GetRoundTensions() for convergence chart
memory/memory.go     TF-IDF RAG: index + cosine similarity search
security/            HMAC signer, input sanitizer, audit logger
telemetry/           Anonymous opt-in startup ping
updater/updater.go   CurrentVersion = "2.5.0", GitHub releases check
dashboard/src/       React + TypeScript SPA (Vite)
```

## REST API

Base path: `/api/v1/` (legacy `/api/*` redirects via HTTP 308)

### Core simulation flow

```
POST /api/v1/simulate              { question, department, rounds, company?, mode? }
GET  /api/v1/simulations           → []JobRow (list with status)
GET  /api/v1/simulations/{id}      → JobRow (live progress)
GET  /api/v1/simulations/{id}/report → FullReport JSON
GET  /api/v1/simulations/{id}/export/markdown → Markdown download
GET  /api/v1/simulations/{id}/export/json    → JSON download
GET  /api/v1/simulations/{id}/events  → []TensionPoint (for convergence chart)
GET  /api/v1/simulations/compare?ids=a,b,c  → ComparisonReport
DELETE /api/v1/simulations/{id}
```

### Feedback & calibration

```
POST /api/v1/feedback              { simulation_id, outcome, notes?, delta_score? }
GET  /api/v1/calibration           → []ArchetypeCalibration
```

### Knowledge base

```
GET  /api/v1/archetypes            (+ ?company_id= to include custom)
GET  /api/v1/rules
GET  /api/v1/rules/domain/{domain}
POST /api/v1/documents             { company_id, doc_type, content, metadata? }
GET  /api/v1/documents/{company_id}
```

### System

```
GET  /api/v1/health
GET  /api/v1/config
PUT  /api/v1/config
GET  /api/v1/audit
```

## Key Types

```go
// engine/world.go
type World struct {
    Rules      map[string]*Rule
    TensionMap map[string]float64
    Evidence   string   // from DeepSearch, read-only in simulation
}

// engine/report.go
type FullReport struct {
    SimulationID  string
    Question      string
    Domain        engine.RuleDomain
    Narrative     string
    FractureEvents []FractureEvent
    Ensemble      EnsembleReport
    Playbook      []string
    // ...
}

// db/db.go
type DomainContextRow struct {
    SimulationID      string
    Domain            string
    Context           string
    Signals           string  // JSON array
    StabilityModifier float64
    Confidence        float64
    AffectedRules     string  // JSON array
    SentimentScore    float64
    CachedAt          int64
}

type JobRow struct {
    ID, Status, Question, Department string
    Rounds     int
    // live progress: CurrentRound, CurrentTension, FractureCount, ...
}
```

## Simulation Pipeline

1. `POST /simulate` → creates JobRow with status=queued
2. Goroutine starts: status=researching → DeepSearch (if API keys set)
3. `SynthesizeDomainContext()` → saves domain contexts to DB
4. Status=running → `engine.RunSimulation()` → round loop
5. Each round: agents react → vote → tension updated → possible fracture
6. After each round: `persistRound()` updates JobRow live progress
7. Simulation ends → `FullReport` generated → saved to simulations table
8. Status=done

## Dashboard Pages

- `NewSimulationPage` — form: question, department, rounds, company, mode
- `SimulationsPage` — list with status badges, checkboxes for comparison, Ver Resultado / Ver Convergência / Dar Feedback buttons
- `ResultPage` — FullReport visualization: narrative, fractures, ensemble, playbook, export buttons
- `FeedbackPage` — outcome selector + delta_score slider (−1..+1)
- `ComparisonPage` — side-by-side tension bars, common/divergent fractures, ConfidenceDelta
- `ConvergencePage` — SVG tension chart, fracture markers, stat cards

## Build

```bash
# Tests only (no dashboard embed needed)
go test ./...

# Dashboard build required for full binary
cd dashboard && npm install && npm run build && cd ..
go build -o fracture .
```

## Database

SQLite at platform data dir:
- Linux: `~/.local/share/FRACTURE/data.db`
- macOS: `~/Library/Application Support/FRACTURE/data.db`
- Windows: `%APPDATA%\FRACTURE\data.db`

Schema applied idempotently on startup. Versioned migrations in `db/migrations.go`.

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `OPENAI_API_KEY` | LLM calls (DeepSearch synthesis + simulation narrative) |
| `TAVILY_API_KEY` | Web search (optional, better results) |
| `SERPAPI_KEY` | Web search fallback (optional) |

Without LLM keys, FRACTURE runs in **heuristic mode** (deterministic only).

## Pending Roadmap

- [ ] Streaming SSE progress endpoint (`/events/stream`)
- [ ] ReplayPage: round-by-round replay of simulation events
- [ ] PDF export
- [ ] Multi-company comparison
- [ ] Custom archetype builder UI
- [ ] Simulation scheduling (run at specific time)
