# Test Deployment Server Webhook

param(
    [Parameter(Mandatory=$false)]
    [string]$ServerIP = "15.134.38.118"
)

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Testing Deployment Server Webhook" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

if ([string]::IsNullOrWhiteSpace($ServerIP)) {
    Write-Host "Please provide your deployment server IP address" -ForegroundColor Yellow
    $ServerIP = Read-Host "Enter your server IP (e.g., 13.123.45.67)"
    
    if ([string]::IsNullOrWhiteSpace($ServerIP)) {
        Write-Host "❌ Server IP is required" -ForegroundColor Red
        exit 1
    }
}

$webhookUrl = "http://${ServerIP}:8080/payments/webhook"

Write-Host "Server IP: $ServerIP" -ForegroundColor White
Write-Host "Webhook URL: $webhookUrl" -ForegroundColor White
Write-Host ""

Write-Host "Step 1: Testing if webhook endpoint is reachable..." -ForegroundColor Cyan
Write-Host "Please wait..." -ForegroundColor Gray
Write-Host ""

try {
    $response = Invoke-WebRequest -Uri $webhookUrl `
        -Method Post `
        -ContentType "application/json" `
        -Headers @{"x-signature" = "test"} `
        -Body '{"test": "data"}' `
        -UseBasicParsing `
        -TimeoutSec 10
    
    Write-Host "✅ Connected! Status: $($response.StatusCode)" -ForegroundColor Green
    
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    
    if ($statusCode -eq 400) {
        Write-Host "============================================" -ForegroundColor Green
        Write-Host "  ✅ SUCCESS! Webhook endpoint is working!" -ForegroundColor Green
        Write-Host "============================================" -ForegroundColor Green
        Write-Host ""
        Write-Host "Status: 400 Bad Request (This is EXPECTED and GOOD!)" -ForegroundColor Gray
        Write-Host "It means the endpoint is reachable and responding." -ForegroundColor Gray
        Write-Host ""
        Write-Host "============================================" -ForegroundColor Cyan
        Write-Host "  Next Step: Configure in PayOS Dashboard" -ForegroundColor Cyan
        Write-Host "============================================" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "1. Open PayOS Dashboard:" -ForegroundColor White
        Write-Host "   https://payos.vn/" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "2. Login to your account" -ForegroundColor White
        Write-Host ""
        Write-Host "3. Navigate to webhook settings:" -ForegroundColor White
        Write-Host "   Settings → Webhook Configuration" -ForegroundColor Gray
        Write-Host "   (or similar menu option)" -ForegroundColor Gray
        Write-Host ""
        Write-Host "4. Copy and paste this webhook URL:" -ForegroundColor White
        Write-Host "   $webhookUrl" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "5. Click Save/Update" -ForegroundColor White
        Write-Host ""
        Write-Host "============================================" -ForegroundColor Cyan
        Write-Host "  Testing Instructions" -ForegroundColor Cyan
        Write-Host "============================================" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "After configuring webhook in PayOS:" -ForegroundColor White
        Write-Host ""
        Write-Host "1. Create a payment using your API" -ForegroundColor White
        Write-Host "2. Complete payment on PayOS checkout page" -ForegroundColor White
        Write-Host "3. Check server logs:" -ForegroundColor White
        Write-Host "   ssh ubuntu@$ServerIP" -ForegroundColor Gray
        Write-Host "   cd /home/ubuntu/whisko-petcare/deployments" -ForegroundColor Gray
        Write-Host "   docker-compose logs -f app | grep WEBHOOK" -ForegroundColor Gray
        Write-Host ""
        Write-Host "4. Look for these messages:" -ForegroundColor White
        Write-Host "   🔔 WEBHOOK RECEIVED from PayOS" -ForegroundColor Gray
        Write-Host "   ✅ Webhook verified! Order Code: XXXXX" -ForegroundColor Gray
        Write-Host "   ✅ Webhook processed successfully!" -ForegroundColor Gray
        Write-Host ""
        Write-Host "5. Verify schedules created automatically" -ForegroundColor White
        Write-Host ""
        Write-Host "============================================" -ForegroundColor Green
        Write-Host "  🎉 Your server is ready for webhooks!" -ForegroundColor Green
        Write-Host "============================================" -ForegroundColor Green
        
    } elseif ($statusCode -eq 404) {
        Write-Host "❌ Error: Webhook endpoint not found (404)" -ForegroundColor Red
        Write-Host "The server is reachable but the webhook endpoint doesn't exist." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Possible issues:" -ForegroundColor White
        Write-Host "  - Server code is not deployed" -ForegroundColor Gray
        Write-Host "  - Application not running" -ForegroundColor Gray
        Write-Host "  - Route configuration issue" -ForegroundColor Gray
        
    } elseif ($null -eq $statusCode) {
        Write-Host "❌ Error: Cannot connect to server" -ForegroundColor Red
        Write-Host ""
        Write-Host "Possible issues:" -ForegroundColor White
        Write-Host "  - Server is not running" -ForegroundColor Gray
        Write-Host "  - Wrong IP address" -ForegroundColor Gray
        Write-Host "  - Port 8080 is not open in security group" -ForegroundColor Gray
        Write-Host "  - Firewall blocking connection" -ForegroundColor Gray
        Write-Host ""
        Write-Host "Troubleshooting steps:" -ForegroundColor Yellow
        Write-Host "1. Verify server IP is correct" -ForegroundColor White
        Write-Host "2. Check EC2 Security Group allows port 8080 inbound" -ForegroundColor White
        Write-Host "3. SSH to server and check if container is running:" -ForegroundColor White
        Write-Host "   ssh ubuntu@$ServerIP" -ForegroundColor Gray
        Write-Host "   cd /home/ubuntu/whisko-petcare/deployments" -ForegroundColor Gray
        Write-Host "   docker-compose ps" -ForegroundColor Gray
        
    } else {
        Write-Host "❌ Unexpected status code: $statusCode" -ForegroundColor Red
        Write-Host "Error details: $_" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Need help? Check these docs:" -ForegroundColor White
Write-Host "  - WEBHOOK_DEPLOYMENT_GUIDE.md" -ForegroundColor Gray
Write-Host "  - docs/PAYOS_WEBHOOK_SETUP.md" -ForegroundColor Gray
Write-Host "============================================" -ForegroundColor Cyan
