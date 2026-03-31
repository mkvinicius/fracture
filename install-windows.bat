@echo off
setlocal EnableDelayedExpansion
chcp 65001 >nul 2>&1
title FRACTURE Installer

:: ============================================================
:: AUTO-ELEVACAO: reinicia como Administrador se nao for admin
:: ============================================================
net session >nul 2>&1
if errorlevel 1 (
    echo Solicitando permissao de Administrador...
    powershell -Command "Start-Process -FilePath '%~f0' -Verb RunAs"
    exit /b
)

echo.
echo  ================================================
echo    FRACTURE - Instalador Automatico Windows
echo  ================================================
echo.

:: ============================================================
:: PASSO 1 - INSTALAR GO
:: ============================================================
go version >nul 2>&1
if not errorlevel 1 (
    echo [OK] Go ja instalado.
    goto :check_git
)

echo [1/4] Instalando Go...

:: Tenta winget primeiro (Windows 10/11 nativo, sem SSL issue)
winget --version >nul 2>&1
if not errorlevel 1 (
    winget install GoLang.Go --silent --accept-package-agreements --accept-source-agreements
    goto :reload_go_path
)

:: Fallback: bitsadmin (built-in, ignora SSL corporativo)
echo     Usando bitsadmin para download...
bitsadmin /transfer "FRACTUREGoDownload" /priority FOREGROUND ^
    "https://go.dev/dl/go1.24.0.windows-amd64.msi" ^
    "%TEMP%\go-fracture.msi" >nul 2>&1

if not exist "%TEMP%\go-fracture.msi" (
    :: Ultimo fallback: PowerShell com SSL bypass
    powershell -Command ^
        "[Net.ServicePointManager]::ServerCertificateValidationCallback={$true}; ^
         (New-Object Net.WebClient).DownloadFile('https://go.dev/dl/go1.24.0.windows-amd64.msi','%TEMP%\go-fracture.msi')"
)

if not exist "%TEMP%\go-fracture.msi" (
    echo.
    echo [ERRO] Nao foi possivel baixar o Go automaticamente.
    echo.
    echo  Solucao manual (2 minutos):
    echo  1. Abra o navegador
    echo  2. Acesse: https://go.dev/dl
    echo  3. Baixe: go1.24.0.windows-amd64.msi
    echo  4. Instale dando duplo clique
    echo  5. Execute este instalador novamente
    echo.
    pause
    exit /b 1
)

msiexec /i "%TEMP%\go-fracture.msi" /quiet /norestart
del "%TEMP%\go-fracture.msi" >nul 2>&1

:reload_go_path
:: Recarrega PATH do registro para este processo
for /f "skip=2 tokens=3*" %%A in ('reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v Path 2^>nul') do set "SYSPATH=%%A %%B"
set "PATH=%SYSPATH%;C:\Program Files\Go\bin"

go version >nul 2>&1
if errorlevel 1 (
    set "PATH=%PATH%;C:\Program Files\Go\bin"
    go version >nul 2>&1
    if errorlevel 1 (
        echo [AVISO] Go instalado mas requer reinicio do terminal.
        echo Fechando e reabrindo automaticamente...
        :: Cria script temporario para continuar apos nova sessao
        echo @echo off > "%TEMP%\fracture-continue.bat"
        echo set PATH=%%PATH%%;C:\Program Files\Go\bin >> "%TEMP%\fracture-continue.bat"
        echo cd /d "%USERPROFILE%\fracture" >> "%TEMP%\fracture-continue.bat"
        echo go build -o fracture.exe . >> "%TEMP%\fracture-continue.bat"
        echo if errorlevel 1 pause >> "%TEMP%\fracture-continue.bat"
        echo start "" fracture.exe >> "%TEMP%\fracture-continue.bat"
        echo timeout /t 3 /nobreak ^>nul >> "%TEMP%\fracture-continue.bat"
        echo start http://localhost:4000 >> "%TEMP%\fracture-continue.bat"
        echo del "%%~f0" >> "%TEMP%\fracture-continue.bat"
        start cmd /k "%TEMP%\fracture-continue.bat"
        exit /b
    )
)
echo [OK] Go instalado.

:: ============================================================
:: PASSO 2 - INSTALAR GIT
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

bitsadmin /transfer "FRACTUREGitDownload" /priority FOREGROUND ^
    "https://github.com/git-for-windows/git/releases/download/v2.44.0.windows.1/Git-2.44.0-64-bit.exe" ^
    "%TEMP%\git-fracture.exe" >nul 2>&1

