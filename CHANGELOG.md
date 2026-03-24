# CHANGELOG

All notable changes to FRACTURE are documented in this file.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [1.10.0] — 2026-03-23

### Added
- **Per-domain reflection in DeepSearch**: cada domínio (market, technology, regulation, behavior, culture, geopolitics, finance) agora tem seu próprio ciclo de busca → síntese → reflexão → busca complementar.
- **Configurable reflection depth**: `MaxReflectionsPerDomain` em Config; defaults: regulation=3, geopolitics=3, technology=2, finance=2, market=2, behavior=1, culture=1 — domínios complexos refletem mais.
- **Gap-driven reflection loop**: `researchDomainWithReflection()` identifica gaps após cada síntese e busca complementar; encerra quando gaps vazios (sem loop infinito).
- **Resumable research state**: pesquisas interrompidas retomam de onde pararam — estado salvo em `domain_research_state` table com hash estável da question.
- **`ResumableState` struct**: Question, Company, Sector, Completed (map de domínios já pesquisados), StartedAt — pronto para retomada.
- **`hashQuestion()` utility**: gera hash SHA256 estável para (question, company, sector) — chave para resumable state.
- **DB helpers**: `SaveResearchState()`, `GetResearchState()`, `DeleteResearchState()` — estado removido automaticamente após sucesso.

### Technical
- `domain_research_state` table com índice em `question_hash`.
- Reflection loop encerra corretamente: `if len(gaps) == 0 { break }` sem timeout.
- Imports: `crypto/sha256`, `encoding/hex` para hash generation.
- Semáforo cap=3 existente mantido; timeout por domínio preservado.

### Tests
- **71 tests passing** (sem regressão). Todas as funções novas compilam e testam.

---

## [1.9.0] — 2026-03-23

### Added
- **Domain-aware DeepSearch**: pesquisa cada um dos 7 domínios (market, technology, regulation, behavior, culture, geopolitics, finance) separadamente com queries otimizadas por domínio.
- **Per-domain stability calibration**: regras ajustadas por confiança da evidência real — `stabilityModifier()` aplica ajuste apenas em `affectedRules` com confidence ≥ 0.6.
- **Evidence field em World struct**: armazena contexto real-world sem virar regra no grafo de tensão — metadado puro para auditoria.
- **`stability_modifier` persistido**: nova coluna em `domain_contexts` table para auditoria histórica e calibração futura — permite comparar modifier aplicado vs resultado da simulação.
- **`SynthesizeDomainContext()` pública**: extrai insights por domínio de `ContextReport` (market, technology, regulation, behavior, culture, finance) e retorna mapa de domain → {ContextText, AffectedRules, Confidence}.
- **Integração em `runWithDeepSearch`**: após DeepSearch completar, persiste domain contexts com stability_modifier calculado como `-0.15 * confidence`.
- **3 novos testes**: `TestSaveDomainContext`, `TestGetDomainContexts`, `TestDefaultWorldForDomainWithContext` — verifica Evidence preenchido, affected rules com stability reduzida, e idempotência.

### Technical
- `domain_contexts` table com índices em `simulation_id` e `domain`.
- Stability clamp [0.05, 0.95] para evitar valores extremos.
- Modifier aplicado APENAS em `affectedRules`, não em todas as regras do domínio.
- Confidence mínima 0.6 para aplicar ajuste; valores menores não modificam estabilidade.
- `DefaultWorldForDomainWithContext()` em `engine/world_domains.go` injeta evidências com ajuste automático.

### Tests
- **71 tests passing** (era 68). Todos os novos testes de domain contexts passam.

---

## [1.8.0] — 2026-03-23

### Fixed
- **`listArchetypes` now returns built-ins + custom merged**: the endpoint previously returned only a hardcoded built-in list. It now queries `db.ListArchetypes(companyID)` and merges results — custom archetypes override built-ins with the same ID; new ones are appended. Pass `?company_id=<id>` to filter.
- **`listRules` and `listRulesByDomain` now merge custom rules from DB**: both endpoints previously returned only `DefaultWorldForDomain(...)` built-ins. They now call `db.ListCustomRules(companyID)` and append active custom rules. Pass `?company_id=<id>` to include company-specific rules.
- **Route conflict resolved**: `GET /rules/{domain}` renamed to `GET /rules/domain/{domain}` to avoid colliding with the new `GET /rules/{id}` endpoint.

