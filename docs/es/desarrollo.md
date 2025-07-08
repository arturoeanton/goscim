# Guía de Desarrollo GoSCIM

## Configuración del Entorno de Desarrollo

### Prerrequisitos

#### Software Requerido
- **Go 1.16+**: Lenguaje de programación
- **Git**: Control de versiones
- **Docker**: Para base de datos local
- **VS Code** o **GoLand**: IDE recomendado
- **Postman** o **Insomnia**: Testing de APIs

#### Instalación en macOS
```bash
# Instalar Homebrew si no está instalado
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Instalar dependencias
brew install go git docker
brew install --cask visual-studio-code
brew install --cask postman
```

#### Instalación en Ubuntu/Debian
```bash
# Actualizar sistema
sudo apt update && sudo apt upgrade -y

# Instalar Go
wget https://go.dev/dl/go1.19.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.19.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Instalar Git y Docker
sudo apt install git docker.io docker-compose -y

# Agregar usuario al grupo docker
sudo usermod -aG docker $USER
```

### Configuración del Proyecto

#### Clonar y Configurar
```bash
# Crear directorio de trabajo
mkdir -p ~/go/src/github.com/arturoeanton
cd ~/go/src/github.com/arturoeanton

# Clonar repositorio
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Configurar Git
git config user.name "Tu Nombre"
git config user.email "tu.email@ejemplo.com"

# Configurar Go modules
go mod tidy
```

#### Configuración de VS Code
```json
{
  "go.toolsManagement.checkForUpdates": "local",
  "go.useLanguageServer": true,
  "go.gopath": "~/go",
  "go.goroot": "/usr/local/go",
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.testFlags": ["-v"],
  "go.coverageDecorator": {
    "type": "gutter",
    "coveredHighlightColor": "rgba(64,128,128,0.5)",
    "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
  }
}
```

#### Extensiones Recomendadas para VS Code
```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-json",
    "humao.rest-client",
    "redhat.vscode-yaml",
    "ms-vscode.hexeditor",
    "eamodio.gitlens"
  ]
}
```

### Base de Datos de Desarrollo

#### Docker Compose para Desarrollo
```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  couchbase:
    image: couchbase:community-7.0.0
    container_name: goscim-couchbase-dev
    ports:
      - "8091-8094:8091-8094"
      - "11210:11210"
    environment:
      - CLUSTER_NAME=goscim-dev
      - CLUSTER_USERNAME=Administrator
      - CLUSTER_PASSWORD=admin123
    volumes:
      - couchbase-data:/opt/couchbase/var
      - ./scripts/init-couchbase.sh:/opt/couchbase/init/init-couchbase.sh
    networks:
      - goscim-network

  redis:
    image: redis:7-alpine
    container_name: goscim-redis-dev
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - goscim-network

volumes:
  couchbase-data:
  redis-data:

networks:
  goscim-network:
    driver: bridge
```

#### Script de Inicialización de Couchbase
```bash
#!/bin/bash
# scripts/init-couchbase.sh

set -euo pipefail

echo "Waiting for Couchbase to start..."
sleep 30

# Configurar cluster
couchbase-cli cluster-init \
    --cluster localhost:8091 \
    --cluster-username Administrator \
    --cluster-password admin123 \
    --cluster-name goscim-dev \
    --services data,query,index \
    --cluster-ramsize 1024 \
    --cluster-index-ramsize 256

echo "Couchbase initialized successfully"
```

#### Iniciar Entorno de Desarrollo
```bash
# Iniciar servicios
docker-compose -f docker-compose.dev.yml up -d

# Verificar que los servicios estén funcionando
docker-compose -f docker-compose.dev.yml ps

# Ver logs
docker-compose -f docker-compose.dev.yml logs -f couchbase
```

## Estructura del Código

