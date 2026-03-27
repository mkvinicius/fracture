# FRACTURE — Quickstart

## O que é

FRACTURE simula como regras de mercado se quebram —
e te ajuda a ser o primeiro a quebrá-las.

166 mentes brilhantes (Buffett, Kahneman, Taiichi Ohno,
Paulo Freire, Nassim Taleb...) debatem o futuro do
seu mercado e entregam um relatório estratégico com
rupturas previstas, confiança e playbook de ação.

---

## Instalação

### Download (recomendado)

1. Baixe o binário em:
   https://github.com/mkvinicius/fracture/releases

2. Configure as variáveis de ambiente:
```bash
export LLM_API_KEY=sk-ant-...
export LLM_BASE_URL=https://api.anthropic.com
export LLM_MODEL_NAME=claude-sonnet-4-20250514

# Busca web (opcional — DuckDuckGo se não configurado)
export TAVILY_API_KEY=tvly-...
```

3. Execute:
```bash
./fracture
```

4. Acesse: http://localhost:8080

### Compilar do código-fonte

Requisitos: Go 1.24+, Node.js 20+, pnpm
```bash
git clone https://github.com/mkvinicius/fracture
cd fracture
make build
./fracture
```

---

## Primeira simulação via dashboard

1. Acesse http://localhost:8080
2. Clique em "Nova Simulação"
3. Digite sua pergunta estratégica
4. Escolha o setor (opcional — auto-detectado)
5. Escolha o modo: Standard ou Premium
6. Clique "Iniciar Simulação"

---

## Primeira simulação via API
```bash
# Criar simulação
curl -X POST http://localhost:8080/api/v1/simulations \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Como o PIX vai impactar meu banco nos próximos 2 anos?",
    "mode": "standard",
    "industry": "fintech",
    "company": "BancoXYZ"
  }'

# Resposta: { "id": "abc-123", "status": "queued" }

# Acompanhar em tempo real
curl -N http://localhost:8080/api/v1/simulations/abc-123/stream

# Buscar resultado
curl http://localhost:8080/api/v1/simulations/abc-123/results

# Exportar PDF
open http://localhost:8080/api/v1/simulations/abc-123/export/pdf
```

---

## Modos de Simulação

| | Standard | Premium |
|---|---|---|
| Rounds | 30 adaptativos | 50 adaptativos |
| Runs | 1 | 2 (ensemble) |
| Resultado | Consensus | Consensus + WeakSignals + Minority |
| Tempo est. | 3-5 min | 8-12 min |
| Custo est. | $0.30-0.80 | $1.50-4.00 |

---

## Skills Verticais

O FRACTURE detecta automaticamente o setor pela pergunta,
ou você pode especificar explicitamente:

| Skill | Slug | Especialistas |
|---|---|---|
| Healthcare | healthcare | Gawande, Topol, Berwick, Rosling |
| Fintech | fintech | Vélez, Fraga, Yunus |
| Retail | retail | Galperin, Trajano, Nader |
| Legal | legal | Sobral Pinto, Kessler |
| Education | education | Freire, Gardner, Khan |
| Agro | agro | Borlaug, Shiva, Conway |
| Construção | construction | Gehl, Fuller, Jane Jacobs |
| Logística | logistics | Hau Lee, Sheffi, Christopher |
| SaaS B2B | saas | Horowitz, Lemkin, Moore |
| Energia | energy | Lovins, Smil, Rifkin |
| Manufatura | manufacturing | Ohno, Deming, Goldratt |
| Mídia | media | McLuhan, Ogilvy, Thompson |
| Turismo | tourism | Kelleher, Conley |

---

## Ciclo de Aprendizado
Simulação → Relatório → Feedback → Calibração → Precisão

1. Após a simulação, clique "Give Feedback"
2. Avalie a precisão (delta_score -1.0 a 1.0)
3. Quando uma ruptura se confirmar: "✓ Esta ruptura se confirmou"
4. Veja sua AccuracyPage — histórico de precisão
5. Agentes que acertam ganham mais peso nas próximas simulações

---

## Documentação completa

- [API Reference](api.md)
- [CHANGELOG](../CHANGELOG.md)
