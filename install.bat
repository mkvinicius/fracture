@echo off
title FRACTURE Installer
echo.
echo  ================================
echo   FRACTURE - Instalador Windows
echo  ================================
echo.

:: Verifica Git
git --version >nul 2>&1
if errorlevel 1 (
    echo [ERRO] Git nao encontrado.
    echo Baixe em: https://git-scm.com/download/win
    pause
    exit /b 1
)

:: Verifica Go
go version >nul 2>&1
if errorlevel 1 (
    echo [INFO] Instalando Go...
    curl -L -o go-installer.msi https://go.dev/dl/go1.24.0.windows-amd64.msi
    msiexec /i go-installer.msi /quiet /norestart
    del go-installer.msi
    set PATH=%PATH%;C:\Program Files\Go\bin
)

:: Verifica Node (necessario apenas para recompilar o dashboard)
node --version >nul 2>&1
if errorlevel 1 (
    echo [INFO] Instalando Node.js...
    curl -L -o node-installer.msi https://nodejs.org/dist/v20.0.0/node-v20.0.0-x64.msi
    msiexec /i node-installer.msi /quiet /norestart
    del node-installer.msi
    set PATH=%PATH%;C:\Program Files\nodejs
)

:: Verifica pnpm
pnpm --version >nul 2>&1
if errorlevel 1 (
    echo [INFO] Instalando pnpm...
    npm install -g pnpm
)

:: Clona ou atualiza o repositorio
if exist "%USERPROFILE%\fracture" (
    echo [INFO] Atualizando FRACTURE...
    cd "%USERPROFILE%\fracture"
    git pull
) else (
    echo [INFO] Baixando FRACTURE...
    git clone https://github.com/mkvinicius/fracture.git "%USERPROFILE%\fracture"
    cd "%USERPROFILE%\fracture"
)

:: Dashboard ja compilado no repositorio (dashboard/dist commitado)
:: Recompila apenas se pnpm estiver disponivel (atualizacoes de desenvolvimento)
pnpm --version >nul 2>&1
if not errorlevel 1 (
    if exist "%USERPROFILE%\fracture\dashboard\package.json" (
        echo [INFO] Atualizando dashboard...
        cd "%USERPROFILE%\fracture\dashboard"
        pnpm install --silent
        pnpm build --silent
        cd "%USERPROFILE%\fracture"
    )
)

:: Compila binario Go
echo [INFO] Compilando FRACTURE...
cd "%USERPROFILE%\fracture"
go build -o fracture.exe .
if errorlevel 1 (
    echo [ERRO] Falha ao compilar. Verifique se o Go esta instalado corretamente.
    pause
    exit /b 1
)

:: Cria atalho na area de trabalho
echo [INFO] Criando atalho...
powershell -Command "$ws = New-Object -ComObject WScript.Shell; $s = $ws.CreateShortcut('%USERPROFILE%\Desktop\FRACTURE.lnk'); $s.TargetPath = '%USERPROFILE%\fracture\fracture.exe'; $s.WorkingDirectory = '%USERPROFILE%\fracture'; $s.IconLocation = '%USERPROFILE%\fracture\fracture.exe'; $s.Save()"

echo.
echo  ================================
echo   FRACTURE instalado com sucesso
echo  ================================
echo.
echo  Clique no icone FRACTURE na area de trabalho para iniciar.
echo  Ou execute: %USERPROFILE%\fracture\fracture.exe
echo.
pause

:: Inicia o FRACTURE
start "" "%USERPROFILE%\fracture\fracture.exe"
timeout /t 3 /nobreak >nul
start http://localhost:4000
