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
        Write-Host "  ✅ Server is reachable via ping" -ForegroundColor Green
    } else {
        Write-Host "  ❌ Server is not responding to ping" -ForegroundColor Red
        Write-Host "     This might be normal if ICMP is blocked" -ForegroundColor Gray
    }
} catch {
    Write-Host "  ⚠️  Ping test inconclusive" -ForegroundColor Yellow
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
            Write-Host "  ✅ Port 8080 is OPEN and accepting connections!" -ForegroundColor Green
            $tcpClient.Close()
        } catch {
            Write-Host "  ❌ Port 8080 connection failed" -ForegroundColor Red
        }
    } else {
        Write-Host "  ❌ Port 8080 is CLOSED or FILTERED" -ForegroundColor Red
        Write-Host "     ⚠️  This is the problem!" -ForegroundColor Yellow
        $tcpClient.Close()
    }
} catch {
    Write-Host "  ❌ Cannot connect to port 8080" -ForegroundColor Red
}
Write-Host ""

# Test 3: Try base URL
Write-Host "Test 3: Testing base URL ($baseUrl)..." -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri $baseUrl -Method Get -TimeoutSec 5 -UseBasicParsing
    Write-Host "  ✅ Server is responding! Status: $($response.StatusCode)" -ForegroundColor Green
} catch {
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "  ⚠️  Server responded with: $statusCode" -ForegroundColor Yellow
    } else {
        Write-Host "  ❌ No response from server" -ForegroundColor Red
    }
}
Write-Host ""

# Test 4: Try health endpoint
Write-Host "Test 4: Testing health endpoint ($healthUrl)..." -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri $healthUrl -Method Get -TimeoutSec 5 -UseBasicParsing
    Write-Host "  ✅ Health endpoint is working! Status: $($response.StatusCode)" -ForegroundColor Green
} catch {
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "  ⚠️  Health endpoint responded with: $statusCode" -ForegroundColor Yellow
    } else {
        Write-Host "  ❌ Health endpoint not reachable" -ForegroundColor Red
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
    Write-Host "  ✅ Webhook endpoint responded! Status: $($response.StatusCode)" -ForegroundColor Green
} catch {
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 400) {
            Write-Host "  ✅ Webhook endpoint is working!" -ForegroundColor Green
            Write-Host "     (400 Bad Request is expected for test payload)" -ForegroundColor Gray
        } else {
            Write-Host "  ⚠️  Webhook endpoint responded with: $statusCode" -ForegroundColor Yellow
        }
    } else {
        Write-Host "  ❌ Webhook endpoint not reachable" -ForegroundColor Red
    }
}
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Diagnosis & Solutions" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "🔍 Most Common Issue: Security Group Configuration" -ForegroundColor Yellow
Write-Host ""
Write-Host "Your EC2 instance needs to allow inbound traffic on port 8080." -ForegroundColor White
Write-Host ""
Write-Host "How to fix:" -ForegroundColor Cyan
Write-Host ""
Write-Host "Option 1: AWS Console (Web Interface)" -ForegroundColor White
Write-Host "--------------------------------------" -ForegroundColor Gray
Write-Host "1. Go to AWS Console: https://console.aws.amazon.com/ec2/" -ForegroundColor White
Write-Host "2. Click 'Instances' in left menu" -ForegroundColor White
Write-Host "3. Find your instance (IP: $ServerIP)" -ForegroundColor White
Write-Host "4. Click on the instance" -ForegroundColor White
Write-Host "5. Click 'Security' tab" -ForegroundColor White
Write-Host "6. Click on the Security Group name (looks like 'sg-xxxxx')" -ForegroundColor White
Write-Host "7. Click 'Edit inbound rules'" -ForegroundColor White
Write-Host "8. Click 'Add rule'" -ForegroundColor White
Write-Host "9. Configure the rule:" -ForegroundColor White
Write-Host "   - Type: Custom TCP" -ForegroundColor Gray
Write-Host "   - Port range: 8080" -ForegroundColor Gray
Write-Host "   - Source: 0.0.0.0/0 (Anywhere IPv4)" -ForegroundColor Gray
Write-Host "   - Description: API server for webhook" -ForegroundColor Gray
Write-Host "10. Click 'Save rules'" -ForegroundColor White
Write-Host ""

