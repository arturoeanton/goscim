# Arquitectura del Sistema GoSCIM

## Visión General de la Arquitectura

GoSCIM implementa una arquitectura modular basada en el patrón MVC (Model-View-Controller) adaptado para servicios REST, siguiendo los principios de Clean Architecture y las especificaciones SCIM 2.0.

### Principios Arquitectónicos

1. **Separación de Responsabilidades**: Cada módulo tiene una responsabilidad específica y bien definida
2. **Inversión de Dependencias**: Las capas superiores no dependen de las inferiores
3. **Configuración Externa**: Toda la configuración es externa al código
4. **Escalabilidad Horizontal**: Diseño stateless para permitir múltiples instancias
5. **Extensibilidad**: Fácil adición de nuevos tipos de recursos y esquemas

## Arquitectura de Alto Nivel

```
┌─────────────────────────────────────────────────────────────┐
│                    Cliente / Integración                    │
│          (AD, LDAP, Aplicaciones, APIs externas)           │
└─────────────────────┬───────────────────────────────────────┘
                      │ HTTPS/TLS
                      │ SCIM 2.0 REST API
┌─────────────────────▼───────────────────────────────────────┐
│                 Load Balancer (Nginx)                      │
│                 - SSL Termination                          │
│                 - Rate Limiting                            │
│                 - Health Checks                            │
└─────────────────────┬───────────────────────────────────────┘
                      │
           ┌──────────┼──────────┐
           │          │          │
┌──────────▼──┐ ┌─────▼─────┐ ┌──▼──────────┐
│  GoSCIM     │ │  GoSCIM   │ │  GoSCIM     │
│  Instance 1 │ │  Instance │ │  Instance N │
│             │ │     2     │ │             │
└──────────┬──┘ └─────┬─────┘ └──┬──────────┘
           │          │          │
           └──────────┼──────────┘
                      │ N1QL/SDK
┌─────────────────────▼───────────────────────────────────────┐
│                Couchbase Cluster                           │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          │
│  │   Node 1    │ │   Node 2    │ │   Node N    │          │
│  │ Data+Query+ │ │ Data+Query+ │ │ Data+Query+ │          │
│  │   Index     │ │   Index     │ │   Index     │          │
│  └─────────────┘ └─────────────┘ └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

## Arquitectura de la Aplicación

### Estructura de Directorios

```
goscim/
├── main.go                 # Punto de entrada
├── scim/                   # Núcleo del sistema SCIM
│   ├── config.go          # Gestión de configuración
│   ├── types.go           # Definición de tipos de datos
│   ├── couchbase.go       # Integración con base de datos
│   ├── discovery.go       # Endpoints de descubrimiento
│   ├── error.go           # Manejo de errores
│   ├── meta.go            # Metadatos de recursos
│   ├── validate.go        # Validación de esquemas
│   ├── validate_role.go   # Validación de roles
│   ├── contains.go        # Utilidades de búsqueda
│   ├── op_*.go           # Operaciones SCIM
│   └── parser/           # Parser ANTLR para filtros
│       ├── *.go         # Código generado por ANTLR
│       └── parser_test.go
├── config/               # Configuraciones JSON
│   ├── schemas/         # Esquemas SCIM
│   ├── resourceType/    # Tipos de recursos
│   ├── bucketSettings/  # Configuración de buckets
│   └── serviceProviderConfig/
├── httpexamples/        # Ejemplos de uso
└── doc/                 # Documentación
```

### Capas Arquitectónicas

#### 1. Capa de Presentación (HTTP/REST)
**Responsabilidad**: Manejo de requests HTTP y responses
**Componentes**:
- Router Gin para manejo de rutas
- Middleware de autenticación y autorización
- Serialización/deserialización JSON
- Manejo de códigos de estado HTTP

```go
// main.go - Configuración de rutas
r := gin.Default()
r.POST(PREFIX+resourceType.Endpoint, Create(resourceType.Endpoint))
r.GET(PREFIX+resourceType.Endpoint+"/:id", Read(resourceType.Endpoint))
r.PUT(PREFIX+resourceType.Endpoint+"/:id", Replace(resourceType.Endpoint))
r.DELETE(PREFIX+resourceType.Endpoint+"/:id", Delete(resourceType.Endpoint))
r.PATCH(PREFIX+resourceType.Endpoint+"/:id", Update(resourceType.Endpoint))
r.GET(PREFIX+resourceType.Endpoint, Search(resourceType.Endpoint))
```

#### 2. Capa de Lógica de Negocio (SCIM Operations)
**Responsabilidad**: Implementación de operaciones SCIM 2.0
**Componentes**:
- Operaciones CRUD (Create, Read, Update, Delete)
- Operaciones de búsqueda y filtrado
- Validación de esquemas y datos
- Control de acceso basado en roles

```go
// Estructura típica de operación
func Create(resource string) func(c *gin.Context) {
    return func(c *gin.Context) {
        // 1. Parsear request
        // 2. Validar esquema
        // 3. Validar permisos
        // 4. Procesar datos
        // 5. Guardar en base de datos
        // 6. Retornar response
    }
}
```

#### 3. Capa de Parser y Filtros
**Responsabilidad**: Procesamiento de filtros SCIM y consultas
**Componentes**:
- Parser ANTLR para gramática SCIM
- Conversión de filtros SCIM a N1QL
- Optimización de consultas

```go
// scim/parser/ - Conversión de filtros
func FilterToN1QL(bucket string, filter string) (string, string) {
    // Parsea filtro SCIM usando ANTLR
    // Convierte a consulta N1QL
    // Retorna query para datos y query para conteo
}
```

#### 4. Capa de Acceso a Datos (Data Access)
**Responsabilidad**: Interacción con Couchbase
**Componentes**:
- Conexión y configuración de Couchbase
- Gestión de buckets y colecciones
- Ejecución de consultas N1QL
- Manejo de transacciones

```go
// scim/couchbase.go - Gestión de base de datos
func InitDB() {
    // Configurar conexión
    // Autenticar
    // Validar conectividad
}

