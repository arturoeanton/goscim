# Changelog

## 1.1.0

Interoperability, storage concurrency, authentication and operational release.

### Added

- **Password hashing** (`SCIM_PASSWORD_MODE`): `writeOnly` `password` attributes
  are hashed before storage with `argon2id` by default, with `bcrypt`, `reject`
  (`400 invalidValue`, advertising `changePassword.supported: false` in
  `/ServiceProviderConfig`) and `plain` (opt-in, logs a warning) also available.
  Omitting `password` on `PUT` or `PATCH` preserves the stored hash.
- **Attribute projection** (`RFC 7644 §3.4.2.5` and `§3.9`): `attributes` and
  `excludedAttributes` are honoured across `GET`, `POST`, `PUT`, `PATCH` and
  `POST /.search`, alongside `returned: always | default | request | never` and
  both `:` and `.` URN delimiters.
- **`POST /.search`** (`RFC 7644 §3.4.3`): `POST /scim/v2/{resource}/.search` and
  server-root `POST /scim/v2/.search` across all configured resource types,
  accepting `SearchRequest` bodies with `filter`, `sortBy`, `sortOrder`,
  `startIndex`, `count`, `attributes` and `excludedAttributes`.
- **Complex array value filters** (`emails[type eq "work"]` and
  `emails[type eq "work"].value`):
  - Translated to N1QL quantified `(ANY ... IN ... SATISFIES ... END)`
    predicates with bound positional parameters in `GET` and `POST /.search`.
  - Evaluated in memory (`scim/parser/eval.go`) on `PATCH` `add`, `replace` and
    `remove` (`RFC 7644 §3.5.2`), including Microsoft Entra ID `add` entry
    synthesis from conjunctive `eq` predicates, empty-array unassignment on
    `remove`, colon-delimited URN paths, and `$writer` / `$reader` / `writeOnly`
    enforcement.
- **EC and OKP signing keys in JWKS**: `SCIM_AUTH=jwt` now verifies `ES256`
  (`P-256`), `ES384` (`P-384`), `ES512` (`P-521`) via `crypto/ecdh` and `EdDSA`
  (`Ed25519` OKP) alongside `RS256` / `RS384` / `RS512`, enforcing strict curve
  and algorithm binding.
- **Structured audit logging and `X-Request-Id`**: `scim.audit` `log/slog`
  records (`SCIM_AUDIT_LOG=json|text|off`) capturing caller `subject`, `roles`,
  action, resource type/id, HTTP status, outcome and latency, correlated via a
  validated `X-Request-Id` header.
- **Health and readiness probes**: `GET /healthz`, `GET /readyz` (`Store.Ping`
  against Couchbase KV and Query services under a 2s deadline), `-healthcheck`
  CLI flag, distroless multi-stage `Dockerfile`, `docker-compose.yml` and
  `.github/dependabot.yml`.

### Changed

- **Atomic uniqueness enforcement**: `CouchbaseStore` now claims `uniqueness:
  server | global` values in a dedicated `_default.scim_unique` KV collection
  (`u:<sha256>`) using atomic `Insert` and document CAS on `PUT` / `PATCH`,
  closing the read-then-write TOCTOU race and folding case when
  `caseExact: false`.
- `/ServiceProviderConfig` advertises `changePassword.supported: true` unless
  `SCIM_PASSWORD_MODE=reject`.

## 1.0.1

Security and dependency maintenance release. Details of every finding and fix
are recorded in `FIX.SECURITY.md`.

### Fixed

- `$writer`, `mutability: readOnly` and `mutability: immutable` could be
  bypassed by sending attribute names or extension URNs in a different case, or
  by sending duplicate JSON keys that differed only in case. `EnforceWriteAccess`
  now resolves keys case-insensitively and rejects case-colliding keys before
  validation.
- Omitting an attribute on `PUT` or targeting it with `PATCH remove` bypassed
  `immutable` and `$writer` checks. Removing an `immutable` attribute is now
  refused with `400` (`scimType: mutability`); `PATCH` checks `$writer` on the
  attributes each operation touches while preserving untouched ones; `PUT`
  preserves non-readable `$writer` attributes without exposing a `200`-vs-`403`
  value oracle.
- `DELETE` did not check `$writer` permissions on the target resource.
- `writeOnly` and `returned: never` attributes (such as `User.password`) could
  be probed through `filter` and `sortBy`, and callers could filter on
  attributes outside their `$reader` roles. Both are now refused.
- Permissive rules in `ScimFilter.g4` accepted `--` and `1.2.3` under `NUMBERS`
  (emitting raw tokens when `strconv.ParseFloat` failed), adjacent expressions
  without `and`/`or`, expression-to-expression comparisons, and unsupported
  bracketed value filters (`emails[type eq "work"]`, which reached Couchbase and
  failed with `500` instead of `400 invalidFilter`). The parse tree is now
  validated before translation.
- JWKS refreshes are serialized and rate-limited across failures too, response
  bodies are capped at 1 MiB, JWK `use` / `key_ops` are enforced, and RSA keys
  shorter than 2048 bits are rejected.
- Malformed or `null` JSON bodies on `POST` and `PUT` now return `400`
  (`scimType: invalidSyntax`); `500` responses no longer echo raw driver errors
  to clients; `hidepassword` no longer logs the first and last character of the
  Couchbase password; `int64` range checks in `validateValue` no longer pass
  `2^63` on `arm64`; `X-Content-Type-Options: nosniff` is set on responses.

