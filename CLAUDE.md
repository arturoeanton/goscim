# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**goscim** is a SCIM 2.0 (RFC 7643/7644) server in Go supporting three `Store` backends (`couchbase`, pure-Go `sqlite`, and `memory`), an embedded HTMX console (`/ui`), Prometheus metrics (`/metrics`), signed CloudEvents webhooks, and B2B multi-tenancy (`/tenants/:tenant/scim/v2`). Its defining trait is that **resource types, endpoints, storage buckets, secondary GSI indexes, and validation are all driven by JSON config, not code**.

The README and `docs/` were rewritten for 1.0 and now describe what the code actually does; earlier versions advertised authentication, metrics, health checks and a compose file that did not exist. They are trustworthy as of 1.0 — keep them that way.

## Commands

```bash
# Build, vet (with -unreachable=false for ANTLR), and run unit tests with -race
make check

# Integration suite against a real Couchbase container (requires Docker)
make integration

# Single test
go test ./scim -run TestCreateElement -v

# Or start Couchbase + GoSCIM via Docker Compose
docker compose up -d --build
```

`SCIM_ADMIN_USER` / `SCIM_ADMIN_PASSWORD` are **Couchbase credentials only** — they never protect the HTTP API.

The HTTP API is protected by `SCIM_AUTH`, which is **required**: `jwt` (RFC 9068 access tokens validated against `SCIM_JWT_JWKS_URL` / `SCIM_JWT_ISSUER` / `SCIM_JWT_AUDIENCE`), `bearer` (static opaque secret tokens via `SCIM_BEARER_TOKEN` or `SCIM_BEARER_TOKENS` for Entra ID Secret Token / Okta API Token), `basic` (`SCIM_BASIC_USERS=user:password:role1,role2;...`, development only), or `none` (`SCIM_ANONYMOUS_ROLES=...`). An unset `SCIM_AUTH` is a startup error rather than an open server.

A live Couchbase is only needed to *run* the server. Tests wire a `MemoryStore` into `scim.DB` instead — see "Storage" below.

## Architecture

### Config-driven startup (the central idea)

`main.go` → `scim.InitDB()` → `scim.ReadResourceType("config", r)`. That single call (`scim/config.go:47`) does four things per file in `config/resourceType/`:

1. Loads every `config/schemas/*.json` into the global `scim.Schemas` map, keyed by schema `id` (a URN).
2. Loads the resource type into `scim.Resources`, keyed by its **`endpoint`** (e.g. `/Elements`).
3. Calls `DB.EnsureBucket(resourceType.Name)` — in production, creates a Couchbase bucket named after **`name`** (e.g. `Element`), its `_default.scim_unique` sentinel collection, and a primary index, optionally configured from `<SCIM_CONFIG_DIR>/bucketSettings/<Name>.json`.
4. Registers the seven per-resource gin routes under `PREFIX = /scim/v2` (`POST`, `GET /:id`, `PUT /:id`, `DELETE /:id`, `PATCH /:id`, `GET`, `POST /.search`), while `NewRouter` also mounts server-root `POST /scim/v2/.search` and `POST /scim/v2/Bulk`.

**The `name` vs `endpoint` split matters everywhere**: `name` is the bucket/`meta.resourceType`, `endpoint` is the URL and the map key. Handlers are closures over the endpoint string (`Create(resourceType.Endpoint)`) that look the type back up via `Resources[resource]`.

Adding a resource type = drop a schema JSON + a resourceType JSON (+ optional bucketSettings JSON) and restart. No code changes.

### Storage

Handlers never touch Couchbase directly. They go through the `Store` interface in `scim/store.go`, held in the package global `scim.DB`:

- `store_couchbase.go` — production. `InitDB()` connects and installs it. One bucket per resource type, one document per resource keyed by UUID in `_default._default`, plus a dedicated `_default.scim_unique` collection holding atomic KV sentinel claims (`u:<sha256>`) for `CreateUnique`, `ReplaceUnique` (with document CAS), and `RemoveUnique`. Because search uses `SELECT * FROM \`Bucket\``, N1QL nests each row under the bucket name, so the store unwraps `item[q.Bucket]` before returning. Driver errors for missing documents are translated to `ErrNotFound`, which the handlers map to 404 — handlers must not import gocb.
- `store_memory.go` — the test fake. It round-trips every document through JSON exactly like the real driver, so `Meta` structs come back as maps and integers as `float64`; without that the fake would hide the type-assertion panics that production hits. It enforces the same mutex-guarded `CreateUnique`/`ReplaceUnique`/`RemoveUnique` contract and evaluates SCIM filters in memory via `parser.CompileMatcherWithOptions` (honouring `ResolveName`, `IsCaseExact`, and `IsMultiValued`); N1QL filter translation is additionally covered by `scim/parser` and `test/integration`.

