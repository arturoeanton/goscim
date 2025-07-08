# Guía de Instalación y Configuración GoSCIM

## Requisitos del Sistema

### Requisitos Mínimos
- **Sistema Operativo**: Linux (Ubuntu 18.04+, CentOS 7+), macOS 10.14+, Windows 10
- **Go**: Versión 1.16 o superior
- **Memoria RAM**: 2GB mínimo, 4GB recomendado
- **Almacenamiento**: 10GB disponibles
- **Red**: Conectividad a internet para descarga de dependencias

### Requisitos de Producción
- **Sistema Operativo**: Linux (Ubuntu 20.04 LTS recomendado)
- **CPU**: 4 cores mínimo, 8 cores recomendado
- **Memoria RAM**: 8GB mínimo, 16GB recomendado
- **Almacenamiento**: 100GB SSD
- **Red**: 1Gbps+ para alta concurrencia

## Instalación en Desarrollo

### 1. Instalación de Go
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# CentOS/RHEL
sudo yum install go

# macOS con Homebrew
brew install go

# Verificar instalación
go version
```

### 2. Instalación de Couchbase

#### Opción A: Docker (Recomendado para desarrollo)
```bash
# Ejecutar Couchbase en contenedor
docker run -d --name couchbase-scim \
    -p 8091-8094:8091-8094 \
    -p 11210:11210 \
    -e CLUSTER_NAME=scim-cluster \
    -e CLUSTER_USERNAME=Administrator \
    -e CLUSTER_PASSWORD=admin123 \
    couchbase:latest

# Esperar que el contenedor esté listo
docker logs -f couchbase-scim
```

#### Opción B: Instalación Nativa
```bash
# Ubuntu/Debian
wget https://packages.couchbase.com/releases/7.0.0/couchbase-server-community_7.0.0-ubuntu20.04_amd64.deb
sudo dpkg -i couchbase-server-community_7.0.0-ubuntu20.04_amd64.deb
sudo systemctl start couchbase-server

# CentOS/RHEL
wget https://packages.couchbase.com/releases/7.0.0/couchbase-server-community-7.0.0-centos8.x86_64.rpm
sudo rpm -i couchbase-server-community-7.0.0-centos8.x86_64.rpm
sudo systemctl start couchbase-server
```

### 3. Configuración de Couchbase

#### Configuración Inicial via Web UI
1. Acceder a `http://localhost:8091`
2. Seleccionar "Setup New Cluster"
3. Configurar:
   - **Cluster Name**: `scim-cluster`
   - **Admin Username**: `Administrator`
   - **Admin Password**: `admin123`
4. Aceptar términos y condiciones
5. Finalizar con configuración por defecto

#### Configuración via CLI
```bash
# Inicializar cluster
couchbase-cli cluster-init \
    --cluster localhost:8091 \
    --cluster-username Administrator \
    --cluster-password admin123 \
    --cluster-name scim-cluster \
    --services data,query,index \
    --cluster-ramsize 1024

# Verificar configuración
couchbase-cli server-info \
    --cluster localhost:8091 \
    --username Administrator \
    --password admin123
```

### 4. Instalación de GoSCIM

#### Clonar Repositorio
```bash
# Crear directorio de trabajo
mkdir -p ~/go/src/github.com/arturoeanton
cd ~/go/src/github.com/arturoeanton

# Clonar repositorio
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Verificar estructura
ls -la
```

#### Instalar Dependencias
```bash
# Descargar dependencias
go mod download

# Verificar dependencias
go mod verify

# Opcional: limpiar módulos
go mod tidy
```

#### Configurar Variables de Entorno
```bash
# Crear archivo de configuración
cat > .env << EOF
SCIM_ADMIN_USER=Administrator
SCIM_ADMIN_PASSWORD=admin123
SCIM_COUCHBASE_URL=localhost
SCIM_PORT=:8080
SCIM_LOG_LEVEL=info
EOF

# Cargar variables
source .env
export $(cat .env | xargs)
```

### 5. Compilar y Ejecutar
```bash
# Compilar aplicación
go build -o goscim main.go

# Ejecutar en modo desarrollo
./goscim

# O ejecutar directamente
go run main.go
```

