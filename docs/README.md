# GoSCIM documentation

GoSCIM is a SCIM 2.0 server backed by Couchbase. Resource types, schemas and
per-attribute permissions are JSON files, so adding a resource type takes no
code.

Documentation is maintained in English only. Keeping several translations in
step with the code proved impossible, and a translation that has fallen behind
is worse than none: it tells a reader something the server no longer does.

## Guides

| Document | What it covers |
|---|---|
| [Getting started](getting-started.md) | A running server and your first user |
| [Installation](installation.md) | Couchbase, building, and deploying |
| [Configuration](configuration.md) | Every environment variable and JSON config file |
| [API reference](api.md) | Endpoints, payloads, status codes, filters |
| [Schemas and permissions](schemas.md) | Defining resource types and who may read or write each attribute |
| [Architecture](architecture.md) | How the pieces fit, and why |
| [Security](security.md) | Authentication, authorization, transport, threat model |
| [Operations](operations.md) | Signals, timeouts, consistency, troubleshooting |
| [Development](development.md) | Building, testing, code layout |

## Project documents

| Document | What it covers |
|---|---|
| [CHANGELOG](../CHANGELOG.md) | Releases, and what breaks between them |
| [RELEASE-1.0](../RELEASE-1.0.md) | The audit behind 1.0, and what is still open |
| [CONTRIBUTING](../CONTRIBUTING.md) | How to propose a change |
| [SECURITY](../SECURITY.md) | How to report a vulnerability |

## What GoSCIM does not do

Worth knowing before you invest time. None of these is implemented:

- `POST /Bulk` (RFC 7644 §3.7)
- Value paths — `emails[type eq "work"]` — in filters or PATCH paths
- The `attributes` and `excludedAttributes` query parameters
- `POST /.search`
- Health, readiness or metrics endpoints
- Password hashing: `password` is stored exactly as sent

The [1.0 audit](../RELEASE-1.0.md) records why, and roughly what each would take.