if exist "%TEMP%\git-fracture.exe" (
    "%TEMP%\git-fracture.exe" /VERYSILENT /NORESTART /NOCANCEL /SP- /CLOSEAPPLICATIONS
    set "PATH=%PATH%;C:\Program Files\Git\cmd"
    del "%TEMP%\git-fracture.exe" >nul 2>&1
) else (
    echo [ERRO] Nao foi possivel instalar Git automaticamente.
    echo Baixe em: https://git-scm.com/download/win
    pause
    exit /b 1
)
echo [OK] Git instalado.

:: ============================================================
:: PASSO 3 - CLONAR OU ATUALIZAR REPOSITORIO
:: ============================================================
:clone_repo
echo [3/4] Baixando FRACTURE...
if exist "%USERPROFILE%\fracture\.git" (
    cd /d "%USERPROFILE%\fracture"
    git pull origin master
) else (
    git clone https://github.com/mkvinicius/fracture.git "%USERPROFILE%\fracture"
    if errorlevel 1 (
        echo [ERRO] Falha ao clonar repositorio.
        pause
        exit /b 1
    )
    cd /d "%USERPROFILE%\fracture"
)
echo [OK] Codigo atualizado.

:: ============================================================
:: PASSO 4 - CONFIGURAR CGO (necessario para SQLite)
:: ============================================================
:: go-sqlite3 requer CGO e um compilador C.
:: Git for Windows ja inclui GCC — detecta automaticamente.
set "CGO_ENABLED=1"
set "CC="

if exist "C:\Program Files\Git\mingw64\bin\gcc.exe" (
    set "CC=C:\PROGRA~1\Git\mingw64\bin\gcc.exe"
    goto :build
)
if exist "C:\Program Files\Git\usr\bin\gcc.exe" (
    set "CC=C:\PROGRA~1\Git\usr\bin\gcc.exe"
    goto :build
)
:: Tenta MinGW standalone
if exist "C:\mingw64\bin\gcc.exe" (
    set "CC=C:\mingw64\bin\gcc.exe"
    set "PATH=%PATH%;C:\mingw64\bin"
    goto :build
)
if exist "C:\TDM-GCC-64\bin\gcc.exe" (
    set "CC=C:\TDM-GCC-64\bin\gcc.exe"
    set "PATH=%PATH%;C:\TDM-GCC-64\bin"
    goto :build
)
:: Tenta gcc no PATH
where gcc >nul 2>&1
if not errorlevel 1 goto :build

:: Instala TDM-GCC via winget se disponivel
winget --version >nul 2>&1
if not errorlevel 1 (
    echo [INFO] Instalando compilador C (TDM-GCC)...
    winget install tdm-gcc.tdm-gcc --silent --accept-package-agreements --accept-source-agreements
    set "PATH=%PATH%;C:\TDM-GCC-64\bin"
)

:build
:: ============================================================
:: PASSO 5 - COMPILAR
:: ============================================================
echo [4/4] Compilando FRACTURE...
cd /d "%USERPROFILE%\fracture"
if defined CC (
    "%CC%" --version >nul 2>&1 && (
        go build -o fracture.exe .
    ) || (
        set "CC="
        go build -o fracture.exe .
    )
) else (
    go build -o fracture.exe .
)
if errorlevel 1 (
    echo.
    echo [ERRO] Falha na compilacao.
    echo.
    echo  O FRACTURE usa SQLite e precisa de um compilador C.
    echo  Solucao mais rapida:
    echo    1. Abra o PowerShell como Admin e execute:
    echo       winget install tdm-gcc.tdm-gcc
    echo    2. Feche e abra um novo CMD como Admin
    echo    3. Execute este instalador novamente
    echo.
    pause
    exit /b 1
)
echo [OK] Compilado com sucesso.

:: ============================================================
:: ATALHO NA AREA DE TRABALHO
:: ============================================================
powershell -Command ^
    "$ws=New-Object -ComObject WScript.Shell; ^
     $s=$ws.CreateShortcut('%USERPROFILE%\Desktop\FRACTURE.lnk'); ^
     $s.TargetPath='%USERPROFILE%\fracture\fracture.exe'; ^
     $s.WorkingDirectory='%USERPROFILE%\fracture'; ^
     $s.Save()"

:: ============================================================
:: INICIAR
:: ============================================================
echo.
echo  ================================================
echo    FRACTURE instalado com sucesso!
echo  ================================================
echo.
echo  Iniciando...
echo.

start "" "%USERPROFILE%\fracture\fracture.exe"
timeout /t 3 /nobreak >nul
start http://localhost:4000

pause
