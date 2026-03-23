# FRACTURE

> **Simulate how market rules break — and be the one to break them first.**
>
> **Simule como as regras do mercado quebram — e seja o primeiro a quebrá-las.**

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![release](https://img.shields.io/badge/release-v1.4.1-red.svg)](https://github.com/mkvinicius/fracture/releases/latest)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20Windows-lightgrey.svg)](https://github.com/mkvinicius/fracture/releases/latest)
[![Tests](https://img.shields.io/badge/tests-20%20passing-brightgreen.svg)](https://github.com/mkvinicius/fracture)

---

## O que é o FRACTURE?

FRACTURE é uma **ferramenta local de simulação estratégica de mercado** que roda no seu computador. Você faz uma pergunta estratégica, o sistema pesquisa automaticamente o mercado com o **DeepSearch**, e coloca **32 agentes de IA** — cada um com personalidade, objetivo e poder distintos — para simular como as regras do seu mercado podem ser reescritas ao longo de **40 rounds**.

Quando a tensão acumula e a pressão estoura, acontece um **FRACTURE POINT**: uma regra fundamental muda. O relatório final com **6 partes** te diz o que vai acontecer, quando vai acontecer, e o que você deve fazer antes que aconteça.

**Seus dados ficam no seu computador. Sem assinatura. Sem nuvem. Sem servidor externo.**

---

## Como funciona

```
1. Você faz uma pergunta estratégica
         ↓
2. 🔍 DeepSearch pesquisa automaticamente o mercado
   (notícias recentes, concorrentes, tendências, mudanças regulatórias)
         ↓
3. FRACTURE constrói um Mundo com 55+ regras em 7 domínios
         ↓
4. 32 agentes de IA interagem por 40 rounds
   — formam alianças, traem uns aos outros, tensionam as regras
         ↓
5. Quando a tensão estoura → FRACTURE POINT (uma regra fundamental muda)
         ↓
6. Relatório completo com 6 partes é gerado
```

---

## DeepSearch — Contexto Real Antes da Simulação

Antes de qualquer simulação começar, o FRACTURE pesquisa automaticamente:

- Notícias recentes sobre o setor e concorrentes
- Movimentos estratégicos de players do mercado
- Tendências emergentes e tecnologias disruptivas
- Mudanças regulatórias relevantes

Os 32 agentes não trabalham com suposições genéricas — eles recebem **contexto real e atualizado** antes de interagir. Isso torna cada simulação específica para o seu mercado.

### Import de Redes Sociais e Sites

Na tela "New Simulation", adicione até 10 URLs de fontes externas:

| Fonte | O que extrai |
|---|---|
| 🌐 Site da empresa | Posicionamento, produtos, diferenciais |
| 💼 LinkedIn | Perfil corporativo, tamanho, crescimento |
| 📸 Instagram | Identidade de marca, engajamento |
| 𝕏 Twitter / X | Posicionamento público, tendências |
| 📘 Facebook | Presença e comunidade |
| ▶️ YouTube | Conteúdo estratégico e narrativa |

---

## Os 32 Agentes

### 20 Conformistas — defendem as regras atuais

| Agente | Papel |
|---|---|
| Skeptical Consumer | Questiona mudanças, exige provas |
| Enthusiast Consumer | Early adopter, amplifica tendências |
| Established Competitor | Protege posição de mercado |
| Emerging Competitor | Desafia incumbentes |
| Regulator | Aplica compliance e regulação |
| Strategic Supplier | Controla insumos críticos |
| Investor | Aloca capital baseado em retorno |
| Key Employee | Molda cultura interna |
| Legacy Media | Controla narrativa e percepção pública |
| Corporate B2B Buyer | Avesso a risco, valoriza estabilidade |
| Distribution Channel Partner | Protege margens de intermediário |
| Labor Union | Defende direitos e salários |
| Secondary Supplier | Redundância na cadeia de suprimentos |
| Industry Analyst | Molda expectativas de mercado |
| Insurance Underwriter | Precifica e transfere risco |
| Pension Fund Manager | Capital de longo prazo, baixa tolerância a risco |
| Platform Ecosystem Partner | Dependente das regras da plataforma |
| Local Government | Aplica regulação local |
| Traditional Retailer | Varejo físico, resiste à digitalização |
| Academic Institution | Valida conhecimento e credenciais |

### 12 Disruptores — desafiam e reescrevem as regras

| Agente | Papel |
|---|---|
| Tech Innovator | Constrói tecnologia que torna regras antigas obsoletas |
| Business Model Changer | Reescreve como valor é criado e capturado |
| Progressive Regulator | Pressiona por novos marcos regulatórios |
| Organized Consumer | Ação coletiva para forçar mudança de mercado |
| Venture Capital Fund | Financia apostas assimétricas em ruptura |
| Big Tech Entrant | Entra em mercados adjacentes com alavancagem de plataforma |
| Social Movement | Muda regras culturais por pressão coletiva |
| International Regulator | Impõe requisitos de compliance transfronteiriços |
| Open Source Community | Commoditiza tecnologia proprietária |
| Sovereign Wealth Fund | Capital estatal com objetivos geopolíticos |
| Adjacent Startup | Ataca de um ângulo inesperado |
| Whistleblower | Expõe regras ocultas que mantêm o status quo |

---

## O Relatório Final — 6 Partes

Após a simulação, você recebe um relatório completo com:

| Parte | O que entrega |
|---|---|
| **1. Futuro Provável** | O que acontece se nada mudar — com probabilidade e prazo |
| **2. Mapa de Tensão** | Quais regras estão mais próximas de quebrar e por quê |
| **3. Cenários de Ruptura** | Os 3 caminhos de FRACTURE POINT mais prováveis |
| **4. Mapa de Coalizões** | Quem se aliou com quem durante a simulação |
| **5. Linha do Tempo** | Quando cada ruptura deve acontecer (90 dias / 1 ano / 3 anos) |
| **6. Playbook de Ação** | O que você deve fazer agora para se posicionar antes da quebra |

Todos os relatórios incluem **watermark** com versão, licença e URL do projeto.

---

## Os 7 Domínios do Mundo

O FRACTURE simula 55+ regras distribuídas em 7 domínios:

| Domínio | Regras | Exemplos |
|---|---|---|
| **Market** | 12 | Switching costs, network effects, poder de precificação |
| **Technology** | 10 | Custos de IA, commoditização open source, edge computing |
| **Regulation** | 8 | Antitruste, auditabilidade de IA, sandboxes regulatórios |
| **Behavior** | 9 | Trabalho remoto, modelos de compensação, liderança |
| **Culture** | 8 | Creator economy, autenticidade, compras orientadas por comunidade |
| **Geopolitics** | 8 | Sanções comerciais, soberania digital, resiliência de supply chain |
| **Finance** | 8 | Alocação de capital, ESG, tokenização, múltiplos de receita |

---

## Notificação Automática de Atualização

Ao abrir o FRACTURE, ele verifica silenciosamente se há uma versão nova no GitHub. Se houver, um banner discreto aparece no canto superior direito com link direto para download. Sem precisar verificar manualmente.

---

## Segurança

- **Proteção contra prompt injection** — todos os dados externos são sanitizados antes de chegar aos agentes
- **Prompts assinados com HMAC** — cada prompt de agente é assinado para detectar adulteração
- **Log de auditoria imutável** — cada evento da simulação é encadeado com assinaturas HMAC
- **Sandboxing de agentes** — agentes não têm acesso ao sistema de arquivos ou rede
- **Local-first** — todos os dados ficam na sua máquina

---

## Telemetria — Opt-in, Transparente

O FRACTURE coleta dados anônimos **apenas se você autorizar** durante o onboarding ou nas configurações:

**O que é coletado:**
- Install ID (UUID anônimo, gerado aleatoriamente, nunca vinculado a você)
- Sistema operacional e arquitetura
- País (derivado do IP, último octeto mascarado)
- Versão do FRACTURE

**O que nunca é coletado:** conteúdo de simulações, chaves de API, dados da empresa ou qualquer informação pessoal.

Você pode ativar ou desativar a qualquer momento em **Settings → Privacy & Telemetry**.

---

## Instalação

### Download direto (recomendado)

Acesse a [página de releases](https://github.com/mkvinicius/fracture/releases/latest) e baixe o arquivo para seu sistema:

| Sistema | Arquivo |
|---|---|
| Linux x86_64 | `fracture-linux-amd64.tar.gz` |
| Linux ARM64 | `fracture-linux-arm64.tar.gz` |
| Windows x86_64 | `fracture-windows-amd64.zip` |

**Linux:**
```bash
tar -xzf fracture-linux-amd64.tar.gz
chmod +x fracture
./fracture
```

**Windows:**
Extraia o `.zip` e execute `fracture-windows-amd64.exe`.

O FRACTURE abre automaticamente no navegador em `http://localhost:7432`.

### Compilar do código-fonte

**Requisitos:** Go 1.22+, Node.js 20+, pnpm

```bash
git clone https://github.com/mkvinicius/fracture.git
cd fracture
make build
./fracture
```

Para desenvolvimento com hot reload:
```bash
make dev
```

---

## Configuração Mínima

| Componente | Mínimo | Recomendado |
|---|---|---|
| OS | Windows 10 64-bit / Ubuntu 20.04+ | Windows 11 / Ubuntu 22.04+ |
| RAM | 4 GB | 8 GB+ |
| CPU | Dual-core 64-bit | 4+ núcleos |
| Disco | 50 MB | 100 MB |
| Internet | Sim (para API de IA e DeepSearch) | Conexão estável |
| API Key | OpenAI, Anthropic, Google ou Ollama | OpenAI GPT-4o mini |

Sem Docker. Sem banco de dados externo. Sem instalação de dependências.

---

## Configuração da API de IA

Na primeira execução, acesse **Settings → API Keys** e adicione sua chave:

| Provedor | Modelo padrão | Custo estimado por simulação |
|---|---|---|
| OpenAI | GPT-4o mini | ~$0,20–0,40 |
| Anthropic | Claude Haiku | ~$0,15–0,30 |
| Google | Gemini Flash | ~$0,10–0,25 |
| Ollama | Llama 3 (local) | Grátis (sem internet) |

As chaves são armazenadas localmente no SQLite da sua máquina e **nunca enviadas a nenhum servidor externo**.

---

## Integração MCP (Cursor, VS Code, Windsurf)

Use o FRACTURE diretamente no seu editor de código via Model Context Protocol:

```json
{
  "mcpServers": {
    "fracture": {
      "command": "/caminho/para/fracture",
      "args": ["--mcp"],
      "env": {}
    }
  }
}
```

Ferramentas disponíveis via MCP:
- `fracture_simulate` — roda uma simulação completa
- `fracture_list_simulations` — lista simulações anteriores
- `fracture_get_report` — recupera o relatório de uma simulação

---

## Diferença para o MiroFish

| | MiroFish | FRACTURE |
|---|---|---|
| **Abordagem** | Prevê tendências com dados históricos | Simula o que ainda não aconteceu |
| **Método** | Análise de dados passados | Simulação multi-agente com DeepSearch |
| **Agentes** | Nenhum | 32 agentes de IA com personalidades distintas |
| **Regras** | Modelo fixo | 55+ regras mutáveis que quebram sob pressão |
| **Output** | Previsão de tendências | Relatório de 6 partes + Playbook de Ação |
| **Posicionamento** | Te diz o que vai acontecer | Te posiciona antes da quebra acontecer |

---

## Arquitetura

```
fracture/
  main.go                  ← Entry point, HTTP server, browser open
  api/handler.go           ← REST API routes + DeepSearch integration
  engine/
    world.go               ← Rule graph with stability weights
    agent.go               ← Agent interface and base types
    simulation.go          ← Main simulation loop (40 rounds)
    voting.go              ← Weighted consensus voting
    report.go              ← Report generation (6 output types + watermark)
    world_domains.go       ← 55+ rules across 7 domains
  archetypes/
    conformists.go         ← 20 Conformist archetypes
    disruptors.go          ← 12 Disruptor archetypes
  deepsearch/
    agent.go               ← Multi-round deep search with reflection
  contextextractor/
    extractor.go           ← URL scraping (site + social media)
  updater/
    updater.go             ← Auto-update check via GitHub API
  memory/
    store.go               ← SQLite-backed agent memory
    calibration.go         ← Feedback loop + archetype calibration
  security/
    sanitizer.go           ← Prompt injection protection
    hmac.go                ← HMAC signing + immutable audit log
  telemetry/
    telemetry.go           ← Anonymous usage tracking (opt-in)
  llm/client.go            ← LLM-agnostic client (semaphore, retry, cache)
  db/
    db.go                  ← SQLite helpers
    schema.sql             ← Database schema
  dashboard/               ← React + Tailwind frontend (embedded in binary)
```

---

## Histórico de Versões

| Versão | Principais mudanças |
|---|---|
| **v1.4.1** | 20 testes unitários, correção DetectSourceType, versão interna corrigida |
| **v1.4.0** | DeepSearch (pesquisa automática de mercado), import de redes sociais, notificação de atualização |
| **v1.3.0** | Import de URLs externas, notificação de atualização automática |
| **v1.2.0** | 32 agentes, 40 rounds, 55+ regras em 7 domínios, relatório com 6 partes + Playbook de Ação |
| **v1.1.0** | Telemetria opt-in, watermark nos relatórios, AGPL-3.0, SECURITY.md |
| **v1.0-alpha** | Lançamento inicial — 12 agentes, 20 rounds, 3 outputs |

---

## Licença

FRACTURE é distribuído sob a [GNU Affero General Public License v3.0 (AGPL-3.0)](LICENSE).

Isso significa:
- Você pode usar, estudar e modificar o FRACTURE livremente
- Se você distribuir uma versão modificada (inclusive como serviço web), deve disponibilizar o código-fonte sob a mesma licença
- Uso comercial requer conformidade com os termos da AGPL-3.0

---

## Links

- 📦 [Releases e Downloads](https://github.com/mkvinicius/fracture/releases/latest)
- 🐛 [Reportar um bug](https://github.com/mkvinicius/fracture/issues)
- 💡 [Sugerir uma feature](https://github.com/mkvinicius/fracture/issues)
- 📄 [Licença AGPL-3.0](LICENSE)
- 🔒 [Política de Segurança](SECURITY.md)

---

© 2025 FRACTURE contributors
