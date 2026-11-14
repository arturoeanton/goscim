# Changelog

## 1.4.0

Turnkey experience, pluggable persistence, HTMX console, observability, webhooks, and B2B multi-tenancy release.

### Added

- **Zero-dependency `-demo` mode & `SCIM_STORE=memory`**: `goscim -demo` starts
  in milliseconds from any working directory using embedded default schemas,
  seeds sample SCIM 2.0 `Users` and nested `Groups` (demonstrating `direct` and
  `indirect` `User.groups` memberships), enables `/ui`, and prints ready-to-run
  `curl` and `goscim conformance` commands.
- **Embedded SQLite store (`SCIM_STORE=sqlite`, `SCIM_SQLITE_PATH`)**: pure-Go
  (`modernc.org/sqlite`, `CGO_ENABLED=0`) ACID single-file persistence in `WAL`
  mode supporting atomic `uniqueness: server` and cross-tenant
  `uniqueness: global` claims without external database infrastructure.
- **Embedded HTMX Web Console & Live Traffic Inspector (`/ui`, `-ui` /
  `SCIM_UI=true`)**: air-gapped single-binary UI (with vendored HTMX 2.0.4 and
  optional `SCIM_UI_TOKEN` guard) featuring live telemetry cards, `Users` /
  `Groups` / `Elements` directory CRUD, an interactive SCIM-to-N1QL filter
  playground, a 200-entry `ngrok`-style live request/response inspector (with
  automatic secret redaction), and a one-click conformance runner.
- **Prometheus metrics endpoint (`GET /metrics`)**: zero-dependency Prometheus
  0.0.4 text exposition (`goscim_build_info`, `goscim_store_up`,
  `goscim_http_requests_total`, `goscim_http_request_duration_seconds`,
  `goscim_auth_failures_total`, `goscim_uniqueness_conflicts_total`,
  `goscim_webhook_deliveries_total`), enabled by default (`SCIM_METRICS`) with
  optional constant-time Bearer protection (`SCIM_METRICS_TOKEN`).
- **Automatic Couchbase secondary N1QL (`GSI`) indexes**: `EnsureResourceIndexes`
  (`SCIM_COUCHBASE_AUTO_INDEX`, enabled by default) inspects loaded JSON schemas
  at startup and creates `IF NOT EXISTS` secondary indexes for `_scim_tenant`,
  `externalId`, `uniqueness: server|global` attributes (`LOWER(...)` when
  `caseExact: false`), and `DISTINCT ARRAY` indexes for multi-valued complex
  `.value` sub-attributes.
- **Signed outbound CloudEvents 1.0 webhooks**: `SCIM_WEBHOOK_URLS`,
  `SCIM_WEBHOOK_SECRET` (`X-SCIM-Signature-256: sha256=<hmac>`),
  `SCIM_WEBHOOK_EVENTS`, and `SCIM_WEBHOOK_MAX_RETRIES`, with per-resource shard
  ordering, `io.goscim.user.deactivated` transition detection, exponential
  retries on `429`/`5xx`, and schema-aware stripping of `returned: "never"`
  fields.
- **Live conformance verifier (`goscim conformance`)**: CLI subcommand (`-url`,
  `-token`, `-basic`, `-profile all|rfc|entra|okta|keycloak`, `-json`) and
  programmatic `scim.RunConformance` runner verifying live SCIM 2.0 servers end
  to end and cleaning up all temporary records.
- **Native B2B SaaS multi-tenancy**: `/tenants/:tenant/scim/v2/...` URL routes
  plus credential-bound tenant isolation via `@<tenant>` in `SCIM_BEARER_TOKENS`,
  `SCIM_BEARER_TENANT`, and `SCIM_JWT_TENANT_CLAIM`, isolating resources,
  searches, `/Bulk`, `Group.members <-> User.groups`, and `uniqueness: server`
  per tenant across `couchbase`, `sqlite`, and `memory` stores.
- **Release, installer, Keycloak lab, and Helm chart**: `.goreleaser.yaml`,
  `.github/workflows/release.yml` (multi-OS/arch binaries + multi-arch
  `ghcr.io/arturoeanton/goscim` image), `scripts/install.sh`,
  `docker-compose.keycloak.yml` + `deploy/keycloak/goscim-realm.json`, and
  `charts/goscim/`.

## 1.3.0

Microsoft Entra ID, Okta, and Keycloak compatibility release.

### Added

- **Static Bearer token authentication (`SCIM_AUTH=bearer`)**: supports opaque
  shared-secret tokens (`SCIM_BEARER_TOKEN` / `SCIM_BEARER_TOKENS`,
  `SCIM_BEARER_SUBJECT`, `SCIM_BEARER_ROLES`) for Microsoft Entra ID's Secret
  Token provisioning mode and Okta's API Token mode, comparing every configured
  token in constant time via SHA-256 without early exit and advertising
  `oauthbearertoken` in `/ServiceProviderConfig`.
- **Extended OAuth 2.0 / OIDC claim extraction (`SCIM_AUTH=jwt`)**:
  - Reads `scp` (Microsoft Entra ID v2.0 space-delimited scopes and Okta scope
    arrays) and `realm_access.roles` (Keycloak realm roles) by default alongside
    `roles`, `groups`, `entitlements`, and `scope`.
  - Supports dotted nested claim paths and comma-separated claim lists in
    `SCIM_JWT_ROLES_CLAIM` (such as `resource_access.goscim-service.roles` for
    Keycloak client roles) while preserving literal claim keys containing dots.
  - Falls back from `sub` through `client_id`, `cid`, `azp`, `oid`, and `appid`
    for client-credentials and service-account tokens that omit `sub`.
