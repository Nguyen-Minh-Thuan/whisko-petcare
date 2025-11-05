# Comprehensive Webhook Troubleshooting Script

param(
    [Parameter(Mandatory=$false)]
    [string]$ServerIP = "15.134.38.118"
)

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Webhook Troubleshooting Diagnostic" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

$webhookUrl = "http://${ServerIP}:8080/payments/webhook"
$healthUrl = "http://${ServerIP}:8080/health"
$baseUrl = "http://${ServerIP}:8080"

Write-Host "Target Server: $ServerIP" -ForegroundColor White
Write-Host ""

# Test 1: Ping server
Write-Host "Test 1: Checking if server is reachable (ping)..." -ForegroundColor Cyan
try {
    $pingResult = Test-Connection -ComputerName $ServerIP -Count 2 -Quiet
    if ($pingResult) {
        Write-Host "  [OK] Server is reachable via ping" -ForegroundColor Green
    } else {
        Write-Host "  [WARN] Server is not responding to ping" -ForegroundColor Red
        Write-Host "     This might be normal if ICMP is blocked" -ForegroundColor Gray
    }
} catch {
    Write-Host "  [INFO] Ping test inconclusive" -ForegroundColor Yellow
}
Write-Host ""

# Test 2: Check port 8080
Write-Host "Test 2: Checking if port 8080 is open..." -ForegroundColor Cyan
try {
    $tcpClient = New-Object System.Net.Sockets.TcpClient
    $connect = $tcpClient.BeginConnect($ServerIP, 8080, $null, $null)
    $wait = $connect.AsyncWaitHandle.WaitOne(3000, $false)
    
    if ($wait) {
        try {
            $tcpClient.EndConnect($connect)
            Write-Host "  [OK] Port 8080 is OPEN and accepting connections!" -ForegroundColor Green
            $tcpClient.Close()
        } catch {
            Write-Host "  [FAIL] Port 8080 connection failed" -ForegroundColor Red
        }
    } else {
        Write-Host "  [FAIL] Port 8080 is CLOSED or FILTERED" -ForegroundColor Red
        Write-Host "     This is the problem! Port 8080 must be opened in Security Group" -ForegroundColor Yellow
        $tcpClient.Close()
    }
} catch {
    Write-Host "  [FAIL] Cannot connect to port 8080" -ForegroundColor Red
}
Write-Host ""

# Test 3: Try base URL
Write-Host "Test 3: Testing base URL ($baseUrl)..." -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri $baseUrl -Method Get -TimeoutSec 5 -UseBasicParsing
    Write-Host "  [OK] Server is responding! Status: $($response.StatusCode)" -ForegroundColor Green
} catch {
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "  [INFO] Server responded with: $statusCode" -ForegroundColor Yellow
    } else {
        Write-Host "  [FAIL] No response from server" -ForegroundColor Red
    }
}
Write-Host ""

# Test 4: Try health endpoint
Write-Host "Test 4: Testing health endpoint ($healthUrl)..." -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri $healthUrl -Method Get -TimeoutSec 5 -UseBasicParsing
    Write-Host "  [OK] Health endpoint is working! Status: $($response.StatusCode)" -ForegroundColor Green
} catch {
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "  [INFO] Health endpoint responded with: $statusCode" -ForegroundColor Yellow
    } else {
        Write-Host "  [FAIL] Health endpoint not reachable" -ForegroundColor Red
    }
}
Write-Host ""

# Test 5: Try webhook endpoint
Write-Host "Test 5: Testing webhook endpoint ($webhookUrl)..." -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri $webhookUrl `
        -Method Post `
        -ContentType "application/json" `
        -Headers @{"x-signature" = "test"} `
        -Body '{"test": "data"}' `
        -TimeoutSec 5 `
        -UseBasicParsing
    Write-Host "  [OK] Webhook endpoint responded! Status: $($response.StatusCode)" -ForegroundColor Green
} catch {
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 400) {
            Write-Host "  [OK] Webhook endpoint is working!" -ForegroundColor Green
            Write-Host "     (400 Bad Request is expected for test payload)" -ForegroundColor Gray
        } else {
            Write-Host "  [INFO] Webhook endpoint responded with: $statusCode" -ForegroundColor Yellow
        }
    } else {
        Write-Host "  [FAIL] Webhook endpoint not reachable" -ForegroundColor Red
    }
}
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  SOLUTION: Open Port 8080 in AWS" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Your EC2 instance needs port 8080 open in Security Group." -ForegroundColor White
Write-Host ""
Write-Host "Steps to fix:" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Go to AWS Console: https://console.aws.amazon.com/ec2/" -ForegroundColor White
Write-Host "2. Click Instances in left menu" -ForegroundColor White
Write-Host "3. Find your instance (IP: $ServerIP)" -ForegroundColor White
Write-Host "4. Click Security tab" -ForegroundColor White
Write-Host "5. Click on Security Group name" -ForegroundColor White
Write-Host "6. Click Edit inbound rules" -ForegroundColor White
Write-Host "7. Click Add rule and configure:" -ForegroundColor White
Write-Host "   - Type: Custom TCP" -ForegroundColor Gray
Write-Host "   - Port: 8080" -ForegroundColor Gray
Write-Host "   - Source: 0.0.0.0/0" -ForegroundColor Gray
Write-Host "   - Description: API webhook endpoint" -ForegroundColor Gray
Write-Host "8. Click Save rules" -ForegroundColor White
Write-Host ""
Write-Host "After fixing, run this script again to verify:" -ForegroundColor Yellow
Write-Host "  powershell -ExecutionPolicy Bypass -File .\troubleshoot_webhook.ps1" -ForegroundColor Gray
Write-Host ""
Write-Host "Once all tests pass, configure in PayOS Dashboard:" -ForegroundColor Yellow
Write-Host "  Webhook URL: $webhookUrl" -ForegroundColor Cyan
Write-Host ""
Write-Host "For detailed instructions, see: FIX_SECURITY_GROUP.md" -ForegroundColor Gray
Write-Host ""
