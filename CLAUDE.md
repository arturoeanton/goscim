# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**goscim** is a SCIM 2.0 (RFC 7643/7644) server in Go, backed by Couchbase. It is a small codebase (~1.4k lines of hand-written Go) whose defining trait is that **resource types, endpoints, storage buckets and validation are all driven by JSON config, not code**.

Note: the README is aspirational marketing. Several features it advertises (OAuth 2.0/JWT auth, Prometheus metrics, health checks, docker-compose, webhooks) do **not** exist in the code. Trust the code over the README and over `docs/`.

## Commands

```bash
# Build / vet
go build ./... && go vet ./...

# All tests — no database required, the suite runs against an in-memory Store
go test ./... -race

# Single test
go test ./scim -run TestCreateElement -v

# Run the server — MUST be run from the repo root: config paths are relative
export SCIM_ADMIN_USER=Administrator    # Couchbase cluster user (default: Administrator)
export SCIM_ADMIN_PASSWORD='admin!'     # Couchbase cluster password (no default)
export SCIM_COUCHBASE_URL=localhost     # default: localhost
export SCIM_PORT=:8080                  # default: :8080
go run .

# Local Couchbase
docker run -d --name db -p 8091-8094:8091-8094 -p 11210:11210 couchbase
# then create the cluster at http://localhost:8091 with the same credentials
```

`SCIM_ADMIN_USER` / `SCIM_ADMIN_PASSWORD` are **Couchbase credentials only** — they never protect the HTTP API.

The HTTP API is protected by `SCIM_AUTH`, which is **required**: `jwt` (RFC 9068 access tokens validated against `SCIM_JWT_JWKS_URL` / `SCIM_JWT_ISSUER` / `SCIM_JWT_AUDIENCE`), `basic` (`SCIM_BASIC_USERS=user:password:role1,role2;...`, development only), or `none` (`SCIM_ANONYMOUS_ROLES=...`). An unset `SCIM_AUTH` is a startup error rather than an open server.

A live Couchbase is only needed to *run* the server. Tests wire a `MemoryStore` into `scim.DB` instead — see "Storage" below.

## Architecture

### Config-driven startup (the central idea)

`main.go` → `scim.InitDB()` → `scim.ReadResourceType("config", r)`. That single call (`scim/config.go:47`) does four things per file in `config/resourceType/`:

1. Loads every `config/schemas/*.json` into the global `scim.Schemas` map, keyed by schema `id` (a URN).
2. Loads the resource type into `scim.Resources`, keyed by its **`endpoint`** (e.g. `/Elements`).
3. Calls `DB.EnsureBucket(resourceType.Name)` — in production, creates a Couchbase bucket named after **`name`** (e.g. `Element`) plus a primary index, optionally configured from `config/bucketSettings/<Name>.json`.
4. Registers the six gin routes under `PREFIX = /scim/v2`.

**The `name` vs `endpoint` split matters everywhere**: `name` is the bucket/`meta.resourceType`, `endpoint` is the URL and the map key. Handlers are closures over the endpoint string (`Create(resourceType.Endpoint)`) that look the type back up via `Resources[resource]`.

Adding a resource type = drop a schema JSON + a resourceType JSON (+ optional bucketSettings JSON) and restart. No code changes.

### Storage

Handlers never touch Couchbase directly. They go through the `Store` interface in `scim/store.go`, held in the package global `scim.DB`:

- `store_couchbase.go` — production. `InitDB()` connects and installs it. One bucket per resource type, one document per resource keyed by UUID in the default collection. Because search uses `SELECT * FROM \`Bucket\``, N1QL nests each row under the bucket name, so the store unwraps `item[q.Bucket]` before returning. Driver errors for missing documents are translated to `ErrNotFound`, which the handlers map to 404 — handlers must not import gocb.
- `store_memory.go` — the test fake. It round-trips every document through JSON exactly like the real driver, so `Meta` structs come back as maps and integers as `float64`; without that the fake would hide the type-assertion panics that production hits. It does **not** evaluate SCIM filters and returns `ErrFilterUnsupported` for a non-empty filter — filter→N1QL correctness is covered by the `scim/parser` tests instead.

`SearchQuery.SortBy` carries a raw SCIM attribute path; translating it (quoting, N1QL) is the store's job, not the handler's.

### Tests

`scim/harness_test.go` builds the **real** router over a `MemoryStore` via `newTestServer(t)`, so routing, validation, meta generation and role filtering all execute. `scim/op_crud_test.go` drives it end to end with `httptest`. Tests that pin behavior known to violate the RFC are marked `TODO(Bn)` referencing `RELEASE-1.0.md` — when you fix bug *n*, those assertions are expected to change.

### Filter parser (`scim/parser/`)

ANTLR 4.7 grammar in `ScimFilter.g4` at the repo root. **Everything in `scim/parser/` is generated except `scimfilter_listener_implement.go`, which is hand-written and must never be overwritten.** It holds `FilterToN1QL` and `AddQuote`.

