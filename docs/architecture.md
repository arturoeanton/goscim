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
              /scim/v2/ServiceProviderConfig, /ResourceTypes, /Schemas
                                       │  (no authentication)
                                       ▼
                        ┌──────────────────────────────┐
                        │ discovery.go                 │
                        └──────────────────────────────┘
                                       │
              /scim/v2/{resource}      │
                                       ▼
                        ┌──────────────────────────────┐
                        │ AuditMiddleware → LimitBody  │  audit.go, limits.go
                        │                → Authenticate│  auth.go
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
        │ validate_role.go   $reader, returned  (on the way out)│
        │ projection.go      attributes / excludedAttributes   │
        └──────────────────────────────┬───────────────────────┘
                                       ▼
                        ┌──────────────────────────────┐
                        │ Store  (scim/store.go)       │
                        ├──────────────┬───────────────┤
                        │ Couchbase    │ in-memory     │
                        │ (production) │ (tests)       │
                        └──────────────┴───────────────┘
```

## Startup

`main.go` does four things in order, and stops at the first failure:

1. Build an `Authenticator` from `SCIM_AUTH`. An unset value is an error.
2. Connect to Couchbase and install a `CouchbaseStore` as the active `Store`.
3. `NewRouter` reads the config directory and registers routes.
4. Serve, with timeouts, until `SIGINT` or `SIGTERM`, then drain.

`ReadResourceType` is where the config-driven part happens. For each file in
`config/resourceType/` it:

- loads every schema in `config/schemas/` into the `Schemas` map, keyed by URN;
- records the resource type in `Resources`, keyed by its **endpoint**;
- calls `Store.EnsureBucket(resourceType.Name)`, creating the bucket, its
  `_default.scim_unique` sentinel collection, and its primary index if they do
  not exist;
- registers the seven per-resource routes (`POST`, `GET /:id`, `PUT /:id`,
  `DELETE /:id`, `PATCH /:id`, `GET`, `POST /.search`) under `/scim/v2` +
  endpoint, after which `NewRouter` mounts server-root `POST /scim/v2/.search`
  and `POST /scim/v2/Bulk`.

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

**`CouchbaseStore`** is production. One bucket per resource type, one document
per resource keyed by its SCIM `id` in `_default._default`, plus a dedicated
`_default.scim_unique` collection holding KV sentinel documents (`u:<sha256>`)
for `uniqueness: server | global`. Keeping sentinels in their own collection
makes `CreateUnique` and `ReplaceUnique` linearizable via KV `Insert` and
document CAS without polluting `SELECT * FROM \`Bucket\`` or `SELECT count(*)`.
Search is N1QL. Because `SELECT * FROM \`Bucket\`` nests each row under the
bucket name, the store unwraps that before returning. Driver errors for a
missing document become `ErrNotFound`, so handlers never import gocb.

**`MemoryStore`** backs the tests. It round-trips every document through JSON
exactly as the driver does, so a `Meta` struct comes back as a map and integers
come back as `float64`. Without that the fake would hide the type-assertion
bugs production hits — which is not hypothetical: two such bugs in the 1.0
audit were only visible because of it.

The fake deliberately does **not** evaluate SCIM filters. Translating filters
is the storage layer's concern, and pretending otherwise would mean a second
filter implementation drifting from the first. A non-empty filter returns
`ErrFilterUnsupported`, and filter behaviour is covered by the parser unit
tests and by the integration suite against a real cluster.

## The filter parser

`ScimFilter.g4` at the repository root, generated with ANTLR 4.13 against
`github.com/antlr4-go/antlr/v4`. Everything in `scim/parser/` is generated
except `scimfilter_listener_implement.go`, `errors.go`, and `eval.go`.

`FilterToN1QL` walks the parse tree and emits N1QL: `eq` becomes `=`,
`co`/`sw`/`ew` become `LIKE ... ESCAPE "\\"` with wildcards applied to the
bound value, `pr` becomes `IS NOT NULL`, and complex array filters
(`emails[type eq "work"]`, `emails[type eq "work"].value co "@corp.com"`)
become quantified `(ANY \`_v0\` IN \`emails\` SATISFIES ... END)` predicates.
Separately, `eval.go` (`CompileMatcher`, `Match`, `EqualitySeeds`) evaluates the
same ANTLR parse tree against decoded JSON maps in memory so `PATCH` can resolve
value-selection paths (`emails[type eq "work"].value`) without N1QL.

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
roles — and the middleware publishes it for the handlers. Three
implementations: JWT against a JWKS, HTTP Basic, and an explicit anonymous
mode.

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
- **No storage abstraction beyond `Store`.** It exists to make the server
  testable, not to support a second database. A PostgreSQL backend lives in a
  [sister project](https://github.com/arturoeanton/go-vorpal-scim).
- **No caching.** Search consistency is `request_plus` precisely so a freshly
  written resource is visible; a cache in front would undo that.
- **The handlers write responses directly.** `*gin.Context` reaches
  `validate.go`, which couples the domain to the framework. It is the main
  thing worth changing if the HTTP layer is ever replaced.