### 6. Verificación de Instalación
```bash
# Verificar que el servidor esté funcionando
curl http://localhost:8080/ServiceProviderConfig

# Verificar endpoints disponibles
curl http://localhost:8080/ResourceTypes

# Verificar esquemas
curl http://localhost:8080/Schemas
```

## Instalación en Producción

### 1. Preparación del Sistema

#### Crear Usuario de Sistema
```bash
# Crear usuario dedicado
sudo useradd -r -s /bin/false -d /opt/goscim goscim

# Crear directorios
sudo mkdir -p /opt/goscim/{bin,config,logs,data}
sudo chown -R goscim:goscim /opt/goscim
```

#### Configurar Firewall
```bash
# Ubuntu/Debian con UFW
sudo ufw allow 8080/tcp
sudo ufw allow 8091:8094/tcp
sudo ufw allow 11210/tcp

# CentOS/RHEL con firewalld
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=8091-8094/tcp
sudo firewall-cmd --permanent --add-port=11210/tcp
sudo firewall-cmd --reload
```

### 2. Instalación de Couchbase Cluster

#### Configuración Multi-Nodo
```bash
# Nodo 1 (master)
couchbase-cli cluster-init \
    --cluster 10.0.1.10:8091 \
    --cluster-username Administrator \
    --cluster-password SecurePassword123 \
    --cluster-name scim-production \
    --services data,query,index \
    --cluster-ramsize 4096

# Nodo 2 (replica)
couchbase-cli server-add \
    --cluster 10.0.1.10:8091 \
    --username Administrator \
    --password SecurePassword123 \
    --server-add 10.0.1.11:8091 \
    --server-add-username Administrator \
    --server-add-password SecurePassword123 \
    --services data,query,index

# Nodo 3 (replica)
couchbase-cli server-add \
    --cluster 10.0.1.10:8091 \
    --username Administrator \
    --password SecurePassword123 \
    --server-add 10.0.1.12:8091 \
    --server-add-username Administrator \
    --server-add-password SecurePassword123 \
    --services data,query,index

# Rebalancear cluster
couchbase-cli rebalance \
    --cluster 10.0.1.10:8091 \
    --username Administrator \
    --password SecurePassword123
```

### 3. Configuración de Aplicación

#### Archivo de Configuración
```bash
# Crear configuración de producción
sudo cat > /opt/goscim/config/production.env << EOF
SCIM_ADMIN_USER=Administrator
SCIM_ADMIN_PASSWORD=SecurePassword123
SCIM_COUCHBASE_URL=10.0.1.10,10.0.1.11,10.0.1.12
SCIM_PORT=:8080
SCIM_LOG_LEVEL=warn
SCIM_LOG_FORMAT=json
SCIM_TLS_ENABLED=true
SCIM_TLS_CERT_FILE=/opt/goscim/config/server.crt
SCIM_TLS_KEY_FILE=/opt/goscim/config/server.key
EOF

sudo chown goscim:goscim /opt/goscim/config/production.env
sudo chmod 600 /opt/goscim/config/production.env
```

#### Configurar TLS
```bash
# Generar certificados SSL
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /opt/goscim/config/server.key \
    -out /opt/goscim/config/server.crt \
    -subj "/C=ES/ST=Madrid/L=Madrid/O=Organization/OU=IT/CN=scim.example.com"

sudo chown goscim:goscim /opt/goscim/config/server.*
sudo chmod 600 /opt/goscim/config/server.*
```

### 4. Compilar para Producción
```bash
# Compilar con optimizaciones
cd /opt/goscim
sudo -u goscim go build -ldflags "-s -w" -o bin/goscim main.go

# Crear script de inicio
sudo cat > /opt/goscim/bin/start.sh << 'EOF'
#!/bin/bash
source /opt/goscim/config/production.env
exec /opt/goscim/bin/goscim
EOF

sudo chmod +x /opt/goscim/bin/start.sh
sudo chown goscim:goscim /opt/goscim/bin/start.sh
```

### 5. Configurar Systemd Service
```bash
# Crear service unit
sudo cat > /etc/systemd/system/goscim.service << EOF
[Unit]
Description=GoSCIM Server
After=network.target
Requires=network.target

[Service]
Type=simple
User=goscim
Group=goscim
WorkingDirectory=/opt/goscim
ExecStart=/opt/goscim/bin/start.sh
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=goscim
KillMode=mixed
TimeoutStopSec=30

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/opt/goscim/logs
ProtectHome=true

[Install]
WantedBy=multi-user.target
EOF

# Habilitar y iniciar servicio
sudo systemctl daemon-reload
sudo systemctl enable goscim
sudo systemctl start goscim
sudo systemctl status goscim
```

