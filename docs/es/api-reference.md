# Referencia de API GoSCIM

## Información General

### URL Base
```
http://localhost:8080
https://scim.example.com
```

### Formato de Datos
- **Content-Type**: `application/json` o `application/scim+json`
- **Encoding**: UTF-8
- **Formato de fechas**: ISO 8601 (`2023-12-01T10:30:00Z`)

### Códigos de Estado HTTP

| Código | Descripción | Uso |
|--------|-------------|-----|
| 200 | OK | Operación exitosa |
| 201 | Created | Recurso creado |
| 204 | No Content | Eliminación exitosa |
| 400 | Bad Request | Error en datos de entrada |
| 401 | Unauthorized | Autenticación requerida |
| 403 | Forbidden | Sin permisos |
| 404 | Not Found | Recurso no encontrado |
| 409 | Conflict | Recurso ya existe |
| 500 | Internal Server Error | Error del servidor |

## Endpoints de Descubrimiento

### Service Provider Configuration
Retorna la configuración del proveedor SCIM.

```http
GET /ServiceProviderConfig
Accept: application/json
```

**Respuesta:**
```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"],
  "documentationUri": "https://tools.ietf.org/html/rfc7644",
  "patch": {
    "supported": true
  },
  "bulk": {
    "supported": false,
    "maxOperations": 0,
    "maxPayloadSize": 0
  },
  "filter": {
    "supported": true,
    "maxResults": 200
  },
  "changePassword": {
    "supported": false
  },
  "sort": {
    "supported": true
  },
  "etag": {
    "supported": false
  },
  "authenticationSchemes": [
    {
      "name": "OAuth Bearer Token",
      "description": "Authentication scheme using the OAuth Bearer Token Standard",
      "specUri": "http://www.rfc-editor.org/info/rfc6750",
      "documentationUri": "http://example.com/help/oauth.html",
      "type": "oauthbearertoken",
      "primary": true
    }
  ]
}
```

### Resource Types
Lista los tipos de recursos disponibles.

```http
GET /ResourceTypes
Accept: application/json
```

**Respuesta:**
```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:ListResponse"],
  "totalResults": 3,
  "Resources": [
    {
      "schemas": ["urn:ietf:params:scim:schemas:core:2.0:ResourceType"],
      "id": "User",
      "name": "User",
      "endpoint": "/Users",
      "description": "User Account",
      "schema": "urn:ietf:params:scim:schemas:core:2.0:User",
      "schemaExtensions": [
        {
          "schema": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
          "required": false
        }
      ]
    }
  ]
}
```

### Schemas
Lista los esquemas disponibles.

```http
GET /Schemas
Accept: application/json
```

**Respuesta:**
```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:ListResponse"],
  "totalResults": 2,
  "Resources": [
    {
      "id": "urn:ietf:params:scim:schemas:core:2.0:User",
      "name": "User",
      "description": "User Account",
      "attributes": [...]
    }
  ]
}
```

## Operaciones CRUD

### CREATE - Crear Recurso

#### Crear Usuario
```http
POST /scim/v2/Users
Content-Type: application/json
Authorization: Bearer {token}

{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User",
    "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
  ],
  "userName": "usuario@ejemplo.com",
  "name": {
    "formatted": "Juan Pérez García",
    "familyName": "Pérez García",
    "givenName": "Juan",
    "middleName": "",
    "honorificPrefix": "Sr.",
    "honorificSuffix": ""
  },
  "displayName": "Juan Pérez",
  "nickName": "Juanito",
  "profileUrl": "https://login.example.com/usuarios/juan",
  "title": "Desarrollador Senior",
  "userType": "Employee",
  "preferredLanguage": "es_ES",
  "locale": "es_ES",
  "timezone": "Europe/Madrid",
  "active": true,
  "password": "PasswordSeguro123!",
  "emails": [
    {
      "value": "juan.perez@ejemplo.com",
      "display": "juan.perez@ejemplo.com",
      "type": "work",
      "primary": true
    },
    {
      "value": "juan@personal.com",
      "display": "juan@personal.com", 
      "type": "home",
      "primary": false
    }
  ],
  "phoneNumbers": [
    {
      "value": "+34-91-555-0123",
      "display": "+34 91 555 0123",
      "type": "work",
      "primary": true
    }
  ],
  "addresses": [
    {
      "formatted": "Calle Mayor 123\n28001 Madrid\nEspaña",
      "streetAddress": "Calle Mayor 123",
      "locality": "Madrid",
      "region": "Madrid",
      "postalCode": "28001",
      "country": "España",
      "type": "work"
    }
  ],
  "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
    "employeeNumber": "EMP001",
    "costCenter": "CC-001",
    "organization": "Tecnología",
    "division": "Desarrollo",
    "department": "Backend",
    "manager": {
      "value": "26118915-6090-4610-87e4-49d8ca9f808d",
      "$ref": "/scim/v2/Users/26118915-6090-4610-87e4-49d8ca9f808d",
      "displayName": "María Directora"
    }
  }
}
```