`FilterToN1QL(resourceName, filter)` returns *two* queries — the page query and a `count(*)` query — by walking the parse tree and doing **token-by-token string substitution** in `VisitTerminal`: `eq`→`=`, `co`/`sw`/`ew`→`LIKE` with `%` wrapping applied to the *next* token via the `prevOperation` field, `pr`→`IS NOT NULL`. There is no AST, no query builder, and **no parameter binding** — literals are concatenated into the SQL text. Parse errors are not checked, so malformed filters still emit a query.

`AddQuote` backtick-quotes attribute paths and splits a URN prefix from the attribute path using a regex. `scim/op_update.go:opPathTopathArray` duplicates that same regex for PATCH paths — change one and you must change the other.

Regenerating (only needed when editing the `.g4`):
```bash
java -jar antlr-4.7-complete.jar -Dlanguage=Go -o scim/parser ScimFilter.g4
```
The `.g4` declares `language = Java` in its `options` block; the `-Dlanguage=Go` flag overrides it. Generated filenames must stay lowercase (`scimfilter_*.go`) to match what is committed.

### Request pipeline

Create/Replace share one shape: unmarshal body → `ValidateFieldSchemas` (checks `schemas` array against `resourceType.Schema` + `schemaExtensions`) → `ValidateSchemas` (walks the core schema, then each extension object nested under its own URN key) → stamp `id`/`meta` → Couchbase upsert/replace.

`scim/validate.go` is a **strict whitelist**: any key present in the payload that is not declared in the schema is a 400. It dispatches per SCIM type (`string`/`boolean`/`decimal`/`integer`/`dateTime`/`binary`/`reference`/`complex`); `binary` and `reference` are no-ops. It also **ignores the `multiValued` flag entirely**, so an attribute declared `multiValued: true` is still validated as if it were scalar.

Update (PATCH) applies `add`/`replace`/`remove` operations by pointer-walking the stored document (`pointValue`), then funnels into the same `replace()` as PUT, so patched documents are re-validated.

### Authentication

`scim/auth.go` defines the `Authenticator` interface and the gin middleware that publishes a `Principal` (subject plus roles) into the request context. `NewRouter` puts every resource endpoint behind it; the discovery endpoints stay outside, which RFC 7644 §2 allows. Implementations live in `auth_jwt.go` (JWKS-backed, RSA only, algorithms restricted to RS256/384/512 so a symmetric or unexpected algorithm cannot be substituted), `auth_basic.go` and the `AnonymousAuthenticator` in `auth.go`.

Roles come from the token via `auth_claims.go`: RFC 9068 §2.2.3.1 reuses SCIM's own `roles`, `groups` and `entitlements` claims from RFC 7643 §4.1.2, and each may be a list of strings, a list of `{"value": ...}` objects, or a space-delimited string. `scope` is read too, plus an optional deployment-specific claim.

### Role-based access control

Two mechanisms, both applied per attribute and both recursing into complex and multi-valued values:

- **Reading** — `scim/validate_role.go`. `ValidateReadRole` drops any attribute whose `$reader` list does not intersect the caller's roles, and any attribute that is `returned: never` or `writeOnly` (`isReturnable`). Dropped means the key is absent, not empty. Every resource response goes through it, not only read and search.
- **Writing** — `scim/validate_write.go`. `EnforceWriteAccess` runs before validation on create and replace (and therefore on patch, which funnels into replace). `$writer` denies with 403; `mutability: readOnly` silently drops the client's value and restores the stored one, because read-modify-write clients echo whole resources back on a PUT; `mutability: immutable` refuses a change with 400 `scimType: mutability`.

`"*"` in either list means any role. Roles come from the authenticated principal via `currentRoles(c)`; a request with no principal has none, so restricted attributes stay hidden.

### Not implemented

`/ServiceProviderConfig`, `/ResourceTypes`, `/Schemas` are registered in `main.go` at the **root**, not under `/scim/v2`, and all three return 501 (`scim/discovery.go`) — `config/serviceProviderConfig/sp_config.json` is never read. `Bulk` returns `{"message":"pong"}` and its route is commented out; the `if resource == "Bulk"` branches in create/replace/delete are dead code (endpoint keys carry a leading `/`).

## Conventions and traps

- `ResoruceType` is misspelled in the source and is the real exported type name.
- The struct field for `$reader` is `Attribute.Read`; for `$writer` it is `Attribute.Writer`.
- Handlers return `func(c *gin.Context)` closures, not plain handlers.
- Responses go through `scim/response.go`: `writeSCIM` sets the `application/scim+json` media type (gin only fills in `Content-Type` when it is unset, so setting it first wins), and `entityTag`/`checkPrecondition` implement `ETag`, `If-Match` and `If-None-Match` over `meta.version`. Errors go through `MakeError` / `MakeTypedError`, the latter carrying the RFC 7644 3.12 `scimType` keyword.
- Config folder paths are hardcoded relative (`"config"`, `FolderBucketSetting = "config/bucketSettings/"`), so the process must start from the repo root.
- Manual API testing lives in `httpexamples/*.http` (REST Client format).
- Docs exist in `docs/{en,es,fr,gr,it,jp,pr,ch}/`; only `en` and `es` are complete.