### Organización de Directorios
```
goscim/
├── main.go                    # Punto de entrada
├── go.mod                     # Dependencias Go
├── go.sum                     # Checksums de dependencias
├── .env.example              # Variables de entorno ejemplo
├── .gitignore                # Archivos a ignorar en Git
├── README.md                 # Documentación básica
├── CLAUDE.md                 # Guía para Claude Code
│
├── scim/                     # Paquete principal SCIM
│   ├── config.go            # Gestión de configuración
│   ├── types.go             # Definición de tipos
│   ├── couchbase.go         # Integración con DB
│   ├── discovery.go         # Endpoints de descubrimiento
│   ├── error.go             # Manejo de errores
│   ├── meta.go              # Metadatos de recursos
│   ├── validate.go          # Validación de esquemas
│   ├── validate_role.go     # Validación de roles
│   ├── contains.go          # Utilidades de búsqueda
│   ├── op_create.go         # Operación CREATE
│   ├── op_read.go           # Operación READ
│   ├── op_update.go         # Operación UPDATE (PATCH)
│   ├── op_replace.go        # Operación REPLACE (PUT)
│   ├── op_delete.go         # Operación DELETE
│   ├── op_search.go         # Operación SEARCH
│   ├── op_bulk.go           # Operaciones BULK (WIP)
│   └── parser/              # Parser ANTLR para filtros
│       ├── *.go            # Archivos generados
│       └── parser_test.go   # Tests del parser
│
├── config/                   # Configuraciones JSON
│   ├── schemas/             # Esquemas SCIM
│   ├── resourceType/        # Tipos de recursos
│   ├── bucketSettings/      # Configuración de buckets
│   └── serviceProviderConfig/
│
├── httpexamples/            # Ejemplos de uso HTTP
│   ├── create.http
│   └── filter.http
│
├── scripts/                 # Scripts de utilidad
│   ├── build.sh
│   ├── test.sh
│   ├── generate-parser.sh
│   └── setup-dev.sh
│
├── test/                    # Tests de integración
│   ├── integration/
│   ├── fixtures/
│   └── utils/
│
└── doc/                     # Documentación
    ├── es/                  # Documentación en español
    ├── api/                 # Documentación de API
    └── diagrams/            # Diagramas arquitectónicos
```

### Convenciones de Código

#### Naming Conventions
```go
// Públicas: PascalCase
type UserResource struct {}
func CreateUser() {}

// Privadas: camelCase
type userValidator struct {}
func validateUser() {}

// Constantes: UPPER_SNAKE_CASE
const MAX_RESULTS_PER_PAGE = 200
const DEFAULT_TIMEOUT = 30 * time.Second

// Interfaces: terminan en -er
type UserCreator interface {
    CreateUser(user User) error
}

// Errores: empiezan con Err
var ErrUserNotFound = errors.New("user not found")
var ErrInvalidSchema = errors.New("invalid schema")
```

#### Comentarios y Documentación
```go
// Package scim implements SCIM 2.0 protocol for identity management.
//
// This package provides a complete implementation of the System for 
// Cross-domain Identity Management (SCIM) specification, including
// support for Users, Groups, and custom resource types.
package scim

// UserResource represents a SCIM User resource as defined in RFC 7643.
//
// Example usage:
//   user := &UserResource{
//       UserName: "john.doe@example.com",
//       Name: Name{
//           FamilyName: "Doe",
//           GivenName:  "John",
//       },
//   }
type UserResource struct {
    // UserName is the unique identifier for the User.
    // This field is required and must be unique across the system.
    UserName string `json:"userName" validate:"required,email"`
    
    // Name contains the components of the user's real name.
    Name Name `json:"name,omitempty"`
}

// CreateUser creates a new user in the system.
//
// Parameters:
//   - ctx: Context for the operation
//   - user: User resource to create
//
// Returns:
//   - Created user with generated ID and metadata
//   - Error if creation fails
//
// Example:
//   createdUser, err := CreateUser(ctx, userResource)
//   if err != nil {
//       return fmt.Errorf("failed to create user: %w", err)
//   }
func CreateUser(ctx context.Context, user UserResource) (*UserResource, error) {
    // Implementation...
}
```

## Desarrollo de Nuevas Funcionalidades

### Agregar un Nuevo Tipo de Recurso

#### 1. Definir el Esquema JSON
```json
// config/schemas/urn+ietf+params+scim+schemas+custom+2.0+Device.json
{
  "id": "urn:ietf:params:scim:schemas:custom:2.0:Device",
  "name": "Device",
  "description": "Device resource for BYOD management",
  "attributes": [
    {
      "name": "deviceId",
      "type": "string",
      "required": true,
      "uniqueness": "server",
      "description": "Unique identifier for the device"
    },
    {
      "name": "deviceType",
      "type": "string",
      "required": true,
      "canonicalValues": ["mobile", "tablet", "laptop", "desktop"],
      "description": "Type of device"
    },
    {
      "name": "owner",
      "type": "complex",
      "description": "Device owner information",
      "subAttributes": [
        {
          "name": "value",
          "type": "string",
          "description": "User ID of device owner"
        },
        {
          "name": "$ref",
          "type": "reference",
          "referenceTypes": ["User"],
          "description": "Reference to User resource"
        }
      ]
    },
    {
      "name": "isManaged",
      "type": "boolean",
      "description": "Whether device is managed by organization"
    }
  ]
}
```

