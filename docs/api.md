# FRACTURE API Reference — v2.5.0

Base URL: `http://localhost:3000/api/v1`

Legacy paths (`/api/*`) redirect permanently to `/api/v1/*` via HTTP 308.

All request bodies are JSON (`Content-Type: application/json`).
All responses are JSON unless noted otherwise.
Errors always return `{ "error": "<message>" }`.

---

## Table of contents

1. [Simulations](#simulations)
2. [Export](#export)
3. [Feedback](#feedback)
4. [Quick Pulse](#quick-pulse)
5. [Templates](#templates)
6. [Archetypes](#archetypes)
7. [Rules](#rules)
8. [Config & Onboarding](#config--onboarding)
9. [Company Profile](#company-profile)
10. [Audit Log](#audit-log)
11. [Telemetry](#telemetry)
12. [System](#system)

---

## Simulations

### POST /simulations

Start a new simulation. Returns immediately with a job ID; the simulation runs asynchronously.

**Request**

| Field        | Type     | Required | Description |
|--------------|----------|----------|-------------|
| `question`   | string   | yes      | Strategic question to simulate |
| `department` | string   | no       | Domain: `market` `technology` `regulation` `behavior` `culture` `geopolitics` `finance` (default `market`) |
| `rounds`     | int      | no       | Ignored when `mode` is set; otherwise number of simulation rounds |
| `mode`       | string   | no       | `standard` (default) or `premium` (ensemble run, more rounds) |
| `context`    | string   | no       | Extra context injected before DeepSearch results |
| `urls`       | []string | no       | Up to 10 URLs — content is extracted and prepended to context |

**Response** `202 Accepted`

```json
{ "id": "uuid", "status": "queued" }
```

---

### GET /simulations

List all simulations (in-memory jobs + completed simulations from DB).

**Response** `200 OK`

```json
[
  {
    "id": "uuid",
    "status": "queued | researching | running | done | error",
    "question": "...",
    "department": "market",
    "rounds": 20,
    "created_at": 1712000000,
    "duration_ms": 45000
  }
]
```

---

### GET /simulations/{id}

Get live status and progress for a running or completed simulation.

**Response** `200 OK`

```json
{
  "id": "uuid",
  "status": "running",
  "question": "...",
  "department": "market",
  "rounds": 20,
  "mode": "standard",
  "created_at": 1712000000,
  "current_round": 12,
  "current_tension": 0.74,
  "fracture_count": 2,
  "last_agent_name": "disruptor-visionary",
  "last_agent_action": "The market is ripe for…",
  "total_tokens": 8400,
  "research_sources": 7,
  "research_tokens": 1200
}
```

---

### GET /simulations/{id}/stream

Server-Sent Events (SSE) stream. Pushes `update` events every 500 ms while the job is running, then a final event when it reaches `done` or `error`.

```
event: update
data: { ...same shape as GET /simulations/{id}... }
```

Closes automatically on `done`, `error`, or after 12 minutes.

---

### GET /simulations/compare?ids=a,b,c

Compare 2–5 completed simulations side by side.

**Query params**

| Param | Description |
|-------|-------------|
| `ids` | Comma-separated list of 2–5 simulation IDs |

**Response** `200 OK`

```json
{
  "common_fractures": ["..."],
  "divergent_fractures": {
    "uuid-a": ["..."],
    "uuid-b": ["..."]
  },
  "tension_deltas": { "rule-id": 0.12 },
  "confidence_delta": 0.08
}
```

---

### GET /simulations/{id}/results

Raw simulation result. Returns a status stub if the simulation is still running.

---

### DELETE /simulations/{id}

Delete a simulation (memory + DB). Returns `{ "ok": true }`.

---

## Export

### GET /simulations/{id}/report

Full structured report as JSON.

**Response** `200 OK` — `FullReport` object:

```json
{
  "simulation_id": "uuid",
  "question": "...",
  "probable_future": {
    "narrative": "...",
    "timeline": [
      { "horizon": "6 months", "description": "...", "confidence": 0.82 }
    ],
    "confidence": 0.78,
    "key_assumptions": ["..."]
  },
  "tension_map": [
    { "rule_id": "rule-1", "description": "...", "domain": "market", "tension": 0.71, "color": "orange" }
  ],
  "rupture_scenarios": [
    {
      "rule_id": "rule-1",
      "rule_description": "...",
      "probability": 0.65,
      "impact_on_company": "...",
      "who_breaks": "...",
      "how_it_happens": "..."
    }
  ],
  "fracture_events": [
    {
      "round": 8,
      "proposed_by": "disruptor-visionary",
      "accepted": true,
      "confidence": 0.84,
      "proposal": {
        "original_rule_id": "rule-1",
        "new_description": "..."
      },
      "vote_breakdown": [
        { "agent_id": "conformist-analyst", "vote": "for", "weight": 0.8, "rationale": "..." }
      ]
    }
  ],
  "coalitions": [...],
  "rupture_timeline": [...],
  "action_playbook": { ... },
  "ensemble_result": null,
  "total_tokens": 14200,
  "duration_ms": 47000
}
```

---

### GET /simulations/{id}/export/markdown

Download the report as a Markdown file.

**Response** `200 OK`
`Content-Type: text/markdown`
`Content-Disposition: attachment; filename="fracture-{id}.md"`

---

### GET /simulations/{id}/export/json

Download the report as a formatted JSON file.

**Response** `200 OK`
`Content-Type: application/json`
`Content-Disposition: attachment; filename="fracture-{id}.json"`

---

### GET /simulations/{id}/events

Tension time-series for the convergence chart.

**Response** `200 OK`

```json
[
  { "round": 1, "tension": 0.31, "fracture": false },
  { "round": 8, "tension": 0.82, "fracture": true }
]
```

---

## Feedback

### POST /simulations/{id}/feedback

Submit real-world outcome feedback. Triggers archetype recalibration and re-indexes the simulation in the RAG store.

**Request**

| Field               | Type    | Required | Description |
|---------------------|---------|----------|-------------|
| `outcome`           | string  | yes      | `accurate` \| `inaccurate` \| `partial` |
| `predicted_fracture`| string  | no       | What the simulation predicted |
| `actual_outcome`    | string  | no       | What actually happened |
| `delta_score`       | float   | no       | Accuracy signal: `-1.0` (wrong) → `0.0` (neutral) → `1.0` (exact) |
| `notes`             | string  | no       | Free text |

**Response** `200 OK` — `{ "ok": true }`

---

## Quick Pulse

### POST /pulse

Fast tension check without running a full simulation. Makes a single LLM call and returns a tension score in seconds.

**Request**

| Field       | Type   | Required | Description |
|-------------|--------|----------|-------------|
| `situation` | string | yes      | Brief description of the business situation |
| `domain`    | string | no       | Domain context hint (e.g. `market`) |

**Response** `200 OK`

```json
{
  "score": 74,
  "level": "high",
  "summary": "Competitive pressure from AI-native entrants is compressing margins.",
  "top_risks": [
    "Price war triggered by free-tier launch",
    "Talent drain to better-funded competitors",
    "Regulatory scrutiny of AI-generated outputs"
  ]
}
```

`level` values: `low` (0–25) · `medium` (26–50) · `high` (51–75) · `critical` (76–100)

---

## Templates

### GET /templates

List built-in scenario templates.

**Response** `200 OK`

```json
[
  {
    "id": "competitor-free-tier",
    "name": "Competitor launches free tier",
    "domain": "market",
    "rounds": 20,
    "question": "What happens if a major competitor launches a free tier targeting our core customers?"
  }
]
```

Available template IDs: `competitor-free-tier` · `ai-disruption` · `regulation-change` · `price-increase` · `talent-war` · `new-entrant`

---

### GET /templates/{id}

Get a single template by ID.

---

## Archetypes

Archetypes are the 56 agent personalities. 12 are built-in (read-only); additional custom archetypes can be created per company.

### GET /archetypes

List built-in + custom archetypes.

**Query params**

| Param        | Description |
|--------------|-------------|
| `company_id` | Include custom archetypes for this company |

**Response** `200 OK`

```json
[
  {
    "id": "visionary",
    "name": "The Visionary",
    "agent_type": "disruptor",
    "description": "Startup founder: contrarian, first-principles, high-risk tolerance",
    "memory_weight": 0.9,
    "is_active": true
  }
]
```

Built-in conformists: `pragmatist` · `loyalist` · `analyst` · `opportunist` · `traditionalist` · `regulator` · `consumer` · `investor`
Built-in disruptors: `visionary` · `rebel` · `tech-accelerator` · `arbitrageur`

---

### POST /archetypes

Create a custom archetype.

**Request**

| Field          | Type   | Required | Description |
|----------------|--------|----------|-------------|
| `name`         | string | yes      | Display name |
| `agent_type`   | string | yes      | `conformist` or `disruptor` |
| `description`  | string | no       | Personality description |
| `memory_weight`| float  | no       | Voting influence `0.0–2.0` (default `1.0`) |
| `company_id`   | string | no       | Scope to a specific company |

**Response** `201 Created` — created `ArchetypeRow`

---

### GET /archetypes/{id}

Get a single archetype (DB first, then built-ins).

---

### PUT /archetypes/{id}

Update a custom archetype's name, description, memory\_weight, or is\_active.

---

### DELETE /archetypes/{id}

Delete a custom archetype. Returns `204 No Content`. Built-ins cannot be deleted (`403`).

---

## Rules

Rules are the market/domain norms that agents vote to mutate. Built-in rules come from `engine/world_domains.go`; custom rules are stored in the DB.

### GET /rules

List all built-in rules (market domain) merged with custom rules.

**Query params**

| Param        | Description |
|--------------|-------------|
| `company_id` | Include active custom rules for this company |

---

### GET /rules/domain/{domain}

Rules for a specific domain. `domain` = `market` · `technology` · `regulation` · `behavior` · `culture` · `geopolitics` · `finance`

---

### POST /rules

Create a custom rule.

**Request**

| Field         | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `description` | string | yes      | Rule statement |
| `domain`      | string | no       | Default `market` |
| `stability`   | float  | no       | `0.0–1.0`, default `0.5` |
| `company_id`  | string | no       | Scope to a company |

**Response** `201 Created`

---

### GET /rules/{id}

Get a custom rule by ID.

---

### PUT /rules/{id}

Update description, domain, stability, or is\_active.

---

### DELETE /rules/{id}

Delete a custom rule. Returns `204 No Content`.

---

## Config & Onboarding

### GET /config

Returns configured key presence and model preferences.

```json
{
  "openai_key_set": "true",
  "anthropic_key_set": "",
  "google_key_set": "",
  "ollama_enabled": "",
  "default_model_conformist": "",
  "default_model_disruptor": "",
  "default_model_synthesis": "",
  "default_rounds": ""
}
```

---

### POST /config

Set one or more config values. Body is a flat `{ "key": "value" }` map.

---

### POST /keys/validate

Save an API key for a provider and mark it as configured.

**Request**

| Field      | Type   | Description |
|------------|--------|-------------|
| `provider` | string | `openai` · `anthropic` · `google` · `ollama` |
| `key`      | string | API key (sanitized before storage) |

**Response** `200 OK` — `{ "valid": true }`

---

### GET /onboarding/status

`{ "complete": false }`

---

### POST /onboarding/complete

Mark onboarding as done. `{ "ok": true }`

---

### POST /extract-context

Fetch and extract text content from URLs (used before creating a simulation).

**Request**

| Field  | Type     | Required | Description |
|--------|----------|----------|-------------|
| `urls` | []string | yes      | 1–10 URLs |

**Response** `200 OK`

```json
{
  "sources": [
    {
      "url": "https://...",
      "source_type": "webpage",
      "title": "...",
      "description": "...",
      "content": "...",
      "error": ""
    }
  ],
  "summary": "Combined extracted text…",
  "has_errors": false
}
```

---

## Company Profile

### GET /company

Returns the saved company JSON, or `null` if not set.

---

### POST /company

Upsert the company profile. Body is a free-form JSON object (name, sector, size, etc.).

---

## Audit Log

### GET /audit

Returns the 50 most recent audit log entries.

```json
[
  {
    "action": "simulation.created",
    "actor": "uuid",
    "details": { "question": "...", "rounds": 20 },
    "created_at": 1712000000
  }
]
```

---

## Telemetry

### GET /telemetry

`{ "enabled": false }`

### POST /telemetry

`{ "enabled": true }` — opt in or out of anonymous startup pings.

---

## System

### GET /health

`{ "status": "ok", "version": "2.5.0" }`

### GET /update-check

```json
{
  "has_update": false,
  "current_version": "2.5.0",
  "latest_version": "2.5.0",
  "release_url": "https://...",
  "release_name": "v2.5.0",
  "release_notes": "..."
}
```