## Configuración Avanzada

### 1. Load Balancer con Nginx

#### Instalación de Nginx
```bash
# Ubuntu/Debian
sudo apt install nginx

# CentOS/RHEL
sudo yum install nginx
```

#### Configuración de Nginx
```bash
sudo cat > /etc/nginx/sites-available/goscim << 'EOF'
upstream goscim_backend {
    server 10.0.1.20:8080;
    server 10.0.1.21:8080;
    server 10.0.1.22:8080;
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
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://goscim_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }

    location /health {
        access_log off;
        proxy_pass http://goscim_backend;
    }
}
EOF

sudo ln -s /etc/nginx/sites-available/goscim /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 2. Monitoreo con Prometheus

#### Configuración de Prometheus
```yaml
# /etc/prometheus/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'goscim'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s
```

### 3. Backup y Recuperación

#### Script de Backup
```bash
#!/bin/bash
# /opt/goscim/bin/backup.sh

BACKUP_DIR="/opt/goscim/backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Crear directorio de backup
mkdir -p $BACKUP_DIR

# Backup de Couchbase
cbbackup http://localhost:8091 $BACKUP_DIR/couchbase_$DATE \
    -u Administrator -p SecurePassword123

# Backup de configuración
tar -czf $BACKUP_DIR/config_$DATE.tar.gz /opt/goscim/config/

# Limpiar backups antiguos (mantener 7 días)
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete
find $BACKUP_DIR -name "couchbase_*" -mtime +7 -exec rm -rf {} \;
```

## Troubleshooting

### Problemas Comunes

#### Error: "connection refused"
```bash
# Verificar que Couchbase esté funcionando
sudo systemctl status couchbase-server

# Verificar conectividad
telnet localhost 8091
```

#### Error: "authentication failed"
```bash
# Verificar credenciales
curl -u Administrator:admin123 http://localhost:8091/pools

# Resetear password de admin
sudo /opt/couchbase/bin/couchbase-cli reset-admin-password \
    --cluster localhost:8091 \
    --new-password NewPassword123
```

#### Error: "bucket not found"
```bash
# Verificar buckets existentes
couchbase-cli bucket-list \
    --cluster localhost:8091 \
    --username Administrator \
    --password admin123

# Crear bucket manualmente
couchbase-cli bucket-create \
    --cluster localhost:8091 \
    --username Administrator \
    --password admin123 \
    --bucket User \
    --bucket-type couchbase \
    --bucket-ramsize 512
```

### Logs y Debugging

#### Logs de Aplicación
```bash
# Seguir logs en tiempo real
sudo journalctl -u goscim -f

# Logs por fecha
sudo journalctl -u goscim --since "2023-12-01" --until "2023-12-02"

# Logs con filtros
sudo journalctl -u goscim -p err
```

#### Logs de Couchbase
```bash
# Logs principales
sudo tail -f /opt/couchbase/var/lib/couchbase/logs/couchbase.log

# Logs de queries
sudo tail -f /opt/couchbase/var/lib/couchbase/logs/query.log
```

## Validación Final

### Checklist de Instalación
- [ ] Go 1.16+ instalado
- [ ] Couchbase funcionando
- [ ] Buckets creados automáticamente
- [ ] Aplicación GoSCIM iniciada
- [ ] Endpoints respondiendo
- [ ] TLS configurado (producción)
- [ ] Firewall configurado
- [ ] Logs funcionando
- [ ] Backup configurado

### Tests de Funcionalidad
```bash
# Test básico de conectividad
curl -k https://scim.example.com/ServiceProviderConfig

# Test de creación de usuario
curl -k -X POST https://scim.example.com/scim/v2/Users \
    -H "Content-Type: application/json" \
    -d '{
        "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
        "userName": "testuser",
        "name": {
            "familyName": "Test",
            "givenName": "User"
        }
    }'

# Test de búsqueda
curl -k "https://scim.example.com/scim/v2/Users?filter=userName eq \"testuser\""
```

La instalación está completa cuando todos los tests de funcionalidad pasan correctamente.