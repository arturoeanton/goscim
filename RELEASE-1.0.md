# GoSCIM — Plan de Release 1.0

Auditoría del código en `main` (commit `ad4661b`). Cada bug fue **verificado ejecutando código**, no solo leyendo. Las pruebas de verificación se ejecutaron y se eliminaron; el repositorio quedó limpio.

Estado de partida: `go build ./...`, `go vet ./...` y `go test ./...` pasan (3 tests, todos en `scim/parser`).

---

## Parte 1 — 11 bugs (10 de la auditoría inicial + 1 hallado durante las correcciones)

### B1 · CRÍTICO · Operadores `gt`/`ge`/`lt`/`le` invertidos — CORREGIDO
`scim/parser/scimfilter_listener_implement.go:82-101`

`ge` genera `>`, `gt` genera `>=`, `le` genera `<`, `lt` genera `<=`. Los cuatro están cruzados: cada búsqueda por rango devuelve el conjunto equivocado, silenciosamente.

Verificado:
```
age gt 10  =>  SELECT * FROM `User` WHERE `age` >= 10     (debe ser >)
age ge 10  =>  SELECT * FROM `User` WHERE `age` > 10      (debe ser >=)
age lt 10  =>  SELECT * FROM `User` WHERE `age` <= 10     (debe ser <)
age le 10  =>  SELECT * FROM `User` WHERE `age` < 10      (debe ser <=)
```

Fix: intercambiar los cuatro `value =` y añadir un test de tabla por operador.

---

### B2 · CRÍTICO · Inyección N1QL vía `filter` (los errores de parseo se ignoran) — CORREGIDO
`scim/parser/scimfilter_listener_implement.go:22-30`

`FilterToN1QL` no instala un `ErrorListener` ni consulta `p.GetNumberOfSyntaxErrors()`. ANTLR reporta el error por stderr, se recupera, y el walker **emite igual todos los tokens** concatenándolos en el SQL.

Verificado:
```
filter=userName eq "a" ,;-- garbage
  => SELECT * FROM `User` WHERE `userName` = "a" ,;-- garbage

filter=userName eq "a" or 1=1
  => SELECT * FROM `User` WHERE `userName` = "a" or 1=1
```

El segundo caso es un bypass completo del filtro. Sin autenticación (B10) esto es lectura arbitraria de la base.

Fix: `ErrorListener` que acumule errores; si hay alguno, `FilterToN1QL` devuelve error y el handler responde `400 invalidFilter`. Adicionalmente, pasar los literales como parámetros posicionales de `gocb` en lugar de concatenarlos.

---

### B3 · CRÍTICO · Inyección N1QL vía `sortBy` (backtick no escapado) — CORREGIDO
`scim/op_search.go:26-37` + `AddQuote`

`AddQuote` envuelve el path en backticks pero no escapa los backticks del input. El saneo posterior solo hace `Trim(" ")` y elimina `;`.

Verificado:
```
AddQuote("id` , (SELECT * FROM `User`) x")  =>  `id` , (SELECT * FROM `User`) x`
```

El atacante sale del quoting y controla la cláusula `ORDER BY`.

Fix: validar `sortBy` contra los atributos declarados del schema (whitelist) y rechazar cualquier cosa fuera de `[-_.:a-zA-Z0-9]`; duplicar los backticks internos.

---

### B11 · CRÍTICO · Un valor del filtro puede romper su propio literal de cadena — CORREGIDO

*Hallazgo posterior a la auditoría inicial, encontrado mientras se corregía B3.*

Los valores del filtro se concatenan en el N1QL como literales entre comillas dobles. La regla `criteria` de la gramática acepta cualquier carácter salvo `"`, `(`, `)`, `[`, `]` — incluido el backslash. Un valor terminado en backslash escapa la comilla de cierre del propio literal.

Verificado:
```
filter=name eq "a\"
  => SELECT * FROM `Element` WHERE `name` = "a\"
     ORDER BY `name` ASC        <- todo esto queda DENTRO del literal
     OFFSET 0
     LIMIT 100
```

