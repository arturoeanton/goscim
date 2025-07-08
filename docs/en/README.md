# GoSCIM Technical Documentation

## Project Overview

**GoSCIM** is a complete SCIM 2.0 (System for Cross-domain Identity Management) implementation built in Go. It provides a robust and scalable solution for identity management in distributed environments, specifically designed to integrate heterogeneous identity systems.

## Key Features

### SCIM 2.0 Compliance
- ✅ Complete CRUD operations (Create, Read, Update, Delete)
- ✅ Advanced search with SCIM filters
- ✅ Pagination and sorting
- ✅ Extensible and customizable schemas
- ✅ Multiple resource type support
- ✅ Bulk operations (in development)

### Technical Architecture
- **Language**: Go 1.16+
- **Web Framework**: Gin (high performance)
- **Database**: Couchbase (distributed NoSQL)
- **Parser**: ANTLR v4 for SCIM filters
- **Data Format**: Native JSON

## System Architecture

### Core Components

#### 1. Server Core (`main.go`)
```go
// Server initialization
func main() {
    scim.InitDB()                    // Couchbase connection
    r := gin.Default()               // HTTP router
    scim.ReadResourceType(config, r) // Dynamic schema loading
    r.Run(port)                      // HTTP server
}
```

#### 2. Configuration Management (`scim/config.go`)
- **Dynamic schema loading** from JSON files
- **Automatic endpoint registration** based on resource types
- **Schema validation** at server startup

#### 3. Database Integration (`scim/couchbase.go`)
- **Secure connection** with authentication
- **Automatic bucket creation** per resource type
- **Customizable bucket configuration**
- **Automatic primary indexes**

#### 4. Filter Parser (`scim/parser/`)
- **ANTLR grammar** for SCIM filters
- **Automatic conversion** to N1QL queries
- **Complete support** for SCIM operators

### SCIM Operations

#### Create (POST)
```http
POST /scim/v2/Users
Content-Type: application/json

{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "userName": "user@example.com",
  "name": {
    "familyName": "Lastname",
    "givenName": "Firstname"
  }
}
```

#### Read (GET)
```http
GET /scim/v2/Users/12345
```

#### Update (PATCH)
```http
PATCH /scim/v2/Users/12345
Content-Type: application/json

{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
  "Operations": [
    {
      "op": "replace",
      "path": "name.familyName",
      "value": "NewLastname"
    }
  ]
}
```

#### Search (GET)
```http
GET /scim/v2/Users?filter=userName sw "admin"&sortBy=userName&sortOrder=ascending&startIndex=1&count=10
```

#### Delete (DELETE)
```http
DELETE /scim/v2/Users/12345
```

## System Configuration

### Environment Variables

#### Required
```bash
export SCIM_ADMIN_USER="Administrator"     # Couchbase admin user
export SCIM_ADMIN_PASSWORD="admin123"     # Couchbase admin password
```

#### Optional
```bash
export SCIM_COUCHBASE_URL="localhost"     # Couchbase server URL
export SCIM_PORT=":8080"                  # SCIM server port
```

### Configuration Structure

```
config/
├── schemas/                    # SCIM schema definitions
│   ├── urn+ietf+params+scim+schemas+core+2.0+User.json
│   ├── urn+ietf+params+scim+schemas+core+2.0+Group.json
│   └── urn+ietf+params+scim+schemas+extension+*.json
├── resourceType/              # Resource types
│   ├── User.json
│   ├── Group.json
│   └── Element.json
├── bucketSettings/            # Couchbase bucket configuration
│   ├── User.json
│   ├── Group.json
│   └── Element.json
└── serviceProviderConfig/     # Provider configuration
    └── sp_config.json
```

## Schemas and Extensions

### Base User Schema
```json
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User",
  "name": "User",
  "description": "User Account",
  "attributes": [
    {
      "name": "userName",
      "type": "string",
      "required": true,
      "uniqueness": "server"
    },
    {
      "name": "name",
      "type": "complex",
      "subAttributes": [
        {
          "name": "familyName",
          "type": "string"
        },
        {
          "name": "givenName", 
          "type": "string"
        }
      ]
    }
  ]
}
```

### Custom Extensions
```json
{
  "id": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
  "name": "Enterprise User",
  "attributes": [
    {
      "name": "employeeNumber",
      "type": "string",
      "uniqueness": "server"
    },
    {
      "name": "department",
      "type": "string"
    }
  ]
}
```

## Access Control

### Roles and Permissions
```json
{
  "name": "sensitiveAttribute",
  "type": "string",
  "$reader": ["admin", "hr"],      # Roles that can read
  "$writer": ["admin"],            # Roles that can write
  "returned": "default"
}
```

### Role Validation
```go
// Automatic validation in searches
roles := []string{"user", "admin", "superadmin"}
element := ValidateReadRole(roles, resourceType, item)
```

## SCIM Filters

### Supported Syntax
```
# Basic comparisons
userName eq "admin"
name.familyName co "Garcia"
userName sw "admin"
active pr

# Temporal comparisons
meta.lastModified gt "2023-01-01T00:00:00Z"
meta.lastModified ge "2023-01-01T00:00:00Z"

# Logical operators
title pr and userType eq "Employee"
title pr or userType eq "Intern"
userType eq "Employee" and (emails co "company.com" or emails co "company.org")
```

### N1QL Conversion
```go
// Example conversion
query, _ := parser.FilterToN1QL("User", 
    "userName eq \"admin\" and active eq true")
// Result: SELECT * FROM `User` WHERE `userName` = "admin" AND `active` = true
```

## Installation and Deployment

### System Requirements

