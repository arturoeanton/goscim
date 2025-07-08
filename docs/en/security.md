# Security Guide

## Security Overview

GoSCIM handles critical identity information, making security paramount. This guide covers implemented security measures, recommended configurations, and best practices for secure deployment.

## Threat Model

### Protected Assets
- **Identity data**: Personal user information
- **Credentials**: Passwords, access tokens
- **Configuration**: System schemas and settings
- **Metadata**: Group and role information

### Identified Threats
1. **Unauthorized access** to identity data
2. **Data manipulation** of users and groups
3. **Privilege escalation** through roles
4. **Denial of service attacks**
5. **Communication interception**
6. **Code injection** in filters and queries

## Security Configuration

### 1. TLS Configuration

#### Enable TLS Properly
```bash
# Configure environment variables
export SCIM_TLS_ENABLED=true
export SCIM_TLS_CERT_FILE=/path/to/server.crt
export SCIM_TLS_KEY_FILE=/path/to/server.key
export SCIM_TLS_MIN_VERSION=1.3
```

#### Generate Production Certificates
```bash
# Let's Encrypt certificate
sudo certbot certonly --nginx -d scim.example.com

# Or self-signed for development
openssl req -x509 -nodes -days 365 -newkey rsa:4096 \
    -keyout server.key \
    -out server.crt \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=scim.example.com"
```

### 2. Authentication

#### OAuth 2.0 / OpenID Connect (Recommended)
```go
// JWT authentication middleware
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

#### OAuth 2.0 Configuration
```yaml
# oauth-config.yaml
oauth:
  issuer: "https://auth.example.com"
  audience: "scim-api"
  jwks_uri: "https://auth.example.com/.well-known/jwks.json"
  algorithms: ["RS256", "ES256"]
  token_validation:
    verify_exp: true
    verify_iat: true
    leeway: 60  # seconds
```

### 3. Authorization

#### Role-Based Access Control (RBAC)

##### Role Definition
```json
{
  "roles": {
    "admin": {
      "description": "System administrator",
      "permissions": ["*"]
    },
    "hr_manager": {
      "description": "HR manager",
      "permissions": [
        "users:read",
        "users:write",
        "groups:read"
      ]
    },
    "user": {
      "description": "Standard user",
      "permissions": [
        "users:read_self",
        "users:update_self"
      ]
    }
  }
}
```

##### Schema Implementation
```json
{
  "name": "salary",
  "type": "number",
  "description": "Employee salary",
  "$reader": ["admin", "hr_manager", "self"],
  "$writer": ["admin", "hr_manager"],
  "returned": "default",
  "mutability": "readWrite"
}
```

### 4. Input Validation

#### Data Sanitization
```go
func sanitizeInput(input string) string {
    // Remove dangerous characters
    input = strings.ReplaceAll(input, "<", "&lt;")
    input = strings.ReplaceAll(input, ">", "&gt;")
    input = strings.ReplaceAll(input, "\"", "&quot;")
    input = strings.ReplaceAll(input, "'", "&#x27;")
    input = strings.ReplaceAll(input, "&", "&amp;")
    return input
}
```

#### Schema Validation
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
        }
    }
    return nil
}
```

#### N1QL Injection Prevention
```go
func sanitizeFilterValue(value string) string {
    // Escape special characters in N1QL
    value = strings.ReplaceAll(value, "'", "''")
    value = strings.ReplaceAll(value, "\\", "\\\\")
    return value
}

func buildSafeQuery(bucket string, filter string) string {
    // Use parameterized queries
    query := "SELECT * FROM `" + bucket + "`"
    if filter != "" {
        sanitizedFilter := sanitizeFilter(filter)
        query += " WHERE " + sanitizedFilter
    }
    return query
}
```

## Security by Environment

### Development
```bash
# .env.development
SCIM_TLS_ENABLED=false
SCIM_LOG_LEVEL=debug
SCIM_AUTH_REQUIRED=false
SCIM_CORS_ENABLED=true
SCIM_RATE_LIMIT_DISABLED=true
```

### Production
```bash
# .env.production
SCIM_TLS_ENABLED=true
SCIM_TLS_MIN_VERSION=1.3
SCIM_LOG_LEVEL=warn
SCIM_AUTH_REQUIRED=true
SCIM_AUTH_ISSUER=https://auth.example.com
SCIM_CORS_ENABLED=false
SCIM_RATE_LIMIT=1000
SCIM_SECURITY_HEADERS=true
SCIM_HSTS_ENABLED=true
```

## HTTP Security Headers

### Implementation
```go
func securityHeadersMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Prevent MIME type sniffing
        c.Header("X-Content-Type-Options", "nosniff")
        
        // Prevent clickjacking
        c.Header("X-Frame-Options", "DENY")
        
        // XSS Protection
        c.Header("X-XSS-Protection", "1; mode=block")
        
        // HSTS
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
        
        // CSP
        c.Header("Content-Security-Policy", "default-src 'self'; script-src 'none'; object-src 'none';")
        
        // Referrer Policy
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        
        c.Next()
    }
}
```

## Rate Limiting and DDoS Protection

### Rate Limiting Implementation
```go
func rateLimitMiddleware() gin.HandlerFunc {
    // Use redis for distributed rate limiting
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

## Security Checklist

### Pre-Deployment
- [ ] TLS 1.3 configured properly
- [ ] OAuth 2.0 authentication implemented
- [ ] Input validation enabled
- [ ] Rate limiting configured
- [ ] Security headers implemented
- [ ] Audit logging activated
- [ ] Secrets management configured

### Post-Deployment
- [ ] Vulnerability scanning completed
- [ ] Penetration testing performed
- [ ] Security monitoring active
- [ ] Incident response procedures documented
- [ ] Encrypted backups validated
- [ ] Disaster recovery plan tested

## Recommended Security Tools

### Static Analysis
```bash
# Gosec - Go security scanner
gosec ./...

# Nancy - Dependency vulnerabilities
nancy sleuth go.sum
```

### Security Testing
```bash
# OWASP ZAP - Security proxy
zap-cli quick-scan --self-contained https://scim.example.com

# SQLMap - SQL injection testing
sqlmap -u "https://scim.example.com/scim/v2/Users?filter=test" --batch
```

### Continuous Monitoring
- **Falco**: Runtime threat detection
- **OSSEC**: Host-based intrusion detection
- **Suricata**: Network IDS/IPS
- **ELK Stack**: Security log analysis

Security is an ongoing process that requires constant vigilance and regular updates to protection measures.