@echo off
setlocal EnableExtensions

set "DRY_RUN=0"
if /I "%~1"=="--dry-run" set "DRY_RUN=1"

set "SCRIPT_DIR=%~dp0"
set "SOURCE_EXE=%SCRIPT_DIR%vial-helperd.exe"

set "INSTALL_DIR=%LOCALAPPDATA%\Programs\vial-helper"
set "TARGET_EXE=%INSTALL_DIR%\vial-helperd.exe"
set "RUNNER_PS1=%INSTALL_DIR%\run-hidden.ps1"

set "TASK_NAME=Vial Helper Daemon"

echo.
echo ========================================
echo  Vial Helper - Windows Installer
echo ========================================
echo.

if not exist "%SOURCE_EXE%" (
    echo [ERROR] vial-helperd.exe not found next to this installer:
    echo %SOURCE_EXE%
    echo.
    exit /b 1
)

echo [1/6] Stopping running daemon...
if "%DRY_RUN%"=="1" (
    echo [dry-run] taskkill /F /IM vial-helperd.exe
) else (
    taskkill /F /IM vial-helperd.exe >nul 2>nul
)

echo [2/6] Creating install directory...
if not exist "%INSTALL_DIR%" (
    if "%DRY_RUN%"=="1" (
        echo [dry-run] mkdir "%INSTALL_DIR%"
    ) else (
        mkdir "%INSTALL_DIR%"
        if errorlevel 1 (
            echo [ERROR] Failed to create:
            echo %INSTALL_DIR%
            echo.
            exit /b 1
        )
    )
)

echo [3/6] Installing binary...
if "%DRY_RUN%"=="1" (
    echo [dry-run] copy /Y "%SOURCE_EXE%" "%TARGET_EXE%"
) else (
    copy /Y "%SOURCE_EXE%" "%TARGET_EXE%" >nul
    if errorlevel 1 (
        echo [ERROR] Failed to copy daemon to:
        echo %TARGET_EXE%
        echo.
        exit /b 1
    )
)

echo [4/6] Initializing config...
if "%DRY_RUN%"=="1" (
    echo [dry-run] "%TARGET_EXE%" --command init
    echo [dry-run] write "%RUNNER_PS1%"
) else (
    "%TARGET_EXE%" --command init
    if errorlevel 1 (
        echo [ERROR] Config initialization failed.
        echo.
        exit /b 1
    )

    > "%RUNNER_PS1%" (
        echo $ErrorActionPreference = 'Stop'
        echo $binary = Join-Path $env:LOCALAPPDATA 'Programs\vial-helper\vial-helperd.exe'
        echo $logDir = Join-Path $env:APPDATA 'vial-helper'
        echo $stderrLog = Join-Path $logDir 'daemon.err.log'
        echo $maxLogSize = 262144
        echo New-Item -ItemType Directory -Force -Path $logDir ^| Out-Null
        echo if ^(Test-Path $stderrLog^) { $size = ^(Get-Item $stderrLog^).Length; if ^($size -ge $maxLogSize^) { Move-Item -Force $stderrLog "$stderrLog.1" } }
        echo $existing = Get-CimInstance Win32_Process -Filter "Name = 'vial-helperd.exe'" ^| Where-Object { $_.ExecutablePath -eq $binary -and $_.CommandLine -match '--command run' }
        echo if ^($existing^) { exit 0 }
        echo Start-Process -FilePath $binary -ArgumentList '--command run' -WindowStyle Hidden -RedirectStandardError $stderrLog
    )
)

echo [5/6] Registering autostart task...
if "%DRY_RUN%"=="1" (
    echo [dry-run] schtasks /Delete /TN "%TASK_NAME%" /F
    echo [dry-run] schtasks /Create /TN "%TASK_NAME%" /SC ONLOGON /TR "powershell.exe -NoLogo -NoProfile -WindowStyle Hidden -ExecutionPolicy Bypass -File \"%RUNNER_PS1%\"" /RL LIMITED /F
) else (
    schtasks /Delete /TN "%TASK_NAME%" /F >nul 2>nul
    schtasks /Create ^
        /TN "%TASK_NAME%" ^
        /SC ONLOGON ^
        /TR "powershell.exe -NoLogo -NoProfile -WindowStyle Hidden -ExecutionPolicy Bypass -File \"%RUNNER_PS1%\"" ^
        /RL LIMITED ^
        /F >nul

    if errorlevel 1 (
        echo [ERROR] Failed to create Scheduled Task:
        echo %TASK_NAME%
        echo.
        exit /b 1
    )
)

echo [6/6] Starting daemon...
if "%DRY_RUN%"=="1" (
    echo [dry-run] schtasks /Run /TN "%TASK_NAME%"
) else (
    schtasks /Run /TN "%TASK_NAME%" >nul
    if errorlevel 1 (
        echo [WARNING] Task was created, but could not be started automatically.
        echo You can run it manually from Task Scheduler.
        echo.
    )
)

echo.
echo ========================================
echo  Installation complete
echo ========================================
echo.
echo Binary:
echo   %TARGET_EXE%
echo.
echo Hidden runner:
echo   %RUNNER_PS1%
echo.
echo Autostart task:
echo   %TASK_NAME%
echo.
echo Config and JSON files:
echo   %APPDATA%\vial-helper
echo.
echo Logs:
echo   %APPDATA%\vial-helper\daemon.err.log
echo.
echo Useful commands:
echo   "%TARGET_EXE%" --command doctor
echo   "%TARGET_EXE%" --command status
echo.
if not "%DRY_RUN%"=="1" pause
