# FRACTURE MCP Server

Use FRACTURE disruption simulation directly inside **Cursor**, **VS Code**, and **Windsurf** via the [Model Context Protocol](https://modelcontextprotocol.io).

Ask your AI assistant things like:
- *"Run a FRACTURE simulation: what happens if a competitor launches a free tier next month?"*
- *"Quick pulse check: we're about to raise prices by 20%, what's the market tension?"*
- *"Show me which marketing rules are most fragile right now"*

---

## Prerequisites

1. **FRACTURE must be running** on your machine (opens at `http://localhost:3000`)
2. **Node.js 20+** installed

---

## Installation

```bash
# From the fracture/mcp directory
npm install
npm run build
```

---

## Setup by Editor

### Cursor

Add to `~/.cursor/mcp.json` (or `.cursor/mcp.json` in your project):

```json
{
  "mcpServers": {
    "fracture": {
      "command": "node",
      "args": ["/path/to/fracture/mcp/dist/index.js"],
      "env": {
        "FRACTURE_URL": "http://localhost:3000"
      }
    }
  }
}
```

Then restart Cursor. The FRACTURE tools appear automatically in the AI chat.

---

### VS Code (with GitHub Copilot or Continue)

**Option A — Continue extension:**

Add to `~/.continue/config.json`:
```json
{
  "mcpServers": [
    {
      "name": "fracture",
      "command": "node",
      "args": ["/path/to/fracture/mcp/dist/index.js"],
      "env": { "FRACTURE_URL": "http://localhost:3000" }
    }
  ]
}
```

**Option B — GitHub Copilot (VS Code 1.99+):**

Add to `.vscode/mcp.json` in your workspace:
```json
{
  "servers": {
    "fracture": {
      "type": "stdio",
      "command": "node",
      "args": ["/path/to/fracture/mcp/dist/index.js"],
      "env": { "FRACTURE_URL": "http://localhost:3000" }
    }
  }
}
```

---

### Windsurf (Codeium)

Add to `~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "fracture": {
      "command": "node",
      "args": ["/path/to/fracture/mcp/dist/index.js"],
      "env": {
        "FRACTURE_URL": "http://localhost:3000"
      }
    }
  }
}
```

---

## Available Tools

| Tool | What it does |
|---|---|
| `fracture_simulate` | Full disruption simulation (12 agents, multiple rounds) |
| `fracture_quick_pulse` | 5-second tension check with score 0-100 |
| `fracture_list_rules` | Show market rules with stability weights |
| `fracture_archetypes` | List all 12 agent archetypes |
| `fracture_history` | Past simulation results |
| `fracture_feedback` | Submit real outcome to calibrate accuracy |

---

## Example Prompts

```
Run a fracture simulation for our marketing domain:
"What happens if TikTok launches a B2B product targeting our customers?"
Use 20 rounds and include context that we're a mid-size SaaS company.
```

```
Quick pulse check: our HR team is about to implement mandatory return-to-office.
What's the tension level?
```

```
List all the finance domain rules and tell me which ones have the lowest stability score.
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `FRACTURE_URL` | `http://localhost:3000` | URL where FRACTURE is running |