Además rompe el caso legítimo: buscar un valor con backslash (una ruta Windows, `DOMAIN\usuario`) generaba SQL inválido.

Fix: escapar el backslash al emitir el contenido de `criteria`, rastreando con `EnterCriteria`/`ExitCriteria` cuándo los terminales pertenecen a un valor. Un test de propiedad comprueba que ninguna consulta generada termina dentro de un literal sin cerrar.

**Queda abierto**: dentro de `co`/`sw`/`ew`, los caracteres `%` y `_` del valor se interpretan como comodines LIKE, así que buscar un `%` literal devuelve todo. Se arregla junto con la mejora 4 (parámetros ligados), que elimina toda esta clase de problema de raíz en vez de escapar caso por caso.

---

### B4 · CRÍTICO · Los atributos `multiValued` son inutilizables — no se puede crear un `User` estándar — CORREGIDO
`scim/validate.go:105-152`

`validateAttribute` nunca mira `Attribute.MultiValued`. Un atributo declarado `multiValued: true, type: complex` recibe un `[]interface{}` y `validateAttributeComplex` exige `map[string]interface{}`.

Verificado con el schema `core:2.0:User` que se distribuye en el repo:
```
POST {"schemas":[...User],"userName":"jane","emails":[{"value":"j@x.com","type":"work"}]}
  => 400  "emails should be complex"
```

Afecta a `emails`, `phoneNumbers`, `ims`, `photos`, `addresses`, `groups`, `entitlements`, `roles`, `x509Certificates` y a `members` de `Group`. En la práctica el servidor no es interoperable con ningún cliente SCIM real.

Fix: en `validateAttribute`, si `attribute.MultiValued` entonces exigir array y validar cada elemento con las reglas del tipo; validar además `primary` único y `type` dentro de los valores canónicos.

---

### B5 · ALTO · Las extensiones opcionales se comportan como obligatorias — CORREGIDO
`scim/validate.go:63-70`

`ValidateSchemas` itera **todas** las `schemaExtensions` del resource type y hace `element[ext.Schema].(map[string]interface{})` sin mirar `ext.Required`. Si la extensión no viene en el payload, el type assert falla y responde 400.

Verificado: `User` + extensión `enterprise:2.0:User` con `required:false`, payload sin la extensión → `400`.

Fix: si la clave no está presente y `!ext.Required`, saltar; si está presente y no es objeto, 400.

---

### B6 · ALTO · Panic (500) en PATCH con un `path` inexistente — CORREGIDO
`scim/op_update.go:76-86`

`pointValue` camina el path haciendo `elemPoint.(map[string]interface{})[field]`. En cuanto un segmento no existe, `elemPoint` es `nil` y el siguiente type assert entra en panic.

Verificado: `patchReplace("noexiste.sub", "x", {...})` → `panic: interface conversion: interface {} is nil, not map[string]interface {}`.

Cualquier cliente puede tumbar el handler con un PATCH trivial. Además, PATCH no soporta paths con filtro (`emails[type eq "work"].value`), que RFC 7644 §3.5.2 exige.

Fix: `pointValue` devuelve `(lastField, parent, error)`; sin panic. Implementar paths con filtro o rechazarlos explícitamente con `400 invalidPath`.

---

### B7 · ALTO · Panic (500) en PUT sin `meta` en el body
`scim/op_replace.go:37`

```go
meta := element["meta"].(map[string]interface{})
```
sin comprobación. RFC 7644 no obliga al cliente a reenviar `meta` en un PUT, así que el caso normal es el que rompe. Mismo patrón de panic que B6.

Fix: leer el `meta` del documento almacenado (`getElementByID`), no del payload — que además es la semántica correcta: `meta.created` no debe ser controlable por el cliente.

---

### B8 · ALTO · El control de acceso de lectura no cubre arrays ni extensiones
`scim/validate_role.go` + `commons/map.go:5-21`

Dos fallos combinados:
1. `WalkMap` solo recursa en `map[string]interface{}`; los elementos dentro de `[]interface{}` no se visitan.
2. `ValidateReadRole` resuelve atributos solo contra `Schemas[resourceType.Schema]`, ignorando `schemaExtensions`.

