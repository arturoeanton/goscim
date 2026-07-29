# Changelog

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
