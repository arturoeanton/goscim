# Guía de Operaciones GoSCIM

## Resumen de Operaciones

Esta guía proporciona información completa para administradores de sistemas, ingenieros DevOps y equipos de operaciones responsables del despliegue, monitoreo y mantenimiento de GoSCIM en entornos de producción.

## Despliegue en Producción

### Arquitectura de Despliegue Recomendada

```
Internet
    │
    ▼
┌─────────────────┐
│  Load Balancer  │ (HAProxy/Nginx)
│  + SSL/TLS      │
└─────┬───────────┘
      │
   ┌──┴──┐
   │     │
   ▼     ▼
┌──────┐ ┌──────┐  ┌──────┐
│GoSCIM│ │GoSCIM│  │GoSCIM│ (Auto-scaling)
│ App  │ │ App  │  │ App  │
│  1   │ │  2   │  │  N   │
└──┬───┘ └──┬───┘  └──┬───┘
   │        │         │
   └────────┼─────────┘
            │
    ┌───────▼───────┐
    │   Couchbase   │
    │    Cluster    │
    └───────────────┘
```

### Configuración de Load Balancer

#### HAProxy Configuration
```bash
# /etc/haproxy/haproxy.cfg
global
    daemon
    user haproxy
    group haproxy
    log 127.0.0.1:514 local0

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    option httplog
    log global

frontend goscim_frontend
    bind *:443 ssl crt /etc/ssl/certs/scim.exemplo.com.pem
    bind *:80
    redirect scheme https if !{ ssl_fc }
    
    # Security headers
    http-response set-header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
    http-response set-header X-Frame-Options "DENY"
    http-response set-header X-Content-Type-Options "nosniff"
    
    # Health check endpoint
    acl health_check path_beg /health
    use_backend goscim_health if health_check
    
    default_backend goscim_backend

backend goscim_backend
    balance roundrobin
    option httpchk GET /health
    http-check expect status 200
    
    server goscim1 10.0.1.10:8080 check inter 30s
    server goscim2 10.0.1.11:8080 check inter 30s
    server goscim3 10.0.1.12:8080 check inter 30s

backend goscim_health
    server goscim1 10.0.1.10:8080 check
```

#### Nginx Configuration
```nginx
# /etc/nginx/sites-available/goscim
upstream goscim_backend {
    least_conn;
    server 10.0.1.10:8080 max_fails=3 fail_timeout=30s;
    server 10.0.1.11:8080 max_fails=3 fail_timeout=30s;
    server 10.0.1.12:8080 max_fails=3 fail_timeout=30s;
}

# Rate limiting
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=auth:10m rate=5r/s;

server {
    listen 80;
    server_name scim.ejemplo.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name scim.ejemplo.com;

    ssl_certificate /etc/letsencrypt/live/scim.ejemplo.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/scim.ejemplo.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
    ssl_prefer_server_ciphers off;

    # Security headers
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload";
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";

    # Health check (no rate limiting)
    location /health {
        access_log off;
        proxy_pass http://goscim_backend;
        proxy_set_header Host $host;
    }

    # Auth endpoints (stricter rate limiting)
    location ~ ^/auth {
        limit_req zone=auth burst=5 nodelay;
        proxy_pass http://goscim_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # API endpoints
    location / {
        limit_req zone=api burst=20 nodelay;
        proxy_pass http://goscim_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
        
        # Buffer settings
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }
}
```

## Monitoreo y Observabilidad

### Métricas de Aplicación

#### Prometheus Metrics
```go
// metrics.go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "goscim_http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "goscim_http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    databaseConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "goscim_database_connections",
            Help: "Number of active database connections",
        },
    )

    scimOperationsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "goscim_operations_total",
            Help: "Total number of SCIM operations",
        },
        []string{"operation", "resource", "status"},
    )
)

func metricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        
        httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
        httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
    }
}
```