**Respuesta (201 Created):**
```json
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User",
    "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
  ],
  "id": "2819c223-7f76-453a-919d-413861904646",
  "externalId": "usuario@ejemplo.com",
  "userName": "usuario@ejemplo.com",
  "name": {
    "formatted": "Juan Pérez García",
    "familyName": "Pérez García",
    "givenName": "Juan"
  },
  "displayName": "Juan Pérez",
  "active": true,
  "emails": [...],
  "meta": {
    "resourceType": "User",
    "created": "2023-12-01T10:30:00Z",
    "lastModified": "2023-12-01T10:30:00Z",
    "version": "W/\"a330bc54f0671c9\"",
    "location": "/scim/v2/Users/2819c223-7f76-453a-919d-413861904646"
  }
}
```

#### Crear Grupo
```http
POST /scim/v2/Groups
Content-Type: application/json
Authorization: Bearer {token}

{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:Group"],
  "displayName": "Desarrolladores Backend",
  "externalId": "dev-backend-001",
  "members": [
    {
      "value": "2819c223-7f76-453a-919d-413861904646",
      "$ref": "/scim/v2/Users/2819c223-7f76-453a-919d-413861904646",
      "type": "User",
      "display": "Juan Pérez"
    }
  ]
}
```

### READ - Obtener Recurso

#### Obtener Usuario por ID
```http
GET /scim/v2/Users/{id}
Accept: application/json
Authorization: Bearer {token}
```

#### Obtener Usuario con Atributos Específicos
```http
GET /scim/v2/Users/{id}?attributes=userName,name,emails
Accept: application/json
Authorization: Bearer {token}
```

#### Excluir Atributos
```http
GET /scim/v2/Users/{id}?excludedAttributes=addresses,phoneNumbers
Accept: application/json
Authorization: Bearer {token}
```

### SEARCH - Buscar Recursos

#### Búsqueda Básica
```http
GET /scim/v2/Users
Accept: application/json
Authorization: Bearer {token}
```

#### Búsqueda con Filtros
```http
GET /scim/v2/Users?filter=userName eq "juan.perez@ejemplo.com"
Accept: application/json
Authorization: Bearer {token}
```

#### Parámetros de Búsqueda

| Parámetro | Descripción | Ejemplo |
|-----------|-------------|---------|
| `filter` | Expresión de filtro SCIM | `userName sw "admin"` |
| `sortBy` | Campo para ordenar | `userName` |
| `sortOrder` | Orden (ascending/descending) | `descending` |
| `startIndex` | Índice inicial (base 1) | `1` |
| `count` | Número de resultados | `10` |
| `attributes` | Atributos a incluir | `userName,name` |
| `excludedAttributes` | Atributos a excluir | `addresses` |

#### Ejemplos de Filtros

##### Filtros de Igualdad
```http
# Usuario específico
GET /scim/v2/Users?filter=userName eq "juan@ejemplo.com"

# Usuarios activos
GET /scim/v2/Users?filter=active eq true

# Por número de empleado
GET /scim/v2/Users?filter=urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:employeeNumber eq "EMP001"
```

##### Filtros de Texto
```http
# Usuarios que empiezan por "admin"
GET /scim/v2/Users?filter=userName sw "admin"

# Usuarios que contienen "perez"
GET /scim/v2/Users?filter=name.familyName co "perez"

# Usuarios que terminan en "ejemplo.com"
GET /scim/v2/Users?filter=emails ew "ejemplo.com"
```

##### Filtros de Presencia
```http
# Usuarios con email
GET /scim/v2/Users?filter=emails pr

# Usuarios con manager asignado
GET /scim/v2/Users?filter=urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:manager pr
```

##### Filtros de Comparación
```http
# Usuarios creados después de fecha
GET /scim/v2/Users?filter=meta.created gt "2023-01-01T00:00:00Z"

# Usuarios modificados recientemente
GET /scim/v2/Users?filter=meta.lastModified ge "2023-12-01T00:00:00Z"
```

##### Filtros Complejos
```http
# Usuarios activos del departamento IT
GET /scim/v2/Users?filter=active eq true and urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department eq "IT"

# Usuarios con email corporativo o personal
GET /scim/v2/Users?filter=emails.type eq "work" or emails.type eq "home"

# Usuarios activos con título específico
GET /scim/v2/Users?filter=active eq true and (title eq "Developer" or title eq "Senior Developer")
```

### UPDATE - Actualizar Recurso

#### PATCH - Actualización Parcial
```http
PATCH /scim/v2/Users/{id}
Content-Type: application/json
Authorization: Bearer {token}

{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
  "Operations": [
    {
      "op": "replace",
      "path": "active",
      "value": false
    },
    {
      "op": "replace",
      "path": "name.familyName",
      "value": "Nuevo Apellido"
    },
    {
      "op": "add",
      "path": "emails",
      "value": {
        "value": "nuevo@ejemplo.com",
        "type": "work",
        "primary": false
      }
    },
    {
      "op": "remove",
      "path": "phoneNumbers[type eq \"home\"]"
    }
  ]
}
```