### Added

- `SCIM_PUBLIC_URL` to build `Location` headers from a configured public origin
  rather than the incoming `Host` header, plus `Host` validation when unset.

### Changed

- Toolchain moved to Go 1.26 (`go1.27.4`), resolving `GO-2026-6859` in
  `net/http.MaxBytesReader`.
- Updated `golang.org/x/net` (`v0.62.0`), `golang.org/x/crypto` (`v0.59.0`),
  `golang.org/x/text` (`v0.44.0`), `google.golang.org/grpc` (`v1.85.0`) and
  `go.opentelemetry.io/otel` (`v1.47.0`), plus `klauspost/compress`,
  `moby/go-archive` and `moby/sys/user` in `test/integration`.

## 1.0.0

First release intended for use. Everything before this was `v0.1` and, as the
audit in `RELEASE-1.0.md` records, could not create a standard SCIM user, had
no authentication, and had two reachable panics and three injection paths.

### Breaking changes

Anything written against the previous code needs to be looked at.

- **Authentication is required.** `SCIM_AUTH` must be set to `jwt`, `basic` or
  `none`; the server refuses to start without it. Every resource endpoint is
  behind it. Discovery stays anonymous.
- **`POST` answers `201 Created`** with a `Location` header, not `200 OK`.
- **Responses use `application/scim+json`**, not `application/json`.
- **Attributes the caller may not read are omitted**, where they used to be
  returned as an empty string. This now applies to create, replace and patch
  responses too, not only read and search.
- **`meta.location` carries the `/scim/v2` prefix**: `/scim/v2/Users/{id}`.
- **Discovery moved under `/scim/v2`** and returns real data instead of `501`.
- **`ResoruceType` is spelled `ResourceType`**, and `FolderResoruceType` is
  `FolderResourceType`.
- **The `commons` package is gone.** It had no callers left.
- **`Store` is a new interface** and `CreateBucket` is now
  `Store.EnsureBucket`. `parser.FilterToN1QL` returns a `parser.Query` with
  bound parameters and takes a `NameResolver`.
- **`compression_mode` was removed from the shipped bucket settings.** It is
  Enterprise Edition only and made the server fail to start against Community.
- **`config/schemas/...core:2.0:Element`**: `name` and `description` were
  `readOnly` while also declaring `$writer: ["*"]`. They are `readWrite`, which
  is what the `$writer` already said.

### Fixed

- `gt`/`ge` and `lt`/`le` were crossed with each other, so every range query
  returned the wrong set.
- Filters that failed to parse were translated anyway, letting arbitrary text
  into the generated N1QL.
- `sortBy` was interpolated into `ORDER BY` without validation.
- A filter value ending in a backslash escaped the closing quote of its own
  string literal.
- `multiValued` attributes were validated as if they were scalar, which made a
  standard SCIM user impossible to create.
- Optional schema extensions were treated as required.
- A `PATCH` naming a path that does not resolve panicked, as did a `PUT`
  without `meta` in the body.
- The read filter skipped arrays and extension attributes entirely.
- `password` was returned in every response despite being declared
  `returned: never`.
- Searches ran with N1QL's default consistency, so a resource that had just
  been created was not found by a search.
- `CREATE PRIMARY INDEX` was issued before the query service could see a
  newly created bucket, failing the first startup against a fresh cluster.
- `totalResults` and `itemsPerPage` were omitted from an empty search result.

### Added

- Authentication: JWT (RFC 9068 access tokens validated against a JWKS),
  HTTP Basic for development, and an explicit anonymous mode.
- Per-attribute authorization: `$reader` and `$writer` role lists, plus SCIM's
  own `mutability` and `returned` keywords.
- `uniqueness: server|global`, answered with `409` and `scimType: uniqueness`.
- `ETag`, `If-Match` and `If-None-Match` over `meta.version`.
- Discovery: `/ServiceProviderConfig`, `/ResourceTypes`, `/ResourceTypes/{id}`,
  `/Schemas`, `/Schemas/{id}`.
- Request timeouts, a 1 MiB body limit, a cap on search page size, and
  graceful shutdown on `SIGTERM`.
- `SCIM_CONFIG_DIR`, `SCIM_TRUSTED_PROXIES`, `SCIM_QUERY_CONSISTENCY`,
  `SCIM_COUCHBASE_TLS`, `SCIM_COUCHBASE_CA_CERT`,
  `SCIM_COUCHBASE_TLS_SKIP_VERIFY`.
- A test suite: unit tests over the whole request path against an in-memory
  store, and `make integration` against a real Couchbase.
- CI, and a `Makefile` that records how to build, check and regenerate the
  parser.

### Changed

- Couchbase TLS certificates are verified by default. Verification used to be
  disabled unconditionally.
- Go 1.16 → 1.25; gin 1.7.7 → 1.12.0; gocb 2.3.5 → 2.12.4; the ANTLR runtime
  moved to the maintained `github.com/antlr4-go/antlr/v4`.

### Not implemented

Named here because the previous service provider config claimed some of them:

- `/Bulk`.
- Filters with value paths, `emails[type eq "work"]`, in filters or patch paths.
- `attributes` and `excludedAttributes`.
- `POST /.search`.
- Health, readiness and metrics endpoints.
