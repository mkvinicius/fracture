# FRACTURE Changelog

## [v2.6.0] — 2026-03-30
### Added
- Interface completamente traduzida para Português Brasileiro (todas as 12 telas)
- Instaladores 2-click: `install-windows.bat` e `install-mac.sh`
- Página de Arquétipos dinâmica: busca os 56 agentes reais da API (antes hardcoded com 12)
- `dashboard/dist` commitado no repositório — sem necessidade de `pnpm build` para instalar

### Fixed
- **Cache infinito do browser**: embed.FS servia arquivos com `Last-Modified: zero`, causando freshness heurístico de ~200 anos. Corrigido servindo `index.html` diretamente via Go com `Cache-Control: no-store`
- Assets movidos de `/assets/` para `/bundle/` para invalidar cache antigo
- `builtinArchetypes()` retornava lista hardcoded de 12 agentes em vez dos 56 do motor real
- Todos os endpoints do dashboard corrigidos para `/api/v1/`
- Downgrade React 19 → 18 para compatibilidade com MetaMask SES (`Map.prototype.getOrInsert`)
- Porta padrão alterada de 3000 para 4000

## [v2.5.0] — 2026-03-25
### Added
- Export: baixa relatório em Markdown ou JSON com botões no ResultPage
- GET /simulations/{id}/report — endpoint dedicado para FullReport
- GET /simulations/{id}/export/markdown — download Markdown estruturado
- GET /simulations/{id}/export/json — download JSON indentado
- Comparação: analisa padrões comuns e divergentes entre 2–5 simulações
- GET /simulations/compare?ids=... — ComparisonReport com common/divergent/delta
- ComparisonPage.tsx com barras side-by-side de tensão e ConfidenceDelta
- Convergência: vê graficamente como a tensão evoluiu por round
- GET /simulations/{id}/events — TensionPoint[] por round (avg + fracture count)
- ConvergencePage.tsx com gráfico SVG puro (sem dependências externas)
- Checkboxes em SimulationsPage para selecionar e comparar simulações
- Botões "Ver Convergência" por simulação em SimulationsPage
- engine/export.go: ReportToMarkdown() com todas as seções
- engine/compare.go: CompareReports() com common/divergent fractures e TensionDelta
- db/rounds.go: GetRoundTensions() — agregação de tensão por round
- 128+ testes passando

## [v2.4.0] — 2026-03-25
### Added
- RAG local com TF-IDF (sem dependências externas)
- Memória de simulações passadas por empresa (similarity search)
- Feedback loop: avaliação → calibração EMA → aprendizado por arquétipo
- ResultPage: visualização completa do FullReport (narrativa, tensão, rupturas, playbook)
- FeedbackPage: formulário de avaliação com delta_score slider (−1..+1)
- Botões "Ver Resultado" e "Dar Feedback" em SimulationsPage
- Roteamento de páginas estendido: result, feedback
- 109 testes passando

## [v2.3.1] — 2026-03-25
### Added
- 23 novos testes: deepsearch, memory, security
- Cobertura expandida para pacotes memory e deepsearch
- 91 testes passando

## [v2.3.0] — 2026-03-25
### Added
- Modos Standard (30 rounds, 1 run) e Premium (50 rounds, 2 runs)
- 56 agentes: 37 conformistas + 19 disruptores por domínio
- Councils por domínio (1 LLM por conselho + retry automático)
- Ensemble tripartite: Consensus / WeakSignals / Minority
- Sentiment score determinístico por domínio
- FractureEvent.Confidence com evidence trail
- 6 bug fixes críticos (Finalize, io.ReadAll, rows.Err, calibração)
- 68 testes passando

## [v2.1.0] — 2026-03-XX
### Added
- Motor híbrido: determinístico + heurístico + LLM
- Rodadas adaptativas com ConvergenceTracker
- Ativação seletiva de agentes por PowerRank e tensão
- Decay dinâmico de tensão por domínio
- CLAUDE.md com contexto do projeto

## [v2.0.0] — 2026-03-XX
### Added
- Histórico completo de eventos por simulação
- Replay endpoints: /events e /replay
- SSE streaming: /events/stream
- ReplayPage, RoundTimeline, EventCard, TensionChart, LiveStream

## [v1.10.0] — 2026-03-XX
### Added
- Reflexão por domínio com profundidade variável
- ResumableState: pesquisas retomam após queda de API

## [v1.9.0] — 2026-03-XX
### Added
- Domain-aware DeepSearch
- Evidence separada das Rules
- stabilityModifier por affected_rules
- stability_modifier como coluna de auditoria

## [v1.8.0] — 2026-03-XX
### Added
- listArchetypes: merge built-ins + custom
- listRules: merge custom rules from DB
- GET/DELETE /archetypes/{id}
- GET/DELETE /rules/{id}

## [v1.0.0] — 2026-03-XX
### Added
- Motor local de simulação com SQLite
- 20 agentes (conformistas e disruptores) em 7 domínios de mercado
- Geração básica de relatórios
- Modo desktop OAuth-free (sem dependência de cloud)
- Auto-updater via GitHub Releases