#### Development
- Go 1.16 or higher
- Couchbase Server 6.0+
- ANTLR 4.7 (for parser regeneration)

#### Production
- CPU: 2+ cores
- RAM: 4GB+ (depending on volume)
- Storage: SSD recommended
- Network: 1Gbps+ for high concurrency

### Local Installation

#### 1. Clone Repository
```bash
git clone https://github.com/arturoeanton/goscim.git
cd goscim
```

#### 2. Install Dependencies
```bash
go mod download
```

#### 3. Configure Couchbase
```bash
# Run Couchbase in Docker
docker run -d --name couchbase-scim \
  -p 8091-8094:8091-8094 \
  -p 11210:11210 \
  couchbase:latest

# Configure cluster via web UI
# http://localhost:8091/ui/index.html
```

#### 4. Configure Environment Variables
```bash
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"
export SCIM_COUCHBASE_URL="localhost"
export SCIM_PORT=":8080"
```

#### 5. Run Server
```bash
go run main.go
```

### Production Deployment

#### Docker Compose
```yaml
version: '3.8'
services:
  couchbase:
    image: couchbase:latest
    ports:
      - "8091-8094:8091-8094"
      - "11210:11210"
    environment:
      - CLUSTER_NAME=scim-cluster
    volumes:
      - couchbase-data:/opt/couchbase/var

  goscim:
    build: .
    ports:
      - "8080:8080"
    environment:
      - SCIM_ADMIN_USER=Administrator
      - SCIM_ADMIN_PASSWORD=admin123
      - SCIM_COUCHBASE_URL=couchbase
    depends_on:
      - couchbase
    restart: unless-stopped

volumes:
  couchbase-data:
```

#### Kubernetes
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
        - name: SCIM_ADMIN_USER
          valueFrom:
            secretKeyRef:
              name: couchbase-secret
              key: username
        - name: SCIM_ADMIN_PASSWORD
          valueFrom:
            secretKeyRef:
              name: couchbase-secret
              key: password
```

## Testing and Development

### Run Tests
```bash
# Unit tests
go test ./...

# Specific tests
go test ./scim/parser -v

# Tests with coverage
go test -cover ./...
```

### Usage Examples
```bash
# Create user
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "testuser",
    "name": {
      "familyName": "Lastname",
      "givenName": "Firstname"
    }
  }'

# Search users
curl "http://localhost:8080/scim/v2/Users?filter=userName sw \"test\""

# Get provider configuration
curl http://localhost:8080/ServiceProviderConfig
```

## Monitoring and Operations

### System Logs
```bash
# Configure structured logging
export SCIM_LOG_LEVEL=debug
export SCIM_LOG_FORMAT=json

# Example log
{"level":"info","timestamp":"2023-12-01T10:00:00Z","message":"GoScim v0.1 started"}
{"level":"debug","timestamp":"2023-12-01T10:00:01Z","message":"Bucket created","bucket":"User"}
```

### Recommended Metrics
- Requests per second (RPS)
- Response time percentiles
- Error rates per endpoint
- Active Couchbase connections
- Memory and CPU usage

### Health Checks
```http
GET /health
{
  "status": "healthy",
  "database": "connected",
  "version": "v0.1"
}
```

## Security

### Security Considerations

#### Authentication
- Implement OAuth 2.0 / OpenID Connect
- JWT token support
- Token validation on each request

#### Authorization
- Granular role-based control
- Resource and operation permissions
- Access audit logs

#### Communication
- Mandatory TLS 1.3 in production
- Valid certificates
- HTTP security headers

#### Validation
- Input sanitization
- Strict schema validation
- Rate limiting per IP/user

## External System Integration

### Identity Providers
- Active Directory
- LDAP
- OAuth 2.0 providers
- SAML 2.0

### Target Systems
- SaaS applications
- User databases
- Directory systems
- Third-party APIs

## Troubleshooting

### Common Issues

#### Couchbase Connection
```bash
# Verify connectivity
telnet localhost 8091

# Verify credentials
curl -u Administrator:admin123 http://localhost:8091/pools
```

#### Schema Errors
```bash
# Validate schema JSON
jq . config/schemas/schema.json

# Verify syntax
go run main.go --validate-config
```

#### Performance Issues
```bash
# Verify indexes
curl -u admin:pass http://localhost:8091/query/service \
  -d 'statement=SELECT * FROM system:indexes WHERE keyspace_id="User"'
```

## Development Roadmap

### Phase 1: Stabilization
- Implement robust authentication
- Complete test suite
- Improve logging and monitoring

### Phase 2: Scalability
- Clustering support
- Distributed cache
- Performance optimizations

### Phase 3: Advanced Features
- Complete Bulk operations
- Webhooks and notifications
- Administration dashboard

## Contributing

### Adding New Resources
1. Create JSON schema in `config/schemas/`
2. Define resource type in `config/resourceType/`
3. Configure bucket in `config/bucketSettings/`
4. Restart server to load changes

### Parser Regeneration
```bash
# Install ANTLR
wget http://www.antlr.org/download/antlr-4.7-complete.jar
alias antlr='java -jar $PWD/antlr-4.7-complete.jar'

# Regenerate parser
antlr -Dlanguage=Go -o scim/parser ScimFilter.g4
```

## Community and Support

For technical support, bug reports, or feature requests:
- **Issues**: [GitHub Issues](https://github.com/arturoeanton/goscim/issues)
- **Discussions**: [GitHub Discussions](https://github.com/arturoeanton/goscim/discussions)
- **Documentation**: [Project Wiki](https://github.com/arturoeanton/goscim/wiki)
- **Examples**: `httpexamples/` directory