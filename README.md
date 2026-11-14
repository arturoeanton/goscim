# GoSCIM

[![CI](https://github.com/arturoeanton/goscim/actions/workflows/ci.yml/badge.svg)](https://github.com/arturoeanton/goscim/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/arturoeanton/goscim)](https://goreportcard.com/report/github.com/arturoeanton/goscim)
[![Go Reference](https://pkg.go.dev/badge/github.com/arturoeanton/goscim.svg)](https://pkg.go.dev/github.com/arturoeanton/goscim)
[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/arturoeanton/goscim)](https://github.com/arturoeanton/goscim/releases)

A schema-driven **SCIM 2.0** (RFC 7643 / RFC 7644) server in Go with pluggable
persistence (**Couchbase**, embedded **SQLite**, or **in-memory**), an embedded
**HTMX Web Console & Live Traffic Inspector**, **Prometheus metrics**, **signed
CloudEvents webhooks**, and **B2B multi-tenancy**.

Resource types, schemas and per-attribute permissions are JSON files, so adding
a resource type takes no code. Authentication is required, filter values are
bound rather than interpolated into queries, and permissions are enforced per
attribute on both read and write.

> Using PostgreSQL instead of Couchbase or SQLite? See the sister project
> [go-vorpal-scim](https://github.com/arturoeanton/go-vorpal-scim).

## Try it in 10 seconds (zero dependencies)

Install the single static binary (Linux or macOS) or run via Go / Docker, and
launch `-demo` mode:

```bash
# Option A: one-line binary install (macOS / Linux, amd64 & arm64)
curl -fsSL https://raw.githubusercontent.com/arturoeanton/goscim/main/scripts/install.sh | sh
goscim -demo

# Option B: with Go installed
go install github.com/arturoeanton/goscim@latest
goscim -demo

# Option C: with Docker (no local toolchain required)
docker run --rm -p 8080:8080 ghcr.io/arturoeanton/goscim:latest -demo
```

Open **<http://127.0.0.1:8080/ui>** to explore the HTMX directory console,
compile SCIM filters to N1QL live, inspect traffic, or run the built-in IdP
conformance suite:

```bash
# Query seeded Users
curl -s -H "Authorization: Bearer demo-secret-token" \
  http://127.0.0.1:8080/scim/v2/Users | jq .

# Run live RFC + Entra ID + Okta + Keycloak conformance checks
goscim conformance -url http://127.0.0.1:8080/scim/v2 -token demo-secret-token
```

## Production & Persistence Modes

- **Single-node persistent (`SCIM_STORE=sqlite`)**: pure-Go embedded SQLite in
  WAL mode (`CGO_ENABLED=0`) — durable single-file storage with zero external
  infrastructure:
  ```bash
  SCIM_STORE=sqlite SCIM_SQLITE_PATH=./goscim.db \
  SCIM_AUTH=bearer SCIM_BEARER_TOKEN='<secret>' \
  goscim -ui
  ```
- **Clustered Couchbase (`SCIM_STORE=couchbase`, default)**: starts Couchbase,
  initialises buckets, `_default.scim_unique` KV sentinels, and secondary GSI
  indexes automatically:
  ```bash
  git clone https://github.com/arturoeanton/goscim.git && cd goscim
  docker compose up -d --build
  ```
- **Keycloak + GoSCIM turnkey lab**:
  ```bash
  docker compose -f docker-compose.keycloak.yml up -d --build
  ```
- **Kubernetes Helm Chart**: `helm install goscim ./charts/goscim` (see
  [installation](docs/installation.md)).

Full walkthrough in [getting started](docs/getting-started.md), and IdP setup in
[identity provider compatibility](docs/idp-compatibility.md).

## What it does

- **CRUD, search (`GET` and `POST /.search`), bulk operations (`POST /Bulk`),
  filtering, sorting, pagination and attribute projection
  (`attributes` / `excludedAttributes`)** over RFC 7643/7644, including
  `externalId` (`caseExact: true`) and explicit JSON `null` clearing
- **Identity provider compatibility**: verified against **Microsoft Entra ID**,
  **Okta**, and **Keycloak** provisioning flows, plus a built-in CLI & UI
  verifier (`goscim conformance -profile all|rfc|entra|okta|keycloak`)
- **Three storage backends behind one `Store` contract**:
  - `couchbase` — clustered KV + N1QL with automatic schema-derived secondary
    GSI indexes and linearizable `_default.scim_unique` sentinel locks
  - `sqlite` — pure-Go single-file ACID persistence (`WAL` mode, static binary)
  - `memory` — zero-dependency ephemeral store for instant demos and CI
- **Embedded HTMX Web Console (`/ui`, `-ui` / `SCIM_UI=true`)**: offline-ready
  directory manager, interactive SCIM-to-N1QL playground, `ngrok`-style live
  request/response inspector (with automatic secret scrubbing), and one-click
  conformance runner
- **Multi-tenancy for B2B SaaS**: `/tenants/:tenant/scim/v2/...` URL namespaces
  and credential-bound tenant isolation (`SCIM_BEARER_TOKENS` `@tenant`,
  `SCIM_BEARER_TENANT`, `SCIM_JWT_TENANT_CLAIM`)
- **Signed outbound webhooks**: asynchronous CloudEvents 1.0 dispatcher
  (`SCIM_WEBHOOK_URLS`) with `X-SCIM-Signature-256` HMAC-SHA256 signing,
  per-resource causal ordering, and exponential-backoff retries
- **Prometheus metrics (`GET /metrics`)**: request counters, latency histograms,
  auth failure and 409 uniqueness conflict counters, webhook outcomes, and store
  health gauge (`goscim_store_up`), optionally guarded by `SCIM_METRICS_TOKEN`
- **Group membership synchronisation**: `Group.members` changes automatically
  materialise `User.groups` (`direct` and `indirect` nested memberships,
  cycle-safe) and clean up dangling references on delete
- **Authentication & per-attribute RBAC**: OAuth 2.0 JWTs against JWKS
  (`RS256`/`384`/`512`, `ES256`/`384`/`512`, `EdDSA`), constant-time opaque
  Bearer tokens, HTTP Basic, `$reader` / `$writer` attribute role lists,
  Argon2id password hashing, and structured `log/slog` audit events

## Adding a resource type

Two files, no code:

```json
// config/schemas/urn+ietf+params+scim+schemas+custom+2.0+Device.json
{
  "id": "urn:ietf:params:scim:schemas:custom:2.0:Device",
  "name": "Device",
  "attributes": [
    { "name": "serialNumber", "type": "string",
      "required": true, "uniqueness": "server", "mutability": "immutable" }
  ]
}
```

```json
// config/resourceType/Device.json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:ResourceType"],
  "id": "Device", "name": "Device", "endpoint": "/Devices",
  "schema": "urn:ietf:params:scim:schemas:custom:2.0:Device",
  "schemaExtensions": []
}
```

Restart, and `/scim/v2/Devices` (plus `/tenants/:tenant/scim/v2/Devices`) exists
with CRUD, `GET` and `POST /.search`, automatic inclusion in root
`POST /.search` and `POST /Bulk`, its own bucket and secondary indexes, and
`serialNumber` enforced unique and refused on update.

## Filters

```
userName eq "bjensen"
name.familyName co "O"
title pr and userType eq "Employee"
emails[type eq "work" and primary eq true]
emails[type eq "work"].value ew "@example.com"
emails.value eq "bjensen@example.com"
userType eq "Employee" and (emails co "example.com" or emails co "example.org")
urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department eq "Sales"
```

All RFC 7644 operators — `eq ne co sw ew gt ge lt le pr`, `and or not`, `( )`,
and `[ ]` complex array value filters. String comparisons honour `caseExact`
(`userName` folds case; `externalId` and `id` match exact bytes). Values are
passed to Couchbase as bound parameters and never enter the query text; a filter
that fails to parse is refused rather than translated.

## Documentation

| | |
|---|---|
| [Getting started](docs/getting-started.md) | A running server and your first user |
| [IdP compatibility](docs/idp-compatibility.md) | Connecting Microsoft Entra ID, Okta, and Keycloak |
| [UI, webhooks & multi-tenancy](docs/ui-and-webhooks.md) | HTMX console, CloudEvents, tenants, conformance CLI |
| [Installation](docs/installation.md) | Binary installer, Docker, Helm, SQLite, Couchbase |
| [Configuration](docs/configuration.md) | Every variable and config file |
| [API reference](docs/api.md) | Endpoints, payloads, status codes, filters |
| [Schemas and permissions](docs/schemas.md) | Attributes, and who may read or write them |
| [Architecture](docs/architecture.md) | How it fits together, and why |
| [Security](docs/security.md) | Threat model and hardening checklist |
| [Operations](docs/operations.md) | Metrics, signals, consistency, troubleshooting |
| [Development](docs/development.md) | Building, testing, code layout |
| [Changelog](CHANGELOG.md) | Releases and breaking changes |

## Requirements

Precompiled static binaries (`linux`/`darwin`/`windows`, `amd64`/`arm64`) need
no external runtime when running with `SCIM_STORE=memory` or `SCIM_STORE=sqlite`.
Building from source requires Go 1.26+ (toolchain 1.27.4+). Clustered mode uses
Couchbase 7.x, Community or Enterprise, with the data, query, and index
services.

## Development

```bash
make demo         # launch instant demo mode with HTMX console on :8080/ui
make check        # build, vet, unit suite with the race detector
make integration  # the suite against a real Couchbase, in a container
make cover        # coverage
```

Contributions are welcome — see [CONTRIBUTING](CONTRIBUTING.md).

## Status

1.4.0. Both the unit conformance suites (over `MemoryStore` and `SQLiteStore`)
and the Couchbase integration suite verify the full request path across RFC
7643/7644, multi-tenant isolation, and the provisioning dialects of Microsoft
Entra ID, Okta, and Keycloak.

## License

[MIT](LICENSE).
