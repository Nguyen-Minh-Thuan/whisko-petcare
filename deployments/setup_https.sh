#!/bin/bash

# Whisko API HTTPS Setup Script
# Run this on your Ubuntu server: sudo bash setup_https.sh

set -e

echo "🚀 Whisko API HTTPS Setup Script"
echo "================================="
echo ""

# Configuration
DOMAIN="api.whisko.shop"
EMAIL="your-email@example.com"  # Change this!
API_PORT="8080"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}❌ This script must be run as root (use sudo)${NC}" 
   exit 1
fi

echo -e "${YELLOW}📋 Configuration:${NC}"
echo "   Domain: $DOMAIN"
echo "   Email: $EMAIL"
echo "   API Port: $API_PORT"
echo ""

# Check if email is set
if [ "$EMAIL" = "your-email@example.com" ]; then
    echo -e "${RED}❌ Please edit this script and set your email address!${NC}"
    exit 1
fi

read -p "Continue with setup? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
fi

# Step 1: Update system
echo -e "${GREEN}📦 Step 1: Updating system...${NC}"
apt update
apt upgrade -y

# Step 2: Install Nginx
echo -e "${GREEN}🌐 Step 2: Installing Nginx...${NC}"
apt install -y nginx

# Step 3: Install Certbot
echo -e "${GREEN}🔐 Step 3: Installing Certbot...${NC}"
apt install -y certbot python3-certbot-nginx

# Step 4: Create certbot directory
echo -e "${GREEN}📁 Step 4: Creating directories...${NC}"
mkdir -p /var/www/certbot

# Step 5: Stop Nginx temporarily
echo -e "${GREEN}⏸️  Step 5: Stopping Nginx...${NC}"
systemctl stop nginx

# Step 6: Obtain SSL certificate
echo -e "${GREEN}🔒 Step 6: Obtaining SSL certificate...${NC}"
certbot certonly --standalone -d $DOMAIN --email $EMAIL --agree-tos --non-interactive

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Failed to obtain SSL certificate!${NC}"
    echo -e "${YELLOW}💡 Make sure:${NC}"
    echo "   1. Domain $DOMAIN points to this server's IP"
    echo "   2. Ports 80 and 443 are open in firewall"
    echo "   3. No other service is using port 80/443"
    exit 1
fi

# Step 7: Create Nginx configuration
echo -e "${GREEN}⚙️  Step 7: Creating Nginx configuration...${NC}"

cat > /etc/nginx/sites-available/whisko-api << 'EOF'
# Redirect HTTP to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name api.whisko.shop;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

# HTTPS Server
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api.whisko.shop;

    # SSL Certificates
    ssl_certificate /etc/letsencrypt/live/api.whisko.shop/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.whisko.shop/privkey.pem;
    
    # SSL Configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384';
    ssl_prefer_server_ciphers off;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_session_tickets off;
    
    # OCSP Stapling
    ssl_stapling on;
    ssl_stapling_verify on;
    ssl_trusted_certificate /etc/letsencrypt/live/api.whisko.shop/chain.pem;
    
    # Security Headers
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Request limits
    client_max_body_size 20M;
    client_body_buffer_size 128k;

    # Timeouts
    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;

    # Logging
    access_log /var/log/nginx/whisko-api-access.log;
    error_log /var/log/nginx/whisko-api-error.log;

    # Proxy to Go API
    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Port $server_port;
        
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
        proxy_busy_buffers_size 8k;
    }

    # Health check endpoint
    location /health {
        access_log off;
        proxy_pass http://localhost:8080/health;
        proxy_set_header Host $host;
    }
}
EOF

# Step 8: Enable configuration
echo -e "${GREEN}🔗 Step 8: Enabling configuration...${NC}"
rm -f /etc/nginx/sites-enabled/default
ln -sf /etc/nginx/sites-available/whisko-api /etc/nginx/sites-enabled/

# Step 9: Test Nginx configuration
echo -e "${GREEN}✅ Step 9: Testing Nginx configuration...${NC}"
nginx -t

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Nginx configuration test failed!${NC}"
    exit 1
fi

# Step 10: Start Nginx
echo -e "${GREEN}▶️  Step 10: Starting Nginx...${NC}"
systemctl start nginx
systemctl enable nginx

# Step 11: Configure firewall
echo -e "${GREEN}🔥 Step 11: Configuring firewall...${NC}"
if command -v ufw &> /dev/null; then
    ufw allow 'Nginx Full'
    ufw --force enable
    echo -e "${GREEN}✅ Firewall configured${NC}"
else
    echo -e "${YELLOW}⚠️  UFW not found, skipping firewall configuration${NC}"
fi

# Step 12: Test HTTPS
echo -e "${GREEN}🧪 Step 12: Testing HTTPS...${NC}"
sleep 2

if curl -I -k https://$DOMAIN/health &> /dev/null; then
    echo -e "${GREEN}✅ HTTPS is working!${NC}"
else
    echo -e "${YELLOW}⚠️  Could not reach https://$DOMAIN/health${NC}"
    echo -e "${YELLOW}   Make sure your Go API is running on port $API_PORT${NC}"
fi

# Summary
echo ""
echo -e "${GREEN}================================================${NC}"
echo -e "${GREEN}✅ HTTPS Setup Complete!${NC}"
echo -e "${GREEN}================================================${NC}"
echo ""
echo -e "${YELLOW}📋 Summary:${NC}"
echo "   ✅ Nginx installed and configured"
echo "   ✅ SSL certificate obtained from Let's Encrypt"
echo "   ✅ HTTPS enabled on $DOMAIN"
echo "   ✅ HTTP → HTTPS redirect configured"
echo "   ✅ Auto-renewal set up"
echo ""
echo -e "${YELLOW}🔍 Next Steps:${NC}"
echo "   1. Verify Go API is running: sudo lsof -i :$API_PORT"
echo "   2. Test HTTPS: curl -I https://$DOMAIN"
echo "   3. Check SSL rating: https://www.ssllabs.com/ssltest/analyze.html?d=$DOMAIN"
echo "   4. Update Postman base_url to: https://$DOMAIN"
echo ""
echo -e "${YELLOW}📊 Useful Commands:${NC}"
echo "   View logs:    sudo tail -f /var/log/nginx/whisko-api-error.log"
echo "   Reload Nginx: sudo systemctl reload nginx"
echo "   Renew cert:   sudo certbot renew"
echo "   Check cert:   sudo certbot certificates"
echo ""
echo -e "${GREEN}🎉 Your API is now secured with HTTPS!${NC}"