### Added
- `GET /archetypes/{id}` — returns a single archetype (custom from DB first, then built-in fallback).
- `DELETE /archetypes/{id}` — removes a custom archetype; built-ins return 403 Forbidden.
- `GET /rules/{id}` — returns a single custom rule by ID.
- `DELETE /rules/{id}` — removes a custom rule.
- `db/archetypes_test.go`: 13 new tests covering `CreateArchetype`, `UpdateArchetype`, `ListArchetypes`, `DeleteArchetype`, built-in protection, `CreateCustomRule`, `UpdateCustomRule`, `ListCustomRules`, `DeleteCustomRule`, progress columns (`current_round`, `current_tension`, `fracture_count`, `last_agent_name`, `last_agent_action`, `total_tokens`), `ListJobs` progress field survival, and full `StartReportGen`/`CompleteReportGen`/`ListReportGens` lifecycle.

### Tests
- **68 tests passing** (was 55). No regressions.

---

## [1.7.0] — 2026-03-23

### Fixed
- **Live progress fields now survive restarts**: `simulation_jobs` table gains 6 new columns (`current_round`, `current_tension`, `fracture_count`, `last_agent_name`, `last_agent_action`, `total_tokens`). `persistJob` writes them after every round; `NewHandler` re-hydrates them on startup so the SSE stream reflects accurate state after a process restart.
- **`StartReportGen` / `CompleteReportGen` now wired to runtime**: `runSimulation` calls both helpers around `GenerateReport`, recording timing, token usage, and error state in `report_generations` for every simulation.
- **`createArchetype` and `updateArchetype` fully implemented**: Both endpoints now parse the request body, validate fields, and persist to the `archetypes` table. Built-in archetypes (`company_id = ''`) are protected from mutation.
- **`createRule` and `updateRule` fully implemented**: Both endpoints persist to a new `custom_rules` table (migration 006), replacing the `{ok: true}` stubs.
- **Vote model semantics corrected**: `ProposalID` is now a stable composite key (`simID:roundN:proposerID`) instead of the proposer name alone. `VoterType` is now `conformist` or `disruptor` (derived from agent ID prefix), not the agent's display name. `Weight` and `Reasoning` are now populated from `VoteRecord`.
- **Migration runner is idempotent for `ALTER TABLE ADD COLUMN`**: New databases created from the updated `schema.sql` already have the progress columns; the runner now ignores "duplicate column" errors so migration 005 does not fail on fresh installs.

### Added
- Migration `005_job_progress.sql`: adds live progress columns to `simulation_jobs` for existing databases.
- Migration `006_custom_rules.sql`: new `custom_rules` table for company-specific world rules.
- `db/archetypes.go`: `CreateArchetype`, `UpdateArchetype`, `GetArchetype`, `ListArchetypes`, `DeleteArchetype`, `CreateCustomRule`, `UpdateCustomRule`, `GetCustomRule`, `ListCustomRules`, `DeleteCustomRule`.

### Tests
- All 55 tests pass (`go test ./... -race`). No regressions.

---

## [1.6.0] — 2026-03-23

### Fixed
- **Watermark version hardcoded**: `engine/report.go` now reads `updater.CurrentVersion` instead of the static string `"1.4.0"` — every generated report reflects the actual binary version.

### Added
- **Versioned migration system** (`db/migrate.go`): a lightweight runner that applies SQL migrations in order, tracks applied migrations in `schema_migrations`, and is idempotent across restarts.
- **Four initial migrations** (`db/migrations/001–004`): codify the full schema history — init tables, simulation jobs, execution trail (rounds/votes/report_generations), and the calibration/causality graph.
- **`simulation_rounds` table**: one row per agent action per round, enabling full replay, per-agent analytics, and future fine-tuning.
- **`fracture_votes` table**: one row per agent vote on each FRACTURE POINT proposal, enabling vote audit and calibration.
- **`report_generations` table**: tracks each report generation attempt with timing, token usage, and error messages.
- **`archetypes` table**: stores per-company archetype overrides and calibration weights — closes the gap entre `memory/calibration.go` e o schema.
- **`causality_nodes` + `causality_edges` tables**: closes the gap between `memory/calibration.go` CausalityGraph and the schema.
- **`db.SaveRound` / `db.ListRounds`**: persist and query simulation rounds.
- **`db.SaveVote` / `db.ListVotes`**: persist and query fracture votes.
- **`db.StartReportGen` / `db.CompleteReportGen` / `db.ListReportGens`**: track report generation lifecycle.
- **Live SSE progress fields** on `simJob`: `current_round`, `current_tension`, `fracture_count`, `last_agent_name`, `last_agent_action`, `total_tokens` — updated after every round and streamed to the UI in real time.
- **`persistRound` helper** in `api/handler.go`: called after each round to persist actions and votes to the DB and update live progress fields atomically.
- **7 new tests** for `db/rounds.go` (rounds, votes, report_generations) — total test count: **55 passing**.

