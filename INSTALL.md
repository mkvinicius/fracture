# Instalação do FRACTURE

## Windows

### Opção 1 — Instalador automático (recomendado)

1. Baixe [`install-windows.bat`](https://github.com/mkvinicius/fracture/raw/main/install-windows.bat)
2. Clique com botão direito → **Executar como administrador**
3. O instalador instala Go, clona o repositório, compila e cria atalho na área de trabalho

### Opção 2 — Executável direto

1. Acesse [github.com/mkvinicius/fracture/releases/latest](https://github.com/mkvinicius/fracture/releases/latest)
2. Baixe `fracture-windows-amd64.exe`
3. Abra o PowerShell na pasta onde baixou e execute:
```powershell
.\fracture-windows-amd64.exe
```

4. O FRACTURE abre automaticamente em `http://localhost:4000`

### Opção 3 — Compilar do código-fonte

**Requisitos:**
- [Go 1.24+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [pnpm](https://pnpm.io/installation) — `npm install -g pnpm`
```powershell
git clone https://github.com/mkvinicius/fracture.git
cd fracture
cd dashboard && pnpm install && pnpm build && cd ..
go build -o fracture.exe .
.\fracture.exe
```

---

## Linux

### Opção 1 — Executável direto
```bash
curl -L https://github.com/mkvinicius/fracture/releases/latest/download/fracture-linux-amd64 -o fracture
chmod +x fracture
./fracture
```

### Opção 2 — Compilar do código-fonte

**Requisitos:** Go 1.24+, Node.js 20+, pnpm
```bash
git clone https://github.com/mkvinicius/fracture.git
cd fracture
cd dashboard && pnpm install && pnpm build && cd ..
go build -o fracture .
./fracture
```

---

## macOS

### Opção 1 — Instalador automático (recomendado)

```bash
curl -L https://github.com/mkvinicius/fracture/raw/main/install-mac.sh | bash
```

Instala Homebrew e Go se necessário, clona o repositório, compila e cria `FRACTURE.command` na área de trabalho.

### Opção 2 — Executável direto
```bash
curl -L https://github.com/mkvinicius/fracture/releases/latest/download/fracture-darwin-amd64 -o fracture
chmod +x fracture
./fracture
```

> Apple Silicon (M1/M2/M3): use `fracture-darwin-arm64`

### Opção 3 — Compilar do código-fonte
```bash
git clone https://github.com/mkvinicius/fracture.git
cd fracture
cd dashboard && pnpm install && pnpm build && cd ..
go build -o fracture .
./fracture
```

---

## Configuração da API de IA

Na primeira execução, acesse **Settings** no dashboard e configure sua chave:

| Provedor | Variável de ambiente | Custo estimado/simulação |
|---|---|---|
| OpenAI | `OPENAI_API_KEY` | ~$0.20–0.40 |
| Anthropic | `ANTHROPIC_API_KEY` | ~$0.15–0.30 |
| Google | `GOOGLE_API_KEY` | ~$0.10–0.25 |
| Ollama | — | Grátis (local) |

### Windows — definir variável de ambiente
```powershell
$env:OPENAI_API_KEY="sk-..."
.\fracture-windows-amd64.exe
```

### Linux/macOS — definir variável de ambiente
```bash
export OPENAI_API_KEY="sk-..."
./fracture
```

---

## Requisitos mínimos

| Componente | Mínimo | Recomendado |
|---|---|---|
| OS | Windows 10 64-bit / Ubuntu 20.04+ / macOS 12+ | Windows 11 / Ubuntu 22.04+ / macOS 14+ |
| RAM | 4 GB | 8 GB+ |
| CPU | Dual-core 64-bit | 4+ núcleos |
| Disco | 100 MB | 200 MB |
| Internet | Sim (para API de IA) | Conexão estável |

---

## Problemas comuns

**Porta 4000 ocupada**

O FRACTURE detecta automaticamente a próxima porta disponível (4000–4099). Se a 4000 estiver ocupada, ele sobe na 4001, 4002, etc. A URL correta é exibida no terminal na inicialização.

**Windows Defender bloqueando o executável**

Clique em "Mais informações" → "Executar assim mesmo". O binário é open source e pode ser auditado em [github.com/mkvinicius/fracture](https://github.com/mkvinicius/fracture).

**Dashboard não abre automaticamente**

Acesse manualmente: `http://localhost:4000`