`SearchQuery.SortBy` carries a raw SCIM attribute path; translating it (quoting, N1QL) is the store's job, not the handler's.

### Tests

There are two suites. `make test` / `make race` run the unit suite; `make integration` runs `test/integration` against a real Couchbase started with testcontainers (build tag `integration`, several minutes). The integration tests live in their own package deliberately: sharing a binary with the unit suite would let `newTestServer` swap the global `scim.DB` for the fake, and they would silently pass against it.

`scim/harness_test.go` builds the **real** router over a `MemoryStore` via `newTestServer(t)`, so routing, validation, meta generation and role filtering all execute. `scim/op_crud_test.go` drives it end to end with `httptest`. Tests that pin behavior known to violate the RFC are marked `TODO(Bn)` referencing `RELEASE-1.0.md` — when you fix bug *n*, those assertions are expected to change.

### Filter parser (`scim/parser/`)

ANTLR grammar in `ScimFilter.g4` at the repo root, generated with ANTLR 4.13 against the `github.com/antlr4-go/antlr/v4` runtime. **Everything in `scim/parser/` is generated except `scimfilter_listener_implement.go`, `errors.go`, and `eval.go`, which are hand-written and must never be overwritten.**

`FilterToN1QL(resourceName, filter, resolve)` returns a `Query` (`Page`, `Count`, `Params`): all literal values are bound via positional `$1`, `$2`, ... parameters (never interpolated into query text), `co`/`sw`/`ew` emit `LIKE ... ESCAPE "\\"` with `%` and `_` escaped, and bracketed value filters (`emails[type eq "work"]`, `emails[type eq "work"].value co "..."`) emit `(ANY \`_v0\` IN \`emails\` SATISFIES ... END)`. Syntax and grammar errors are collected by `errors.go` (`Validate`, `ReferencedAttributes`, `rewriteBracketSubPaths`); a filter that fails validation is never translated.

`eval.go` (`CompileMatcher`, `Match`, `EqualitySeeds`) evaluates the same ANTLR parse tree against decoded JSON maps in memory for `PATCH` value-selection paths (`emails[type eq "work"].value`). `SplitURNPath` in `scimfilter_listener_implement.go` is shared across N1QL quoting, `sortBy`, projection, and `PATCH`.

Regenerating (only needed when editing the `.g4`):
```bash
make generate   # runs ANTLR in a container, no JDK needed on the machine
make check
```
The jar is downloaded into `build/` on demand rather than committed. The generator writes only the four `scimfilter_{lexer,parser,listener,base_listener}.go` files and the `.tokens`; it does not touch the three hand-written files (`scimfilter_listener_implement.go`, `errors.go`, `eval.go`).

The entry rule is `p.Start_()` — ANTLR suffixes it because `Start` collides with a method on the generated parser.

`make vet` disables the `unreachable` analyser project-wide. The ANTLR Go target emits unreachable statements in the generated parser, and vet surfaces a dependency's diagnostics in every package importing it, so it cannot be disabled for that package alone. Plain `go vet ./...` therefore exits 1; use `make vet`.

### Request pipeline

Create/Replace share one shape: `RequestID` → `AuditMiddleware` → `LimitBody` → `Authenticate` → unmarshal body → `ValidateFieldSchemas` → `EnforceWriteAccess` (`$writer` + `mutability`) → `ValidateSchemas` → `ApplyPasswordPolicy` (`SCIM_PASSWORD_MODE`: `argon2id` default, `bcrypt`, `reject`, `plain`) → `ExtractUniqueClaims` → `Store.CreateUnique` / `Store.ReplaceUnique` → `syncAfterCreate` / `syncAfterReplace` (`scim/groups_sync.go`, keeping `User.groups` in sync with `Group.members` including `direct` and `indirect` nested groups) → `FormatResource` (`ValidateReadRole` + `ApplyProjection` over `attributes` / `excludedAttributes` + `returned`). Delete cleans up dangling `Group.members` references via `syncAfterDelete`.

Update (`PATCH`) supports both plain paths and RFC 7644 §3.5.2 value-selection paths (`emails[type eq "work"].value`, `members[value eq "..."]`, plus Entra ID `add` entry synthesis via `parser.EqualitySeeds` and colon-delimited URN paths), checks `$writer` / `$reader` / `writeOnly` on target and filter paths, then funnels into `replaceWithMode()`. Search is served on `GET /scim/v2/{resource}`, `POST /scim/v2/{resource}/.search`, and server-root `POST /scim/v2/.search`.

`POST /scim/v2/Bulk` (`scim/op_bulk.go`, RFC 7644 §3.7) validates the `BulkRequest` envelope (`MaxBulkOperations = 1000`, `maxPayloadSize = DefaultMaxBodyBytes`), resolves `bulkId:<token>` cross-references in paths, `$ref` URIs, and `PATCH` filter literals, forwards `version` as `If-Match`, honours `failOnErrors`, and dispatches each sub-operation through an internal router sharing the outer `Principal` and `X-Request-Id`.

