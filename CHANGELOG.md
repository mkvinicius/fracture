# FRACTURE Changelog

## [v2.9.0] — 2026-03-26
### Added
- **Accuracy & Calibration**: página `AccuracyPage` com estatísticas de feedback por empresa
  - GET /company/accuracy — delta médio, contagem por tipo, calibração por arquétipo
  - GET /simulations/{id}/accuracy — feedback + confirmações de ruptura de uma simulação
  - Barras de distribuição (preciso / parcial / impreciso) e tabela de peso por arquétipo
  - Badges de outcome em `SimulationsPage` carregados assincronamente
  - Link "Accuracy" na Sidebar
- **Rupturas Confirmadas**: registro de ruptura real diretamente no relatório
  - POST /simulations/{id}/confirm-rupture — registra ruptura com rule_id, descrição e notas
  - GET /simulations/{id}/confirmations — lista confirmações da simulação
  - GET /company/confirmations — lista todas as confirmações da empresa
  - Botão "✓ Esta ruptura se confirmou" em cada card de Ruptura do ResultPage
  - Badge "✓ CONFIRMADA" aparece após confirmação; seção na AccuracyPage
  - Migration 011_confirmed_ruptures.sql
- **Feedback delta**: migration 010_feedback_delta.sql adiciona `predicted_fracture`,
  `actual_outcome` e `delta_score` à tabela `feedback`; `SaveFeedback()` atualizado

## [v2.8.0] — 2026-03-26
### Added
- PDF export via HTML otimizado para impressão (print CSS, zero dependências novas)
  - GET /simulations/{id}/export/pdf — serve HTML auto-contido
  - `ReportToHTML()` em engine/export.go: todas as seções, XSS-safe, skill badge
  - Botão "📄 PDF" no ResultPage (abre em nova aba → Ctrl+P → Salvar como PDF)
  - 5 novos testes em engine/export_test.go

## [v2.5.1] — 2026-03-26
### Added
- 73 novas mentes no sistema: 19 conformistas + 9 disruptores + 13 especialistas de skill
  - Conformistas: Frederick Taylor, Henri Fayol, Max Weber, Mary Parker Follett, Chester Barnard,
    Douglas McGregor, Herbert Simon, Adam Smith, Joseph Schumpeter, Ronald Coase, Karl Marx,
    Alfred Chandler, Abraham Maslow, Theodore Levitt, Everett Rogers, Jay Forrester,
    Donella Meadows, W. Brian Arthur, Roger Martin
  - Disruptores: Thomas Kuhn, Nassim Taleb, Adam Smith (anti-monopólio), Karl Marx (revolução),
    Shoshana Zuboff, Mariana Mazzucato, Kate Raworth, Kai-Fu Lee, Carlota Perez
  - Skills: Hans Rosling (healthcare), Muhammad Yunus (fintech), Gordon Conway (agro),
    Jane Jacobs (construction), Yossi Sheffi (logistics), Jeremy Rifkin (energy),
    W. Edwards Deming + Vilfredo Pareto + Joseph Juran (manufacturing),
    Walter Lippmann (media)
- Enriquecimento biográfico de Andy Grove, Jeff Immelt, John Bogle, Vint Cerf

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
