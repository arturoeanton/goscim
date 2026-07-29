# GoSCIM

[![CI](https://github.com/arturoeanton/goscim/actions/workflows/ci.yml/badge.svg)](https://github.com/arturoeanton/goscim/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/arturoeanton/goscim)](https://goreportcard.com/report/github.com/arturoeanton/goscim)
[![Go Reference](https://pkg.go.dev/badge/github.com/arturoeanton/goscim.svg)](https://pkg.go.dev/github.com/arturoeanton/goscim)
[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/arturoeanton/goscim)](https://github.com/arturoeanton/goscim/releases)

A **SCIM 2.0** server in Go, backed by Couchbase.

Resource types, schemas and per-attribute permissions are JSON files, so adding
a resource type takes no code. Authentication is required, filter values are
bound rather than interpolated into queries, and permissions are enforced per
attribute on both read and write.

> Using PostgreSQL instead of Couchbase? See the sister project
> [go-vorpal-scim](https://github.com/arturoeanton/go-vorpal-scim).

## Quick start

```bash
docker run -d --name goscim-db -p 8091-8094:8091-8094 -p 11210:11210 \
  couchbase:community-7.1.1
# Set up the cluster at http://localhost:8091 with the credentials below.

git clone https://github.com/arturoeanton/goscim.git && cd goscim

export SCIM_AUTH=basic                       # how clients authenticate
export SCIM_BASIC_USERS='admin:secret:admin'
export SCIM_ADMIN_USER=Administrator         # how the server reaches Couchbase
export SCIM_ADMIN_PASSWORD=password
export SCIM_COUCHBASE_TLS=false              # local cluster, no certificate

go run .
```

```bash
curl -s -u admin:secret -X POST http://localhost:8080/scim/v2/Users \
  -H 'Content-Type: application/scim+json' \
  -d '{"schemas":["urn:ietf:params:scim:schemas:core:2.0:User"],
       "userName":"jane.doe@example.com",
       "emails":[{"value":"jane.doe@example.com","type":"work","primary":true}],
       "active":true}'
```

`201 Created`, with the resource in the body, its URI in `Location` and its
revision in `ETag`.

`SCIM_AUTH` has no default: the server refuses to start until you pick `jwt`,
`basic`, or `none`. Full walkthrough in
[getting started](docs/getting-started.md).

## What it does

- **CRUD, search, filtering, sorting and pagination** over RFC 7643/7644
- **Discovery**: `ServiceProviderConfig`, `ResourceTypes`, `Schemas`, derived
  from the running code rather than a file that can drift from it
- **Authentication**: OAuth 2.0 access tokens (RFC 9068) verified against a
  JWKS, HTTP Basic for development, or an explicit anonymous mode
- **Authorization per attribute**: `$reader` / `$writer` role lists alongside
  SCIM's own `mutability`, `returned` and `uniqueness`
- **Concurrency control**: `ETag` with `If-Match` and `If-None-Match`
- **Schema-driven everything**: add a resource type by dropping in two JSON
  files and restarting

## What it does not do

Worth knowing before you invest time:

- `POST /Bulk`
- Value paths — `emails[type eq "work"]` — in filters or PATCH paths
- `attributes` and `excludedAttributes`
- `POST /.search`
- Health, readiness or metrics endpoints
- Password hashing: `password` is stored exactly as sent

The [1.0 audit](RELEASE-1.0.md) records why each is open and roughly what it
would take.

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

Restart, and `/scim/v2/Devices` exists with all six verbs, its own bucket,
`serialNumber` enforced unique and refused on update.

## Filters

```
userName eq "bjensen"
name.familyName co "O"
title pr and userType eq "Employee"
userType eq "Employee" and (emails co "example.com" or emails co "example.org")
urn:ietf:params:scim:schemas:extension:enterprise:2.0:User.department eq "Sales"
```

All RFC 7644 operators — `eq ne co sw ew gt ge lt le pr`, `and or not`, `( )`.
Values are passed to Couchbase as bound parameters and never enter the query
text; a filter that fails to parse is refused rather than translated.

## Documentation

| | |
|---|---|
| [Getting started](docs/getting-started.md) | A running server and your first user |
| [Installation](docs/installation.md) | Couchbase, building, deploying |
| [Configuration](docs/configuration.md) | Every variable and config file |
| [API reference](docs/api.md) | Endpoints, payloads, status codes, filters |
| [Schemas and permissions](docs/schemas.md) | Attributes, and who may read or write them |
| [Architecture](docs/architecture.md) | How it fits together, and why |
| [Security](docs/security.md) | Threat model and hardening checklist |
| [Operations](docs/operations.md) | Signals, consistency, troubleshooting |
| [Development](docs/development.md) | Building, testing, code layout |
| [Changelog](CHANGELOG.md) | Releases and breaking changes |

## Requirements

Go 1.25+ to build. Couchbase 7.x, Community or Enterprise, with the data,
query and index services.

## Development

```bash
make check        # build, vet, unit suite with the race detector
make integration  # the suite against a real Couchbase, in a container
make cover        # coverage
```

The unit suite runs the real router over an in-memory store; the integration
suite starts Couchbase in a container and drives the same server through it.
Contributions are welcome — see [CONTRIBUTING](CONTRIBUTING.md), and
[RELEASE-1.0](RELEASE-1.0.md) for what is open.

## Status

1.0.0. The test suite covers the request path end to end, and the integration
suite exercises it against a real Couchbase. It has not yet been run against a
production identity provider such as Okta or Entra ID; if you do, reports are
welcome — that is where the remaining interoperability gaps will show up.

## License

[MIT](LICENSE).
