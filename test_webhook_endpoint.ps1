# Quick Webhook Endpoint Test

$serverIP = "15.134.38.118"
$webhookUrl = "http://${serverIP}:8080/payments/webhook"

Write-Host "Testing webhook endpoint..." -ForegroundColor Cyan
Write-Host "URL: $webhookUrl" -ForegroundColor White
Write-Host ""

# Try with minimal payload
Write-Host "Attempt 1: Minimal test..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri $webhookUrl `
        -Method Post `
        -ContentType "application/json" `
        -Headers @{"x-signature" = "test"} `
        -Body '{}' `
        -UseBasicParsing `
        -TimeoutSec 5
    
    Write-Host "[OK] Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Body: $($response.Content)" -ForegroundColor Gray
} catch {
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "[Response] Status: $statusCode" -ForegroundColor Yellow
        
        if ($statusCode -eq 400) {
            Write-Host "[SUCCESS] Webhook endpoint is working!" -ForegroundColor Green
            Write-Host "400 Bad Request is expected for test payload" -ForegroundColor Gray
        } elseif ($statusCode -eq 404) {
            Write-Host "[ERROR] Webhook endpoint not found (404)" -ForegroundColor Red
            Write-Host "The route /payments/webhook might not be registered" -ForegroundColor Yellow
        } else {
            Write-Host "[INFO] Unexpected status: $statusCode" -ForegroundColor Yellow
        }
    } else {
        Write-Host "[ERROR] Connection failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "Testing other payment endpoints..." -ForegroundColor Cyan

# Test base payments endpoint
Write-Host "Testing GET /payments..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://${serverIP}:8080/payments" -Method Get -UseBasicParsing -TimeoutSec 5
    Write-Host "[OK] GET /payments works: $($response.StatusCode)" -ForegroundColor Green
} catch {
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "[Response] GET /payments: $statusCode" -ForegroundColor Yellow
    } else {
        Write-Host "[FAIL] GET /payments failed" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "DIAGNOSIS" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "If webhook endpoint returns 404:" -ForegroundColor Yellow
Write-Host "  - Check if server code is deployed correctly" -ForegroundColor White
Write-Host "  - Verify route is registered in main.go" -ForegroundColor White
Write-Host "  - Check deployment logs for errors" -ForegroundColor White
Write-Host ""
Write-Host "If connection closes unexpectedly:" -ForegroundColor Yellow
Write-Host "  - Server might be crashing on webhook request" -ForegroundColor White
Write-Host "  - Check server logs: docker-compose logs app" -ForegroundColor White
Write-Host "  - Verify webhook handler code" -ForegroundColor White
Write-Host ""
Write-Host "If webhook returns 400 Bad Request:" -ForegroundColor Yellow
Write-Host "  - SUCCESS! Webhook endpoint is working correctly" -ForegroundColor Green
Write-Host "  - Configure this URL in PayOS Dashboard:" -ForegroundColor White
Write-Host "    $webhookUrl" -ForegroundColor Cyan
Write-Host ""