#### 2. Definir el Resource Type
```json
// config/resourceType/Device.json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:ResourceType"],
  "id": "Device",
  "name": "Device",
  "endpoint": "/Devices",
  "description": "Device resource for BYOD management",
  "schema": "urn:ietf:params:scim:schemas:custom:2.0:Device",
  "meta": {
    "location": "/v2/ResourceTypes/Device",
    "resourceType": "ResourceType"
  }
}
```

#### 3. Configurar Bucket (Opcional)
```json
// config/bucketSettings/Device.json
{
  "ram_quota_mb": 512,
  "num_replicas": 1,
  "bucket_type": "couchbase",
  "flush_enabled": false
}
```

#### 4. Agregar Validaciones Personalizadas
```go
// scim/validate_device.go
package scim

import (
    "fmt"
    "strings"
)

// ValidateDevice performs custom validation for Device resources
func ValidateDevice(device map[string]interface{}) error {
    // Validar deviceId format
    if deviceId, ok := device["deviceId"].(string); ok {
        if !isValidDeviceId(deviceId) {
            return fmt.Errorf("deviceId must be in format ORG-TYPE-SERIAL")
        }
    }
    
    // Validar que dispositivos móviles tengan owner
    if deviceType, ok := device["deviceType"].(string); ok {
        if deviceType == "mobile" || deviceType == "tablet" {
            if _, hasOwner := device["owner"]; !hasOwner {
                return fmt.Errorf("mobile and tablet devices must have an owner")
            }
        }
    }
    
    return nil
}

func isValidDeviceId(deviceId string) bool {
    parts := strings.Split(deviceId, "-")
    return len(parts) == 3 && len(parts[0]) > 0 && len(parts[1]) > 0 && len(parts[2]) > 0
}
```

#### 5. Registrar Validación Personalizada
```go
// scim/validate.go - agregar a la función existente
func ValidateResourceData(resourceType string, data map[string]interface{}) error {
    // Validación de esquema existente...
    
    // Validaciones personalizadas por tipo de recurso
    switch resourceType {
    case "Device":
        return ValidateDevice(data)
    case "User":
        return ValidateUser(data)
    // Agregar más tipos según sea necesario
    }
    
    return nil
}
```

### Implementar una Nueva Operación

#### Ejemplo: Operación de Activación de Dispositivo
```go
// scim/op_device_activate.go
package scim

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
)

// DeviceActivateRequest representa la solicitud de activación
type DeviceActivateRequest struct {
    ActivationCode string `json:"activationCode" validate:"required"`
    UserReference  string `json:"userRef" validate:"required"`
}

// ActivateDevice maneja la activación de dispositivos
func ActivateDevice(c *gin.Context) {
    deviceId := c.Param("id")
    
    var request DeviceActivateRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        MakeError(c, http.StatusBadRequest, "Invalid request body")
        return
    }
    
    // Validar entrada
    if err := validateStruct(request); err != nil {
        MakeError(c, http.StatusBadRequest, err.Error())
        return
    }
    
    // Obtener dispositivo actual
    device, err := getDeviceById(deviceId)
    if err != nil {
        MakeError(c, http.StatusNotFound, "Device not found")
        return
    }
    
    // Validar código de activación
    if !isValidActivationCode(request.ActivationCode, deviceId) {
        MakeError(c, http.StatusBadRequest, "Invalid activation code")
        return
    }
    
    // Actualizar dispositivo
    device["isManaged"] = true
    device["owner"] = map[string]interface{}{
        "value": request.UserReference,
        "$ref":  "/scim/v2/Users/" + request.UserReference,
    }
    device["activatedAt"] = time.Now().Format(time.RFC3339)
    
    // Guardar cambios
    if err := saveDevice(deviceId, device); err != nil {
        MakeError(c, http.StatusInternalServerError, "Failed to activate device")
        return
    }
    
    // Log de auditoría
    logSecurityEvent(c, "DEVICE_ACTIVATED", true, 
        fmt.Sprintf("Device %s activated by user %s", deviceId, request.UserReference))
    
    c.JSON(http.StatusOK, device)
}

// Registrar endpoint en main.go
// r.POST("/scim/v2/Devices/:id/activate", scim.ActivateDevice)
```

