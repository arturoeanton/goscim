# GoSCIM — Auditoría y Correcciones de Seguridad (1.0.1)

Revisión completa de vulnerabilidades reportadas sobre las dependencias del repositorio (`go.mod` y `test/integration/go.mod`), del toolchain de Go, y auditoría manual del código fuente sobre `1.0.0`. Cada hallazgo fue reproducido con pruebas automatizadas antes de darlo por cerrado.

---

## Parte 1 — Dependencias y Toolchain (GitHub / Dependabot / Go VulnDB)

Se verificaron los árboles completos de dependencias contra la base de GitHub Security Advisories (OSV) y `govulncheck`. Resultado final en ambos módulos: **0 vulnerabilidades**.

### Módulo principal (`go.mod`)

| Componente | Versión anterior | Versión nueva | Avisos resueltos |
|---|---|---|---|
| Toolchain / stdlib Go | `1.25.0` | `1.26.0` (`go1.27.4`) | `GO-2026-6859` (condición de carrera en `net/http.MaxBytesReader`, invocado en `scim/limits.go`) |
| `golang.org/x/net` | `v0.53.0` | `v0.62.0` | `GHSA-5cv4-jp36-h3mw`, `GO-2026-6617`, `GO-2026-6612`, `GO-2026-6611`, `GO-2026-6603` y 10 avisos adicionales en `http2`/`html`/`net` |
| `google.golang.org/grpc` | `v1.82.1` | `v1.85.0` | `GHSA-2v4p-qf9q-27wj` (`GO-2026-6348`, agotamiento de memoria en HTTP/2), `GHSA-qc2q-p7wx-3px3` (`GO-2026-6443`), `GHSA-vp52-pcj8-j9qc` (`GO-2026-6441`) |
| `golang.org/x/crypto` | `v0.51.0` | `v0.59.0` | 13 avisos `GHSA-*` y 17 `GO-*` en `ssh`, `x509` y primitivas auxiliares |
| `go.opentelemetry.io/otel` (`+metric`, `+trace`) | `v1.43.0` | `v1.47.0` | `GO-2026-5158` |
| `golang.org/x/text` | `v0.37.0` | `v0.44.0` | `GO-2026-5970`, `GO-2026-6629` |

### Módulo de integración (`test/integration/go.mod`)

Además de alinear las versiones anteriores con el módulo raíz, se actualizaron dependencias propias de `testcontainers-go` y el cliente de Docker:

| Componente | Versión anterior | Versión nueva | Avisos resueltos |
|---|---|---|---|
| `github.com/klauspost/compress` | `v1.18.5` | `v1.20.4` | `GHSA-259r-337f-4rfw`, `GO-2026-5841`, `GO-2026-6865` |
| `github.com/moby/go-archive` | `v0.2.0` | `v0.3.3` | `GHSA-hfg8-hc9c-6c3h` (`GO-2026-6253`) |
| `github.com/moby/sys/user` | `v0.4.0` | `v0.5.0` | `GHSA-mjcv-p78q-w5fw` (`GO-2026-6686`) |

En `.github/workflows/ci.yml` se actualizó `GO_VERSION` a `1.27.4` y se añadió la ejecución de `govulncheck -tags integration ./...` sobre `test/integration`.

---

## Parte 2 — Vulnerabilidades encontradas en el código fuente

### S1 · CRÍTICO · Bypass de `$writer`, `readOnly` e `immutable` mediante capitalización distinta y claves JSON duplicadas
Archivos: `scim/caseinsensitive.go`, `scim/validate_write.go`, `scim/validate.go`, `scim/op_update.go`

`EnforceWriteAccess` se ejecutaba antes de `ValidateSchemas` y buscaba los atributos en el mapa entrante por coincidencia exacta (`incoming[attribute.Name]`, `incoming[extension.Schema]`, `delete(incoming, attribute.Name)`). Justo después, `ValidateSchemas` renombraba las claves de forma case-insensitive a su nombre canónico y guardaba el documento.

Consecuencias verificadas:
- Enviar `"DESCRIPTION": "x"` en lugar de `"description"` evitaba el control de `$writer` y persistía `"description": "x"`.
- Enviar `"GROUPS": [{"value": "admins"}]` en `POST /Users` evitaba el descarte de `mutability: readOnly` y persistía `"groups"`.
- Enviar `"IDENTIFIER": "nuevo"` evitaba el bloqueo de `mutability: immutable`.
- Enviar un URN de extensión en mayúsculas (`"URN:IETF:PARAMS:SCIM:SCHEMAS:EXTENSION:..."`) saltaba por completo la revisión de permisos de toda la extensión.
- Enviar dos claves que solo difieren en mayúsculas (`{"groups": [...], "GROUPS": [...]}`) borraba únicamente la coincidencia exacta y dejaba sobrevivir la segunda durante la canonización.

