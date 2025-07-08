# Guía de Seguridad GoSCIM

## Resumen de Seguridad

GoSCIM maneja información crítica de identidades, por lo que la seguridad es fundamental. Esta guía cubre las medidas de seguridad implementadas, configuraciones recomendadas y mejores prácticas para un despliegue seguro.

## Modelo de Amenazas

### Activos Protegidos
- **Datos de identidad**: Información personal de usuarios
- **Credenciales**: Passwords, tokens de acceso
- **Configuración**: Esquemas y configuración del sistema
- **Metadatos**: Información de grupos y roles

### Amenazas Identificadas
1. **Acceso no autorizado** a datos de identidad
2. **Manipulación de datos** de usuarios y grupos
3. **Escalamiento de privilegios** a través de roles
4. **Ataques de denegación de servicio**
5. **Intercepción de comunicaciones**
6. **Inyección de código** en filtros y queries

## Configuración de Seguridad

### 1. Configuración TLS

#### Habilitar TLS Correctamente
```bash
# Configurar variables de entorno
export SCIM_TLS_ENABLED=true
export SCIM_TLS_CERT_FILE=/path/to/server.crt
export SCIM_TLS_KEY_FILE=/path/to/server.key
export SCIM_TLS_MIN_VERSION=1.3
```

#### Generar Certificados de Producción
```bash
# Certificado de Let's Encrypt
sudo certbot certonly --nginx -d scim.ejemplo.com

# O certificado autofirmado para desarrollo
openssl req -x509 -nodes -days 365 -newkey rsa:4096 \
    -keyout server.key \
    -out server.crt \
    -subj "/C=ES/ST=Madrid/L=Madrid/O=Empresa/CN=scim.ejemplo.com"
```

#### Configuración TLS en Código
```go
// Configuración TLS segura
func configureTLS() *tls.Config {
    return &tls.Config{
        MinVersion:               tls.VersionTLS13,
        PreferServerCipherSuites: true,
        CipherSuites: []uint16{
            tls.TLS_AES_256_GCM_SHA384,
            tls.TLS_AES_128_GCM_SHA256,
            tls.TLS_CHACHA20_POLY1305_SHA256,
        },
    }
}
```

### 2. Autenticación

#### OAuth 2.0 / OpenID Connect (Recomendado)
```go
// Middleware de autenticación JWT
func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := extractToken(c)
        if tokenString == "" {
            c.AbortWithStatusJSON(401, gin.H{
                "error": "Authorization token required",
            })
            return
        }

        claims, err := validateJWT(tokenString)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{
                "error": "Invalid token",
            })
            return
        }

        c.Set("user", claims)
        c.Next()
    }
}
```

#### Configuración OAuth 2.0
```yaml
# oauth-config.yaml
oauth:
  issuer: "https://auth.ejemplo.com"
  audience: "scim-api"
  jwks_uri: "https://auth.ejemplo.com/.well-known/jwks.json"
  algorithms: ["RS256", "ES256"]
  token_validation:
    verify_exp: true
    verify_iat: true
    leeway: 60  # seconds
```

### 3. Autorización

#### Control de Acceso Basado en Roles (RBAC)

##### Definición de Roles
```json
{
  "roles": {
    "admin": {
      "description": "Administrador del sistema",
      "permissions": ["*"]
    },
    "hr_manager": {
      "description": "Gestor de RRHH",
      "permissions": [
        "users:read",
        "users:write",
        "groups:read"
      ]
    },
    "user": {
      "description": "Usuario estándar",
      "permissions": [
        "users:read_self",
        "users:update_self"
      ]
    }
  }
}
```

##### Implementación en Esquemas
```json
{
  "name": "salary",
  "type": "number",
  "description": "Salario del empleado",
  "$reader": ["admin", "hr_manager", "self"],
  "$writer": ["admin", "hr_manager"],
  "returned": "default",
  "mutability": "readWrite"
}
```

