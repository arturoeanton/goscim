# Architecture

Around 2,000 lines of hand-written Go. The defining trait is that resource
types, endpoints, storage buckets, validation and permissions all come from
JSON rather than from code.

## Request path

```
                        ┌──────────────────────────────┐
  HTTP ────────────────▶│ gin router  (scim/router.go) │
                        └──────────────┬───────────────┘
                                       │
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
                        │ LimitBody  → Authenticate    │  auth.go, limits.go
                        └──────────────┬───────────────┘
                                       ▼
                        ┌──────────────────────────────┐
                        │ op_create / op_read / ...    │
                        └──────────────┬───────────────┘
                                       ▼
        ┌──────────────────────────────────────────────────────┐
        │ validate.go        types, required, multiValued      │
        │ validate_write.go  $writer, mutability               │
        │ uniqueness.go      uniqueness: server | global       │
        │ validate_role.go   $reader, returned  (on the way out)│
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
- calls `Store.EnsureBucket(resourceType.Name)`, creating the bucket and its
  primary index if they do not exist;
- registers the six verbs under `/scim/v2` + endpoint.

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
	Search(q SearchQuery) (total int, resources []map[string]interface{}, err error)
	FindIDByAttribute(bucket, attributePath string, value interface{}) (string, error)
	Close() error
}
```

**`CouchbaseStore`** is production. One bucket per resource type, one document
per resource keyed by its SCIM `id`. Search is N1QL. Because `SELECT * FROM
`Bucket`` nests each row under the bucket name, the store unwraps that before
returning. Driver errors for a missing document become `ErrNotFound`, so
handlers never import gocb.

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
except `scimfilter_listener_implement.go` and `errors.go`.

`FilterToN1QL` walks the parse tree and emits N1QL by token substitution:
`eq` becomes `=`, `co`/`sw`/`ew` become `LIKE` with the wildcards applied to
the value, `pr` becomes `IS NOT NULL`. There is no AST and no query builder.

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
