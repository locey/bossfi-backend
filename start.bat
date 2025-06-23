@echo off
echo Starting BossFi Backend Server...
echo.

echo Generating Swagger documentation...
swag init -g cmd/server/main.go --output ./docs
if %ERRORLEVEL% NEQ 0 (
    echo Warning: Failed to generate Swagger docs
    echo Please ensure 'swag' is installed and in PATH
    echo.
)

echo Building application...
go build -o main.exe ./cmd/server
if %ERRORLEVEL% NEQ 0 (
    echo Error: Failed to build application
    pause
    exit /b 1
)

echo Starting server...
echo.
.\main.exe

pause 