### Changed
- `db.Open()` now calls `Migrate()` after applying the base schema, ensuring all migrations run on startup.
- `openTestDB` in `db_test.go` now also applies migrations so new tables are available in tests.
- Test count: **33 → 55 passing**.

---

## [1.5.0] — 2025-03-23

### Fixed
- **Version consistency** — internal telemetry and updater now correctly report `1.5.0`; previously the binary self-reported `1.0.0` in some code paths.
- **`openBrowser` now opens the real browser** — replaced the stub that only logged the URL with platform-specific implementations using `os/exec` (`xdg-open` on Linux, `open` on macOS, `cmd /c start` on Windows).
- **HMAC secret generation uses `crypto/rand`** — replaced `time.Now().UnixNano()` seed (predictable) with 32 bytes of cryptographically secure randomness.

### Added
- **Persistent simulation job state** — introduced `simulation_jobs` table in SQLite. Every status transition (`queued → researching → running → done/error`) is now written to the database immediately, so the UI correctly reflects job state after a process restart.
- **Startup resilience** — `NewHandler` calls `MarkInterruptedJobsFailed()` at boot and re-hydrates the in-memory map from the DB, so no job is silently lost across restarts.
- **Structured HTML parsing in context extractor** — replaced fragile regex-only stripping with `golang.org/x/net/html` tree walker. Correctly skips `<script>`, `<style>`, `<nav>`, `<footer>`, `<header>`, `<noscript>`, `<iframe>`, `<svg>`, `<form>` and other noise elements.
- **Per-URL timeout and retry in context extractor** — each URL now has a 12-second context deadline and up to 2 retries with exponential backoff (500 ms, 1 s) for transient network errors. HTTP 4xx/5xx errors are not retried.
- **Social network graceful degradation** — LinkedIn, Instagram, Twitter/X, and Facebook URLs that return login walls or bot blocks are flagged as `Limited: true` in the extracted context, with a clear message injected into the LLM prompt instead of a silent empty string.
- **HTML entity decoding via stdlib** — uses `html.UnescapeString` (standard library) for correct decoding of `&amp;`, `&lt;`, `&gt;`, `&quot;`, `&#39;`, `&nbsp;` and all numeric entities.
- **13 new DB tests** — full coverage of `UpsertJob`, `GetJob`, `ListJobs`, `DeleteJob`, `MarkInterruptedJobsFailed`, `SetConfig`, `GetConfig`, `SaveSimulation`, `GetSimulation`, `ListSimulations`, `DeleteSimulation`.
- **17 new extractor tests** — covers HTML node extraction, title/meta parsing, script stripping, truncation, HTTP success/404/500, URL normalization, concurrent extraction, race detection, social network limited flag, and `cleanRawText` fallback.

### Changed
- `go.mod` toolchain upgraded to Go 1.25.0 (required by `golang.org/x/net v0.52.0`).
- `deleteSimulation` handler now also removes the corresponding row from `simulation_jobs` to keep both tables in sync.
- Test count: **20 → 33 passing**.

---

## [1.4.1] — 2025-03-20

### Added
- DeepSearch integration: 32 agents now receive real-world market context before simulation begins.
- Import of up to 10 external URLs (company website, LinkedIn, Instagram, Twitter/X, Facebook, YouTube) as simulation context.
- 6-part structured report: Executive Summary, Fracture Scenarios, Strategic Recommendations, Market Dynamics, Risk Matrix, Action Plan.
- Audit log with HMAC-signed entries for tamper detection.
- Archetype calibration table for feedback-driven accuracy improvement.

### Changed
- Agent count increased from 20 to 32 (20 conformists + 12 disruptors).
- Simulation rounds increased to 40.
- Report generator now uses a dedicated synthesis LLM role.

---

## [1.0.0] — 2025-02-01

### Added
- Initial public release.
- Local-first simulation engine with SQLite persistence.
- 20 agents (conformists and disruptors) across 7 market domains.
- Basic report generation.
- Manus OAuth-free desktop mode (no cloud dependency).
- Auto-updater via GitHub Releases.
