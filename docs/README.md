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
| [Identity provider compatibility](idp-compatibility.md) | Connecting Microsoft Entra ID, Okta, and Keycloak |
| [UI, webhooks & multi-tenancy](ui-and-webhooks.md) | HTMX console, CloudEvents webhooks, multi-tenancy, conformance CLI |
| [Installation](installation.md) | Binary installer, Docker, Helm, SQLite, and Couchbase |
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

## Historical notes

The [1.0 audit](../RELEASE-1.0.md) records the original security and RFC
remediation work behind 1.0; every gap listed there (including `/Bulk`,
`Group.members <-> User.groups`, ETag preconditions, and Prometheus `/metrics`)
has since been implemented.
