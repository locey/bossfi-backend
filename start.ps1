Write-Host "Starting BossFi Backend..." -ForegroundColor Green
Write-Host ""

Write-Host "Generating Swagger documentation..." -ForegroundColor Yellow
try {
    go generate ./cmd/server/main.go
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Warning: Failed to generate Swagger docs. Continuing anyway..." -ForegroundColor Yellow
    } else {
        Write-Host "Swagger documentation generated successfully!" -ForegroundColor Green
    }
} catch {
    Write-Host "Warning: Could not generate Swagger docs. Continuing anyway..." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Starting server..." -ForegroundColor Green
Write-Host "Server will be available at: http://localhost:8080" -ForegroundColor Cyan
Write-Host "Swagger UI will be available at: http://localhost:8080/swagger/index.html" -ForegroundColor Cyan
Write-Host ""
Write-Host "Press Ctrl+C to stop the server" -ForegroundColor Yellow
Write-Host ""

go run cmd/server/main.go 