- **RFC 7643 §3.1 `externalId` common attribute**: supported end to end across
  create, read, replace, patch, attribute projection (`attributes` /
  `excludedAttributes`), `sortBy`, and case-exact (`caseExact: true`) search
  filtering in both Couchbase N1QL and `MemoryStore`.
- **RFC 7643 §4.1.2 `addresses.primary` sub-attribute**: added to the shipped
  `urn:ietf:params:scim:schemas:core:2.0:User` schema so default address
  mappings from Microsoft Entra ID and Okta validate cleanly.
- **Explicit JSON `null` clearing**: setting optional attributes, sub-attributes,
  or non-required schema extensions (`EnterpriseUser: null`) to `null` on `POST`
  or `PUT` unassigns them cleanly while continuing to reject `null` on
  `required: true` attributes.
- **Microsoft Entra ID provisioning dialect support**:
  - Expands dotted (`name.givenName`) and colon-URN
    (`urn:...:EnterpriseUser:department`) keys inside pathless `PATCH` `add` and
    `replace` value maps, automatically attaching referenced extension URNs to
    `schemas`.
  - Supports element-selective removal on `PATCH remove` when `path` names a
    multi-valued attribute without a bracket filter and `value` supplies the
    elements to remove (`{"op": "Remove", "path": "members", "value":
    [{"value": "<id>"}]}`).
  - Normalises legacy stringified booleans (`"True"` / `"False"`) on `boolean`
    attributes and stringified `EnterpriseUser.manager` IDs (`"manager": "<id>"`
    to `{"value": "<id>"}`).
- **IdP conformance and integration suites**: added `scim/compat_entra_test.go`,
  `scim/compat_okta_test.go`, `scim/compat_keycloak_test.go`,
  `test/integration/idp_compat_test.go`, `docs/idp-compatibility.md`, and
  `httpexamples/idp_{entra,okta,keycloak}.http`.

### Changed

- **Case-aware and multi-valued SCIM filter engine**:
  - `FilterToN1QLWithOptions` folds case (`LOWER(...)`) on string comparisons
    when the schema declares `caseExact: false` (such as `userName`) while
    preserving exact comparisons when `caseExact: true` (such as `id` and
    `externalId`).
  - Dotted paths over multi-valued complex attributes (`emails.value eq "..."`,
    `members.value eq "..."`) now expand to quantified `ANY _v0 IN ... SATISFIES
    ... END` predicates in N1QL and traverse slices in `MemoryStore`.
  - `SplitURNPath` accepts both `:` (RFC 7644 §3.4.2.2 / §3.10) and `.` URN
    delimiters across `filter`, `sortBy`, `PATCH`, and projection, and strips
    core schema URN prefixes (`urn:ietf:params:scim:schemas:core:2.0:User:userName`)
    to top-level attributes.
- **Discovery completeness**: `/scim/v2/Schemas` now exposes `canonicalValues`
  and `referenceTypes` when declared, and `/scim/v2/Schemas/{id}` resolves
  schema URNs case-insensitively.

## 1.2.0

Bulk provisioning and referential group membership release.

### Added

- **`POST /scim/v2/Bulk`** (`RFC 7644 §3.7`): executes up to `1000`
  (`MaxBulkOperations`) `POST`, `PUT`, `PATCH` and `DELETE` operations within a
  single 1 MiB (`maxPayloadSize`) request:
  - Resolves `bulkId:<id>` transient cross-references in operation `path`
    (`/Users/bulkId:u1`), `$ref` URIs, complex/extension attribute values and
    `PATCH` value-filter literals (`members[value eq "bulkId:u1"]`), answering
    per-operation `400` (`scimType: invalidValue`) on forward, unknown or failed
    references.
  - Honours `version` (`If-Match` precondition, answering per-operation `412` on
    mismatch) and `failOnErrors` early termination.
  - Answers HTTP `413 Payload Too Large` when either `maxOperations` (`1000`) or
    `maxPayloadSize` (`1048576` bytes) is exceeded (`RFC 7644 §3.7.2`).
  - Dispatches every sub-operation through the full validation, `$writer` /
    `$reader` RBAC, password-hashing, atomic-uniqueness and audit pipeline under
    the outer caller's `Principal` and `X-Request-Id`.
- **Referential `Group.members` ↔ `User.groups` synchronisation**
  (`RFC 7643 §4.1.2` and `§4.2`):
  - Creating, replacing, patching or deleting a `Group` differentially updates
    the server-owned `readOnly` `User.groups` array (`value`, `$ref`, `display`,
    `type: "direct" | "indirect"`) in storage, supporting nested groups with
    cycle-safe traversal and N1QL filtering on `User.groups`.
  - Deleting a `User` or `Group` removes dangling entries from `Group.members`
    (unassigning `members` when empty) and refreshes `Group.meta`.
- **`Group.members.display` sub-attribute**: declared in the shipped
  `urn:ietf:params:scim:schemas:core:2.0:Group` schema (`RFC 7643 §8.3` / Errata
  4769) so group payloads from Okta and Microsoft Entra ID carrying member
  display names validate cleanly.

### Changed

- `/ServiceProviderConfig` now advertises
  `"bulk": { "supported": true, "maxOperations": 1000, "maxPayloadSize": 1048576 }`.

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
