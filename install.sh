#!/bin/bash
clear
echo ""
echo " ================================"
echo "  FRACTURE - Instalador macOS"
echo " ================================"
echo ""

# Instala Homebrew se necessario
if ! command -v brew &>/dev/null; then
    echo "[INFO] Instalando Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
fi

# Instala Go se necessario
if ! command -v go &>/dev/null; then
    echo "[INFO] Instalando Go..."
    brew install go
fi

# Clona ou atualiza
if [ -d "$HOME/fracture" ]; then
    echo "[INFO] Atualizando FRACTURE..."
    cd "$HOME/fracture" && git pull
else
    echo "[INFO] Baixando FRACTURE..."
    git clone https://github.com/mkvinicius/fracture.git "$HOME/fracture"
    cd "$HOME/fracture"
fi

# Dashboard ja compilado no repositorio (dashboard/dist commitado)
# Recompila apenas se pnpm estiver disponivel (atualizacoes de desenvolvimento)
if command -v pnpm &>/dev/null; then
    echo "[INFO] Atualizando dashboard..."
    cd "$HOME/fracture/dashboard" && pnpm install --silent && pnpm build --silent
    cd "$HOME/fracture"
elif command -v npm &>/dev/null; then
    if ! command -v pnpm &>/dev/null; then
        npm install -g pnpm --silent
    fi
fi

# Compila binario Go
echo "[INFO] Compilando FRACTURE..."
cd "$HOME/fracture"
go build -o fracture .
if [ $? -ne 0 ]; then
    echo "[ERRO] Falha ao compilar. Verifique se o Go esta instalado corretamente."
    exit 1
fi

# Cria atalho na area de trabalho
echo "[INFO] Criando atalho..."
cat > "$HOME/Desktop/FRACTURE.command" << 'SHORTCUT'
#!/bin/bash
cd "$HOME/fracture"
./fracture &
sleep 2
open http://localhost:4000
SHORTCUT
chmod +x "$HOME/Desktop/FRACTURE.command"

echo ""
echo " ================================"
echo "  FRACTURE instalado com sucesso!"
echo " ================================"
echo ""
echo " Clique em FRACTURE.command na area de trabalho para iniciar."
echo " Ou execute: cd ~/fracture && ./fracture"
echo ""

# Inicia agora
./fracture &
sleep 2
open http://localhost:4000
