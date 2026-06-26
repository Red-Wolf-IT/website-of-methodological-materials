@echo off
cd /d "%~dp0..\web"
call npm run dev
if errorlevel 1 (
    echo.
    echo [ERROR] Frontend failed to start.
    pause
)