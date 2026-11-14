# Development

## Layout

```
main.go                       startup, subcommands (conformance, version), -demo, -ui, -store, -healthcheck
ScimFilter.g4                 the filter grammar
Makefile                      build, demo, check, integration, generate
Dockerfile                    distroless container image
docker-compose.yml            local Couchbase + GoSCIM stack
docker-compose.keycloak.yml   Keycloak 26 OIDC/JWKS + GoSCIM lab
.goreleaser.yaml              multi-OS / multi-arch release packaging
scripts/install.sh            1-line binary installer
charts/goscim/                Kubernetes Helm chart
config/                       schemas, resource types, bucket settings, embed.go
scim/
  router.go                   route registration (/scim/v2, /tenants/:tenant/scim/v2, /metrics, /ui)
  tenant.go                   multi-tenant path, credential scoping, and bucket partitioning
  audit.go                    X-Request-Id, structured log/slog audit, metrics & inspector hooks
  metrics.go                  Prometheus text exposition registry (/metrics)
  inspector.go                redacted in-memory request/response ring buffer
  webhooks.go                 CloudEvents 1.0 sharded webhook dispatcher + HMAC-SHA256 signing
  conformance.go              programmatic & CLI RFC + IdP conformance runner
  demo.go                     -demo seed dataset (Users + nested Groups)
  ui.go ui/                   embedded HTMX 2.0.4 Web Console, filter playground, inspector
  health.go                   /healthz and /readyz handlers
  config.go                   loading config (disk + embedded fallback), registering endpoints
  types.go                    SCIM data structures
  store.go store_init.go      the Store interface and backend selector
  store_couchbase.go          Couchbase storage + _default.scim_unique sentinels
  store_couchbase_indexes.go  schema-derived secondary N1QL GSI index builder
  store_couchbase_tls.go      connection string and TLS settings
  store_sqlite.go             pure-Go embedded SQLite WAL store (modernc.org/sqlite)
  store_memory.go             in-memory store (-demo and unit tests)
  auth.go auth_jwt.go auth_bearer.go auth_basic.go auth_claims.go
  op_create.go op_read.go op_search.go op_replace.go op_update.go op_delete.go op_bulk.go
  groups_sync.go              Group.members ↔ User.groups synchronisation
  validate.go                 types, required, multiValued
  validate_write.go           $writer, mutability
  validate_role.go            $reader, returned
  projection.go               attributes and excludedAttributes projection
  password.go                 Argon2id / bcrypt / reject password handling
  uniqueness.go               atomic uniqueness claim extraction
  sortby.go                   sortBy whitelist, attribute path canonicalisation
  caseinsensitive.go          case-insensitive resolution
  discovery.go                ServiceProviderConfig, ResourceTypes, Schemas
  response.go                 media type, ETag, preconditions, tenant-aware Location
  limits.go error.go meta.go contains.go
  parser/                     ANTLR output, plus three hand-written files
test/integration/             the suite against a real Couchbase (incl. GSI indexes & multi-tenancy)
```

## Commands

```bash
make check        # build, vet, unit suite with the race detector
make test         # unit suite only
make race         # unit suite with the race detector
make cover        # coverage
make integration  # against a real Couchbase, in a container
make generate     # regenerate the filter parser
```

`make vet` disables the `unreachable` analyser. The ANTLR Go target emits
unreachable statements in the generated parser, and `go vet` reports a
dependency's diagnostics in every package that imports it, so it cannot be
switched off for the generated package alone. Every other analyser stays on.
Plain `go vet ./...` therefore exits 1 — use `make vet`.

## Tests

Two suites, deliberately separate.

**Unit** — `scim/*_test.go`. `newTestServer(t)` builds the **real** router over
a `MemoryStore`, so routing, validation, meta generation, authentication and
permission filtering all execute. Tests drive it with `httptest`.

**Integration** — `test/integration/`, behind the `integration` build tag.
Starts Couchbase with testcontainers and drives the same server through it.
Slow, and the only thing that covers startup, bucket and index creation, and
the N1QL translation.

