@echo off
title FRACTURE Installer
echo.
echo  ================================
echo   FRACTURE - Instalador Windows
echo  ================================
echo.

:: Requer privilegio de administrador
net session >nul 2>&1
if errorlevel 1 (
    echo [AVISO] Execute como Administrador para instalar Go e Node.js automaticamente.
    echo Clique com botao direito no arquivo .bat e escolha "Executar como administrador".
    echo.
    pause
    exit /b 1
)

:: -------------------------------------------------------
:: Funcao auxiliar: download via PowerShell (evita SSL curl)
:: -------------------------------------------------------

:: Verifica Git
git --version >nul 2>&1
if errorlevel 1 (
    echo [INFO] Instalando Git...
    powershell -Command "& { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://github.com/git-for-windows/git/releases/download/v2.44.0.windows.1/Git-2.44.0-64-bit.exe' -OutFile '%TEMP%\git-installer.exe' }"
    if not exist "%TEMP%\git-installer.exe" (
        echo [ERRO] Falha ao baixar Git. Instale manualmente: https://git-scm.com/download/win
        pause
        exit /b 1
    )
    "%TEMP%\git-installer.exe" /VERYSILENT /NORESTART /NOCANCEL /SP- /CLOSEAPPLICATIONS /RESTARTAPPLICATIONS /COMPONENTS="icons,ext\reg\shellhere,assoc,assoc_sh"
    set "PATH=%PATH%;C:\Program Files\Git\cmd"
    del "%TEMP%\git-installer.exe"
)

:: Verifica Go
go version >nul 2>&1
if errorlevel 1 (
    echo [INFO] Baixando Go (pode demorar alguns minutos)...
    powershell -Command "& { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://go.dev/dl/go1.24.0.windows-amd64.msi' -OutFile '%TEMP%\go-installer.msi' }"
    if not exist "%TEMP%\go-installer.msi" (
        echo [ERRO] Falha ao baixar Go. Instale manualmente: https://go.dev/dl
        pause
        exit /b 1
    )
    echo [INFO] Instalando Go...
    msiexec /i "%TEMP%\go-installer.msi" /quiet /norestart
    del "%TEMP%\go-installer.msi"
    :: Recarrega PATH do registro
    for /f "tokens=2*" %%A in ('reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v Path 2^>nul') do set "SYSPATH=%%B"
    for /f "tokens=2*" %%A in ('reg query "HKCU\Environment" /v Path 2^>nul') do set "USERPATH=%%B"
    set "PATH=%SYSPATH%;%USERPATH%;C:\Program Files\Go\bin"
)

:: Verifica Go novamente
go version >nul 2>&1
if errorlevel 1 (
    echo [ERRO] Go nao encontrado mesmo apos instalacao.
    echo Reinicie o computador e execute o instalador novamente.
    pause
    exit /b 1
)

:: Clona ou atualiza o repositorio
if exist "%USERPROFILE%\fracture\.git" (
    echo [INFO] Atualizando FRACTURE...
    cd /d "%USERPROFILE%\fracture"
    git pull
) else (
    echo [INFO] Baixando FRACTURE...
    git clone https://github.com/mkvinicius/fracture.git "%USERPROFILE%\fracture"
    if errorlevel 1 (
        echo [ERRO] Falha ao clonar repositorio.
        pause
        exit /b 1
    )
    cd /d "%USERPROFILE%\fracture"
)

:: Compila binario Go (dashboard/dist ja esta incluso no repositorio)
echo [INFO] Compilando FRACTURE (aguarde)...
cd /d "%USERPROFILE%\fracture"
go build -o fracture.exe .
if errorlevel 1 (
    echo.
    echo [ERRO] Falha na compilacao.
    echo Possivel causa: Go recem instalado precisa de nova sessao.
    echo.
    echo Solucao: feche este terminal, abra um novo como Administrador e execute:
    echo   cd %USERPROFILE%\fracture
    echo   go build -o fracture.exe .
    echo   fracture.exe
    pause
    exit /b 1
)

:: Cria atalho na area de trabalho
echo [INFO] Criando atalho na area de trabalho...
powershell -Command "$ws = New-Object -ComObject WScript.Shell; $s = $ws.CreateShortcut('%USERPROFILE%\Desktop\FRACTURE.lnk'); $s.TargetPath = '%USERPROFILE%\fracture\fracture.exe'; $s.WorkingDirectory = '%USERPROFILE%\fracture'; $s.Save()"

echo.
echo  ================================
echo   FRACTURE instalado com sucesso
echo  ================================
echo.
echo  Atalho criado na area de trabalho.
echo  Ou execute diretamente: %USERPROFILE%\fracture\fracture.exe
echo.

:: Inicia o FRACTURE
start "" "%USERPROFILE%\fracture\fracture.exe"
timeout /t 3 /nobreak >nul
start http://localhost:4000

pause