Verificado, con un rol sin ningún permiso:
```json
{"description":"","lista":[{"name":"secreto-en-array"}],"name":""}
```
Los campos escalares se censuran; el valor dentro del array se filtra intacto.

Además, censurar poniendo `""` en vez de omitir la clave hace que el cliente no distinga «sin permiso» de «vacío», y rompe el tipo para atributos no-string.

Fix: recursar en arrays, resolver el atributo contra el schema correcto según el prefijo URN, y **omitir** la clave en vez de vaciarla.

---

### B9 · MEDIO · Desviaciones de RFC 7644 en las respuestas
- `scim/op_create.go:49` devuelve `200 OK`; debe ser `201 Created` con cabecera `Location`.
- Content-Type es `application/json`; debe ser `application/scim+json`.
- `scim/meta.go:16,29` genera `meta.location` como `/Elements/<id>` — sin el prefijo `/scim/v2` y sin URI absoluta, por lo que el valor no es navegable.
- `meta.version` es un UUID nuevo por escritura, pero no se emite como `ETag` ni se soporta `If-Match`, así que no hay control de concurrencia: dos PUT simultáneos se pisan sin aviso.
- `op_search.go` fija `itemsPerPage` al `count` solicitado, no al número de recursos devueltos.

---

### B10 · CRÍTICO · No existe autenticación, y la conexión a Couchbase no valida el certificado
`main.go` no registra ningún middleware de auth: los seis verbos de cada resource type son públicos. Los roles están hardcodeados en `op_read.go:21` y `op_search.go:107`, y `$writer` no se aplica en ningún punto (tres `//TODO: Validate _write`).

En paralelo, `scim/couchbase.go:52` fija `TLSSkipVerify: true` sobre un esquema `couchbases://`, lo que anula la protección del TLS que se está pidiendo (MITM entre servidor y base).

Un release 1.0 no debería publicarse sin esto resuelto.

---

## Parte 2 — 20 mejoras para un 1.0 profesional

### Seguridad y corrección (bloqueantes)

1. **Middleware de autenticación** con estrategias intercambiables: Bearer/JWT (JWKS, validación de `iss`/`aud`/`exp`) y HTTP Basic para desarrollo. Poner los seis verbos detrás del middleware.
2. **Extraer roles del token** y eliminar los dos arrays hardcodeados; propagar el sujeto por `context` para autorización y auditoría.
3. **Aplicar `$writer`** en create/replace/update y `$remove` (o `$writer`) en delete — hoy son tres TODO que dejan el modelo de permisos a medias.
4. **Consultas parametrizadas** con `gocb.QueryOptions{PositionalParameters: ...}` en toda la ruta de búsqueda; ningún literal debe concatenarse en el N1QL.
5. **TLS configurable de verdad**: `SCIM_COUCHBASE_TLS_SKIP_VERIFY` (por defecto `false`) y ruta de CA; permitir `couchbase://` para desarrollo local. Añadir TLS al listener HTTP.
6. **Rate limiting y límite de tamaño de body** (`http.MaxBytesReader`); hoy un POST de 1 GB se lee entero a memoria con `buf.ReadFrom(c.Request.Body)`.
7. **Unicidad de atributos**: honrar `uniqueness: server|global` (mínimo `userName`) con índice único en Couchbase y respuesta `409 uniqueness`.

### Cumplimiento SCIM

8. **Implementar los tres endpoints de discovery** bajo `/scim/v2`, sirviendo `config/serviceProviderConfig/sp_config.json` (que ya existe sin usarse) y proyectando `Schemas`/`Resources` en memoria. Es lo primero que consulta cualquier IdP.
9. **Implementar `/Bulk`** (RFC 7644 §3.7) con `failOnErrors`, `bulkId` y resolución de referencias, o retirarlo del README hasta que exista.
10. **`POST /.search`** en cada resource type y a nivel raíz.
11. **`attributes` / `excludedAttributes`** en GET y search, más el atributo `returned: never` (hoy `password` se devuelve en las respuestas).
12. **Comparación case-insensitive** de nombres de atributo y de URNs de schema, como exige RFC 7643 §2.1.
13. **`ETag` + `If-Match`/`If-None-Match`** sobre `meta.version`, cerrando el hueco de concurrencia de B9.
14. **Filtros con valor complejo** (`emails[type eq "work"]`): la gramática los acepta (`LBRAC_EXPR_RBRAC`) pero el listener no los traduce a N1QL.