They live in different packages on purpose. Sharing a binary would let the unit
suite's `newTestServer` swap the global `scim.DB` for the fake, and the
integration tests would quietly pass against it — a false green, which is worse
than no test.

### Writing tests

Prefer driving the router over calling functions directly: it is the contract
clients depend on, and it catches middleware ordering mistakes that a direct
call cannot.

```go
func TestSomething(t *testing.T) {
	r, store := newTestServer(t)

	w := do(t, r, http.MethodPost, usersPath, `{"schemas":[...],"userName":"jane"}`)
	requireStatus(t, w, http.StatusCreated)

	stored, err := store.Get("User", decode(t, w)["id"].(string))
	...
}
```

Helpers in `scim/harness_test.go`: `newTestServer`, `newTestServerAs` for a
specific authenticator, `do`, `serve`, `decode`, `requireStatus`,
`requireSCIMError`, `testContext`.

Anything touching the N1QL filter translation, bucket creation or Couchbase
query behaviour belongs in the integration suite as well (`test/integration/`),
even though `MemoryStore` evaluates filters in memory for unit conformance
tests.

### A note on what a test is worth

The 1.0 work used mutation testing to check the suite: break the code on
purpose, confirm a test fails. It caught two tests that did not test what they
claimed — one asserting a 200 from `sortBy` without checking the ordering,
another exercising the JWT library's type safety rather than the algorithm
allow-list it was named for. A green suite is not evidence until you have seen
it go red.

## The filter parser

`ScimFilter.g4` is the grammar. Everything in `scim/parser/` is generated
**except** `scimfilter_listener_implement.go`, `errors.go`, and `eval.go`, which
are hand-written and must not be overwritten.

```bash
make generate   # runs ANTLR 4.13 in a container; no JDK needed
make check
```

The jar is downloaded into `build/` on demand rather than committed. The
generator writes only the four `scimfilter_{lexer,parser,listener,base_listener}.go`
files and the `.tokens`.

Two invariants to preserve if you touch the translation:

1. **No client value reaches the query text.** Values are bound parameters.
   `scim/parser/binding_test.go` asserts it for every operator.
2. **A filter that does not parse produces no query.** ANTLR recovers from
   syntax errors and keeps walking, so the error listener has to be consulted.

The entry rule is `p.Start_()` — ANTLR suffixes it to avoid a name collision.

## Adding a resource type

No code. Drop a schema in `config/schemas/` and a resource type in
`config/resourceType/`, optionally a `config/bucketSettings/<Name>.json`, and
restart. See [getting started](getting-started.md#7-add-a-resource-type-without-writing-code).

## Adding an authenticator

Implement three methods and wire it into `NewAuthenticatorFromEnv`:

```go
type Authenticator interface {
	Challenge() string
	AuthenticationSchemes() []AuthenticationScheme
	Authenticate(r *http.Request) (*Principal, error)
}
```

Return `ErrUnauthenticated` rather than a specific reason: the middleware owns
the response so that every scheme fails identically from the client's point of
view. `AuthenticationSchemes` feeds `/ServiceProviderConfig`, which is why
discovery cannot drift from the scheme actually in force.

## Conventions

- Handlers are `func(c *gin.Context)` closures over the endpoint string, and
  look the resource type back up from `Resources`.
- Errors go through `MakeError` / `MakeTypedError`, the latter carrying the
  RFC 7644 §3.12 `scimType`.
- Responses go through `writeSCIM`, which sets the SCIM media type.
- Storage errors are normalised: a missing document is `ErrNotFound`, so
  handlers never import gocb.
- `name` is the bucket, `endpoint` is the URL. They are not interchangeable.

## Before opening a pull request

```bash
make check
make integration   # if you touched storage, filters or startup
gofmt -l .         # generated parser files excepted
```

CI runs the same checks plus `govulncheck` and a `go mod tidy` check. See
[CONTRIBUTING](../CONTRIBUTING.md).