#### Operaciones PATCH Soportadas

##### Replace - Reemplazar Valor
```json
{
  "op": "replace",
  "path": "userName",
  "value": "nuevo.usuario@ejemplo.com"
}
```

##### Add - Agregar Valor
```json
{
  "op": "add",
  "path": "emails",
  "value": {
    "value": "nuevo@ejemplo.com",
    "type": "work"
  }
}
```

##### Remove - Eliminar Valor
```json
{
  "op": "remove",
  "path": "emails[type eq \"personal\"]"
}
```

#### PUT - Reemplazo Completo
```http
PUT /scim/v2/Users/{id}
Content-Type: application/json
Authorization: Bearer {token}

{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "2819c223-7f76-453a-919d-413861904646",
  "userName": "usuario.actualizado@ejemplo.com",
  "name": {
    "familyName": "Nuevo Apellido",
    "givenName": "Juan"
  },
  "active": true
}
```

### DELETE - Eliminar Recurso

```http
DELETE /scim/v2/Users/{id}
Authorization: Bearer {token}
```

**Respuesta (204 No Content):**
```
HTTP/1.1 204 No Content
```

## Operaciones Bulk (En Desarrollo)

### Crear Múltiples Recursos
```http
POST /scim/v2/Bulk
Content-Type: application/json
Authorization: Bearer {token}

{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:BulkRequest"],
  "Operations": [
    {
      "method": "POST",
      "path": "/Users",
      "bulkId": "qwerty",
      "data": {
        "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
        "userName": "usuario1@ejemplo.com",
        "name": {
          "familyName": "Apellido1",
          "givenName": "Nombre1"
        }
      }
    },
    {
      "method": "POST",
      "path": "/Users",
      "bulkId": "ytrewq",
      "data": {
        "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
        "userName": "usuario2@ejemplo.com",
        "name": {
          "familyName": "Apellido2",
          "givenName": "Nombre2"
        }
      }
    }
  ]
}
```

## Manejo de Errores

### Formato de Error SCIM
```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"],
  "detail": "Attribute 'userName' is required",
  "status": "400"
}
```

### Errores Comunes

#### 400 Bad Request
```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"],
  "detail": "Invalid JSON syntax in request body",
  "status": "400",
  "scimType": "invalidSyntax"
}
```

#### 401 Unauthorized
```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"],
  "detail": "Authentication token required",
  "status": "401"
}
```

#### 404 Not Found
```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"],
  "detail": "Resource 2819c223-7f76-453a-919d-413861904646 not found",
  "status": "404",
  "scimType": "notFound"
}
```

#### 409 Conflict
```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"],
  "detail": "User with userName 'juan@ejemplo.com' already exists",
  "status": "409",
  "scimType": "uniqueness"
}
```

## Autenticación y Autorización

### Bearer Token
```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Headers de Seguridad Recomendados
```http
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
```

## Ejemplos de Integración

### cURL Examples

#### Crear Usuario
```bash
curl -X POST https://scim.ejemplo.com/scim/v2/Users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "test@ejemplo.com",
    "name": {
      "familyName": "Test",
      "givenName": "User"
    },
    "active": true
  }'
```

#### Buscar Usuarios
```bash
curl -G https://scim.ejemplo.com/scim/v2/Users \
  -H "Authorization: Bearer ${TOKEN}" \
  -d filter='userName sw "test"' \
  -d startIndex=1 \
  -d count=10
```

#### Actualizar Usuario
```bash
curl -X PATCH https://scim.ejemplo.com/scim/v2/Users/123 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${TOKEN}" \
  -d '{
    "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
    "Operations": [
      {
        "op": "replace",
        "path": "active",
        "value": false
      }
    ]
  }'
```

### Postman Collection
```json
{
  "info": {
    "name": "GoSCIM API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "variable": [
    {
      "key": "baseUrl",
      "value": "https://scim.ejemplo.com"
    },
    {
      "key": "token",
      "value": "{{bearer_token}}"
    }
  ],
  "auth": {
    "type": "bearer",
    "bearer": [
      {
        "key": "token",
        "value": "{{token}}"
      }
    ]
  }
}
```

## Límites y Restricciones

### Límites por Defecto
- **Resultados por página**: 100 (máximo 200)
- **Tamaño máximo de request**: 10MB
- **Longitud máxima de filtro**: 1000 caracteres
- **Operaciones bulk máximas**: 100 (cuando esté implementado)

### Rate Limiting
- **Requests por minuto**: 1000 por IP
- **Requests por hora**: 10000 por token
- **Concurrent connections**: 100 por IP

### Headers de Rate Limiting
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1638360000
Retry-After: 60
```