func CreateBucket(name string) {
    // Crear bucket si no existe
    // Configurar índices
    // Aplicar configuración personalizada
}
```

#### 5. Capa de Configuración (Configuration)
**Responsabilidad**: Gestión de esquemas y configuración
**Componentes**:
- Carga de esquemas JSON
- Validación de configuración
- Registro dinámico de endpoints

```go
// scim/config.go - Gestión de configuración
func ReadResourceType(folderConfig string, r *gin.Engine) {
    // Cargar esquemas desde archivos JSON
    // Validar estructura
    // Registrar endpoints dinámicamente
}
```

## Patrones de Diseño Implementados

### 1. Repository Pattern
Aunque no está explícitamente implementado, la capa de acceso a datos actúa como un repository:

```go
// Patrón implícito en operaciones
type ResourceRepository interface {
    Create(resource Resource) error
    Read(id string) (Resource, error)
    Update(id string, resource Resource) error
    Delete(id string) error
    Search(filter string) ([]Resource, error)
}
```

### 2. Strategy Pattern
Para diferentes tipos de operaciones:

```go
// Diferentes estrategias de operación
func Create(resourceType string) gin.HandlerFunc { /* ... */ }
func Update(resourceType string) gin.HandlerFunc { /* ... */ }
func Delete(resourceType string) gin.HandlerFunc { /* ... */ }
```

### 3. Builder Pattern
Para construcción de consultas N1QL:

```go
// Parser construye consultas paso a paso
query := "SELECT * FROM `" + bucket + "`"
if filter != "" {
    query += " WHERE " + processFilter(filter)
}
query += " ORDER BY " + sortBy + " " + sortOrder
```

### 4. Factory Pattern
Para creación de recursos basada en tipo:

```go
// Creación dinámica basada en resource type
resourceType := Resources[endpoint]
bucket := resourceType.Name
schema := Schemas[resourceType.Schema]
```

## Flujo de Datos

### 1. Request de Creación de Usuario

```
1. Cliente → HTTP POST /scim/v2/Users
2. Gin Router → Create() handler
3. Validar JSON y esquema SCIM
4. Validar permisos de escritura
5. Generar ID y metadatos
6. Convertir a formato interno
7. Guardar en Couchbase bucket "User"
8. Retornar response con código 201
```

### 2. Request de Búsqueda con Filtros

```
1. Cliente → HTTP GET /scim/v2/Users?filter=userName sw "admin"
2. Gin Router → Search() handler
3. Parser ANTLR procesa filtro SCIM
4. Convertir a consulta N1QL
5. Ejecutar consulta en Couchbase
6. Aplicar validación de permisos de lectura
7. Formatear resultados como lista SCIM
8. Retornar response con código 200
```

### 3. Request de Actualización PATCH

```
1. Cliente → HTTP PATCH /scim/v2/Users/123
2. Gin Router → Update() handler
3. Validar operaciones PATCH
4. Obtener recurso actual
5. Aplicar operaciones (add/remove/replace)
6. Validar esquema resultante
7. Guardar cambios en Couchbase
8. Retornar recurso actualizado
```

## Gestión de Esquemas

### Esquemas Dinámicos

GoSCIM permite definir esquemas completamente a través de archivos JSON sin modificar código:

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
      "uniqueness": "server",
      "$reader": ["admin", "user"],
      "$writer": ["admin"]
    }
  ]
}
```

### Extensiones de Esquema

Soporte completo para extensiones empresariales:

```json
{
  "id": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
  "attributes": [
    {
      "name": "employeeNumber",
      "type": "string",
      "uniqueness": "server"
    }
  ]
}
```

### Tipos de Recursos Personalizados

Capacidad de definir nuevos tipos de recursos:

