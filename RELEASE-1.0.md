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

### B7 · ALTO · Panic (500) en PUT sin `meta` en el body — CORREGIDO
`scim/op_replace.go:37`

```go
meta := element["meta"].(map[string]interface{})
```
sin comprobación. RFC 7644 no obliga al cliente a reenviar `meta` en un PUT, así que el caso normal es el que rompe. Mismo patrón de panic que B6.

Fix: leer el `meta` del documento almacenado (`getElementByID`), no del payload — que además es la semántica correcta: `meta.created` no debe ser controlable por el cliente.

---

### B8 · ALTO · El control de acceso de lectura no cubre arrays ni extensiones — CORREGIDO
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

### B9 · MEDIO · Desviaciones de RFC 7644 en las respuestas — CORREGIDO
- `scim/op_create.go:49` devuelve `200 OK`; debe ser `201 Created` con cabecera `Location`.
- Content-Type es `application/json`; debe ser `application/scim+json`.
- `scim/meta.go:16,29` genera `meta.location` como `/Elements/<id>` — sin el prefijo `/scim/v2` y sin URI absoluta, por lo que el valor no es navegable.
- `meta.version` es un UUID nuevo por escritura, pero no se emite como `ETag` ni se soporta `If-Match`, así que no hay control de concurrencia: dos PUT simultáneos se pisan sin aviso.
- `op_search.go` fija `itemsPerPage` al `count` solicitado, no al número de recursos devueltos.

**Queda abierto**: `meta.location` es ahora una ruta absoluta del servidor (`/scim/v2/Elements/<id>`), no un URI absoluto. Para emitir un URI absoluto correcto hace falta una URL pública configurable, porque derivarla de la conexión da un valor equivocado detrás de un proxy que termina el TLS. La cabecera `Location` sí es absoluta.

---

### B10 · CRÍTICO · No existe autenticación, y la conexión a Couchbase no valida el certificado — CORREGIDO
`main.go` no registra ningún middleware de auth: los seis verbos de cada resource type son públicos. Los roles están hardcodeados en `op_read.go:21` y `op_search.go:107`, y `$writer` no se aplica en ningún punto (tres `//TODO: Validate _write`).

En paralelo, `scim/couchbase.go:52` fija `TLSSkipVerify: true` sobre un esquema `couchbases://`, lo que anula la protección del TLS que se está pidiendo (MITM entre servidor y base).

Un release 1.0 no debería publicarse sin esto resuelto.

**Hallazgo colateral**: el schema `core:2.0:Element` que se distribuye declaraba `name` y `description` como `mutability: readOnly` y a la vez con `$writer: ["*"]`. Las dos cosas se contradicen, y con `readOnly` aplicado de verdad el `Element` de ejemplo dejaba de poder crearse (`name` es `required`). Se corrigieron a `readWrite`, que es lo que el `$writer` ya decía.

**Queda abierto**: `mutability: readOnly` se aplica **ignorando** el valor del cliente, no rechazándolo. Es lo que permite el patrón leer-modificar-escribir: un cliente que hace PUT devuelve el recurso entero, incluidos los atributos que no le pertenecen, y rechazarlo haría imposible cualquier actualización normal.

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

15. ~~**Cobertura de tests**~~ **Hecho**. La suite unitaria cubre el 76 % de `scim/`, y `make integration` levanta una Couchbase real con `testcontainers-go` para cubrir lo que el fake no puede: arranque, creación de buckets e índices, y la traducción a N1QL completa.

    Los tests de integración encontraron **cuatro bugs de producción** en su primera ejecución:
    - `EnsureBucket` enviaba siempre `compression_mode`, que **Couchbase Community rechaza**: el servidor no arrancaba contra la edición gratuita que indica el propio README.
    - `CREATE PRIMARY INDEX` se pedía inmediatamente después de crear el bucket, antes de que el servicio de consultas lo viera. El primer arranque contra un cluster nuevo fallaba con *service not available*.
    - Las búsquedas usaban la consistencia por defecto de N1QL (`not_bounded`), así que **un recurso recién creado no aparecía al buscarlo**. Para una API de aprovisionamiento eso no es un caso límite: es el flujo normal. Ahora es `request_plus`, configurable con `SCIM_QUERY_CONSISTENCY`.
    - `totalResults` e `itemsPerPage` llevaban `omitempty`, así que una búsqueda sin resultados respondía sin ellos. RFC 7644 §3.4.2 los exige.
