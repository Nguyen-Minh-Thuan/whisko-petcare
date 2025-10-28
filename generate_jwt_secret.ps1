# Generate Secure JWT Secret Key
# This script generates a cryptographically secure random string for JWT_SECRET_KEY

Write-Host "=== JWT Secret Key Generator ===" -ForegroundColor Cyan
Write-Host ""

# Method 1: Generate 64-character random string (recommended)
Write-Host "Method 1: Alphanumeric (64 characters)" -ForegroundColor Yellow
$chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
$key1 = -join ((1..64) | ForEach-Object { $chars[(Get-Random -Maximum $chars.Length)] })
Write-Host $key1 -ForegroundColor Green
Write-Host ""

# Method 2: Base64 encoded random bytes (cryptographically secure)
Write-Host "Method 2: Base64 Encoded (stronger)" -ForegroundColor Yellow
$bytes = [byte[]]::new(48)
[Security.Cryptography.RNGCryptoServiceProvider]::Create().GetBytes($bytes)
$key2 = [Convert]::ToBase64String($bytes)
Write-Host $key2 -ForegroundColor Green
Write-Host ""

# Method 3: Hexadecimal (64 characters)
Write-Host "Method 3: Hexadecimal (64 characters)" -ForegroundColor Yellow
$hexBytes = [byte[]]::new(32)
[Security.Cryptography.RNGCryptoServiceProvider]::Create().GetBytes($hexBytes)
$key3 = ($hexBytes | ForEach-Object { $_.ToString("x2") }) -join ''
Write-Host $key3 -ForegroundColor Green
Write-Host ""

# Method 4: UUID-based (simple but less secure)
Write-Host "Method 4: UUID-based (simple)" -ForegroundColor Yellow
$uuid1 = [guid]::NewGuid().ToString("N")
$uuid2 = [guid]::NewGuid().ToString("N")
$key4 = $uuid1 + $uuid2
Write-Host $key4 -ForegroundColor Green
Write-Host ""

# Recommendation
Write-Host "=== Recommendation ===" -ForegroundColor Cyan
Write-Host "Use Method 2 (Base64) for maximum security" -ForegroundColor White
Write-Host "Minimum length: 32 characters" -ForegroundColor White
Write-Host ""

# Copy to clipboard option
Write-Host "Would you like to copy Method 2 to clipboard? (Y/N)" -ForegroundColor Yellow
$response = Read-Host
if ($response -eq 'Y' -or $response -eq 'y') {
    $key2 | Set-Clipboard
    Write-Host "✅ Copied to clipboard!" -ForegroundColor Green
    Write-Host "Paste this into GitHub Secrets as JWT_SECRET_KEY" -ForegroundColor White
} else {
    Write-Host "Copy one of the keys above and add it to GitHub Secrets" -ForegroundColor White
}

Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Go to GitHub Repository → Settings → Secrets → Actions" -ForegroundColor White
Write-Host "2. Click 'New repository secret'" -ForegroundColor White
Write-Host "3. Name: JWT_SECRET_KEY" -ForegroundColor White
Write-Host "4. Value: (paste the generated key)" -ForegroundColor White
Write-Host "5. Click 'Add secret'" -ForegroundColor White
Write-Host ""
Write-Host "⚠️  Never commit this key to Git!" -ForegroundColor Red