##### Validación de Permisos
```go
func validatePermissions(userRoles []string, operation string, resource string, attribute string) bool {
    requiredPermission := fmt.Sprintf("%s:%s", resource, operation)
    
    for _, role := range userRoles {
        permissions := getRolePermissions(role)
        for _, permission := range permissions {
            if permission == "*" || permission == requiredPermission {
                return true
            }
        }
    }
    return false
}

func ValidateReadRole(roles []string, resourceType ResoruceType, element map[string]interface{}) map[string]interface{} {
    schema := Schemas[resourceType.Schema]
    result := make(map[string]interface{})
    
    for key, value := range element {
        attribute := GetAttribute(schema.Attributes, key)
        if attribute.Read == nil || hasPermission(roles, attribute.Read) {
            result[key] = value
        }
    }
    return result
}
```

### 4. Validación de Entrada

#### Sanitización de Datos
```go
func sanitizeInput(input string) string {
    // Remover caracteres peligrosos
    input = strings.ReplaceAll(input, "<", "&lt;")
    input = strings.ReplaceAll(input, ">", "&gt;")
    input = strings.ReplaceAll(input, "\"", "&quot;")
    input = strings.ReplaceAll(input, "'", "&#x27;")
    input = strings.ReplaceAll(input, "&", "&amp;")
    return input
}
```

#### Validación de Esquemas
```go
func validateSchema(data map[string]interface{}, schema Schema) error {
    for _, attribute := range schema.Attributes {
        value, exists := data[attribute.Name]
        
        if attribute.Required && !exists {
            return fmt.Errorf("attribute '%s' is required", attribute.Name)
        }
        
        if exists {
            if err := validateAttributeType(value, attribute); err != nil {
                return err
            }
            
            if err := validateAttributeValue(value, attribute); err != nil {
                return err
            }
        }
    }
    return nil
}
```

#### Prevención de Inyección N1QL
```go
func sanitizeFilterValue(value string) string {
    // Escapar caracteres especiales en N1QL
    value = strings.ReplaceAll(value, "'", "''")
    value = strings.ReplaceAll(value, "\\", "\\\\")
    return value
}

func buildSafeQuery(bucket string, filter string) string {
    // Usar consultas parametrizadas
    query := "SELECT * FROM `" + bucket + "`"
    if filter != "" {
        // Validar y sanitizar filtro antes de usar
        sanitizedFilter := sanitizeFilter(filter)
        query += " WHERE " + sanitizedFilter
    }
    return query
}
```

### 5. Seguridad en Base de Datos

#### Configuración Segura de Couchbase
```go
func initSecureDB() {
    cluster, err := gocb.Connect("couchbases://"+endpoint, gocb.ClusterOptions{
        Authenticator: gocb.PasswordAuthenticator{
            Username: username,
            Password: password,
        },
        SecurityConfig: gocb.SecurityConfig{
            TLSSkipVerify: false,  // NEVER skip in production
            TLSRootCAs:    loadCACerts(),
        },
        TimeoutsConfig: gocb.TimeoutsConfig{
            ConnectTimeout: 10 * time.Second,
            QueryTimeout:   30 * time.Second,
        },
    })
}
```

#### Encriptación en Reposo
```bash
# Configurar encriptación en Couchbase
couchbase-cli setting-encryption \
    --cluster localhost:8091 \
    --username Administrator \
    --password password \
    --enable-encryption true \
    --encryption-key-path /opt/couchbase/var/lib/couchbase/encryption/keys
```

#### Backup Seguro
```bash
#!/bin/bash
# Script de backup seguro

BACKUP_DIR="/secure/backups"
ENCRYPTION_KEY="/secure/keys/backup.key"

# Backup con encriptación
cbbackup http://localhost:8091 $BACKUP_DIR/$(date +%Y%m%d) \
    -u Administrator -p $COUCHBASE_PASSWORD

# Encriptar backup
gpg --cipher-algo AES256 --compress-algo 1 --s2k-mode 3 \
    --s2k-digest-algo SHA512 --s2k-count 65536 \
    --symmetric --output $BACKUP_DIR/$(date +%Y%m%d).gpg \
    $BACKUP_DIR/$(date +%Y%m%d)

# Limpiar backup sin encriptar
rm -rf $BACKUP_DIR/$(date +%Y%m%d)
```

## Configuración de Seguridad por Ambiente

### Desarrollo
```bash
# .env.development
SCIM_TLS_ENABLED=false
SCIM_LOG_LEVEL=debug
SCIM_AUTH_REQUIRED=false
SCIM_CORS_ENABLED=true
SCIM_RATE_LIMIT_DISABLED=true
```

