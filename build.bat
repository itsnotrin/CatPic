@echo off
echo Building for Linux (Debian x64)...

REM Set environment variables for cross-compilation
set GOOS=linux
set GOARCH=amd64

REM Build the binary
go build -o catpic-linux-amd64 main.go

if %errorlevel% neq 0 (
    echo Build failed!
    exit /b %errorlevel%
)

echo Build complete: catpic-linux-amd64
pause