Write-Host "Option 2: AWS CLI" -ForegroundColor White
Write-Host "--------------------------------------" -ForegroundColor Gray
Write-Host "If you have AWS CLI configured:" -ForegroundColor White
Write-Host ""
Write-Host "# First, find your security group ID:" -ForegroundColor Gray
Write-Host 'aws ec2 describe-instances --filters "Name=ip-address,Values=' -NoNewline -ForegroundColor DarkGray
Write-Host $ServerIP -NoNewline -ForegroundColor DarkGray
Write-Host '" --query "Reservations[0].Instances[0].SecurityGroups[0].GroupId" --output text' -ForegroundColor DarkGray
Write-Host ""
Write-Host "# Then add the rule (replace sg-xxxxx with your security group ID):" -ForegroundColor Gray
Write-Host "aws ec2 authorize-security-group-ingress \" -ForegroundColor DarkGray
Write-Host "    --group-id sg-xxxxx \" -ForegroundColor DarkGray
Write-Host "    --protocol tcp \" -ForegroundColor DarkGray
Write-Host "    --port 8080 \" -ForegroundColor DarkGray
Write-Host "    --cidr 0.0.0.0/0" -ForegroundColor DarkGray
Write-Host ""

Write-Host "Option 3: Check with Server Admin" -ForegroundColor White
Write-Host "--------------------------------------" -ForegroundColor Gray
Write-Host "If you don't have AWS access, contact your server administrator" -ForegroundColor White
Write-Host "and ask them to:" -ForegroundColor White
Write-Host "  - Open port 8080 in the EC2 Security Group" -ForegroundColor Gray
Write-Host "  - Allow inbound traffic from 0.0.0.0/0 (or at least PayOS IPs)" -ForegroundColor Gray
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Alternative: Check if Server is Running" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "SSH to your server and check:" -ForegroundColor White
Write-Host ""
Write-Host "ssh ubuntu@$ServerIP" -ForegroundColor Yellow
Write-Host ""
Write-Host "# Check if Docker containers are running:" -ForegroundColor Gray
Write-Host "cd /home/ubuntu/whisko-petcare/deployments" -ForegroundColor Yellow
Write-Host "docker-compose ps" -ForegroundColor Yellow
Write-Host ""
Write-Host "# Expected output: Container 'app' should show 'Up'" -ForegroundColor Gray
Write-Host ""
Write-Host "# If container is not running, start it:" -ForegroundColor Gray
Write-Host "docker-compose up -d" -ForegroundColor Yellow
Write-Host ""
Write-Host "# Check logs:" -ForegroundColor Gray
Write-Host "docker-compose logs -f app" -ForegroundColor Yellow
Write-Host ""
Write-Host "# Should see: 'Server starting on port 8080'" -ForegroundColor Gray
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  After Fixing Security Group" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. Wait 1-2 minutes for changes to take effect" -ForegroundColor White
Write-Host "2. Run this script again to verify:" -ForegroundColor White
Write-Host "   .\troubleshoot_webhook.ps1" -ForegroundColor Yellow
Write-Host ""
Write-Host "3. Once all tests pass, configure webhook in PayOS:" -ForegroundColor White
Write-Host "   - Go to: https://payos.vn/" -ForegroundColor Gray
Write-Host "   - Settings → Webhook Configuration" -ForegroundColor Gray
Write-Host "   - Webhook URL: $webhookUrl" -ForegroundColor Yellow
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Need more help? Check:" -ForegroundColor White
Write-Host "  - WEBHOOK_DEPLOYMENT_GUIDE.md" -ForegroundColor Gray
Write-Host "  - docs/PAYOS_WEBHOOK_SETUP.md" -ForegroundColor Gray
Write-Host "============================================" -ForegroundColor Cyan
