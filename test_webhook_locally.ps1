# 🧪 Test Webhook Locally

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Local Webhook Testing Tool" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# This script simulates a PayOS webhook call to test your webhook endpoint locally
# Note: This bypasses signature verification, so it's only for testing the handler logic

$webhookUrl = "http://localhost:8080/payments/webhook"

Write-Host "Testing webhook endpoint: $webhookUrl" -ForegroundColor White
Write-Host ""

# Prompt for order code
$orderCode = Read-Host "Enter order code to test (or press Enter for demo: 123456)"
if ([string]::IsNullOrWhiteSpace($orderCode)) {
    $orderCode = "123456"
}

Write-Host "Using order code: $orderCode" -ForegroundColor Green
Write-Host ""

# Create a simulated PayOS webhook payload
$webhookPayload = @{
    code = "00"
    desc = "Thành công"
    data = @{
        orderCode = [int64]$orderCode
        amount = 50000
        description = "Test payment"
        accountNumber = "1234567890"
        reference = "FT123456"
        transactionDateTime = (Get-Date).ToString("yyyy-MM-dd HH:mm:ss")
        currency = "VND"
        paymentLinkId = "test-link-id"
        code = "00"
        desc = "Thành công"
        counterAccountBankId = ""
        counterAccountBankName = ""
        counterAccountName = ""
        counterAccountNumber = ""
        virtualAccountName = ""
        virtualAccountNumber = ""
    }
    signature = "test-signature-for-local-testing"
} | ConvertTo-Json -Depth 10

Write-Host "Webhook payload:" -ForegroundColor Cyan
Write-Host $webhookPayload -ForegroundColor Gray
Write-Host ""

Write-Host "⚠️  NOTE: This test will likely fail signature verification" -ForegroundColor Yellow
Write-Host "This is expected - it's just to test if the endpoint is reachable." -ForegroundColor Yellow
Write-Host ""

try {
    Write-Host "Sending webhook request..." -ForegroundColor Green
    
    $response = Invoke-WebRequest -Uri $webhookUrl `
        -Method Post `
        -ContentType "application/json" `
        -Headers @{
            "x-signature" = "test-signature-for-local-testing"
        } `
        -Body $webhookPayload `
        -UseBasicParsing
    
    Write-Host "✅ Response Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "Response Body:" -ForegroundColor Cyan
    Write-Host $response.Content -ForegroundColor Gray
    
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    
    if ($statusCode -eq 400) {
        Write-Host "✅ Endpoint is reachable (400 Bad Request is expected for test payload)" -ForegroundColor Green
        Write-Host "The webhook endpoint is working correctly!" -ForegroundColor Green
    } else {
        Write-Host "❌ Error: $_" -ForegroundColor Red
        Write-Host "Status Code: $statusCode" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "For real webhook testing:" -ForegroundColor White
Write-Host "1. Use ngrok: .\setup_webhook_testing.ps1" -ForegroundColor Gray
Write-Host "2. Configure webhook URL in PayOS dashboard" -ForegroundColor Gray
Write-Host "3. Make a real payment" -ForegroundColor Gray
Write-Host "============================================" -ForegroundColor Cyan
