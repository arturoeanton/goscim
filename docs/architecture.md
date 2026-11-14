# Architecture

A compact Go codebase whose defining trait is that resource types, endpoints,
storage buckets, validation and permissions all come from JSON rather than from
code.

## Request path

```
                        ┌──────────────────────────────┐
  HTTP ────────────────▶│ gin router + RequestID       │  router.go, audit.go
                        └──────────────┬───────────────┘
                                       │
              /healthz, /readyz        ├──▶ health.go     (no authentication)
              /metrics                 ├──▶ metrics.go    (optional token)
              /ui                      ├──▶ ui.go         (optional HTMX console)
              /(tenants/:t/)?scim/v2/ServiceProviderConfig, /ResourceTypes, /Schemas
                                       │  (no authentication)
                                       ▼
                        ┌──────────────────────────────┐
                        │ discovery.go                 │
                        └──────────────────────────────┘
                                       │
              /(tenants/:t/)?scim/v2/{resource}
                                       ▼
                        ┌──────────────────────────────┐
                        │ AuditMiddleware → LimitBody  │  audit.go, limits.go
                        │                → Authenticate│  auth.go, tenant.go
                        └──────────────┬───────────────┘
                                       ▼
                        ┌──────────────────────────────┐
                        │ op_create / op_read / ...    │
                        │ op_bulk (sub-request router) │
                        └──────────────┬───────────────┘
                                       ▼
        ┌──────────────────────────────────────────────────────┐
        │ validate.go        types, required, multiValued      │
        │ validate_write.go  $writer, mutability               │
        │ password.go        Argon2id / bcrypt / reject        │
        │ uniqueness.go      atomic uniqueness claims          │
        │ groups_sync.go     Group.members ↔ User.groups sync  │
        │ webhooks.go        CloudEvents 1.0 sharded dispatch  │
        │ validate_role.go   $reader, returned  (on the way out)│
        │ projection.go      attributes / excludedAttributes   │
        └──────────────────────────────┬───────────────────────┘
                                       ▼
                     ┌───────────────────────────────────┐
                     │ Store  (scim/store.go)            │
                     ├───────────┬───────────┬───────────┤
                     │ Couchbase │  SQLite   │ in-memory │
                     │ (cluster) │(appliance)│(demo/test)│
                     └───────────┴───────────┴───────────┘
```

## Startup

`main.go` handles CLI subcommands (`conformance`, `version`, `-healthcheck`,
`-demo`) and otherwise starts the HTTP server in four steps, stopping at the
first failure:

1. Build an `Authenticator` from `SCIM_AUTH`. An unset value is an error (unless
   `-demo` supplies defaults).
2. Initialise the configured `Store` (`NewStoreFromEnv` in `scim/store_init.go`:
   `couchbase`, `sqlite`, or `memory`) and optionally start the outbound
   `WebhookDispatcher` (`scim/webhooks.go`).
3. `NewRouter` reads `SCIM_CONFIG_DIR` (falling back to the embedded
   `config/embed.go` FS when no directory exists on disk) and registers
   `/scim/v2/...`, `/tenants/:tenant/scim/v2/...`, `/metrics`, and optional
   `/ui`.
4. Serve, with timeouts, until `SIGINT` or `SIGTERM`, then drain HTTP
   connections, flush outbound webhook workers, and close the store.

`ReadResourceType` is where the config-driven part happens. For each file in
`config/resourceType/` it:

- loads every schema in `config/schemas/` into the `Schemas` map, keyed by URN;
- records the resource type in `Resources`, keyed by its **endpoint**;
- calls `Store.EnsureBucket(resourceType.Name)` and, on Couchbase with
  `SCIM_COUCHBASE_AUTO_INDEX=true`, `EnsureResourceIndexes(resourceType)` to
  build secondary GSI N1QL indexes (`idx_scim_*`) for declared scalar,
  complex, extension, `meta.*`, and multi-valued array sub-attributes;
- registers the seven per-resource routes (`POST`, `GET /:id`, `PUT /:id`,
  `DELETE /:id`, `PATCH /:id`, `GET`, `POST /.search`) under both `/scim/v2` and
  `/tenants/:tenant/scim/v2`, after which `NewRouter` mounts server-root
  `POST /.search` and `POST /Bulk` on both prefixes.

### name versus endpoint

The one piece of vocabulary worth internalising. A resource type has both, and
they are not interchangeable:

- **`name`** — the Couchbase bucket, and `meta.resourceType`. `User`.
- **`endpoint`** — the URL path, and the key in `Resources`. `/Users`.

Handlers are closures over the endpoint string and look the type back up, so a
handler always knows which resource it serves without a global.

## Storage

Handlers never touch Couchbase. They go through `Store`:

