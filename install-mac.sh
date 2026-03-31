#!/bin/bash
set -e

clear
echo ""
echo " ================================================"
echo "  FRACTURE - Instalador Automatico macOS"
echo " ================================================"
echo ""
echo " Este instalador vai:"
echo "   1. Instalar Homebrew (se necessario)"
echo "   2. Instalar Go (se necessario)"
echo "   3. Baixar/atualizar o FRACTURE"
echo "   4. Compilar e iniciar automaticamente"
echo ""
echo " Aguarde cada etapa concluir..."
echo ""

# ============================================================
# PASSO 1 - HOMEBREW
# ============================================================
if ! command -v brew &>/dev/null; then
    echo "[1/4] Instalando Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    # Adiciona brew ao PATH (Apple Silicon e Intel)
    if [ -f "/opt/homebrew/bin/brew" ]; then
        eval "$(/opt/homebrew/bin/brew shellenv)"
    elif [ -f "/usr/local/bin/brew" ]; then
        eval "$(/usr/local/bin/brew shellenv)"
    fi
    echo "[OK] Homebrew instalado."
else
    echo "[OK] Homebrew ja instalado."
fi

# ============================================================
# PASSO 2 - GO
# ============================================================
if ! command -v go &>/dev/null; then
    echo "[2/4] Instalando Go..."
    brew install go
    # Garante que go esta no PATH desta sessao
    export PATH="$(brew --prefix go)/bin:$PATH"
    echo "[OK] Go instalado."
else
    echo "[OK] Go ja instalado: $(go version)"
fi

# ============================================================
# PASSO 3 - CLONAR OU ATUALIZAR
# ============================================================
echo "[3/4] Baixando FRACTURE..."
if [ -d "$HOME/fracture/.git" ]; then
    cd "$HOME/fracture"
    git pull origin master
    echo "[OK] Codigo atualizado."
else
    git clone https://github.com/mkvinicius/fracture.git "$HOME/fracture"
    cd "$HOME/fracture"
    echo "[OK] FRACTURE baixado."
fi

# ============================================================
# PASSO 4 - DEPENDENCIAS + COMPILAR
# ============================================================
echo "[4/4] Compilando FRACTURE..."
cd "$HOME/fracture"

# go mod tidy baixa dependencias (modernc.org/sqlite - puro Go, sem CGO)
go mod tidy

go build -o fracture .
if [ $? -ne 0 ]; then
    echo ""
    echo "[ERRO] Falha ao compilar."
    echo ""
    echo " Tente manualmente:"
    echo "   cd ~/fracture && go mod tidy && go build -o fracture ."
    echo ""
    exit 1
fi
echo "[OK] Compilado com sucesso."

# ============================================================
# ATALHO NA AREA DE TRABALHO
# ============================================================
cat > "$HOME/Desktop/FRACTURE.command" << 'SHORTCUT'
#!/bin/bash
cd "$HOME/fracture"
./fracture &
sleep 2
open http://localhost:4000
SHORTCUT
chmod +x "$HOME/Desktop/FRACTURE.command"
echo "[OK] Atalho criado na area de trabalho."

# ============================================================
# INICIAR
# ============================================================
echo ""
echo " ================================================"
echo "  FRACTURE instalado com sucesso!"
echo " ================================================"
echo ""
echo " Iniciando FRACTURE..."
echo ""

cd "$HOME/fracture"
./fracture &
sleep 3
open http://localhost:4000

echo " Abra o navegador em: http://localhost:4000"
echo " Atalho salvo em: ~/Desktop/FRACTURE.command"
echo ""
