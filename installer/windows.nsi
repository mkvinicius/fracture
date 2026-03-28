!include "MUI2.nsh"
!include "FileFunc.nsh"

; Configurações gerais
Name "FRACTURE"
OutFile "FRACTURE-Setup.exe"
InstallDir "$PROGRAMFILES64\FRACTURE"
InstallDirRegKey HKLM "Software\FRACTURE" "InstallDir"
RequestExecutionLevel admin
Unicode True

; Variáveis para API keys
Var OpenAIKey
Var AnthropicKey
Var GoogleKey

; Interface visual
!define MUI_ABORTWARNING
!define MUI_ICON "dashboard\public\favicon.ico"
!define MUI_UNICON "dashboard\public\favicon.ico"
!define MUI_WELCOMEPAGE_TITLE "Bem-vindo ao FRACTURE"
!define MUI_WELCOMEPAGE_TEXT "FRACTURE simula como as regras do seu mercado vão quebrar — antes que aconteça.$\n$\nEste assistente vai instalar o FRACTURE no seu computador."
!define MUI_FINISHPAGE_RUN "$INSTDIR\fracture.exe"
!define MUI_FINISHPAGE_RUN_TEXT "Iniciar o FRACTURE agora"
!define MUI_FINISHPAGE_SHOWREADME "$INSTDIR\INSTALL.md"
!define MUI_FINISHPAGE_SHOWREADME_TEXT "Ver guia de instalação"

; Páginas do instalador
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "LICENSE"
!insertmacro MUI_PAGE_DIRECTORY
Page custom APIKeysPage APIKeysLeave
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

; Páginas do desinstalador
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

; Idioma
!insertmacro MUI_LANGUAGE "PortugueseBR"

; Página customizada de API Keys
Function APIKeysPage
  nsDialogs::Create 1018
  Pop $0

  ${NSD_CreateLabel} 0 0 100% 24u "Configure suas chaves de API de IA"
  Pop $0

  ${NSD_CreateLabel} 0 36u 100% 12u "OpenAI API Key (recomendado — GPT-4o mini)"
  Pop $0
  ${NSD_CreateText} 0 50u 100% 14u ""
  Pop $OpenAIKey

  ${NSD_CreateLabel} 0 76u 100% 12u "Anthropic API Key (Claude — opcional)"
  Pop $0
  ${NSD_CreateText} 0 90u 100% 14u ""
  Pop $AnthropicKey

  ${NSD_CreateLabel} 0 116u 100% 12u "Google API Key (Gemini — opcional)"
  Pop $0
  ${NSD_CreateText} 0 130u 100% 14u ""
  Pop $GoogleKey

  ${NSD_CreateLabel} 0 156u 100% 24u "Você pode configurar ou alterar as chaves depois em Settings no dashboard."
  Pop $0

  nsDialogs::Show
FunctionEnd

Function APIKeysLeave
  ${NSD_GetText} $OpenAIKey $0
  ${NSD_GetText} $AnthropicKey $1
  ${NSD_GetText} $GoogleKey $2

  StrCpy $OpenAIKey $0
  StrCpy $AnthropicKey $1
  StrCpy $GoogleKey $2
FunctionEnd

; Seção principal de instalação
Section "FRACTURE" SecMain
  SetOutPath "$INSTDIR"

  ; Copia os arquivos
  File "fracture-windows-amd64.exe"
  Rename "$INSTDIR\fracture-windows-amd64.exe" "$INSTDIR\fracture.exe"
  File "README.md"
  File "LICENSE"
  File "INSTALL.md"

  ; Salva API keys no arquivo de configuração
  FileOpen $0 "$INSTDIR\fracture.env" w
  FileWrite $0 "OPENAI_API_KEY=$OpenAIKey$\n"
  FileWrite $0 "ANTHROPIC_API_KEY=$AnthropicKey$\n"
  FileWrite $0 "GOOGLE_API_KEY=$GoogleKey$\n"
  FileClose $0

  ; Cria script launcher que carrega o .env e abre o navegador
  FileOpen $0 "$INSTDIR\launch.bat" w
  FileWrite $0 "@echo off$\n"
  FileWrite $0 "for /f $\"tokens=1,2 delims==$\" %%a in ($\"$INSTDIR\fracture.env$\") do set %%a=%%b$\n"
  FileWrite $0 "start $\"$\" $\"$INSTDIR\fracture.exe$\"$\n"
  FileWrite $0 "timeout /t 2 /nobreak >nul$\n"
  FileWrite $0 "start http://localhost:3000$\n"
  FileClose $0

  ; Cria ícone na área de trabalho
  CreateShortcut "$DESKTOP\FRACTURE.lnk" "$INSTDIR\launch.bat" "" "$INSTDIR\fracture.exe" 0

  ; Cria entrada no Menu Iniciar
  CreateDirectory "$SMPROGRAMS\FRACTURE"
  CreateShortcut "$SMPROGRAMS\FRACTURE\FRACTURE.lnk" "$INSTDIR\launch.bat" "" "$INSTDIR\fracture.exe" 0
  CreateShortcut "$SMPROGRAMS\FRACTURE\Desinstalar FRACTURE.lnk" "$INSTDIR\Uninstall.exe"

  ; Registra no Painel de Controle (Adicionar/Remover Programas)
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\FRACTURE" \
    "DisplayName" "FRACTURE — Market Disruption Simulation Engine"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\FRACTURE" \
    "UninstallString" "$INSTDIR\Uninstall.exe"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\FRACTURE" \
    "DisplayIcon" "$INSTDIR\fracture.exe"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\FRACTURE" \
    "Publisher" "FRACTURE"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\FRACTURE" \
    "DisplayVersion" "2.5.0"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\FRACTURE" \
    "URLInfoAbout" "https://github.com/mkvinicius/fracture"
  ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
  IntFmt $0 "0x%08X" $0
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\FRACTURE" \
    "EstimatedSize" "$0"

  ; Salva diretório de instalação
  WriteRegStr HKLM "Software\FRACTURE" "InstallDir" "$INSTDIR"

  ; Cria desinstalador
  WriteUninstaller "$INSTDIR\Uninstall.exe"
SectionEnd

; Seção de desinstalação
Section "Uninstall"
  ; Remove arquivos
  Delete "$INSTDIR\fracture.exe"
  Delete "$INSTDIR\fracture.env"
  Delete "$INSTDIR\launch.bat"
  Delete "$INSTDIR\README.md"
  Delete "$INSTDIR\LICENSE"
  Delete "$INSTDIR\INSTALL.md"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir "$INSTDIR"

  ; Remove atalhos
  Delete "$DESKTOP\FRACTURE.lnk"
  Delete "$SMPROGRAMS\FRACTURE\FRACTURE.lnk"
  Delete "$SMPROGRAMS\FRACTURE\Desinstalar FRACTURE.lnk"
  RMDir "$SMPROGRAMS\FRACTURE"

  ; Remove registro
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\FRACTURE"
  DeleteRegKey HKLM "Software\FRACTURE"
SectionEnd
