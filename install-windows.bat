@echo off
setlocal EnableDelayedExpansion
chcp 65001 >nul 2>&1
title FRACTURE Installer

:: ============================================================
:: AUTO-ELEVACAO: reinicia como Administrador automaticamente
:: ============================================================
net session >nul 2>&1
if errorlevel 1 (
    powershell -Command "Start-Process -FilePath '%~f0' -Verb RunAs"
    exit /b
)

cls
echo.
echo  ================================================
echo    FRACTURE - Instalador Automatico Windows
echo  ================================================
echo.
echo  Este instalador vai:
echo    1. Instalar Go (se necessario)
echo    2. Instalar Git (se necessario)
echo    3. Baixar/atualizar o FRACTURE
echo    4. Compilar e iniciar automaticamente
echo.
echo  Aguarde cada etapa concluir...
echo.

:: ============================================================
:: PASSO 1 - GO
:: ============================================================
go version >nul 2>&1
if not errorlevel 1 (
    echo [OK] Go ja instalado.
    goto :check_git
)

echo [1/4] Instalando Go...

:: Tenta winget (Windows 10/11)
winget --version >nul 2>&1
if not errorlevel 1 (
    winget install GoLang.Go --silent --accept-package-agreements --accept-source-agreements
    :: Recarrega PATH
    for /f "skip=2 tokens=*" %%A in ('reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v Path 2^>nul') do set "NEWPATH=%%A"
    set "PATH=!NEWPATH:~10!;C:\Program Files\Go\bin"
    go version >nul 2>&1
    if not errorlevel 1 goto :check_git
)

:: Fallback: bitsadmin
echo     Baixando instalador do Go...
bitsadmin /transfer "GoDownload" /priority FOREGROUND ^
    "https://go.dev/dl/go1.24.0.windows-amd64.msi" ^
    "%TEMP%\go-fracture.msi" >nul 2>&1

if not exist "%TEMP%\go-fracture.msi" (
    powershell -Command "[Net.ServicePointManager]::ServerCertificateValidationCallback={$true}; (New-Object Net.WebClient).DownloadFile('https://go.dev/dl/go1.24.0.windows-amd64.msi','%TEMP%\go-fracture.msi')"
)

if not exist "%TEMP%\go-fracture.msi" (
    echo.
    echo [ERRO] Nao foi possivel baixar o Go automaticamente.
    echo.
    echo  Instale manualmente (2 minutos):
    echo    1. Abra o navegador
    echo    2. Acesse: https://go.dev/dl
    echo    3. Baixe go1.24.0.windows-amd64.msi
    echo    4. Instale e execute este arquivo novamente
    echo.
    pause
    exit /b 1
)

msiexec /i "%TEMP%\go-fracture.msi" /quiet /norestart
del "%TEMP%\go-fracture.msi" >nul 2>&1
set "PATH=%PATH%;C:\Program Files\Go\bin"

go version >nul 2>&1
if errorlevel 1 (
    echo [AVISO] Go instalado. Abrindo nova janela para continuar...
    start cmd /k "set PATH=%PATH%;C:\Program Files\Go\bin && cd /d %~dp0 && %~f0"
    exit /b
)

echo [OK] Go instalado.

:: ============================================================
:: PASSO 2 - GIT
:: ============================================================
:check_git
git --version >nul 2>&1
if not errorlevel 1 (
    echo [OK] Git ja instalado.
    goto :clone_repo
)

echo [2/4] Instalando Git...
winget --version >nul 2>&1
if not errorlevel 1 (
    winget install Git.Git --silent --accept-package-agreements --accept-source-agreements
    set "PATH=%PATH%;C:\Program Files\Git\cmd"
    goto :clone_repo
)

bitsadmin /transfer "GitDownload" /priority FOREGROUND ^
    "https://github.com/git-for-windows/git/releases/download/v2.44.0.windows.1/Git-2.44.0-64-bit.exe" ^
    "%TEMP%\git-fracture.exe" >nul 2>&1

if exist "%TEMP%\git-fracture.exe" (
    "%TEMP%\git-fracture.exe" /VERYSILENT /NORESTART /NOCANCEL /SP- /CLOSEAPPLICATIONS
    set "PATH=%PATH%;C:\Program Files\Git\cmd"
    del "%TEMP%\git-fracture.exe" >nul 2>&1
) else (
    echo [ERRO] Nao foi possivel instalar o Git.
    echo Baixe em: https://git-scm.com/download/win e execute novamente.
    pause
    exit /b 1
)
echo [OK] Git instalado.

:: ============================================================
:: PASSO 3 - CLONAR OU ATUALIZAR
:: ============================================================
:clone_repo
echo [3/4] Baixando FRACTURE...
if exist "%USERPROFILE%\fracture\.git" (
    cd /d "%USERPROFILE%\fracture"
    git pull origin master
) else (
    git clone https://github.com/mkvinicius/fracture.git "%USERPROFILE%\fracture"
    if errorlevel 1 (
        echo [ERRO] Falha ao baixar o FRACTURE. Verifique sua conexao com a internet.
        pause
        exit /b 1
    )
)
cd /d "%USERPROFILE%\fracture"
echo [OK] Codigo atualizado.

:: ============================================================
:: PASSO 4 - DEPENDENCIAS + COMPILAR
:: ============================================================
echo [4/4] Compilando FRACTURE...
cd /d "%USERPROFILE%\fracture"

:: go mod tidy baixa dependencias (incluindo modernc.org/sqlite - puro Go, sem GCC)
go mod tidy
if errorlevel 1 (
    echo [AVISO] go mod tidy com aviso, tentando compilar mesmo assim...
)

go build -o fracture.exe .
if errorlevel 1 (
    echo.
    echo [ERRO] Falha na compilacao.
    echo.
    echo  Tente manualmente:
    echo    cd %USERPROFILE%\fracture
    echo    go mod tidy
    echo    go build -o fracture.exe .
    echo.
    pause
    exit /b 1
)
echo [OK] Compilado com sucesso.

:: ============================================================
:: ATALHO NA AREA DE TRABALHO
:: ============================================================
powershell -Command "$ws=New-Object -ComObject WScript.Shell; $s=$ws.CreateShortcut('%USERPROFILE%\Desktop\FRACTURE.lnk'); $s.TargetPath='%USERPROFILE%\fracture\fracture.exe'; $s.WorkingDirectory='%USERPROFILE%\fracture'; $s.Save()"
echo [OK] Atalho criado na area de trabalho.

:: ============================================================
:: INICIAR
:: ============================================================
echo.
echo  ================================================
echo    FRACTURE instalado com sucesso!
echo  ================================================
echo.
echo  Iniciando FRACTURE...
echo.

start "" "%USERPROFILE%\fracture\fracture.exe"
timeout /t 3 /nobreak >nul
start http://localhost:4000

echo  Abra o navegador em: http://localhost:4000
echo  Atalho salvo na area de trabalho.
echo.
pause
