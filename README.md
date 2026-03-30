# FRACTURE

> **Simulate how market rules break — and be the one to break them first.**

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![release](https://img.shields.io/badge/release-v2.6.0-red.svg)](https://github.com/mkvinicius/fracture/releases/latest)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey.svg)](https://github.com/mkvinicius/fracture/releases/latest)
[![Tests](https://img.shields.io/badge/tests-115%20passing-brightgreen.svg)](https://github.com/mkvinicius/fracture)

---

## What is FRACTURE?

FRACTURE is a **local strategic market simulation tool** that runs on your computer. You ask a strategic question, the system automatically researches the market with **DeepSearch**, and then runs **56 AI agents** — each with a distinct personality, objective, and power level — to simulate how your market's rules might be rewritten over **20–50 rounds**.

When tension accumulates and pressure explodes, a **FRACTURE POINT** occurs: a fundamental rule changes. The final report tells you what will happen, when, and what you should do before it does.

Everything runs **100% locally**. No cloud, no subscription, no data leaves your machine.

---

## Key Features

- **Multi-Agent Simulation Engine** — 56 agents across 7 domains (Market, Technology, Regulation, Behavior, Culture, Geopolitics, Finance). Conformist and Disruptor archetypes with configurable power weights.
- **DeepSearch Integration** — Real-world context from web research injected into simulation before it runs. Per-domain research with reflection loops and resumable state.
- **Domain Calibration** — Feedback loop: rate simulation accuracy → calibrate archetype weights per domain via EMA → future simulations improve automatically.
- **RAG Memory** — Local TF-IDF semantic search over past simulations per company. No external embedding APIs required.
- **Export** — Download full reports as structured Markdown or JSON.
- **Comparison** — Compare 2–5 simulations side-by-side: common fractures, divergent outcomes, tension delta, confidence delta.
- **Convergence Chart** — SVG tension chart per simulation showing how pressure built and where fracture points occurred.
- **Persistent Job State** — Simulations survive server restarts. Progress tracked per round with live UI updates.
- **Audit Log** — HMAC-signed audit trail for all simulation events.
- **React Dashboard** — Embedded SPA served from the Go binary. No separate web server needed.
- **Auto-updater** — Checks GitHub Releases at startup and notifies of new versions.

---

## Architecture

```
fracture/
├── main.go                  # HTTP server, router, graceful shutdown
├── api/
│   └── handler.go           # REST API handlers (/api/v1/*)
├── engine/
│   ├── engine.go            # Simulation loop (rounds, voting, tension)
│   ├── agents.go            # Agent types, personalities, permissions
│   ├── world.go             # World state: Rules, TensionMap, Evidence
│   ├── world_domains.go     # Domain-specific rule sets (7 domains)
│   ├── report.go            # FullReport generation
│   ├── export.go            # ReportToMarkdown()
│   └── compare.go           # CompareReports()
├── deepsearch/
│   ├── agent.go             # DeepSearch agent (web search + LLM synthesis)
│   └── domain_research.go   # DomainResearcher with semaphore + cache
├── db/
│   ├── db.go                # SQLite helpers (simulations, jobs, contexts)
│   ├── schema.sql           # Database schema
│   ├── migrations.go        # Versioned schema migrations
│   └── rounds.go            # Round-level persistence + tension aggregation
├── memory/
│   └── memory.go            # TF-IDF RAG: index + similarity search
├── security/
│   ├── signer.go            # HMAC-SHA256 audit signing
│   ├── sanitizer.go         # Input sanitization
│   └── audit.go             # Audit logger
├── telemetry/
│   └── telemetry.go         # Anonymous opt-in ping
├── updater/
│   └── updater.go           # GitHub Releases version check
└── dashboard/               # React + TypeScript SPA (Vite build)
    └── src/
        ├── App.tsx
        └── pages/
            ├── NewSimulationPage.tsx
            ├── SimulationsPage.tsx
            ├── ResultPage.tsx
            ├── FeedbackPage.tsx
            ├── ComparisonPage.tsx
            └── ConvergencePage.tsx
```

---

## Simulation Modes

| Mode     | Rounds | Runs | Agents | Use Case                          |
|----------|--------|------|--------|-----------------------------------|
| Standard | 30     | 1    | 56     | Quick analysis, daily use         |
| Premium  | 50     | 2    | 56     | Deep research, critical decisions |

---

## REST API

All endpoints are under `/api/v1/`. Legacy `/api/*` paths redirect to `/api/v1/*` with HTTP 308 (method-preserving).

### Simulations

| Method   | Path                                    | Description                        |
|----------|-----------------------------------------|------------------------------------|
| `POST`   | `/api/v1/simulate`                      | Start a new simulation             |
| `GET`    | `/api/v1/simulations`                   | List all simulations (with status) |
| `GET`    | `/api/v1/simulations/{id}`              | Get simulation status/progress     |
| `DELETE` | `/api/v1/simulations/{id}`              | Delete a simulation                |
| `GET`    | `/api/v1/simulations/{id}/report`       | Full FullReport JSON               |
| `GET`    | `/api/v1/simulations/{id}/export/markdown` | Download report as Markdown     |
| `GET`    | `/api/v1/simulations/{id}/export/json` | Download report as JSON            |
| `GET`    | `/api/v1/simulations/{id}/events`       | Tension per round (convergence)    |
| `GET`    | `/api/v1/simulations/compare?ids=a,b`  | Compare 2–5 simulations            |

### Feedback & Calibration

| Method | Path                    | Description                            |
|--------|-------------------------|----------------------------------------|
| `POST` | `/api/v1/feedback`      | Submit accuracy feedback               |
| `GET`  | `/api/v1/calibration`   | Get archetype calibration weights      |

### Knowledge Base

| Method   | Path                              | Description                         |
|----------|-----------------------------------|-------------------------------------|
| `GET`    | `/api/v1/archetypes`              | List agent archetypes               |
| `GET`    | `/api/v1/rules`                   | List world rules                    |
| `GET`    | `/api/v1/rules/domain/{domain}`   | List rules for a specific domain    |
| `POST`   | `/api/v1/documents`               | Upload RAG document for a company   |
| `GET`    | `/api/v1/documents/{company_id}`  | List RAG documents for a company    |

### System

| Method | Path             | Description              |
|--------|------------------|--------------------------|
| `GET`  | `/api/v1/health` | Health check             |
| `GET`  | `/api/v1/config` | Get configuration        |
| `PUT`  | `/api/v1/config` | Update configuration     |
| `GET`  | `/api/v1/audit`  | Recent audit log entries |

---

## Install

### Pre-built binary (recommended)

Download the latest release from [GitHub Releases](https://github.com/mkvinicius/fracture/releases/latest):

```bash
# Linux (amd64)
curl -L https://github.com/mkvinicius/fracture/releases/latest/download/fracture-linux-amd64 -o fracture
chmod +x fracture
./fracture
```

The dashboard opens automatically in your browser at `http://localhost:4000`.

### Build from source

Requirements: Go 1.21+ (Node.js not required — `dashboard/dist` is committed)

```bash
git clone https://github.com/mkvinicius/fracture
cd fracture
go build -o fracture .
./fracture
```

> macOS installer: `bash install-mac.sh` — Windows installer: run `install-windows.bat` as administrator

---

## Configuration

FRACTURE stores data in the platform data directory:
- **Linux**: `~/.local/share/FRACTURE/data.db`
- **macOS**: `~/Library/Application Support/FRACTURE/data.db`
- **Windows**: `%APPDATA%\FRACTURE\data.db`

### Environment Variables

| Variable          | Description                                                       |
|-------------------|-------------------------------------------------------------------|
| `OPENAI_API_KEY`  | OpenAI API key (LLM calls in DeepSearch + simulation narrative)   |
| `TAVILY_API_KEY`  | Tavily search key (optional, improves DeepSearch quality)         |
| `SERPAPI_KEY`     | SerpAPI key (optional, fallback web search)                       |

If no LLM API key is set, FRACTURE runs in **heuristic mode** — simulations complete using deterministic rules without LLM narrative generation.

---

## Agents

56 agents across 7 domains:

**Conformists (37)** — defend existing rules, resist change, vote against fractures.
**Disruptors (19)** — challenge the status quo, propose new rules, push for fractures.

Each agent has:
- `PowerWeight` — influence on voting (calibrated by feedback loop)
- `Personality` — bias profile (risk tolerance, contrarianism, etc.)
- `Domain` — specialization area
- `AgentMemory` — stores past actions for behavioral consistency

Agent calibration is updated automatically when you submit feedback: accurate predictions increase an archetype's weight, inaccurate ones decrease it.

---

## Development

```bash
# Run all tests (115 tests)
go test ./...

# Run specific package
go test ./engine/... -v

# Build packages (no dashboard embed needed)
go build ./api/... ./engine/...
```

### Adding a new domain

1. Add a constant to `engine/world.go`: `DomainMyDomain RuleDomain = "mydomain"`
2. Add a rule set function in `engine/world_domains.go`: `func myDomainRules() []*Rule { ... }`
3. Add a case in `DefaultWorldForDomain` switch statement
4. Add domain-specific agents in `engine/agents.go`

---

## License

AGPL-3.0 — see [LICENSE](LICENSE).