### Testing
```bash
# .env.testing
SCIM_TLS_ENABLED=true
SCIM_LOG_LEVEL=info
SCIM_AUTH_REQUIRED=true
SCIM_AUTH_ISSUER=https://test-auth.ejemplo.com
SCIM_CORS_ENABLED=true
SCIM_RATE_LIMIT=100
```

### Producción
```bash
# .env.production
SCIM_TLS_ENABLED=true
SCIM_TLS_MIN_VERSION=1.3
SCIM_LOG_LEVEL=warn
SCIM_AUTH_REQUIRED=true
SCIM_AUTH_ISSUER=https://auth.ejemplo.com
SCIM_CORS_ENABLED=false
SCIM_RATE_LIMIT=1000
SCIM_SECURITY_HEADERS=true
SCIM_HSTS_ENABLED=true
```

## Headers de Seguridad HTTP

### Implementación de Security Headers
```go
func securityHeadersMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Prevenir MIME type sniffing
        c.Header("X-Content-Type-Options", "nosniff")
        
        // Prevenir clickjacking
        c.Header("X-Frame-Options", "DENY")
        
        // XSS Protection
        c.Header("X-XSS-Protection", "1; mode=block")
        
        // HSTS
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
        
        // CSP
        c.Header("Content-Security-Policy", "default-src 'self'; script-src 'none'; object-src 'none';")
        
        // Referrer Policy
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        
        // Permissions Policy
        c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
        
        c.Next()
    }
}
```

## Rate Limiting y Protección DDoS

### Implementación de Rate Limiting
```go
func rateLimitMiddleware() gin.HandlerFunc {
    // Usar redis para rate limiting distribuido
    store := redis_rate.NewStore(redisClient)
    limiter := redis_rate.NewLimiter(store)
    
    return func(c *gin.Context) {
        key := getClientIdentifier(c)
        limit := getRateLimit(c)
        
        result, err := limiter.Allow(key, limit)
        if err != nil {
            c.AbortWithStatusJSON(500, gin.H{"error": "Rate limiting error"})
            return
        }
        
        c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit.Rate))
        c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
        c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetTime.Unix()))
        
        if result.Allowed == 0 {
            c.Header("Retry-After", fmt.Sprintf("%d", result.RetryAfter.Seconds()))
            c.AbortWithStatusJSON(429, gin.H{
                "error": "Rate limit exceeded",
                "retry_after": result.RetryAfter.Seconds(),
            })
            return
        }
        
        c.Next()
    }
}
```

## Auditoría y Logging

### Logging de Seguridad
```go
type SecurityEvent struct {
    Timestamp   time.Time `json:"timestamp"`
    Event       string    `json:"event"`
    User        string    `json:"user"`
    IP          string    `json:"ip"`
    UserAgent   string    `json:"user_agent"`
    Resource    string    `json:"resource"`
    Action      string    `json:"action"`
    Success     bool      `json:"success"`
    ErrorCode   string    `json:"error_code,omitempty"`
    Details     string    `json:"details,omitempty"`
}

func logSecurityEvent(c *gin.Context, event string, success bool, details string) {
    secEvent := SecurityEvent{
        Timestamp: time.Now(),
        Event:     event,
        User:      getUserFromContext(c),
        IP:        c.ClientIP(),
        UserAgent: c.GetHeader("User-Agent"),
        Resource:  c.Request.URL.Path,
        Action:    c.Request.Method,
        Success:   success,
        Details:   details,
    }
    
    // Log como JSON estructurado
    logJSON, _ := json.Marshal(secEvent)
    log.Printf("SECURITY_EVENT: %s", logJSON)
    
    // Enviar a SIEM si está configurado
    if siemEnabled {
        sendToSIEM(secEvent)
    }
}
```

### Eventos de Auditoría
- **Autenticación**: Login exitoso/fallido
- **Autorización**: Acceso denegado, escalamiento de privilegios
- **Operaciones**: Creación, modificación, eliminación de recursos
- **Configuración**: Cambios en esquemas o configuración
- **Errores**: Errores de validación, excepciones

## Gestión de Secretos