Corrección:
- `EnforceWriteAccess` resuelve y elimina claves de manera case-insensitive en el schema principal, sub-atributos complejos, elementos `multiValued` y extensiones.
- Cualquier objeto JSON con dos claves que colisionen al ignorar mayúsculas/minúsculas se rechaza con `400` (`scimType: invalidSyntax`) antes de procesar permisos o validar tipos.
- La resolución de paths en `PATCH` (`op_update.go`) usa la misma comparación case-insensitive tanto en el documento almacenado como en la definición del schema.

---

### S2 · ALTO · Eliminación no autorizada de atributos `immutable` y `$writer` en `PUT` y `PATCH`, y oráculo 403
Archivos: `scim/validate_write.go`, `scim/op_replace.go`, `scim/op_update.go`

En `enforceObjectWriteAccess`, cuando un atributo no venía presente en el payload (`if !present { continue }`) se omitían todas las comprobaciones de `immutable` y `$writer`, incluso si el documento almacenado sí tenía valor para ese atributo.

Consecuencias verificadas:
1. Un atributo `immutable` no podía modificarse en un paso, pero sí borrarse omitiéndolo en un `PUT` (o con `PATCH remove`) y escribirse con un valor nuevo en una segunda petición.
2. Un llamador sin el rol `$writer` podía eliminar atributos protegidos mediante `PATCH {"op": "remove", "path": "description"}`.
3. A la inversa, como `PATCH` desembocaba en `replace()` reevaluando el documento completo, un usuario sin `$writer` sobre `description` recibía `403` al intentar hacer `PATCH` sobre un atributo distinto (`name`) si el documento ya contenía `description`.
4. En `PUT`, comparar `reflect.DeepEqual(rawValue, previous)` sobre un atributo oculto por `$reader` permitía distinguir con `200` vs `403` cuándo el valor enviado coincidía con el almacenado.

Corrección:
- Eliminar un atributo `immutable` que ya tenía valor responde `400` (`scimType: mutability`).
- `PATCH` valida `$writer` sobre las rutas tocadas por cada operación (`enforcePatchOperationWriteAccess`) y conserva intactos los atributos que la petición no modificó.
- En `PUT`, si el llamador no tiene `$reader` ni `$writer` sobre un atributo existente, no puede enviarlo (siempre `403`, sin comparar valores) y, si lo omite en su ciclo normal de lectura-modificación-escritura, el servidor preserva el valor almacenado. Si sí puede leerlo pero no escribirlo, modificarlo o quitarlo responde `403`.

---

### S3 · ALTO · Ausencia de control de autorización en `DELETE`
Archivos: `scim/op_delete.go`, `scim/validate_write.go`

`Delete` autenticaba la petición pero borraba directamente el documento en `DB.Remove` sin comprobar ningún permiso `$writer` del schema ni de las extensiones presentes en el recurso. Cualquier cuenta autenticada sin roles podía borrar cualquier usuario, grupo o recurso.

Corrección:
- Se añadió `EnforceDeleteAccess`, que carga el documento previo (devolviendo `404` si no existe, respetando `If-Match` cuando aplica) y exige permisos `$writer` sobre los atributos requeridos del tipo de recurso y sobre los atributos presentes en el documento antes de permitir el borrado.

---

### S4 · ALTO · Exfiltración ciega de secretos (`password`) y atributos restringidos vía `filter` y `sortBy`, y reglas permisivas en ANTLR
Archivos: `scim/op_search.go`, `scim/sortby.go`, `scim/parser/errors.go`, `scim/parser/scimfilter_listener_implement.go`

Problemas detectados:
1. `User.password` (`mutability: writeOnly`, `returned: never`) nunca se devolvía en el JSON de salida, pero seguía siendo aceptado en `filter` y `sortBy`. Un cliente podía inferir la contraseña carácter por carácter con `GET /Users?filter=password sw "a"` o deducir su orden relativo con `sortBy=password`.
2. Un cliente sin rol en `$reader` podía filtrar por el contenido de atributos que tenía prohibido leer (`GET /Elements?filter=description co "secreto"`), obteniendo confirmación por presencia en los resultados.
3. En `ScimFilter.g4`, la regla léxica `NUMBERS : '-'? [0-9.-]+` aceptaba cadenas como `--` o `1.2.3` sin error de sintaxis. Al fallar `strconv.ParseFloat` en `scimfilter_listener_implement.go`, el código caía al caso por defecto y concatenaba el token `--` crudo dentro del texto N1QL (`WHERE \`age\` = --`).
4. La gramática también aceptaba sin error expresiones adyacentes sin operador booleano (`(a eq "x")(b eq "y")`), comparaciones entre expresiones (`a eq b pr`) y filtros con corchetes aún no soportados (`emails[type eq "work"]`), generando N1QL inválido que terminaba en `500` desde Couchbase en vez de `400 invalidFilter`.

