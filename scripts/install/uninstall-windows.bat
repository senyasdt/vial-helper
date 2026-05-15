@echo off
setlocal EnableExtensions

set "DRY_RUN=0"
if /I "%~1"=="--dry-run" set "DRY_RUN=1"

set "INSTALL_DIR=%LOCALAPPDATA%\Programs\vial-helper"
set "TARGET_EXE=%INSTALL_DIR%\vial-helperd.exe"
set "RUNNER_PS1=%INSTALL_DIR%\run-hidden.ps1"
set "TASK_NAME=Vial Helper Daemon"

echo.
echo ========================================
echo  Vial Helper - Windows Uninstaller
echo ========================================
echo.

echo [1/4] Stopping running daemon...
if "%DRY_RUN%"=="1" (
    echo [dry-run] taskkill /F /IM vial-helperd.exe
) else (
    taskkill /F /IM vial-helperd.exe >nul 2>nul
)

echo [2/4] Removing autostart task...
if "%DRY_RUN%"=="1" (
    echo [dry-run] schtasks /Delete /TN "%TASK_NAME%" /F
) else (
    schtasks /Delete /TN "%TASK_NAME%" /F >nul 2>nul
)

echo [3/4] Removing binary...
if exist "%TARGET_EXE%" (
    if "%DRY_RUN%"=="1" (
        echo [dry-run] del "%TARGET_EXE%"
    ) else (
        del /F /Q "%TARGET_EXE%"
    )
)

if exist "%RUNNER_PS1%" (
    if "%DRY_RUN%"=="1" (
        echo [dry-run] del "%RUNNER_PS1%"
    ) else (
        del /F /Q "%RUNNER_PS1%"
    )
)

echo [4/4] Finished.
echo.
echo Removed:
echo   %TARGET_EXE%
echo   %TASK_NAME%
echo.
echo Config and JSON files were left intact:
echo   %APPDATA%\vial-helper
echo.
if not "%DRY_RUN%"=="1" pause
