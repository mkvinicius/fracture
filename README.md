# FRACTURE

> Simulate how market rules break — and be the one to break them first.

[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://github.com/mkvinicius/fracture/blob/master/LICENSE)
[![Release](https://img.shields.io/github/v/release/mkvinicius/fracture)](https://github.com/mkvinicius/fracture/releases/latest)

FRACTURE is a **local-first desktop application** that runs a market disruption simulation engine on your machine. It uses 12 AI agents (8 Conformists + 4 Disruptors) to simulate how fundamental market rules could be rewritten — and who would rewrite them first.

---

## How It Works

1. You ask a strategic question: *"If our main competitor went free tomorrow, what would happen in 12 months?"*
2. FRACTURE builds a **World** — a graph of the rules that govern your market
3. **12 agents** with distinct personalities, goals, and power levels interact over multiple rounds
4. When tension accumulates, a **FRACTURE POINT** triggers — an agent proposes rewriting a rule
5. Other agents vote. If the proposal passes, the world changes and the simulation continues with new rules
6. You receive three outputs:
   - **Probable Future** — what happens if no rules break
   - **Tension Map** — which rules are under the most pressure
   - **Rupture Scenarios** — the top 3 ways the market could be disrupted, and how *you* could do it first

---

## Installation

### Linux (amd64)

```bash
curl -L https://github.com/mkvinicius/fracture/releases/latest/download/fracture-linux-amd64.tar.gz | tar xz
chmod +x fracture
./fracture
```

### Linux (arm64 — Apple Silicon via Rosetta or native ARM servers)

```bash
curl -L https://github.com/mkvinicius/fracture/releases/latest/download/fracture-linux-arm64.tar.gz | tar xz
chmod +x fracture
./fracture
```

### Windows (amd64)

Download `fracture-windows-amd64.zip` from the [releases page](https://github.com/mkvinicius/fracture/releases/latest), extract, and run `fracture.exe`.

### macOS

Build from source (see below) — macOS binaries require code signing for distribution.

FRACTURE opens your browser automatically at `http://localhost:3000`.

---

## Build From Source

**Requirements:** Go 1.22+, Node.js 20+, pnpm

```bash
git clone https://github.com/mkvinicius/fracture
cd fracture
make build   # builds dashboard + Go binary
./fracture   # starts the server and opens browser
```

For development with hot reload:

```bash
make dev
```

---

## AI Provider Keys

FRACTURE uses your own AI provider keys. They are stored locally in SQLite on your machine and **never sent to any external server**.

Recommended configuration:

| Provider | Model | Role |
|----------|-------|------|
| **OpenAI** | GPT-4o | Conformist agents + synthesis reports |
| **Anthropic** | Claude Sonnet | Disruptor agents (best creativity) |
| **Google** | Gemini 1.5 | Optional third model for diversity |

You can also use **Ollama** for fully offline operation (no API costs).

---

## Privacy & Telemetry

FRACTURE collects **anonymous usage data** to understand how the tool is being used and improve future versions.

**What is collected:**
- Anonymous install ID (UUID — randomly generated, never linked to you)
- OS and architecture
- Country (derived from IP, last octet masked)
- FRACTURE version

**What is never collected:** simulation content, API keys, company data, or any personally identifiable information.

You can **opt out at any time** during the onboarding wizard or in **Settings → Privacy & Telemetry**.

---

## Architecture

```
fracture/
  main.go              ← Entry point, HTTP server, browser open
  api/handler.go       ← REST API routes
  engine/
    world.go           ← Rule graph with stability weights
    agent.go           ← Agent interface and base types
    simulation.go      ← Main simulation loop
    voting.go          ← Weighted consensus voting
    report.go          ← Report generation (3 output types + watermark)
  archetypes/
    conformists.go     ← 8 Conformist archetypes
    disruptors.go      ← 4 Disruptor archetypes
  memory/
    store.go           ← SQLite-backed agent memory
    calibration.go     ← Feedback loop + archetype calibration
  security/
    sanitizer.go       ← Prompt injection protection
    hmac.go            ← HMAC signing + immutable audit log
  telemetry/
    telemetry.go       ← Anonymous usage tracking (opt-out)
  llm/client.go        ← LLM-agnostic client with smart routing
  db/
    db.go              ← SQLite helpers
    schema.sql         ← Database schema
  dashboard/           ← React + Tailwind frontend (embedded in binary)
```

---

## Security

- **Prompt injection protection** — all external data is sanitized before reaching agents
- **HMAC-signed prompts** — every agent prompt is signed to detect tampering
- **Immutable audit log** — every simulation event is chained with HMAC signatures
- **Agent sandboxing** — agents have no access to filesystem or network
- **Local-first** — all data stays on your machine

To report a security vulnerability, see [SECURITY.md](SECURITY.md).

---

## License

FRACTURE is licensed under the [GNU Affero General Public License v3.0 (AGPL-3.0)](LICENSE).

This means:
- You can use, study, and modify FRACTURE freely
- If you distribute a modified version (including running it as a network service), you must release the source code under the same license
- Commercial use requires compliance with AGPL-3.0 terms

© 2025 FRACTURE contributors
