#!/bin/bash
# SSL Setup Script for Whisko PetCare

set -e

echo "========================================="
echo "  SSL Certificate Setup with Let's Encrypt"
echo "========================================="
echo ""

# Configuration
read -p "Enter your domain name (e.g., api.whisko.com): " DOMAIN
read -p "Enter your email address: " EMAIL

echo ""
echo "Domain: $DOMAIN"
echo "Email: $EMAIL"
echo ""
read -p "Is this correct? (y/n): " confirm

if [ "$confirm" != "y" ]; then
    echo "Cancelled."
    exit 1
fi

# Update nginx.conf with domain
echo "Updating nginx configuration..."
sed -i "s/your-domain.com/$DOMAIN/g" nginx/nginx.conf

# Create directories
echo "Creating directories..."
mkdir -p certbot/conf
mkdir -p certbot/www
mkdir -p nginx/ssl

# Create initial nginx config for HTTP only (for certbot verification)
echo "Creating temporary HTTP-only nginx config..."
cat > nginx/nginx-temp.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    upstream app {
        server app:8080;
    }

    server {
        listen 80;
        server_name DOMAIN_PLACEHOLDER;
        
        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }

        location / {
            proxy_pass http://app;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
    }
}
EOF

sed -i "s/DOMAIN_PLACEHOLDER/$DOMAIN/g" nginx/nginx-temp.conf

# Start nginx with temp config
echo "Starting nginx with HTTP-only config..."
docker compose up -d nginx

# Wait for nginx to start
sleep 5

# Get SSL certificate
echo "Obtaining SSL certificate from Let's Encrypt..."
docker compose run --rm certbot certonly \
    --webroot \
    --webroot-path=/var/www/certbot \
    --email $EMAIL \
    --agree-tos \
    --no-eff-email \
    -d $DOMAIN

# Replace nginx config with HTTPS version
echo "Switching to HTTPS configuration..."
docker compose down
sleep 2

# Start all services with HTTPS
echo "Starting all services with HTTPS..."
docker compose up -d

echo ""
echo "========================================="
echo "  ✅ SSL Setup Complete!"
echo "========================================="
echo ""
echo "Your API is now available at:"
echo "  https://$DOMAIN"
echo ""
echo "HTTP requests will automatically redirect to HTTPS"
echo ""
echo "Certificate will auto-renew every 12 hours"
echo "========================================="
