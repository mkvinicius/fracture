# FRACTURE

> **Simulate how market rules break — and be the one to break them first.**

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![release](https://img.shields.io/badge/release-v2.6.0-red.svg)](https://github.com/mkvinicius/fracture/releases/latest)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8.svg)](https://golang.org)

---

## What is FRACTURE?

FRACTURE is a local strategic market simulation tool. You ask a strategic question, the system automatically researches the market with **DeepSearch**, and runs **56 AI agents** — each with a distinct personality, objective, and power level — to simulate how your market's rules might be rewritten over **40 rounds**.

When tension accumulates and pressure explodes, a **FRACTURE POINT** occurs: a fundamental rule changes. The final 6-part report tells you what will happen, when, and what you should do before it does.

**Your data stays on your machine. No subscription. No cloud. No external server.**

---

## ⚡ Quick Install / Instalação Rápida

**English:**
Download and double-click the installer for your system — no terminal required.

| System | Download | Instructions |
|---|---|---|
| **Windows** | [`install-windows.bat`](install-windows.bat) | Right-click → Run as administrator |
| **macOS** | [`install-mac.sh`](install-mac.sh) | Double-click → Allow execution |
| **Linux** | See below | Terminal required |

The installer will automatically:
- Check and install Go if missing
- Clone and build FRACTURE
- Create a desktop shortcut
- Open the dashboard in your browser

---

**Português:**
Baixe e clique duas vezes no instalador para o seu sistema — sem terminal necessário.

| Sistema | Download | Instruções |
|---|---|---|
| **Windows** | [`install-windows.bat`](install-windows.bat) | Clique direito → Executar como administrador |
| **macOS** | [`install-mac.sh`](install-mac.sh) | Clique duplo → Permitir execução |
| **Linux** | Veja abaixo | Requer terminal |

O instalador faz automaticamente:
- Verifica e instala Go se necessário
- Clona e compila o FRACTURE
- Cria atalho na área de trabalho
- Abre o dashboard no navegador

---

## How it works
```
1. You ask a strategic question
         ↓
2. DeepSearch automatically researches the market
   (recent news, competitors, trends, regulatory changes)
         ↓
3. FRACTURE builds a World with 55+ rules across 7 domains
         ↓
4. 56 AI agents interact over 40 rounds
   — form alliances, betray each other, tension the rules
         ↓
5. When tension explodes → FRACTURE POINT (a fundamental rule changes)
         ↓
6. Full 6-part report generated in parallel
```

---

## Key Features

- **56-Agent Simulation Engine** — 37 Conformists + 19 Disruptors across 7 domains with configurable power weights
- **DeepSearch Integration** — Real-world market context injected before simulation runs
- **Semantic Memory** — Embeddings-based similarity search across past simulations. Simulations get smarter over time.
- **JUDGE Learning Cycle** — After each simulation, agents are scored (hit/partial/miss) and weights adjusted directionally
- **EWC++ Consolidation** — Elastic Weight Consolidation prevents forgetting patterns from past simulations
- **Causal Graph** — DeepSearch findings are extracted as causal triples and persist across simulations
- **Parallel Report Generation** — All 6 report sections generated concurrently (~10s vs ~40s sequential)
- **Domain Calibration** — Rate simulation accuracy → calibrate archetype weights → future simulations improve
- **Comparison** — Compare 2–5 simulations side-by-side
- **Convergence Chart** — SVG tension chart showing pressure build-up and fracture points
- **Persistent Job State** — Simulations survive server restarts
- **Audit Log** — HMAC-signed immutable audit trail
- **React Dashboard** — Embedded SPA served from the Go binary. No separate web server needed.
- **Auto-updater** — Checks GitHub Releases at startup

---

## Install

### macOS (2 clicks)
```bash
bash install-mac.sh
```

### Windows (2 clicks)
Download and run `install-windows.bat` as administrator.

Or use the GUI installer: download `FRACTURE-Setup.exe` from [Releases](https://github.com/mkvinicius/fracture/releases/latest).

### Linux
```bash
curl -L https://github.com/mkvinicius/fracture/releases/latest/download/fracture-linux-amd64 -o fracture
chmod +x fracture && ./fracture
```

### Build from source
Requirements: Go 1.24+ only (`dashboard/dist` is pre-built and committed)
```bash
git clone https://github.com/mkvinicius/fracture
cd fracture
go build -o fracture .
./fracture
```

Opens at `http://localhost:4000`

---

## Configuration

Data stored at:
- Linux: `~/.local/share/FRACTURE/data.db`
- macOS: `~/Library/Application Support/FRACTURE/data.db`
- Windows: `%APPDATA%\FRACTURE\data.db`

| Variable | Description |
|---|---|
| `OPENAI_API_KEY` | LLM calls + embeddings (recommended) |
| `ANTHROPIC_API_KEY` | Claude models (optional) |
| `GOOGLE_API_KEY` | Gemini models (optional) |
| `TAVILY_API_KEY` | Web search for DeepSearch (optional) |
| `SERPAPI_KEY` | Web search fallback (optional) |

Without API keys: runs in heuristic mode (deterministic only).

---

## The 56 Agents

37 Conformists defend the status quo. 19 Disruptors challenge and rewrite the rules.
Agents learn from every simulation — weights adjust via JUDGE verdicts and EWC++ consolidation.

---

## The 6-Part Report

| Part | Delivers |
|---|---|
| **1. Probable Future** | What happens if nothing changes — with probability and timeline |
| **2. Tension Map** | Which rules are closest to breaking and why |
| **3. Rupture Scenarios** | The 3 most likely FRACTURE POINT paths |
| **4. Coalition Map** | Who allied with whom during the simulation |
| **5. Timeline** | When each rupture should happen (90 days / 1 year / 3 years) |
| **6. Action Playbook** | What you should do now to position before the break |

---

## License

FRACTURE is distributed under [AGPL-3.0](LICENSE).
