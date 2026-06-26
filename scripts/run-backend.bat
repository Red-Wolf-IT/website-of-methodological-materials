@echo off
cd /d "%~dp0.."
go run ./cmd/app
if errorlevel 1 (
    echo.
    echo [ERROR] Backend failed to start. Check PostgreSQL and .env
    pause
)