### Authentication and observability

`scim/auth.go` defines the `Authenticator` interface and the gin middleware that publishes a `Principal` (subject plus roles) into the request context. `NewRouter` puts every resource endpoint and `POST /.search` behind `AuditMiddleware`, `LimitBody`, and `Authenticate`; `/healthz`, `/readyz` (`scim/health.go`), and `/scim/v2/{ServiceProviderConfig,ResourceTypes,Schemas}` (`scim/discovery.go`) stay unauthenticated. Implementations live in `auth_jwt.go` (JWKS-backed across RSA `RS256`/`384`/`512`, NIST EC `ES256`/`384`/`512` validated via `crypto/ecdh`, and Ed25519 OKP `EdDSA` with strict curve/algorithm binding, plus Entra `scp`, Okta `scp`/`cid`, and Keycloak `realm_access.roles` / dotted `SCIM_JWT_ROLES_CLAIM`), `auth_bearer.go` (constant-time SHA-256 static opaque tokens), `auth_basic.go`, and `AnonymousAuthenticator`.

`scim/audit.go` emits structured `log/slog` audit events (`SCIM_AUDIT_LOG=json|text|off`) with `X-Request-Id` correlation (`RequestID()`).

### Role-based access control

Two mechanisms, both applied per attribute and both recursing into complex and multi-valued values:

- **Reading** — `scim/validate_role.go` and `scim/projection.go`. `ValidateReadRole` drops any attribute whose `$reader` list does not intersect the caller's roles, and any attribute that is `returned: never` or `writeOnly` (`isReturnable`). `ApplyProjection` then enforces `returned: default | always | request` alongside `attributes` / `excludedAttributes`.
- **Writing** — `scim/validate_write.go`. `EnforceWriteAccess` runs before validation on create, replace, patch, and delete. `$writer` denies with 403; `mutability: readOnly` silently drops the client's value and restores the stored one; `mutability: immutable` refuses a change or removal with 400 `scimType: mutability`.

`"*"` in either list means any role. Roles come from the authenticated principal via `currentRoles(c)`; a request with no principal has none, so restricted attributes stay hidden.

### Observability, UI, Webhooks, and Multi-Tenancy

- **Prometheus metrics** (`scim/metrics.go`): `GET /metrics` (enabled unless `SCIM_METRICS=false`, optionally guarded by `SCIM_METRICS_TOKEN`).
- **HTMX Web Console & Live Inspector** (`scim/ui.go`, `scim/inspector.go`, `scim/ui/`): mounted at `/ui` when `-ui`, `-demo`, or `SCIM_UI=true` is set (optionally guarded by `SCIM_UI_TOKEN`).
- **Signed CloudEvents Webhooks** (`scim/webhooks.go`): `SCIM_WEBHOOK_URLS`, `SCIM_WEBHOOK_SECRET`, `SCIM_WEBHOOK_EVENTS`, `SCIM_WEBHOOK_MAX_RETRIES`.
- **Multi-tenancy** (`scim/tenant.go`): `/tenants/:tenant/scim/v2/...` routes + credential scoping (`SCIM_BEARER_TOKENS` `@tenant`, `SCIM_BEARER_TENANT`, `SCIM_JWT_TENANT_CLAIM`), encoded into `Store` bucket keys via `ScopedBucket(tenant, bucket)`.
- **Conformance CLI** (`scim/conformance.go`): `goscim conformance -url ... -profile all|rfc|entra|okta|keycloak`.

## Conventions and traps

- The struct field for `$reader` is `Attribute.Read`; for `$writer` it is `Attribute.Writer`.
- Handlers return `func(c *gin.Context)` closures, not plain handlers.
- Responses go through `scim/response.go`: `writeSCIM` sets the `application/scim+json` media type (gin only fills in `Content-Type` when it is unset, so setting it first wins), and `entityTag`/`checkPrecondition` implement `ETag`, `If-Match` and `If-None-Match` over `meta.version`. Errors go through `MakeError` / `MakeTypedError`, the latter carrying the RFC 7644 3.12 `scimType` keyword.
- Config folder paths default to `"config"` (`SCIM_CONFIG_DIR` overrides both the schema/resourceType directory and `FolderBucketSetting`).
- Manual API testing lives in `httpexamples/*.http` (REST Client format: `create.http`, `filter.http`, `bulk.http`, `idp_entra.http`, `idp_okta.http`, `idp_keycloak.http`).
- Documentation is English-only, in `docs/`, and is authoritative: `getting-started`, `idp-compatibility`, `installation`, `configuration`, `api`, `schemas`, `architecture`, `security`, `operations`, `development`. The six partial translations were removed in 1.0 because every one had fallen behind the code.
- A behaviour change needs the docs updated in the same commit. Several statements in this file and in `docs/` were false for months because a fix landed without them.
