# Documentación Técnica GoSCIM

## Descripción General

**GoSCIM** es una implementación completa del protocolo SCIM 2.0 (System for Cross-domain Identity Management) desarrollada en Go. Proporciona una solución robusta y escalable para la gestión de identidades en entornos distribuidos, especialmente diseñada para integrar sistemas de identidad heterogéneos.

## Características Principales

### Cumplimiento SCIM 2.0
- ✅ Operaciones CRUD completas (Create, Read, Update, Delete)
- ✅ Búsqueda avanzada con filtros SCIM
- ✅ Paginación y ordenamiento
- ✅ Esquemas extensibles y personalizables
- ✅ Soporte para múltiples tipos de recursos
- ✅ Operaciones Bulk (en desarrollo)

### Arquitectura Técnica
- **Lenguaje**: Go 1.16+
- **Framework Web**: Gin (alta performance)
- **Base de Datos**: Couchbase (NoSQL distribuida)
- **Parser**: ANTLR v4 para filtros SCIM
- **Formato de Datos**: JSON nativo

## Arquitectura del Sistema

### Componentes Principales

#### 1. Núcleo del Servidor (`main.go`)
```go
// Inicialización del servidor
func main() {
    scim.InitDB()                    // Conexión a Couchbase
    r := gin.Default()               // Router HTTP
    scim.ReadResourceType(config, r) // Carga dinámica de esquemas
    r.Run(port)                      // Servidor HTTP
}
```

#### 2. Gestión de Configuración (`scim/config.go`)
- **Carga dinámica de esquemas** desde archivos JSON
- **Registro automático de endpoints** basado en tipos de recursos
- **Validación de esquemas** al inicio del servidor

#### 3. Integración con Base de Datos (`scim/couchbase.go`)
- **Conexión segura** con autenticación
- **Creación automática de buckets** por tipo de recurso
- **Configuración personalizable** de buckets
- **Índices primarios automáticos**

#### 4. Parser de Filtros (`scim/parser/`)
- **Gramática ANTLR** para filtros SCIM
- **Conversión automática** a consultas N1QL
- **Soporte completo** para operadores SCIM

### Operaciones SCIM

#### Create (POST)
```http
POST /scim/v2/Users
Content-Type: application/json

{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "userName": "usuario@ejemplo.com",
  "name": {
    "familyName": "Apellido",
    "givenName": "Nombre"
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
      "value": "NuevoApellido"
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

## Configuración del Sistema

### Variables de Entorno

#### Obligatorias
```bash
export SCIM_ADMIN_USER="Administrator"     # Usuario admin de Couchbase
export SCIM_ADMIN_PASSWORD="admin123"     # Contraseña admin de Couchbase
```

#### Opcionales
```bash
export SCIM_COUCHBASE_URL="localhost"     # URL del servidor Couchbase
export SCIM_PORT=":8080"                  # Puerto del servidor SCIM
```

### Estructura de Configuración

```
config/
├── schemas/                    # Definiciones de esquemas SCIM
│   ├── urn+ietf+params+scim+schemas+core+2.0+User.json
│   ├── urn+ietf+params+scim+schemas+core+2.0+Group.json
│   └── urn+ietf+params+scim+schemas+extension+*.json
├── resourceType/              # Tipos de recursos
│   ├── User.json
│   ├── Group.json
│   └── Element.json
├── bucketSettings/            # Configuración de buckets Couchbase
│   ├── User.json
│   ├── Group.json
│   └── Element.json
└── serviceProviderConfig/     # Configuración del proveedor
    └── sp_config.json
```

## Esquemas y Extensiones

### Esquema Base de Usuario
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

### Extensiones Personalizadas
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

## Control de Acceso

### Roles y Permisos
```json
{
  "name": "sensitiveAttribute",
  "type": "string",
  "$reader": ["admin", "hr"],      # Roles que pueden leer
  "$writer": ["admin"],            # Roles que pueden escribir
  "returned": "default"
}
```

### Validación de Roles
```go
// Validación automática en búsquedas
roles := []string{"user", "admin", "superadmin"}
element := ValidateReadRole(roles, resourceType, item)
```

## Filtros SCIM

### Sintaxis Soportada
```
# Comparaciones básicas
userName eq "admin"
name.familyName co "García"
userName sw "admin"
active pr

# Comparaciones temporales
meta.lastModified gt "2023-01-01T00:00:00Z"
meta.lastModified ge "2023-01-01T00:00:00Z"

# Operadores lógicos
title pr and userType eq "Employee"
title pr or userType eq "Intern"
userType eq "Employee" and (emails co "empresa.com" or emails co "empresa.org")
```

### Conversión a N1QL
```go
// Ejemplo de conversión
query, _ := parser.FilterToN1QL("User", 
    "userName eq \"admin\" and active eq true")
