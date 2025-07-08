# Installation Guide

## Quick Installation

### Option 1: Docker (Recommended)

The fastest way to get GoSCIM running:

```bash
# Clone the repository
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Start with Docker Compose
docker-compose up -d

# Verify installation
curl http://localhost:8080/ServiceProviderConfig
```

### Option 2: Binary Release

```bash
# Download latest release
wget https://github.com/arturoeanton/goscim/releases/latest/download/goscim-linux-amd64
chmod +x goscim-linux-amd64

# Set environment variables
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"

# Run
./goscim-linux-amd64
```

### Option 3: Build from Source

```bash
# Prerequisites: Go 1.16+
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Build and run
go build -o goscim main.go
./goscim
```

## System Requirements

### Minimum Requirements
- **CPU**: 1 core
- **Memory**: 512MB RAM
- **Storage**: 1GB available space
- **OS**: Linux, macOS, or Windows

### Recommended for Production
- **CPU**: 2+ cores
- **Memory**: 4GB+ RAM  
- **Storage**: 20GB+ SSD
- **OS**: Ubuntu 20.04 LTS or CentOS 8+

## Database Setup

### Couchbase Installation

#### Docker (Development)
```bash
docker run -d --name couchbase-scim \
  -p 8091-8094:8091-8094 \
  -p 11210:11210 \
  couchbase:community-7.0.0
```

#### Ubuntu/Debian
```bash
# Add Couchbase repository
wget -O - https://packages.couchbase.com/couchbase.key | sudo apt-key add -
echo "deb https://packages.couchbase.com/repos/deb/ubuntu20.04 focal focal/main" | sudo tee /etc/apt/sources.list.d/couchbase.list

# Install
sudo apt update
sudo apt install couchbase-server-community

# Start service
sudo systemctl start couchbase-server
sudo systemctl enable couchbase-server
```

#### CentOS/RHEL
```bash
# Add repository
sudo wget -O /etc/yum.repos.d/couchbase.repo https://packages.couchbase.com/rpm/couchbase-centos8.repo

# Install
sudo yum install couchbase-server-community

# Start service
sudo systemctl start couchbase-server
sudo systemctl enable couchbase-server
```

### Database Configuration

#### Web Console Setup
1. Open `http://localhost:8091`
2. Choose "Setup New Cluster"
3. Set cluster name: `goscim-cluster`
4. Create admin user: `Administrator` / `admin123`
5. Accept terms and finish setup

#### CLI Setup
```bash
# Initialize cluster
couchbase-cli cluster-init \
    --cluster localhost:8091 \
    --cluster-username Administrator \
    --cluster-password admin123 \
    --cluster-name goscim-cluster \
    --services data,query,index \
    --cluster-ramsize 1024

# Verify setup
couchbase-cli server-info \
    --cluster localhost:8091 \
    --username Administrator \
    --password admin123
```

## Environment Configuration

### Required Variables
```bash
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"
export SCIM_COUCHBASE_URL="localhost"
```

### Optional Variables
```bash
export SCIM_PORT=":8080"
export SCIM_LOG_LEVEL="info"
export SCIM_TLS_ENABLED="false"
```

### Production Variables
```bash
export SCIM_TLS_ENABLED="true"
export SCIM_TLS_CERT_FILE="/path/to/cert.pem"
export SCIM_TLS_KEY_FILE="/path/to/key.pem"
export SCIM_AUTH_REQUIRED="true"
export SCIM_RATE_LIMIT="1000"
```

## Service Installation

### Systemd Service (Linux)

#### Create Service File
```bash
sudo tee /etc/systemd/system/goscim.service << 'EOF'
[Unit]
Description=GoSCIM Server
After=network.target couchbase-server.service
Requires=network.target

[Service]
Type=simple
User=goscim
Group=goscim
WorkingDirectory=/opt/goscim
ExecStart=/opt/goscim/goscim
Restart=always
RestartSec=10

# Environment variables
Environment=SCIM_ADMIN_USER=Administrator
Environment=SCIM_ADMIN_PASSWORD=admin123
Environment=SCIM_COUCHBASE_URL=localhost
Environment=SCIM_PORT=:8080

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/opt/goscim/logs
ProtectHome=true

[Install]
WantedBy=multi-user.target
EOF
```

#### Install and Start
```bash
# Create user
sudo useradd -r -s /bin/false goscim

# Create directories
sudo mkdir -p /opt/goscim/{bin,config,logs}
sudo cp goscim /opt/goscim/bin/
sudo cp -r config /opt/goscim/
sudo chown -R goscim:goscim /opt/goscim

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable goscim
sudo systemctl start goscim

# Check status
sudo systemctl status goscim
```