### Variables de Entorno Seguras
```bash
# Usar herramientas como HashiCorp Vault
vault kv put secret/goscim \
    db_password="SecurePassword123!" \
    jwt_secret="super-secret-key" \
    encryption_key="32-char-encryption-key-here"

# Cargar secretos en runtime
export SCIM_DB_PASSWORD=$(vault kv get -field=db_password secret/goscim)
```

### Rotación de Credenciales
```bash
#!/bin/bash
# Script de rotación automática

# Generar nueva contraseña
NEW_PASSWORD=$(openssl rand -base64 32)

# Actualizar en Vault
vault kv put secret/goscim db_password="$NEW_PASSWORD"

# Actualizar en Couchbase
couchbase-cli user-manage \
    --cluster localhost:8091 \
    --username Administrator \
    --password $OLD_PASSWORD \
    --set --rbac-username scim_user \
    --rbac-password $NEW_PASSWORD

# Reiniciar aplicación para usar nueva contraseña
systemctl restart goscim
```

## Respuesta a Incidentes

### Detección de Amenazas
```go
// Detector de patrones sospechosos
func detectSuspiciousActivity(c *gin.Context) {
    ip := c.ClientIP()
    
    // Múltiples intentos fallidos
    if getFailedAttempts(ip) > 5 {
        logSecurityEvent(c, "BRUTE_FORCE_DETECTED", false, 
            fmt.Sprintf("IP %s exceeded failed attempts", ip))
        blockIP(ip, 1*time.Hour)
    }
    
    // Patrones de inyección
    if containsSQLInjection(c.Request.URL.RawQuery) {
        logSecurityEvent(c, "INJECTION_ATTEMPT", false,
            "SQL injection pattern detected")
        c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request"})
        return
    }
}
```

### Plan de Respuesta
1. **Detección**: Monitoreo automático de logs y métricas
2. **Contención**: Bloqueo automático de IPs maliciosas
3. **Erradicación**: Parcheo de vulnerabilidades
4. **Recuperación**: Restauración desde backups seguros
5. **Lecciones aprendidas**: Actualización de medidas de seguridad

## Compliance y Estándares

### GDPR (Reglamento General de Protección de Datos)
- **Derecho al olvido**: Implementar endpoint de eliminación completa
- **Portabilidad**: Exportación de datos en formato estándar
- **Consentimiento**: Tracking de consentimientos de usuario

### SOC 2 Type II
- **Control de acceso**: Documentación de roles y permisos
- **Encriptación**: Datos encriptados en tránsito y en reposo
- **Monitoreo**: Logs de acceso y cambios
- **Backup**: Procedimientos de respaldo y recuperación

### ISO 27001
- **Gestión de riesgos**: Análisis regular de amenazas
- **Controles de seguridad**: Implementación de controles técnicos
- **Formación**: Capacitación en seguridad para desarrolladores

## Checklist de Seguridad

### Pre-Despliegue
- [ ] TLS 1.3 configurado correctamente
- [ ] Autenticación OAuth 2.0 implementada
- [ ] Validación de entrada habilitada
- [ ] Rate limiting configurado
- [ ] Security headers implementados
- [ ] Logging de auditoría activado
- [ ] Secrets management configurado

### Post-Despliegue
- [ ] Vulnerability scanning completado
- [ ] Penetration testing realizado
- [ ] Monitoring de seguridad activo
- [ ] Procedimientos de respuesta documentados
- [ ] Backups encriptados y validados
- [ ] Plan de recuperación de desastres probado

## Herramientas de Seguridad Recomendadas

### Análisis Estático
```bash
# Gosec - Security scanner para Go
gosec ./...

# Nancy - Vulnerabilidades en dependencias
nancy sleuth go.sum
```

### Testing de Seguridad
```bash
# OWASP ZAP - Proxy de seguridad
zap-cli quick-scan --self-contained https://scim.ejemplo.com

# SQLMap - Testing de inyección SQL
sqlmap -u "https://scim.ejemplo.com/scim/v2/Users?filter=test" --batch
```

### Monitoreo Continuo
- **Falco**: Detección de amenazas en runtime
- **OSSEC**: HIDS (Host-based Intrusion Detection)
- **Suricata**: IDS/IPS de red
- **ELK Stack**: Análisis de logs de seguridad

La seguridad es un proceso continuo que requiere vigilancia constante y actualizaciones regulares de las medidas de protección.