16. **CI en GitHub Actions**: `build`, `vet`, `staticcheck`, `go test -race -cover`, `govulncheck` y `gosec` en cada PR. `.github/` hoy solo tiene plantillas.
17. **Actualizar dependencias y toolchain** — *hecho en parte*. `go` subió de 1.16 a 1.25; gin 1.7.7 → 1.12.0, gocb 2.3.5 → 2.12.4, gocbcore 10.0.7 → 10.9.3, uuid 1.3.0 → 1.6.0, y grpc/quic-go/otel (transitivas de gocb y gin) a versiones sin vulnerabilidades alcanzables. `govulncheck` pasa de 25 vulnerabilidades en módulos requeridos a **0 alcanzables desde nuestro código**.

    El parser se regeneró con ANTLR 4.13.2 y se migró al runtime mantenido `github.com/antlr4-go/antlr/v4` (el path viejo `github.com/antlr/antlr4/runtime/Go/antlr` está sin mantenimiento desde 2022). La herramienta ANTLR es Java y no hay JDK en la máquina, así que `make generate` la ejecuta en un contenedor y descarga el jar bajo demanda.

    **Coste**: el generador Go de ANTLR emite código inalcanzable, y `go vet` propaga los diagnósticos de una dependencia a todo paquete que la importe, así que no se puede desactivar solo para el paquete generado. `make vet` desactiva el analizador `unreachable` en todo el proyecto; el resto siguen activos. `go vet ./...` a secas sale con 1.

    **Queda abierto**: migrar `ioutil` a `os`.
18. **Dockerfile multi-stage + docker-compose** con Couchbase y provisión automática del cluster. El README promete `docker-compose up -d` y el archivo no existe.
19. **Configuración por entorno, no por CWD**: `SCIM_CONFIG_DIR` en lugar de las rutas relativas `"config"` y `"config/bucketSettings/"`, que obligan a lanzar el binario desde la raíz del repo.
20. **Arranque y apagado de nivel producción**: `http.Server` con `ReadTimeout`/`WriteTimeout`/`IdleTimeout` (hoy `r.Run` no fija ninguno → slowloris), graceful shutdown por `SIGTERM`, y sustituir los `log.Fatalln` de `CreateBucket`/`InitDB` por errores propagados. Añadir `/healthz` y `/readyz`, logging estructurado con request-id, y métricas Prometheus — las tres cosas que el README ya anuncia.

### Extras de pulido (no bloqueantes, alto retorno)

- ~~**Higiene del repositorio**: `antlr-4.7-complete.jar` (2 MB) versionado, y `.antlr/` y `.scannerwork/` commiteados.~~ **Hecho**: los tres fuera del repo y en `.gitignore`; el jar lo descarga `make generate`.
- **Alinear README con la realidad**: OAuth 2.0/JWT, Prometheus, health checks, webhooks y la tabla de rendimiento («10.000 req/s») no tienen respaldo en el código. Para un 1.0 «premium», mover eso a un roadmap explícito.
- **Corregir el typo `ResoruceType`** en un solo commit mecánico antes de congelar la API pública del paquete — después del 1.0 será un breaking change.
- **Unificar el regex URN duplicado** entre `AddQuote`, `opPathTopathArray` y `splitURNPath` en un único helper.
- **`commons.WalkMap` quedó sin uso** al reescribir el filtrado por rol (no podía expresar «omitir la clave»). El paquete `commons` entero puede borrarse.
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