Corrección:
- `NormalizeSortBy` rechaza atributos no retornables (`writeOnly` o `returned: never`) con `400` (`scimType: invalidValue`).
- `ValidateFilterAttributes` extrae todos los atributos referenciados en el árbol del filtro y rechaza con `400` (`scimType: invalidFilter`) cualquier atributo inexistente, de un URN ajeno al recurso, `writeOnly` / `returned: never`, o fuera de los roles `$reader` del llamador.
- `treeValidator` en `scim/parser/errors.go` inspecciona el árbol de ANTLR antes de traducir y rechaza literales numéricos mal formados (`--`, `-`, `1.2.3`), expresiones adyacentes sin `and`/`or`, reglas `ATTR_OPER_EXPR` / `EXPR_OPER_EXPR` y filtros complejos `LBRAC_EXPR_RBRAC` aún no implementados. Además, un fallo en `strconv.ParseFloat` ya nunca emite texto literal en la consulta.

---

### S5 · MEDIO · Endurecimiento de validación JWT y refresco de JWKS
Archivo: `scim/auth_jwt.go`

Problemas detectados:
- Una ráfaga concurrente de peticiones con un `kid` inexistente ejecutaba descargas paralelas contra `SCIM_JWT_JWKS_URL` antes de que `fetchedAt` se actualizara; si el servidor JWKS fallaba, `fetchedAt` nunca se actualizaba y cada petición reintentaba sin límite de tasa.
- `io.ReadAll(resp.Body)` leía la respuesta JWKS sin cota de memoria.
- Se aceptaban entradas del JWKS marcadas para cifrado (`"use": "enc"` o `key_ops` sin `"verify"`) y claves RSA menores a 2048 bits o con exponentes inválidos.

Corrección:
- Serialización de refrescos con un mutex dedicado (`refreshMu`) y registro de `lastAttempt` antes de la llamada HTTP para respetar la ventana mínima incluso ante errores del emisor.
- Lectura del cuerpo JWKS acotada a 1 MiB (`io.LimitReader`).
- Verificación de `use == "sig"` (cuando `use` está presente), presencia de `"verify"` en `key_ops` (cuando `key_ops` está presente), módulo RSA de al menos 2048 bits (RFC 7518 §3.3) y exponente público impar `>= 3`.

---

### S6 · MEDIO / BAJO · Higiene de entrada JSON, redacción de errores 500 y credenciales en logs, cabecera `Location` y desbordamiento `int64` en ARM64
Archivos: `scim/op_create.go`, `scim/op_replace.go`, `scim/op_read.go`, `scim/error.go`, `scim/uniqueness.go`, `scim/response.go`, `scim/store_couchbase.go`, `scim/validate.go`

- **JSON malformado en `POST` y `PUT`**: `json.Unmarshal` ignoraba el error retornado, tratando un JSON truncado o `null` como un mapa vacío o nulo en lugar de responder `400` (`scimType: invalidSyntax`). Corregido.
- **Fuga de detalles internos en respuestas `500`**: los errores crudos del driver de Couchbase (direcciones internas, nombres de buckets, mensajes N1QL) se enviaban en `detail` al cliente HTTP. Ahora se registran en el log del servidor mediante `internalError` y al cliente se le devuelve un mensaje genérico.
- **Enmascaramiento de password en logs de arranque**: `hidepassword` dejaba visibles el primer y último carácter (`a****!`) o la cadena entera si tenía 1 carácter. Ahora devuelve siempre `[REDACTED]` (o `(empty)` si está vacío).
- **Cabecera `Location` y `X-Content-Type-Options`**: `absoluteLocation` construía la URL a partir de `c.Request.Host` sin validarlo. Se añadió la variable de entorno `SCIM_PUBLIC_URL` para fijar el origen público detrás de proxies, validación estricta de `Host` cuando no está definida, y `X-Content-Type-Options: nosniff` en todas las respuestas SCIM.
- **Saturación `float64 -> int64` en ARM64**: en `validateValue`, `float64(int64(v)) == v` daba falso positivo para `9223372036854775808` (`2^63`) en `arm64`, donde `int64(2^63)` satura a `math.MaxInt64` y al reconvertir a `float64` redondea otra vez a `2^63`. Se añadieron límites explícitos antes de convertir.

---

## Verificación

Suites añadidas en `scim/security_test.go` y `scim/parser/security_test.go`:

- `make check` (`go build ./...`, `go vet -unreachable=false ./...`, `go test ./... -count=1 -race`): OK.
- `make cover`: 76.7 % en `scim`, 61.6 % en `scim/parser` (68.9 % global).
- `make integration` (contenedor real `couchbase:community-7.1.1` vía `testcontainers-go`): 12 suites end-to-end en verde.
- `govulncheck ./...` y `cd test/integration && govulncheck -tags integration ./...`: sin vulnerabilidades.
