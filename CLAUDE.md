# FRACTURE — Contexto para Claude Code

## O que é
Motor de simulação estratégica de mercado.
Local-first, sem Docker, sem banco externo.
Binário único Go + dashboard React embedado.
Repositório: https://github.com/mkvinicius/fracture

## Stack
- Go 1.24+, CGO habilitado (SQLite)
- chi/v5 para routing
- SQLite via mattn/go-sqlite3
- React + Tailwind no dashboard
- GoReleaser para builds

## Versão atual: v2.6.0
- 125+ testes passando
- 166 mentes brilhantes (61 conformistas + 28 disruptores + 77 skills)
- 13 skills verticais
- Modos Standard e Premium
- Ciclo completo: simulação → feedback → calibração → precisão

## Arquitetura

fracture/
  main.go
  api/
    handler.go              ← Handler struct, Routes, helpers
    handler_simulations.go  ← createSimulation, listSimulations,
                               getSimulation, deleteSimulation,
                               getResults, getReport, streamSimulation,
                               getSimulationEvents, compareSimulations
    handler_export.go       ← exportMarkdown, exportJSON, exportPDF
    handler_feedback.go     ← submitFeedback, getSimulationAccuracy,
                               confirmRupture, getSimulationConfirmations,
                               getCompanyAccuracy, getCompanyConfirmations
    handler_config.go       ← getConfig, setConfig, onboardingStatus,
                               completeOnboarding, getCompany, upsertCompany,
                               validateKey, getAuditLog, getTelemetry,
                               setTelemetry, checkForUpdate
    handler_knowledge.go    ← listArchetypes, createArchetype, getArchetype,
                               updateArchetype, deleteArchetype,
                               listRules, listRulesByDomain, createRule,
                               getCustomRule, updateRule, deleteCustomRule,
                               listTemplates, getTemplate
    handler_pulse.go        ← quickPulse, extractContext
  engine/
    world.go           ← Rule, Agent, World structs
    world_domains.go   ← DefaultWorldForDomain*
    simulation.go      ← RunSimulation, motor híbrido
    councils.go        ← BuildCouncils, RunCouncilDebate
    ensemble.go        ← RunEnsemble, tripartite classification
    report.go          ← ReportGenerator, FullReport
    export.go          ← ReportToMarkdown, ReportToHTML
    compare.go         ← CompareReports
    config.go          ← SimulationMode, ModeConfig
  archetypes/
    conformists.go     ← 61 conformistas com personas reais
    disruptors.go      ← 28 disruptores com personas reais
  skills/
    skill.go           ← Skill interface, Registry, Detect()
    init.go            ← Register() para 13 skills
    healthcare.go, fintech.go, retail.go, legal.go, education.go
    agro.go, construction.go, logistics.go, saas.go, energy.go
    manufacturing.go, media.go, tourism.go
  deepsearch/
    agent.go           ← DeepSearch, EnrichWithHistory
    domain_research.go ← DomainResearcher, DomainResearchResult
    sentiment.go       ← CalculateSentimentScore
  memory/
    store.go           ← Store, SaveRound, GetSimulationHistory
    calibration.go     ← Calibrator, CausalityGraph
    rag.go             ← RAGStore, TF-IDF local
  db/
    db.go              ← DB struct, all queries
    schema.sql         ← base schema
    migrations/        ← 001-011 SQL migrations
  security/
    hmac.go, sanitizer.go
  llm/
    client.go
  contextextractor/
  telemetry/
  updater/
  dashboard/src/pages/
    HomePage.tsx, NewSimulationPage.tsx, SimulationsPage.tsx
    ResultPage.tsx, FeedbackPage.tsx, ComparisonPage.tsx
    ConvergencePage.tsx, AccuracyPage.tsx

## Motor de Simulação

Pipeline:
1. DeepSearch pesquisa 7 domínios com reflexão
2. Skill detectada → carrega regras + agentes especializados
3. 89+ agentes recebem contexto calibrado
4. Motor híbrido: determinístico + heurístico + LLM
5. Rounds adaptativos (30 Standard / 50 Premium)
6. Councils debatem a cada 5 rounds (1 LLM + retry)
7. Premium: 2 runs → Ensemble tripartite
8. Relatório com confidence, evidence trail, playbook

Modos:
- Standard: 30 rounds, 1 run
- Premium: 50 rounds, 2 runs, ensemble tripartite

## As 166 Mentes