## Container Deployment

### Docker

#### Simple Container
```bash
# Build image
docker build -t goscim:latest .

# Run container
docker run -d --name goscim \
  -p 8080:8080 \
  -e SCIM_ADMIN_USER=Administrator \
  -e SCIM_ADMIN_PASSWORD=admin123 \
  -e SCIM_COUCHBASE_URL=couchbase-host \
  goscim:latest
```

#### Docker Compose
```yaml
version: '3.8'

services:
  couchbase:
    image: couchbase:community-7.0.0
    container_name: goscim-couchbase
    ports:
      - "8091-8094:8091-8094"
      - "11210:11210"
    environment:
      - CLUSTER_NAME=goscim-cluster
    volumes:
      - couchbase-data:/opt/couchbase/var
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8091/pools"]
      interval: 30s
      timeout: 10s
      retries: 5

  goscim:
    build: .
    container_name: goscim-app
    ports:
      - "8080:8080"
    environment:
      - SCIM_ADMIN_USER=Administrator
      - SCIM_ADMIN_PASSWORD=admin123
      - SCIM_COUCHBASE_URL=couchbase
      - SCIM_LOG_LEVEL=info
    depends_on:
      couchbase:
        condition: service_healthy
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  couchbase-data:
```

### Kubernetes

#### Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: goscim
  labels:
    app: goscim
spec:
  replicas: 3
  selector:
    matchLabels:
      app: goscim
  template:
    metadata:
      labels:
        app: goscim
    spec:
      containers:
      - name: goscim
        image: goscim:latest
        ports:
        - containerPort: 8080
        env:
        - name: SCIM_COUCHBASE_URL
          value: "couchbase-service"
        - name: SCIM_ADMIN_USER
          valueFrom:
            secretKeyRef:
              name: couchbase-credentials
              key: username
        - name: SCIM_ADMIN_PASSWORD
          valueFrom:
            secretKeyRef:
              name: couchbase-credentials
              key: password
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: goscim-service
spec:
  selector:
    app: goscim
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: goscim-ingress
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - scim.example.com
    secretName: goscim-tls
  rules:
  - host: scim.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: goscim-service
            port:
              number: 80
```

## Load Balancer Setup

### Nginx
```nginx
upstream goscim_backend {
    server 10.0.1.10:8080;
    server 10.0.1.11:8080;
    server 10.0.1.12:8080;
}

server {
    listen 80;
    server_name scim.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name scim.example.com;

    ssl_certificate /etc/ssl/certs/scim.example.com.crt;
    ssl_certificate_key /etc/ssl/private/scim.example.com.key;

    location / {
        proxy_pass http://goscim_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        access_log off;
        proxy_pass http://goscim_backend;
    }
}
```

## Verification

### Health Check
```bash
curl http://localhost:8080/health
```

### SCIM Endpoints
```bash
# Service provider configuration
curl http://localhost:8080/ServiceProviderConfig

# Resource types
curl http://localhost:8080/ResourceTypes

# Schemas
curl http://localhost:8080/Schemas
```

### Create Test User
```bash
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "test@example.com",
    "name": {
      "familyName": "User",
      "givenName": "Test"
    },
    "active": true
  }'
```

## Troubleshooting

### Common Issues

#### Connection Refused
```bash
# Check if service is running
sudo systemctl status goscim

# Check logs
sudo journalctl -u goscim -f

# Verify port is listening
sudo netstat -tlnp | grep :8080
```

#### Database Connection Failed
```bash
# Test Couchbase connectivity
curl -f http://localhost:8091/pools

# Check Couchbase service
sudo systemctl status couchbase-server

# Verify credentials
curl -u Administrator:admin123 http://localhost:8091/pools
```

#### Permission Denied
```bash
# Check file permissions
ls -la /opt/goscim/

# Fix ownership
sudo chown -R goscim:goscim /opt/goscim/
```

### Log Analysis
```bash
# View application logs
sudo journalctl -u goscim --since "1 hour ago"

# Follow logs in real-time
sudo journalctl -u goscim -f

# View Couchbase logs
sudo tail -f /opt/couchbase/var/lib/couchbase/logs/couchbase.log
```

## Next Steps

After successful installation:

1. **Configure Authentication**: Set up OAuth 2.0 or JWT authentication
2. **Enable TLS**: Configure SSL/TLS certificates for production
3. **Set up Monitoring**: Install monitoring and alerting
4. **Create Backups**: Configure automated backups
5. **Load Test**: Verify performance under expected load

For detailed configuration guides, see:
- [Security Guide](security.md)
- [Operations Guide](operations.md)
- [Integration Guide](integrations.md)