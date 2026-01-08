@echo off
REM Interactive Artwork Upload Agent - Windows

echo.
echo 🎨 Interactive Artwork Upload ^& Knowledge Management Agent
echo ============================================================
echo.
echo This script will:
echo   1. Upload images from ~/Downloads/artwork_images
echo   2. Categorize by folder (original/AI-generated)
echo   3. Enter interactive chat mode for searching
echo.

REM Check if server is running
curl -s http://localhost:8080/health >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Server is not running!
    echo.
    echo Please start the server first:
    echo   Terminal 1: make run
    echo.
    echo Then run this script again in Terminal 2.
    pause
    exit /b 1
)

echo ✅ Server detected, starting upload agent...
echo.

REM Run the Go program
go run scripts/interactive_upload.go

pause
