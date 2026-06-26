@echo off
setlocal EnableDelayedExpansion

cd /d "%~dp0"

rem Node/Go often missing in PATH when launched from Explorer
set "PATH=%PATH%;%ProgramFiles%\nodejs;%ProgramFiles(x86)%\nodejs%;%APPDATA%\npm;%LOCALAPPDATA%\Programs\Go\bin;C:\Program Files\Go\bin"

echo ============================================
echo   Methodological Materials - start
echo ============================================
echo.

where go >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go not found in PATH.
    echo Install: https://go.dev/dl/
    goto :fail
)

where npm >nul 2>&1
if errorlevel 1 (
    echo [ERROR] npm not found in PATH.
    echo Install Node.js: https://nodejs.org/
    goto :fail
)

if not exist ".env" (
    if exist ".env.example" (
        echo [.env missing - copying from .env.example]
        copy /Y ".env.example" ".env" >nul
        echo Edit .env if needed ^(DB_PASSWORD, ADMIN_TOKEN^)
        echo.
    ) else (
        echo [ERROR] .env and .env.example not found
        goto :fail
    )
)

if not exist "web\node_modules" (
    echo [Installing frontend dependencies - may take a minute...]
    pushd "web"
    call npm install
    if errorlevel 1 (
        echo [ERROR] npm install failed
        popd
        goto :fail
    )
    popd
    echo.
)

echo [1/3] Starting API backend on :8080...
start "API-Backend" cmd /k call "%~dp0scripts\run-backend.bat"

echo [2/3] Waiting for API...
set RETRIES=30
:wait_api
curl.exe -s http://localhost:8080/health 2>nul | findstr /C:"ok" >nul
if not errorlevel 1 goto api_ready
set /a RETRIES-=1
if !RETRIES! leq 0 (
    echo [WARN] API did not respond in 30 sec.
    echo Check: PostgreSQL running, migrations applied, .env configured.
    echo See "API-Backend" window for errors.
    echo.
    goto start_frontend
)
timeout /t 1 /nobreak >nul 2>nul
if errorlevel 1 ping -n 2 127.0.0.1 >nul
goto wait_api

:api_ready
echo       API is ready.
echo.

:start_frontend
echo [3/3] Starting web UI on :5173...
start "Web-Frontend" cmd /k call "%~dp0scripts\run-frontend.bat"

echo       Waiting for frontend...
ping -n 5 127.0.0.1 >nul

start "" http://localhost:5173

echo.
echo ============================================
echo   Ready!
echo   Site: http://localhost:5173
echo   API:  http://localhost:8080/health
echo.
echo   To stop: close "API-Backend" and "Web-Frontend" windows
echo ============================================
echo.
pause
exit /b 0

:fail
echo.
pause
exit /b 1