## Testing

### Configuración de Tests

#### Test Setup
```go
// test/setup_test.go
package test

import (
    "os"
    "testing"
    
    "github.com/arturoeanton/goscim/scim"
)

func TestMain(m *testing.M) {
    // Setup test environment
    setupTestDB()
    setupTestConfig()
    
    // Run tests
    code := m.Run()
    
    // Cleanup
    cleanupTestDB()
    
    os.Exit(code)
}

func setupTestDB() {
    os.Setenv("SCIM_ADMIN_USER", "Administrator")
    os.Setenv("SCIM_ADMIN_PASSWORD", "admin123")
    os.Setenv("SCIM_COUCHBASE_URL", "localhost")
    
    scim.InitDB()
}

func setupTestConfig() {
    scim.ReadResourceType("../config", nil)
}

func cleanupTestDB() {
    // Limpiar datos de test
}
```

#### Unit Tests
```go
// scim/validate_test.go
package scim

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestValidateUser(t *testing.T) {
    tests := []struct {
        name    string
        user    map[string]interface{}
        wantErr bool
    }{
        {
            name: "valid user",
            user: map[string]interface{}{
                "userName": "test@example.com",
                "name": map[string]interface{}{
                    "familyName": "Test",
                    "givenName":  "User",
                },
                "active": true,
            },
            wantErr: false,
        },
        {
            name: "missing userName",
            user: map[string]interface{}{
                "name": map[string]interface{}{
                    "familyName": "Test",
                },
            },
            wantErr: true,
        },
        {
            name: "invalid email format",
            user: map[string]interface{}{
                "userName": "invalid-email",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateUser(tt.user)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestDeviceValidation(t *testing.T) {
    device := map[string]interface{}{
        "deviceId":   "CORP-MOBILE-001",
        "deviceType": "mobile",
        "owner": map[string]interface{}{
            "value": "user123",
        },
    }
    
    err := ValidateDevice(device)
    assert.NoError(t, err)
}
```

#### Integration Tests
```go
// test/integration/user_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/arturoeanton/goscim/scim"
)

func TestCreateUser(t *testing.T) {
    router := gin.New()
    router.POST("/scim/v2/Users", scim.Create("/Users"))
    
    user := map[string]interface{}{
        "schemas":  []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
        "userName": "test@example.com",
        "name": map[string]interface{}{
            "familyName": "Test",
            "givenName":  "User",
        },
        "active": true,
    }
    
    jsonData, _ := json.Marshal(user)
    req, _ := http.NewRequest("POST", "/scim/v2/Users", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.NotEmpty(t, response["id"])
    assert.Equal(t, "test@example.com", response["userName"])
}

func TestSearchUsers(t *testing.T) {
    router := gin.New()
    router.GET("/scim/v2/Users", scim.Search("/Users"))
    
    req, _ := http.NewRequest("GET", "/scim/v2/Users?filter=userName sw \"test\"", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Contains(t, response["schemas"], "urn:ietf:params:scim:api:messages:2.0:ListResponse")
}
```

#### Ejecutar Tests
```bash
# Tests unitarios
go test ./scim -v

# Tests de integración
go test ./test/integration -v

# Tests con coverage
go test -cover ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Debugging y Profiling

### Configuración de Debug

#### Habilitar Debug Mode
```bash
export SCIM_LOG_LEVEL=debug
export GIN_MODE=debug
go run main.go
```

#### Usar Delve Debugger
```bash
# Instalar Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug con Delve
dlv debug main.go

# En Delve:
(dlv) break main.main
(dlv) continue
(dlv) step
(dlv) print variable_name
```

#### VS Code Debug Configuration
```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch GoSCIM",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/main.go",
            "env": {
                "SCIM_ADMIN_USER": "Administrator",
                "SCIM_ADMIN_PASSWORD": "admin123",
                "SCIM_LOG_LEVEL": "debug",
                "GIN_MODE": "debug"
            },
            "args": []
        },
        {
            "name": "Attach to Process",
            "type": "go",
            "request": "attach",
            "mode": "local",
            "processId": 0
        }
    ]
}
```

### Profiling de Performance

#### Habilitar pprof
```go
// main.go - agregar al inicio
import _ "net/http/pprof"