```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:ResourceType"],
  "id": "CustomResource",
  "name": "CustomResource",
  "endpoint": "/CustomResources",
  "schema": "urn:ietf:params:scim:schemas:custom:2.0:CustomResource"
}
```

## Gestión de Estado

### Stateless Design

GoSCIM está diseñado para ser completamente stateless:

- **Sin sesiones del servidor**: Toda la información necesaria está en el token
- **Sin cache local**: Todos los datos están en Couchbase
- **Instancias intercambiables**: Cualquier instancia puede manejar cualquier request

### Gestión de Metadatos

Cada recurso incluye metadatos SCIM estándar:

```go
type Meta struct {
    ResourceType string `json:"resourceType"`
    Created      string `json:"created"`
    LastModified string `json:"lastModified"`
    Version      string `json:"version"`
    Location     string `json:"location"`
}
```

## Seguridad en la Arquitectura

### Capas de Seguridad

1. **Transporte**: TLS 1.3 obligatorio
2. **Autenticación**: OAuth 2.0/JWT tokens
3. **Autorización**: RBAC a nivel de atributo
4. **Validación**: Sanitización de entrada
5. **Auditoría**: Logging de acciones

### Control de Acceso

```go
// Validación de permisos por atributo
func ValidateReadRole(roles []string, resourceType ResoruceType, element map[string]interface{}) {
    schema := Schemas[resourceType.Schema]
    for _, attribute := range schema.Attributes {
        if attribute.Read != nil {
            // Validar si el rol puede leer este atributo
        }
    }
}
```

## Escalabilidad y Performance

### Escalabilidad Horizontal

- **Load Balancing**: Múltiples instancias detrás de load balancer
- **Database Clustering**: Couchbase cluster distribuido
- **Stateless Design**: Fácil auto-scaling

### Optimizaciones

1. **Índices de Base de Datos**: Índices automáticos en Couchbase
2. **Consultas Optimizadas**: Conversión eficiente de filtros SCIM a N1QL
3. **Paginación**: Límites en resultados para evitar sobrecarga
4. **Connection Pooling**: Reutilización de conexiones a base de datos

### Métricas de Performance

- **Response Time**: < 100ms para operaciones simples
- **Throughput**: 1000+ requests/segundo por instancia
- **Concurrency**: 100+ conexiones simultáneas
- **Memory Usage**: < 512MB por instancia

## Monitoreo y Observabilidad

### Logging

```go
// Logging estructurado
log.Println("Operation completed", 
    "operation", "create",
    "resource", "User",
    "id", userId,
    "duration", duration)
```

### Métricas

- Request count por endpoint
- Response time percentiles
- Error rates
- Database connection status
- Memory y CPU usage

### Health Checks

```go
// Health check endpoint
func healthCheck(c *gin.Context) {
    // Verificar conectividad a Couchbase
    // Verificar estado de la aplicación
    // Retornar status de salud
}
```

## Extensibilidad

### Agregar Nuevos Recursos

1. Crear esquema JSON en `config/schemas/`
2. Crear resource type en `config/resourceType/`
3. Reiniciar aplicación
4. Endpoints se registran automáticamente

### Agregar Nuevas Operaciones

1. Crear handler function
2. Registrar ruta en main.go
3. Implementar lógica específica

### Customizar Validaciones

```go
// Validaciones personalizadas por tipo de recurso
func customValidation(resourceType string, data map[string]interface{}) error {
    switch resourceType {
    case "User":
        return validateUser(data)
    case "Group":
        return validateGroup(data)
    default:
        return validateGeneric(data)
    }
}
```

## Consideraciones de Diseño

### Trade-offs Arquitectónicos

1. **Simplicidad vs Flexibilidad**: Se eligió flexibilidad con esquemas JSON dinámicos
2. **Performance vs Funcionalidad**: Se priorizó funcionalidad SCIM completa
3. **Monolito vs Microservicios**: Se eligió monolito modular para simplicidad

### Decisiones Técnicas

1. **Go como lenguaje**: Performance y simplicidad
2. **Gin como framework**: Velocidad y facilidad de uso
3. **Couchbase como DB**: Flexibilidad de esquemas y escalabilidad
4. **ANTLR para parser**: Robustez y mantenibilidad

### Limitaciones Actuales

1. **Single-tenant**: No hay soporte multi-tenancy nativo
2. **Cache limitado**: No hay cache distribuido implementado
3. **Bulk operations**: Implementación parcial
4. **Event streaming**: No hay notificaciones en tiempo real

## Roadmap Arquitectónico

### Corto Plazo
- Implementar cache Redis
- Completar operaciones Bulk
- Mejorar observabilidad

### Medio Plazo
- Multi-tenancy
- Event streaming
- Microservicios opcionales

### Largo Plazo
- Cloud-native deployment
- Auto-scaling avanzado
- Machine learning para optimizaciones