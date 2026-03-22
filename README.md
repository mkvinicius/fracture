# FRACTURE

> Simulate how market rules break — and be the one to break them first.

FRACTURE is a desktop application that runs a market disruption simulation engine locally on your machine. It uses 12 AI agents (8 Conformists + 4 Disruptors) to simulate how fundamental market rules could be rewritten — and who would rewrite them.

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

### macOS
```bash
# Download the latest release
curl -L https://github.com/your-org/fracture/releases/latest/download/fracture_macos_amd64.tar.gz | tar xz
./fracture
```

### Windows
Download `fracture_windows_amd64.zip` from the [releases page](https://github.com/your-org/fracture/releases), extract, and run `fracture.exe`.

### Linux
```bash
curl -L https://github.com/your-org/fracture/releases/latest/download/fracture_linux_amd64.tar.gz | tar xz
./fracture
```

FRACTURE opens your browser automatically at `http://localhost:3000`.

---

## Build From Source

**Requirements:** Go 1.22+, Node.js 20+, pnpm

```bash
git clone https://github.com/your-org/fracture
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
- **OpenAI** (GPT-4o) — Conformist agents + synthesis reports
- **Anthropic** (Claude Sonnet) — Disruptor agents (best creativity)
- **Google** (Gemini) — Optional third model for diversity

You can also use **Ollama** for fully offline operation (no API costs).

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
    report.go          ← Report generation (3 output types)
  archetypes/
    conformists.go     ← 8 Conformist archetypes
    disruptors.go      ← 4 Disruptor archetypes
  memory/
    store.go           ← SQLite-backed agent memory
    calibration.go     ← Feedback loop + archetype calibration
  security/
    sanitizer.go       ← Prompt injection protection
    hmac.go            ← HMAC signing + immutable audit log
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

---

## License

MIT