### Ingeniería y operación

15. **Cobertura de tests**: hoy hay 3 tests, todos del parser, y `scim/` está a 0 %. Mínimo para un 1.0: tabla de casos por operador y por tipo en el parser, tests de `validate.go` con `httptest`, y tests de integración de los seis verbos contra Couchbase con `testcontainers-go`. Objetivo razonable: 70 % en `scim/`.
16. **CI en GitHub Actions**: `build`, `vet`, `staticcheck`, `go test -race -cover`, `govulncheck` y `gosec` en cada PR. `.github/` hoy solo tiene plantillas.
17. **Actualizar dependencias y toolchain**: `go 1.16` está fuera de soporte; `gin v1.7.7` arrastra CVEs conocidos (corregidos en 1.9.x). Subir a Go 1.22+, gin 1.10.x, gocb v2 actual, y migrar `ioutil` a `os`.
18. **Dockerfile multi-stage + docker-compose** con Couchbase y provisión automática del cluster. El README promete `docker-compose up -d` y el archivo no existe.
19. **Configuración por entorno, no por CWD**: `SCIM_CONFIG_DIR` en lugar de las rutas relativas `"config"` y `"config/bucketSettings/"`, que obligan a lanzar el binario desde la raíz del repo.
20. **Arranque y apagado de nivel producción**: `http.Server` con `ReadTimeout`/`WriteTimeout`/`IdleTimeout` (hoy `r.Run` no fija ninguno → slowloris), graceful shutdown por `SIGTERM`, y sustituir los `log.Fatalln` de `CreateBucket`/`InitDB` por errores propagados. Añadir `/healthz` y `/readyz`, logging estructurado con request-id, y métricas Prometheus — las tres cosas que el README ya anuncia.

### Extras de pulido (no bloqueantes, alto retorno)

- **Higiene del repositorio**: `antlr-4.7-complete.jar` (2 MB) versionado, y `.antlr/` y `.scannerwork/` commiteados. Sacarlos y añadirlos a `.gitignore`; descargar ANTLR desde un `Makefile`.
- **Alinear README con la realidad**: OAuth 2.0/JWT, Prometheus, health checks, webhooks y la tabla de rendimiento («10.000 req/s») no tienen respaldo en el código. Para un 1.0 «premium», mover eso a un roadmap explícito.
- **Corregir el typo `ResoruceType`** en un solo commit mecánico antes de congelar la API pública del paquete — después del 1.0 será un breaking change.
- **Unificar el regex URN duplicado** entre `AddQuote` y `opPathTopathArray` en un único helper.
- **`bucket_type: "membase"`** en `config/bucketSettings/*.json` no coincide con ninguno de los valores que acepta `CreateBucket` (`couchbase`/`memcached`/`ephemeral`), así que se ignora en silencio y cae al default. Validar el config al cargarlo y fallar con un mensaje claro.

---

## Orden sugerido

| Fase | Contenido | Criterio de salida |
|---|---|---|
| 1 — Parar la hemorragia | B1, B2, B3, B4, B5, B6, B7 | Un `User` SCIM estándar se crea, lee, filtra y parchea sin panics |
| 2 — Seguridad | B10, B8, mejoras 1-7 | Nada es accesible sin token; sin inyección; sin TLS inseguro |
| 3 — Cumplimiento | B9, mejoras 8-14 | Un IdP real (Okta/Entra) completa un ciclo de aprovisionamiento |
| 4 — Producción | Mejoras 15-20 + pulido | CI en verde con cobertura, imagen publicada, apagado limpio |

Las fases 1 y 2 son el mínimo innegociable para poner la etiqueta 1.0.