```go
type Store interface {
	EnsureBucket(name string) error
	Get(bucket, id string) (map[string]interface{}, error)
	Upsert(bucket, id string, doc map[string]interface{}) error
	Replace(bucket, id string, doc map[string]interface{}) error
	Remove(bucket, id string) error
	CreateUnique(bucket, id string, doc map[string]interface{}, claims []UniqueClaim) error
	ReplaceUnique(bucket, id string, doc map[string]interface{}, newClaims []UniqueClaim, extractClaims func(map[string]interface{}) []UniqueClaim) error
	RemoveUnique(bucket, id string, claims []UniqueClaim) error
	Search(q SearchQuery) (total int, resources []map[string]interface{}, err error)
	FindIDByAttribute(bucket, attributePath string, value interface{}) (string, error)
	Ping(ctx context.Context) error
	Close() error
}
```

**`CouchbaseStore`** targets clustered production. One bucket per resource type,
one document per resource keyed by `<id>` (or `t:<tenant>:<id>` under
multi-tenancy) in `_default._default`, plus a dedicated `_default.scim_unique`
collection holding KV sentinel documents (`u:<sha256>`) for
`uniqueness: server | global`. Keeping sentinels in their own collection makes
`CreateUnique` and `ReplaceUnique` linearizable via KV `Insert` and document CAS
without polluting `SELECT * FROM \`Bucket\`` or `SELECT count(*)`. On startup,
`store_couchbase_indexes.go` walks the JSON schemas and issues
`CREATE INDEX IF NOT EXISTS` statements (`idx_scim_*`) covering `externalId`,
`meta.created`, `meta.lastModified`, declared scalar/complex attributes, and
`DISTINCT ARRAY` indexes on multi-valued `value` sub-attributes (`emails`,
`groups`, `members`).

**`SQLiteStore`** (`scim/store_sqlite.go`, `SCIM_STORE=sqlite`) provides durable
single-file storage via pure-Go `modernc.org/sqlite` (`CGO_ENABLED=0`, WAL
journal mode, `busy_timeout=5000`). Resources and uniqueness claims are updated
inside serialised SQL transactions (`scim_resources` and `scim_unique` tables),
guaranteeing atomic `uniqueness: server | global` enforcement and full
multi-tenant isolation (`ScopedBucket`) without external infrastructure.

**`MemoryStore`** (`SCIM_STORE=memory`) backs `-demo` and the unit test suite.
It round-trips every document through JSON exactly as Couchbase does, so a
`Meta` struct comes back as a map and integers come back as `float64`. Without
that the fake would hide the type-assertion bugs production hits — which is not
hypothetical: two such bugs in the 1.0 audit were only visible because of it.
Both `SQLiteStore` and `MemoryStore` evaluate SCIM filters through
`parser.CompileMatcherWithOptions` (honouring `Resolve`, `IsCaseExact`, and
multi-valued slice traversal), while N1QL translation is independently verified
by `scim/parser` unit tests and by the integration suite against a real
Couchbase cluster.

## Multi-tenancy, Webhooks, and Embedded UI

- **Multi-tenancy (`scim/tenant.go`)**: `ScopedBucket(tenant, bucket)` encodes
  tenant context as `t:<tenant>:<bucket>` across all three `Store` engines.
  `uniqueness: server` claims hash the scoped bucket (isolating `userName` per
  tenant), whereas `uniqueness: global` claims hash `*` across the whole server.
  `Authenticate` enforces credential-to-tenant pins (`403` on cross-tenant path
  access).
- **CloudEvents 1.0 Webhooks (`scim/webhooks.go`)**: mutations emit
  `io.goscim.<resource>.created|updated|activated|deactivated|deleted` events
  onto worker queues sharded by `fnv32a(tenant + ":" + subject)`, preserving
  causal per-resource ordering even across HTTP `429`/`5xx` exponential-backoff
  retries, signed with `X-SCIM-Signature-256: sha256=<hex>` and stripped of
  `writeOnly` / `returned: never` secrets.
- **HTMX Web Console (`scim/ui.go`)**: embeds HTMX 2.0.4 and server-rendered
  `html/template` views in the Go binary. UI mutations and 1-click conformance
  runs dispatch directly through the real `gin.Engine` via an unexported Go
  context key (`internalUICtxKey{}`), exercising validation, RBAC, password
  hashing, uniqueness, audit, metrics, inspector capture, and webhooks without
  opening any network auth bypass.

## The filter parser

`ScimFilter.g4` at the repository root, generated with ANTLR 4.13 against
`github.com/antlr4-go/antlr/v4`. Everything in `scim/parser/` is generated
except `scimfilter_listener_implement.go`, `errors.go`, and `eval.go`.