#### Prometheus Configuration
```yaml
# /etc/prometheus/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "goscim_rules.yml"

scrape_configs:
  - job_name: 'goscim'
    static_configs:
      - targets: 
        - '10.0.1.10:8080'
        - '10.0.1.11:8080'
        - '10.0.1.12:8080'
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'couchbase'
    static_configs:
      - targets: ['10.0.2.10:8091', '10.0.2.11:8091', '10.0.2.12:8091']
    metrics_path: '/_prometheusMetrics'
    scrape_interval: 30s

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

#### Grafana Dashboard
```json
{
  "dashboard": {
    "id": null,
    "title": "GoSCIM Monitoring",
    "tags": ["goscim", "identity"],
    "timezone": "browser",
    "panels": [
      {
        "title": "Requests per Second",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(goscim_http_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "Response Time Percentiles",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(goscim_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(goscim_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(goscim_http_requests_total{status=~\"4..|5..\"}[5m]) / rate(goscim_http_requests_total[5m])",
            "legendFormat": "Error Rate"
          }
        ]
      }
    ]
  }
}
```

### Alertas

#### Prometheus Alerting Rules
```yaml
# /etc/prometheus/goscim_rules.yml
groups:
  - name: goscim.rules
    rules:
      - alert: GoSCIMHighErrorRate
        expr: rate(goscim_http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "GoSCIM high error rate"
          description: "GoSCIM error rate is {{ $value }} errors per second"

      - alert: GoSCIMHighResponseTime
        expr: histogram_quantile(0.95, rate(goscim_http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "GoSCIM high response time"
          description: "95th percentile response time is {{ $value }} seconds"

      - alert: GoSCIMDatabaseConnectionsHigh
        expr: goscim_database_connections > 80
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "GoSCIM high database connections"
          description: "Database connections: {{ $value }}"

      - alert: GoSCIMInstanceDown
        expr: up{job="goscim"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "GoSCIM instance down"
          description: "GoSCIM instance {{ $labels.instance }} is down"
```

### Logging

#### Structured Logging Configuration
```go
// logging.go
package main

import (
    "github.com/sirupsen/logrus"
    "github.com/gin-gonic/gin"
)

func initLogging() {
    logrus.SetFormatter(&logrus.JSONFormatter{
        TimestampFormat: time.RFC3339,
    })
    
    level := os.Getenv("SCIM_LOG_LEVEL")
    switch level {
    case "debug":
        logrus.SetLevel(logrus.DebugLevel)
    case "info":
        logrus.SetLevel(logrus.InfoLevel)
    case "warn":
        logrus.SetLevel(logrus.WarnLevel)
    case "error":
        logrus.SetLevel(logrus.ErrorLevel)
    default:
        logrus.SetLevel(logrus.InfoLevel)
    }
}

func loggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        raw := c.Request.URL.RawQuery
        
        c.Next()
        
        end := time.Now()
        latency := end.Sub(start)
        
        logrus.WithFields(logrus.Fields{
            "status":     c.Writer.Status(),
            "method":     c.Request.Method,
            "path":       path,
            "query":      raw,
            "ip":         c.ClientIP(),
            "user_agent": c.Request.UserAgent(),
            "latency":    latency,
            "user_id":    getUserID(c),
        }).Info("HTTP Request")
    }
}
```

#### Log Aggregation con ELK Stack
```yaml
# filebeat.yml
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /var/log/goscim/*.log
  fields:
    service: goscim
  fields_under_root: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "goscim-logs-%{+yyyy.MM.dd}"

logging.level: info
logging.to_files: true
logging.files:
  path: /var/log/filebeat
  name: filebeat
```

## Gestión de Configuración

### Configuración por Ambiente

#### Development
```yaml
# config/development.yaml
server:
  port: 8080
  tls_enabled: false
  
database:
  host: localhost
  port: 8091
  username: Administrator
  password: admin123
  
logging:
  level: debug
  format: text
  
auth:
  required: false
  
features:
  bulk_operations: true
  metrics: true
```

#### Production
```yaml
# config/production.yaml
server:
  port: 8080
  tls_enabled: true
  tls_cert_file: /opt/goscim/certs/server.crt
  tls_key_file: /opt/goscim/certs/server.key
  
database:
  hosts:
    - couchbase-1.internal:8091
    - couchbase-2.internal:8091
    - couchbase-3.internal:8091
  username: ${COUCHBASE_USERNAME}
  password: ${COUCHBASE_PASSWORD}
  tls_enabled: true
  
logging:
  level: warn
  format: json
  output: /var/log/goscim/app.log
  
auth:
  required: true
  issuer: https://auth.ejemplo.com
  audience: scim-api
  
rate_limiting:
  enabled: true
  requests_per_minute: 1000
  
features:
  bulk_operations: false
  metrics: true
```

### Gestión de Secretos

#### HashiCorp Vault Integration
```go
// vault.go
package main

import (
    "github.com/hashicorp/vault/api"
)

func loadSecretsFromVault() error {
    config := api.DefaultConfig()
    config.Address = os.Getenv("VAULT_ADDR")
    
    client, err := api.NewClient(config)
    if err != nil {
        return err
    }
    
    client.SetToken(os.Getenv("VAULT_TOKEN"))
    
    secret, err := client.Logical().Read("secret/data/goscim")
    if err != nil {
        return err
    }
    
    data := secret.Data["data"].(map[string]interface{})
    
    os.Setenv("SCIM_ADMIN_PASSWORD", data["db_password"].(string))
    os.Setenv("SCIM_JWT_SECRET", data["jwt_secret"].(string))
    
    return nil
}
```

## Backup y Recuperación

### Estrategia de Backup

#### Script de Backup Automatizado
```bash
#!/bin/bash
# /opt/goscim/scripts/backup.sh

set -euo pipefail

BACKUP_DIR="/opt/backups/goscim"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# Crear directorio de backup
mkdir -p $BACKUP_DIR/$DATE

# Backup de Couchbase
cbbackup http://couchbase-1.internal:8091 $BACKUP_DIR/$DATE/couchbase \
    -u $COUCHBASE_USERNAME -p $COUCHBASE_PASSWORD \
    --verbose

# Backup de configuración
tar -czf $BACKUP_DIR/$DATE/config.tar.gz /opt/goscim/config/

# Backup de logs (últimos 7 días)
find /var/log/goscim -name "*.log" -mtime -7 -exec tar -czf $BACKUP_DIR/$DATE/logs.tar.gz {} +

# Calcular checksum
cd $BACKUP_DIR/$DATE
sha256sum * > checksums.sha256

# Encriptar backup
gpg --cipher-algo AES256 --compress-algo 1 --s2k-mode 3 \
    --s2k-digest-algo SHA512 --s2k-count 65536 \
    --symmetric --output ../goscim_backup_$DATE.gpg .

# Limpiar backup sin encriptar
cd ..
rm -rf $DATE

# Subir a S3
aws s3 cp goscim_backup_$DATE.gpg s3://backups-bucket/goscim/

# Limpiar backups antiguos
find $BACKUP_DIR -name "goscim_backup_*.gpg" -mtime +$RETENTION_DAYS -delete

# Notificar éxito
echo "Backup completed successfully: goscim_backup_$DATE.gpg"
```

#### Configurar Cron para Backups
```bash
# Backup diario a las 2:00 AM
0 2 * * * /opt/goscim/scripts/backup.sh >> /var/log/goscim/backup.log 2>&1

# Backup de configuración cada 6 horas
0 */6 * * * tar -czf /opt/backups/config_$(date +\%H).tar.gz /opt/goscim/config/
```

### Procedimiento de Recuperación

#### Recuperación Completa
```bash
#!/bin/bash
# /opt/goscim/scripts/restore.sh

BACKUP_FILE=$1
TEMP_DIR="/tmp/goscim_restore"

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file.gpg>"
    exit 1
fi

# Crear directorio temporal
mkdir -p $TEMP_DIR
cd $TEMP_DIR

# Desencriptar backup
gpg --decrypt $BACKUP_FILE | tar -xz

# Verificar integridad
sha256sum -c checksums.sha256

# Parar servicios
systemctl stop goscim

# Restaurar configuración
tar -xzf config.tar.gz -C /

# Restaurar Couchbase
cbrestore couchbase/ http://couchbase-1.internal:8091 \
    -u $COUCHBASE_USERNAME -p $COUCHBASE_PASSWORD

# Iniciar servicios
systemctl start goscim

# Verificar funcionamiento
sleep 30
curl -f http://localhost:8080/health || exit 1

echo "Restore completed successfully"
```

## Actualizaciones y Mantenimiento

### Procedimiento de Actualización

#### Rolling Update
```bash
#!/bin/bash
# /opt/goscim/scripts/rolling_update.sh

NEW_VERSION=$1
INSTANCES=(goscim-1 goscim-2 goscim-3)

if [ -z "$NEW_VERSION" ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

for instance in "${INSTANCES[@]}"; do
    echo "Updating $instance to version $NEW_VERSION"
    
    # Drenar tráfico del load balancer
    curl -X POST "http://lb.internal/api/drain/$instance"
    
    # Esperar que las conexiones activas terminen
    sleep 60
    
    # Parar instancia
    ssh $instance "systemctl stop goscim"
    
    # Actualizar binario
    scp goscim-$NEW_VERSION $instance:/opt/goscim/bin/goscim
    
    # Actualizar configuración si es necesario
    # scp config/production.yaml $instance:/opt/goscim/config/
    
    # Iniciar instancia
    ssh $instance "systemctl start goscim"
    
    # Verificar salud
    for i in {1..30}; do
        if ssh $instance "curl -f http://localhost:8080/health"; then
            echo "$instance is healthy"
            break
        fi
        sleep 10
    done
    
    # Reactivar en load balancer
    curl -X POST "http://lb.internal/api/enable/$instance"
    
    echo "$instance updated successfully"
    sleep 30
done

echo "Rolling update completed"
```

### Mantenimiento Programado

#### Script de Mantenimiento
```bash
#!/bin/bash
# /opt/goscim/scripts/maintenance.sh

echo "Starting maintenance tasks..."

# Limpiar logs antiguos
find /var/log/goscim -name "*.log" -mtime +7 -delete

# Limpiar archivos temporales
find /tmp -name "goscim_*" -mtime +1 -delete

# Compactar base de datos
curl -X POST http://couchbase-1.internal:8091/controller/compactBucket \
    -u $COUCHBASE_USERNAME:$COUCHBASE_PASSWORD \
    -d bucket=User

# Rebalancear cluster si es necesario
# couchbase-cli rebalance --cluster couchbase-1.internal:8091 \
#     --username $COUCHBASE_USERNAME --password $COUCHBASE_PASSWORD

# Verificar estado del sistema
/opt/goscim/scripts/health_check.sh

echo "Maintenance completed"
```

## Troubleshooting

### Problemas Comunes

#### High Memory Usage
```bash
# Investigar uso de memoria
ps aux | grep goscim
pmap -d $(pgrep goscim)

# Analizar heap
go tool pprof http://localhost:8080/debug/pprof/heap

# Configurar límites de memoria
systemctl edit goscim
# [Service]
# MemoryLimit=1G
# MemoryAccounting=yes
```

#### Database Connection Issues
```bash
# Verificar conectividad
telnet couchbase-1.internal 8091

# Verificar logs de Couchbase
tail -f /opt/couchbase/var/lib/couchbase/logs/couchbase.log

# Verificar consultas lentas
curl -u admin:password http://couchbase-1.internal:8091/admin/vitals
```

#### Performance Issues
```bash
# Analizar CPU
perf top -p $(pgrep goscim)

# Analizar I/O
iotop -p $(pgrep goscim)

# Profiles de Go
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
```

### Scripts de Diagnóstico

#### Health Check Completo
```bash
#!/bin/bash
# /opt/goscim/scripts/health_check.sh

echo "=== GoSCIM Health Check ==="

# Verificar proceso
if pgrep goscim > /dev/null; then
    echo "✓ GoSCIM process is running"
else
    echo "✗ GoSCIM process is not running"
    exit 1
fi

# Verificar puerto
if netstat -ln | grep :8080 > /dev/null; then
    echo "✓ Port 8080 is listening"
else
    echo "✗ Port 8080 is not listening"
fi

# Verificar endpoint de salud
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "✓ Health endpoint is responding"
else
    echo "✗ Health endpoint is not responding"
fi

# Verificar base de datos
if curl -f http://couchbase-1.internal:8091/pools > /dev/null 2>&1; then
    echo "✓ Couchbase is responding"
else
    echo "✗ Couchbase is not responding"
fi

# Verificar espacio en disco
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -lt 90 ]; then
    echo "✓ Disk usage is acceptable ($DISK_USAGE%)"
else
    echo "⚠ Disk usage is high ($DISK_USAGE%)"
fi

# Verificar memoria
MEM_USAGE=$(free | grep Mem | awk '{printf("%.0f", $3/$2 * 100.0)}')
if [ $MEM_USAGE -lt 90 ]; then
    echo "✓ Memory usage is acceptable ($MEM_USAGE%)"
else
    echo "⚠ Memory usage is high ($MEM_USAGE%)"
fi

echo "Health check completed"
```

## Escalabilidad

### Auto-scaling con Kubernetes

#### Deployment YAML
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: goscim
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
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
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
  type: LoadBalancer

---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: goscim-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: goscim
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

La gestión operacional efectiva de GoSCIM requiere una combinación de monitoreo proactivo, mantenimiento preventivo y procedimientos de respuesta bien definidos.