func main() {
    // Habilitar profiling en development
    if os.Getenv("SCIM_ENABLE_PPROF") == "true" {
        go func() {
            log.Println("Starting pprof server on :6060")
            log.Println(http.ListenAndServe("localhost:6060", nil))
        }()
    }
    
    // Resto de la aplicación...
}
```

#### Analizar Profiles
```bash
# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profiling
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Dentro de pprof:
(pprof) top10
(pprof) list function_name
(pprof) web
```

## Scripts de Desarrollo

### Build Script
```bash
#!/bin/bash
# scripts/build.sh

set -euo pipefail

VERSION=${1:-dev}
BUILD_DIR="./build"
BINARY_NAME="goscim"

echo "Building GoSCIM version $VERSION..."

# Crear directorio de build
mkdir -p $BUILD_DIR

# Build flags
BUILD_FLAGS="-ldflags=-X main.Version=$VERSION -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Build para diferentes plataformas
GOOS=linux GOARCH=amd64 go build $BUILD_FLAGS -o $BUILD_DIR/${BINARY_NAME}-linux-amd64 .
GOOS=darwin GOARCH=amd64 go build $BUILD_FLAGS -o $BUILD_DIR/${BINARY_NAME}-darwin-amd64 .
GOOS=windows GOARCH=amd64 go build $BUILD_FLAGS -o $BUILD_DIR/${BINARY_NAME}-windows-amd64.exe .

echo "Build completed successfully"
ls -la $BUILD_DIR/
```

### Test Script
```bash
#!/bin/bash
# scripts/test.sh

set -euo pipefail

echo "Running GoSCIM tests..."

# Lint
echo "Running linter..."
golangci-lint run

# Unit tests
echo "Running unit tests..."
go test -v -race -cover ./scim/...

# Integration tests
echo "Running integration tests..."
go test -v ./test/integration/...

# Generate coverage report
echo "Generating coverage report..."
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

echo "Tests completed successfully"
echo "Coverage report: coverage.html"
```

### Development Setup Script
```bash
#!/bin/bash
# scripts/setup-dev.sh

set -euo pipefail

echo "Setting up GoSCIM development environment..."

# Verificar dependencias
command -v go >/dev/null 2>&1 || { echo "Go is not installed"; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "Docker is not installed"; exit 1; }

# Instalar herramientas de desarrollo
echo "Installing development tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install golang.org/x/tools/cmd/goimports@latest

# Configurar pre-commit hooks
echo "Setting up pre-commit hooks..."
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/sh
echo "Running pre-commit checks..."
make lint
make test
EOF
chmod +x .git/hooks/pre-commit

# Iniciar servicios de desarrollo
echo "Starting development services..."
docker-compose -f docker-compose.dev.yml up -d

# Configurar variables de entorno
if [ ! -f .env ]; then
    cp .env.example .env
    echo "Created .env file from template"
fi

echo "Development environment setup completed!"
echo "Run 'go run main.go' to start the development server"
```

## Makefile

```makefile
# Makefile
.PHONY: build test lint clean run dev setup

# Variables
BINARY_NAME=goscim
VERSION?=dev
BUILD_DIR=./build

# Default target
all: lint test build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) .

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean

# Run development server
run:
	@echo "Starting development server..."
	@go run main.go

# Setup development environment
dev:
	@echo "Setting up development environment..."
	@docker-compose -f docker-compose.dev.yml up -d
	@go mod tidy

# Initial setup
setup:
	@echo "Initial setup..."
	@scripts/setup-dev.sh

# Generate parser (requires ANTLR)
parser:
	@echo "Generating SCIM filter parser..."
	@scripts/generate-parser.sh

# Database operations
db-reset:
	@echo "Resetting development database..."
	@docker-compose -f docker-compose.dev.yml down -v
	@docker-compose -f docker-compose.dev.yml up -d

# Coverage report
coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
```

Esta guía proporciona todo lo necesario para comenzar a desarrollar con GoSCIM, desde la configuración del entorno hasta las mejores prácticas de desarrollo y testing.