`FilterToN1QLWithOptions` walks the parse tree and emits N1QL: `eq` becomes `=`
(or `LOWER(attr) = LOWER($n)` when `caseExact: false`), `co`/`sw`/`ew` become
`LIKE ... ESCAPE "\\"` (folded with `LOWER(...)` when `caseExact: false`) with
wildcards applied to the bound value, `pr` becomes `IS NOT NULL`, and both
bracketed array filters (`emails[type eq "work"]`,
`emails[type eq "work"].value co "@corp.com"`) and dotted multi-valued paths
(`emails.value eq "x"`, `members.value eq "u1"`) become quantified
`(ANY \`_v0\` IN \`emails\` SATISFIES ... END)` predicates.
Separately, `eval.go` (`CompileMatcherWithOptions`, `Match`, `EqualitySeeds`)
evaluates the same ANTLR parse tree against decoded JSON maps in memory for
`PATCH` value-selection paths and `MemoryStore.Search`.

Two properties matter more than the mechanism:

**Values never enter the query text.** They are collected and returned as bound
parameters. Escaping can be got wrong one case at a time — this release fixed
three separate holes that were each an escaping mistake — whereas a bound value
cannot become syntax whatever it contains. A property test asserts that no
filter value appears in the generated text, for any operator.

**A filter that does not parse is rejected.** ANTLR recovers from syntax errors
and keeps walking, so an unchecked tree emits whatever the client sent. An
error listener collects the failures and the translation refuses to produce a
query.

Attribute paths are not values, so they cannot be bound; they are resolved
against the schema through a `NameResolver` and quoted with backticks doubled.
`sortBy` is separately restricted to declared attributes.

## Authentication and authorization

`Authenticator` turns a request into a `Principal` — a subject and a set of
roles — and the middleware publishes it for the handlers. Four
implementations: JWT against a JWKS (`SCIM_AUTH=jwt`), static opaque Bearer
tokens (`SCIM_AUTH=bearer`), HTTP Basic (`SCIM_AUTH=basic`), and an explicit
anonymous mode (`SCIM_AUTH=none`).

Authorization is per attribute rather than per endpoint, and runs on both
sides:

- **Inbound**, `EnforceWriteAccess`: `$writer` roles, and SCIM's `mutability`.
- **Outbound**, `ValidateReadRole`: `$reader` roles, and `returned`/`writeOnly`.

Both walk the attribute definitions rather than string paths, so complex
values, multi-valued elements and extension objects are covered by
construction rather than by remembering to handle each case.

## Group membership synchronisation and Bulk

Two features sit above single-document CRUD without leaking storage details into
the handlers:

- **`Group.members` ↔ `User.groups` (`scim/groups_sync.go`)**: RFC 7643 §4.1.2
  makes `User.groups` `readOnly` and derived from `Group.members`. Computing it
  lazily on `GET /Users` would break N1QL filters on `User.groups` (since
  Couchbase buckets are separate and SCIM queries do not join across them) and
  penalise reads. Instead, `Group` mutations compute the before/after membership
  graph in memory (resolving nested groups via a cycle-safe BFS where `"direct"`
  takes precedence over `"indirect"`) and differentially update only the `User`
  documents whose `groups` array actually changed. Conversely, deleting a `User`
  or `Group` cleans up dangling references in `Group.members`.
- **`POST /scim/v2/Bulk` (`scim/op_bulk.go`)**: rather than maintaining a second
  copy of validation, RBAC, password hashing, uniqueness and audit logic,
  `NewRouter` builds an internal `gin.Engine` (`newBulkEngine`) mounting `POST`,
  `PUT`, `PATCH` and `DELETE` for every configured resource type. `BulkHandler`
  validates the envelope once, resolves `bulkId:<token>` references across paths,
  `$ref` URIs, `data` maps and `PATCH` filter literals as `201 Created` IDs
  accumulate, attaches the outer `Principal` and `X-Request-Id` to each
  sub-request's context, and dispatches through `subRouter.ServeHTTP`. Every
  guarantee of direct CRUD therefore holds inside `/Bulk` by construction.

## Deliberate omissions

- **No ORM, no query builder.** The data is schemaless JSON and the queries are
  simple. A layer would earn its keep only once joins appear, and SCIM has none.
- **No storage abstraction beyond `Store`.** `CouchbaseStore`, `SQLiteStore`,
  and `MemoryStore` implement one narrow 11-method interface. A PostgreSQL
  backend lives in a [sister project](https://github.com/arturoeanton/go-vorpal-scim).
- **No caching.** Search consistency is `request_plus` precisely so a freshly
  written resource is visible; a cache in front would undo that.
- **The handlers write responses directly.** `*gin.Context` reaches
  `validate.go`, which couples the domain to the framework. It is the main
  thing worth changing if the HTTP layer is ever replaced.
