# FRACTURE API Reference

## Base URL
http://localhost:8080/api/v1

## Autenticação
Sem autenticação no modo local.

---

## Simulations

### POST /simulations
Cria uma nova simulação.

**Request Body:**
```json
{
  "question": "Como o PIX vai impactar meu banco?",
  "mode": "standard",
  "industry": "fintech",
  "company": "MinhaEmpresa",
  "department": "market",
  "context": "contexto adicional opcional",
  "urls": ["https://exemplo.com"]
}
```

Campos:
- question (obrigatório): pergunta estratégica
- mode: "standard" (padrão) ou "premium"
- industry: slug da skill vertical (auto-detectado se omitido)
  Valores: healthcare, fintech, retail, legal, education,
           agro, construction, logistics, saas, energy,
           manufacturing, media, tourism
- company: nome da empresa (habilita memória e calibração)
- department: domínio principal (padrão: "market")
- context: contexto adicional para a simulação
- urls: URLs para scraping automático

**Response:**
```json
{ "id": "uuid", "status": "queued" }
```

---

### GET /simulations
Lista todas as simulações.

**Response:** array de simulation jobs com status.

---

### GET /simulations/{id}
Retorna status e metadados de uma simulação.

**Response:**
```json
{
  "id": "uuid",
  "status": "queued|researching|running|done|error",
  "question": "...",
  "mode": "standard|premium",
  "skill": "healthcare",
  "current_round": 12,
  "current_tension": 0.73,
  "fracture_count": 2,
  "last_agent_name": "Warren Buffett",
  "total_tokens": 45230,
  "duration_ms": 92000
}
```

---

### GET /simulations/{id}/stream
SSE stream de eventos em tempo real.

**Content-Type:** text/event-stream

**Eventos:**
- round_complete: round finalizado
- fracture_proposed: proposta de ruptura detectada
- council_debate: resultado de debate do council
- simulation_complete: simulação finalizada

**Exemplo:**
```bash
curl -N http://localhost:8080/api/v1/simulations/{id}/stream
```

---

### GET /simulations/{id}/results
Retorna o FullReport completo.

**Response:** objeto FullReport com todas as seções:
probable_future, tension_map, rupture_scenarios,
coalitions, rupture_timeline, action_playbook,
fracture_events, ensemble_result (Premium)

---

### DELETE /simulations/{id}
Remove uma simulação.

---

## Export

### GET /simulations/{id}/export/markdown
Exporta relatório em Markdown.

**Headers:** Content-Disposition: attachment; filename="fracture-{id}.md"

---

### GET /simulations/{id}/export/json
Exporta relatório em JSON formatado.

**Headers:** Content-Disposition: attachment; filename="fracture-{id}.json"

---

### GET /simulations/{id}/export/pdf
Abre relatório HTML otimizado para impressão.

**Headers:** Content-Disposition: inline; filename="fracture-{id}.html"

Use Ctrl+P → Salvar como PDF no browser.

---

## Comparison

### GET /simulations/compare?ids={id1},{id2}
Compara 2 a 5 simulações.

**Query params:** ids (obrigatório, separados por vírgula)

**Response:**
```json
{
  "simulation_ids": ["id1", "id2"],
  "common_fractures": ["ruptura presente em todas"],
  "divergent_fractures": { "id1": ["..."], "id2": ["..."] },
  "tension_delta": [{ "rule_id": "...", "delta": 0.3 }],
  "confidence_delta": 0.15,
  "summary": "texto descritivo do padrão observado"
}
```

---

## Feedback & Accuracy

### POST /simulations/{id}/feedback
Avalia o resultado de uma simulação.

**Request Body:**
```json
{
  "outcome": "accurate|partial|inaccurate",
  "predicted_fracture": "qual ruptura foi prevista",
  "actual_outcome": "o que realmente aconteceu",
  "delta_score": 0.8,
  "notes": "observações opcionais"
}
```

delta_score: -1.0 (completamente errado) a 1.0 (perfeito)

---

### GET /simulations/{id}/accuracy
Retorna o impacto do feedback nos arquétipos.

---

### POST /simulations/{id}/confirm-rupture
Registra que uma ruptura prevista se confirmou.

**Request Body:**
```json
{
  "predicted_rule": "qual ruptura se confirmou",
  "confirmation_notes": "como se confirmou no mundo real",
  "confirmed_at": 1711234567
}
```

**Response:**
```json
{ "id": "uuid", "days_to_confirm": 47 }
```

---

### GET /company/accuracy?company={name}
Score de precisão histórico da empresa.

**Response:**
```json
{
  "company_id": "MinhaEmpresa",
  "total_simulations": 12,
  "simulations_with_feedback": 8,
  "average_accuracy": 73.5,
  "accuracy_trend": "improving|stable|declining",
  "top_accurate_archetypes": [],
  "recent_feedback": []
}
```

---

### GET /company/confirmations?company={name}
Lista rupturas confirmadas da empresa.

---

## Knowledge Base

### GET /archetypes
Lista todos os arquétipos (166 mentes).

### GET /rules
Lista todas as regras ativas.

### GET /rules/domain/{domain}
Lista regras por domínio.

Domínios: market, regulation, finance, behavior,
          culture, geopolitics, technology

---

## Config

### GET /config
Retorna configuração atual (sem API keys).

### POST /config
Atualiza configuração.

**Request Body:**
```json
{
  "llm_api_key": "sk-...",
  "llm_base_url": "https://api.anthropic.com",
  "llm_model_name": "claude-sonnet-4-20250514",
  "tavily_api_key": "tvly-..."
}
```

---

### POST /pulse
Análise rápida sem simulação completa.

**Request Body:**
```json
{
  "question": "pergunta estratégica",
  "company": "MinhaEmpresa"
}
```

**Response:** análise em 30 segundos sem rounds completos.

---

## Tipos de Erro
```json
{ "error": "mensagem descritiva" }
```

HTTP 400: request inválido
HTTP 404: simulação não encontrada
HTTP 500: erro interno
