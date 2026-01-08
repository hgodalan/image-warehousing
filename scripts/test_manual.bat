@echo off
REM Manual Testing Helper Script for Windows

echo.
echo 🧪 Manual Testing Helper for Image Warehousing
echo ==============================================
echo.

REM Check if server is running
echo Checking if server is running...
curl -s http://localhost:8080/health >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ Server is running
) else (
    echo ⚠ Server is not running. Start it with: make run
    exit /b 1
)

echo.
echo Available test commands:
echo ------------------------
echo.

echo 1. Upload a test image:
echo    curl -X POST http://localhost:8080/api/v1/images/upload ^
echo      -F "image=@your_image.jpg" ^
echo      -F "title=Test Image" ^
echo      -F "artist=Your Name" ^
echo      -F "tags=[\"test\",\"sample\"]"
echo.

echo 2. Search for images:
echo    curl -X POST http://localhost:8080/api/v1/search ^
echo      -H "Content-Type: application/json" ^
echo      -d "{\"query\": \"your search query\", \"limit\": 10}"
echo.

echo 3. Check server health:
echo    curl http://localhost:8080/health
echo.

echo 4. View the index file:
echo    type data\index.md
echo.

if "%~1"=="" (
    echo Usage: %0 [image_file] [title] [artist]
    echo.
    echo Example:
    echo   %0 photo.jpg "Beach Sunset" "John Doe"
    echo.
    echo Or run automated test:
    echo   go run scripts/test_e2e.go
    exit /b 0
)

echo Running quick upload test with: %1
echo.

set IMAGE_FILE=%1
set TITLE=%~2
set ARTIST=%~3
if "%TITLE%"=="" set TITLE=Test Upload
if "%ARTIST%"=="" set ARTIST=Test User

echo Uploading %IMAGE_FILE%...

curl -X POST http://localhost:8080/api/v1/images/upload ^
    -F "image=@%IMAGE_FILE%" ^
    -F "title=%TITLE%" ^
    -F "artist=%ARTIST%" ^
    -F "tags=[\"test\",\"manual\"]"

echo.
echo ✓ Upload complete!
echo Wait 30 seconds for AI processing, then try searching.
echo.
echo Search example:
echo   curl -X POST http://localhost:8080/api/v1/search ^
echo     -H "Content-Type: application/json" ^
echo     -d "{\"query\": \"%TITLE%\", \"limit\": 10}"
