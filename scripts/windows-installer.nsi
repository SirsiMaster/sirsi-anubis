; Sirsi Pantheon — Windows Installer (NSIS)
; Builds SirsiPantheon-VERSION-windows-setup.exe
;
; Usage: makensis -DVERSION=0.17.2 windows-installer.nsi
; Requires: NSIS 3.x (https://nsis.sourceforge.io)

!include "MUI2.nsh"
!include "EnvVarUpdate.nsh"

; --- Metadata ---
Name "Sirsi Pantheon ${VERSION}"
OutFile "..\bin\SirsiPantheon-${VERSION}-windows-setup.exe"
InstallDir "$LOCALAPPDATA\Sirsi\Pantheon"
InstallDirRegKey HKCU "Software\Sirsi\Pantheon" "InstallDir"
RequestExecutionLevel user

; --- UI ---
!define MUI_ICON "..\cmd\sirsi-menubar\bundle\icon.ico"
!define MUI_ABORTWARNING
!define MUI_WELCOMEPAGE_TITLE "Sirsi Pantheon ${VERSION}"
!define MUI_WELCOMEPAGE_TEXT "Find and fix infrastructure waste on your machine.$\r$\n$\r$\n81 rules. Zero config. Zero telemetry.$\r$\n$\r$\nThis installer will add the sirsi CLI to your PATH."

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_LANGUAGE "English"

; --- Install ---
Section "Sirsi Pantheon" SecMain
    SetOutPath "$INSTDIR"

    ; Copy all binaries
    File "..\bin\windows\sirsi.exe"
    File "..\bin\windows\sirsi-agent.exe"
    File "..\bin\windows\sirsi-anubis.exe"
    File "..\bin\windows\sirsi-guard.exe"
    File "..\bin\windows\sirsi-maat.exe"
    File "..\bin\windows\sirsi-scarab.exe"
    File "..\bin\windows\sirsi-thoth.exe"

    ; Copy configs
    SetOutPath "$INSTDIR\configs"
    File "..\configs\default_rules.yaml"

    ; Copy license
    SetOutPath "$INSTDIR"
    File "..\LICENSE"

    ; Write uninstaller
    WriteUninstaller "$INSTDIR\uninstall.exe"

    ; Registry keys for Add/Remove Programs
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\SirsiPantheon" \
        "DisplayName" "Sirsi Pantheon ${VERSION}"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\SirsiPantheon" \
        "UninstallString" "$INSTDIR\uninstall.exe"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\SirsiPantheon" \
        "DisplayVersion" "${VERSION}"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\SirsiPantheon" \
        "Publisher" "Sirsi Technologies"
    WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\SirsiPantheon" \
        "URLInfoAbout" "https://sirsi.ai/pantheon"

    ; Add to user PATH
    ${EnvVarUpdate} $0 "PATH" "A" "HKCU" "$INSTDIR"

    ; Store install dir
    WriteRegStr HKCU "Software\Sirsi\Pantheon" "InstallDir" "$INSTDIR"
SectionEnd

; --- Uninstall ---
Section "Uninstall"
    ; Remove from PATH
    ${un.EnvVarUpdate} $0 "PATH" "R" "HKCU" "$INSTDIR"

    ; Remove files
    Delete "$INSTDIR\sirsi.exe"
    Delete "$INSTDIR\sirsi-agent.exe"
    Delete "$INSTDIR\sirsi-anubis.exe"
    Delete "$INSTDIR\sirsi-guard.exe"
    Delete "$INSTDIR\sirsi-maat.exe"
    Delete "$INSTDIR\sirsi-scarab.exe"
    Delete "$INSTDIR\sirsi-thoth.exe"
    Delete "$INSTDIR\configs\default_rules.yaml"
    Delete "$INSTDIR\LICENSE"
    Delete "$INSTDIR\uninstall.exe"
    RMDir "$INSTDIR\configs"
    RMDir "$INSTDIR"

    ; Remove registry keys
    DeleteRegKey HKCU "Software\Microsoft\Windows\CurrentVersion\Uninstall\SirsiPantheon"
    DeleteRegKey HKCU "Software\Sirsi\Pantheon"
SectionEnd