61 Conformistas por domínio:
  Market: Buffett, Porter, Kotler, Collins, Welch, Schultz,
          Walton, Gerstner, Iger, Immelt, Ellison
  Regulation: Lagarde, Khan, Powell, Stiglitz, Roubini, Gensler, Draghi
  Finance: Dalio, Munger, Graham, Bogle, Soros
  Behavior: Drucker, Kahneman, Thaler, Grant, Lencioni, McGregor, Simon, Maslow
  Culture: Godin, Gladwell, Brown, Sinek
  Geopolitics: Kissinger, Bremmer, Acemoglu, Harari, Ferguson, Fukuyama
  Technology: Gates, Grove, Cerf, Berners-Lee
  Management: Taylor, Fayol, Weber, Follett, Barnard, Chandler
  Marketing: Levitt, Rogers
  Systems: Forrester, Meadows, Brian Arthur
  Strategy: Adam Smith, Schumpeter, Coase, Marx, Roger Martin

28 Disruptores:
  Musk, Huang, Altman, Andreessen, Thiel, Bezos, Naval,
  Hastings, Chesky, Collison, Ek, Hormozi, Wood, Srinivasan,
  Dixon, Kurzweil, Christensen, Dorsey, Buterin,
  Kuhn, Taleb, Zuboff, Mazzucato, Kate Raworth,
  Kai-Fu Lee, Carlota Perez, Adam Smith (disruptor), Marx (disruptor)

77 Especialistas de Skill:
  Healthcare: Farmer, Rosenthal, Gawande, Topol, Etges, Berwick, Rosling
  Fintech: Vélez, Muszkat, Esteves, Mejía, Fraga, De Soto, Yunus
  Retail: Galperin, Trajano, Skora, Diniz, Zemel, Nader
  Legal: Kessler, Carvalhosa, Sobral Pinto, Bermudes
  Education: Khan, Gardner, Koller, Freire
  Agro: Jank, Shiva, Maggi, Raj Patel, Borlaug, Conway
  Construction: Horn, Fischer, Gehl, Fuller, Cano, Jane Jacobs
  Logistics: Schvartsman, Hau Lee, Glaeser, Palmaka, Christopher, Sheffi
  SaaS: Horowitz, Lemkin, Dubugras, Levie, Sacks, Moore
  Energy: Prates, Wajsman, Lovins, Smil, Rifkin
  Manufacturing: Skaf, Ohno, Rial, Goldratt, Deming, Pareto, Juran
  Media: Marinho, MrBeast, McLuhan, Ogilvy, Thompson, Lippmann
  Tourism: Kelleher, Chammas, Trippe, Paulus, Conley

## API — Endpoints (/api/v1/)

POST   /simulations
GET    /simulations
GET    /simulations/{id}
DELETE /simulations/{id}
GET    /simulations/{id}/results
GET    /simulations/{id}/report
GET    /simulations/{id}/stream       ← SSE
GET    /simulations/{id}/events
GET    /simulations/{id}/export/markdown
GET    /simulations/{id}/export/json
GET    /simulations/{id}/export/pdf   ← HTML print-optimized
GET    /simulations/compare?ids=...
POST   /simulations/{id}/feedback
GET    /simulations/{id}/accuracy
POST   /simulations/{id}/confirm-rupture
GET    /simulations/{id}/confirmations
GET    /company/accuracy
GET    /company/confirmations

## Banco de Dados (SQLite, migrations 001-011)

simulation_jobs       ← jobs com mode, skill, status
simulation_rounds     ← rounds com agent actions
simulation_events     ← SSE events
fracture_votes        ← votações
domain_contexts       ← contexto por domínio + sentiment
domain_research_state ← ResumableState
rag_documents         ← RAG local TF-IDF
agent_memory          ← memória por agente
feedback              ← outcome + delta_score + predicted/actual
archetype_calibration ← accuracy_weight EMA por arquétipo
confirmed_ruptures    ← rupturas confirmadas no mundo real
audit_log             ← HMAC signed events

## Regras de Desenvolvimento

1. NUNCA criar agentes genéricos — toda mente deve ser real
   com obra documentada e legado verificável
2. NUNCA usar math.Abs em sentiment — fórmula aprovada:
   adjusted = base * (1.0 - sentiment*0.3), clamp [0.05, 0.50]
3. NUNCA criar tabela report_ir
4. Councils injetam Evidence no World, NÃO como Rule
5. Cada skill tem: rules + agents + context + queries
6. go test ./... deve passar antes de qualquer commit
7. Adam Smith e Karl Marx aparecem nos dois pools
   (conformistas e disruptores) — isso é intencional

## Decisões de Arquitetura

- Local-first: SQLite, sem Docker, sem banco externo
- Motor híbrido: reduz LLM calls ~60% vs puro LLM
- TF-IDF local: RAG sem dependência de embedding API
- ResumableState: DeepSearch retoma após queda de API
- Councils: 1 LLM call por conselho + retry, semaphore cap=3
- Ensemble: Consensus ≥60%, WeakSignals >1 run, Minority 1 run
- personalityFactor: 1.1-1.5 para disruptores
- PDF: HTML print-optimized, zero deps externas

## Fluxo de Trabalho

Claude (chat) → arquitetura, decisões, briefings
Claude Code → executa no repositório
Manus → merge, releases, GitHub UI