// Resultado: SELECT * FROM `User` WHERE `userName` = "admin" AND `active` = true
```

## Instalación y Despliegue

### Requisitos del Sistema

#### Desarrollo
- Go 1.16 o superior
- Couchbase Server 6.0+
- ANTLR 4.7 (para regenerar parser)

#### Producción
- CPU: 2+ cores
- RAM: 4GB+ (dependiendo del volumen)
- Almacenamiento: SSD recomendado
- Red: 1Gbps+ para alta concurrencia

### Instalación Local

#### 1. Clonar Repositorio
```bash
git clone https://github.com/arturoeanton/goscim.git
cd goscim
```

#### 2. Instalar Dependencias
```bash
go mod download
```

#### 3. Configurar Couchbase
```bash
# Ejecutar Couchbase en Docker
docker run -d --name couchbase-scim \
  -p 8091-8094:8091-8094 \
  -p 11210:11210 \
  couchbase:latest

# Configurar cluster via web UI
# http://localhost:8091/ui/index.html
```

#### 4. Configurar Variables de Entorno
```bash
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"
export SCIM_COUCHBASE_URL="localhost"
export SCIM_PORT=":8080"
```

#### 5. Ejecutar Servidor
```bash
go run main.go
```

### Despliegue en Producción

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

## Testing y Desarrollo

### Ejecutar Tests
```bash
# Tests unitarios
go test ./...

# Tests específicos
go test ./scim/parser -v

# Tests con coverage
go test -cover ./...
```

### Ejemplos de Uso HTTP
```bash
# Crear usuario
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "testuser",
    "name": {
      "familyName": "Apellido",
      "givenName": "Nombre"
    }
  }'

# Buscar usuarios
curl "http://localhost:8080/scim/v2/Users?filter=userName sw \"test\""

# Obtener configuración del proveedor
curl http://localhost:8080/ServiceProviderConfig
```

## Monitoreo y Operaciones

### Logs del Sistema
```bash
# Configurar logging estructurado
export SCIM_LOG_LEVEL=debug
export SCIM_LOG_FORMAT=json

# Ejemplo de log
{"level":"info","timestamp":"2023-12-01T10:00:00Z","message":"GoScim v0.1 started"}
{"level":"debug","timestamp":"2023-12-01T10:00:01Z","message":"Bucket created","bucket":"User"}
```

### Métricas Recomendadas
- Requests per second (RPS)
- Response time percentiles
- Error rates por endpoint
- Conexiones activas a Couchbase
- Memoria y CPU usage

### Health Checks
```http
GET /health
{
  "status": "healthy",
  "database": "connected",
  "version": "v0.1"
}
```

## Seguridad

### Consideraciones de Seguridad

#### Autenticación
- Implementar OAuth 2.0 / OpenID Connect
- Soporte para JWT tokens
- Validación de tokens en cada request

#### Autorización
- Control granular basado en roles
- Permisos por recurso y operación
- Audit logs de acceso

#### Comunicación
- TLS 1.3 obligatorio en producción
- Certificados válidos
- HTTP security headers

#### Validación
- Sanitización de entrada
- Validación de esquemas estricta
- Rate limiting por IP/usuario

## Integración con Sistemas Externos

### Proveedores de Identidad
- Active Directory
- LDAP
- OAuth 2.0 providers
- SAML 2.0

### Sistemas de Destino
- Aplicaciones SaaS
- Bases de datos de usuarios
- Sistemas de directorio
- APIs de terceros

## Troubleshooting

### Problemas Comunes

#### Conexión a Couchbase
```bash
# Verificar conectividad
telnet localhost 8091

# Verificar credenciales
curl -u Administrator:admin123 http://localhost:8091/pools
```

#### Errores de Esquema
```bash
# Validar JSON de esquema
jq . config/schemas/schema.json

# Verificar sintaxis
go run main.go --validate-config
```

#### Performance Issues
```bash
# Verificar índices
curl -u admin:pass http://localhost:8091/query/service \
  -d 'statement=SELECT * FROM system:indexes WHERE keyspace_id="User"'
```

## Roadmap de Desarrollo

### Fase 1: Estabilización
- Implementar autenticación robusta
- Completar suite de tests
- Mejorar logging y monitoreo

### Fase 2: Escalabilidad
- Soporte para clustering
- Cache distribuido
- Optimizaciones de performance

### Fase 3: Funcionalidades Avanzadas
- Operaciones Bulk completas
- Webhooks y notificaciones
- Dashboard de administración

## Contribución

### Desarrollo de Nuevos Recursos
1. Crear esquema JSON en `config/schemas/`
2. Definir tipo de recurso en `config/resourceType/`
3. Configurar bucket en `config/bucketSettings/`
4. Reiniciar servidor para cargar cambios

### Regeneración del Parser
```bash
# Instalar ANTLR
wget http://www.antlr.org/download/antlr-4.7-complete.jar
alias antlr='java -jar $PWD/antlr-4.7-complete.jar'

# Regenerar parser
antlr -Dlanguage=Go -o scim/parser ScimFilter.g4
```

## Soporte y Contacto

Para soporte técnico, reportes de bugs o solicitudes de funcionalidades:
- **Issues**: [GitHub Issues](https://github.com/arturoeanton/goscim/issues)
- **Documentación**: [Wiki del proyecto](https://github.com/arturoeanton/goscim/wiki)
- **Ejemplos**: Directorio `httpexamples/`