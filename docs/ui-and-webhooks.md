# HTMX Console, Webhooks, Multi-Tenancy, and Conformance CLI

This guide covers the interactive HTMX Web Console (`/ui`), signed outbound
CloudEvents 1.0 webhooks, B2B SaaS multi-tenancy (`/tenants/:tenant/scim/v2`),
and the built-in `goscim conformance` CLI.

## Embedded HTMX Web Console (`/ui`)

GoSCIM ships an optional, zero-dependency management and live traffic inspection
console embedded directly inside the Go binary (including a vendored copy of
HTMX 2.0.4, so no external CDN or internet access is required).

### Enabling the console

- **Quick demo** (`goscim -demo`): enables the console automatically alongside
  the in-memory store, sample `Users` and nested `Groups`, and a ready-to-use
  Bearer token (`demo-secret-token`).
- **Any store** (`goscim -ui` or `SCIM_UI=true`): mounts `/ui` over `couchbase`,
  `sqlite`, or `memory`.

To protect `/ui` with a dedicated token in shared staging or production
environments:

```bash
export SCIM_UI=true
export SCIM_UI_TOKEN='<high-entropy-console-secret>'
```

When `SCIM_UI_TOKEN` is set, unauthenticated browser visits redirect to
`/ui/login` (`HttpOnly`, `SameSite=Strict` session cookie) and programmatic
requests may pass `Authorization: Bearer <SCIM_UI_TOKEN>`.

### Console capabilities

1. **Operational Overview (`/ui`)**: real-time polling counters (active Users,
   Groups, HTTP requests, 409 uniqueness locks, webhook deliveries, store
   health), ready-to-copy `curl` commands, and pre-filled Microsoft Entra ID,
   Okta, and Keycloak connector settings.
2. **Directory Resources (`/ui/resources`)**: browse and filter `Users`,
   `Groups`, and `Elements` using live RFC 7644 filter expressions, create Users
   and Groups, toggle `active`, delete records, and inspect formatted SCIM JSON
   alongside computed `direct` and `indirect` `User.groups` badges.
3. **SCIM Filter & N1QL Lab (`/ui/playground`)**: type any SCIM 2.0 filter
   expression (or pick an Entra, Okta, or Keycloak preset) to inspect both the
   parameterized Couchbase N1QL translation (`$1..$n` positional arguments) and
   the live matched resources from the active store.
4. **Live Traffic Inspector (`/ui/inspector`)**: `ngrok`-style stream of recent
   SCIM API requests and responses from an in-memory ring buffer (200 entries),
   automatically scrubbing `password`, `secret`, and `token` fields to
   `[REDACTED]`.
5. **Webhooks & Conformance (`/ui/conformance`)**: run the full live IdP
   conformance suite (`all`, `rfc`, `entra`, `okta`, `keycloak`) against any
   tenant with one click and inspect recent outbound webhook deliveries.

## Signed Outbound Webhooks (CloudEvents 1.0)

When `SCIM_WEBHOOK_URLS` is configured, GoSCIM asynchronously delivers
CloudEvents 1.0 (`application/cloudevents+json`) notifications whenever SCIM
resources are created, updated, deactivated, or deleted.

```bash
export SCIM_WEBHOOK_URLS='https://app.example.com/webhooks/scim'
export SCIM_WEBHOOK_SECRET='whsec_replace_with_32_byte_random_secret'
export SCIM_WEBHOOK_EVENTS='user.*,group.*'
export SCIM_WEBHOOK_MAX_RETRIES=3
```

### Event types

| Mutation | Emitted `type` |
|---|---|
| `POST /Users` | `io.goscim.user.created` |
| `PUT` / `PATCH /Users/:id` (`active` transitions `true -> false`) | `io.goscim.user.deactivated` |
| `PUT /Users/:id` (other profile updates) | `io.goscim.user.replaced` |
| `PATCH /Users/:id` (other profile updates) | `io.goscim.user.patched` |
| `DELETE /Users/:id` | `io.goscim.user.deleted` |
| `POST /Groups` | `io.goscim.group.created` |
| `PUT` / `PATCH /Groups/:id` | `io.goscim.group.updated` |
| `DELETE /Groups/:id` | `io.goscim.group.deleted` |

Write-only and `returned: "never"` attributes (such as `password`) are stripped
before serialization. Events targeting the same `(tenant, resourceID)` pair are
sharded onto the same worker queue so causal lifecycle ordering (`created ->
deactivated -> deleted`) holds even across exponential-backoff retries on HTTP
`429` and `5xx`.

### Verifying `X-SCIM-Signature-256`

When `SCIM_WEBHOOK_SECRET` is set, every request carries `X-SCIM-Timestamp`
(Unix epoch seconds) and `X-SCIM-Signature-256`:

```
X-SCIM-Signature-256: sha256=<hex(HMAC-SHA256(secret, timestamp + "." + rawBody))>
```

Receivers should reject timestamps outside a narrow replay window (for example
300 seconds) and compare the computed digest in constant time.

## Multi-Tenancy (`/tenants/:tenant/scim/v2`)

B2B SaaS deployments provisioning multiple customer directories through one
GoSCIM cluster can isolate tenants without running separate deployments:

- **URL routing**: every discovery and resource endpoint is mounted at both
  `/scim/v2/...` (default namespace) and `/tenants/:tenant/scim/v2/...`
  (`[A-Za-z0-9._-]{1,64}`).
- **Credential-bound tenant scoping**:
  - **Opaque Bearer tokens (`SCIM_AUTH=bearer`)**: append `@<tenant>` to the
    roles field in `SCIM_BEARER_TOKENS`
    (`entra-acme:secret-tok-1:admin@acme;okta-globex:secret-tok-2:admin@globex`),
    or set `SCIM_BEARER_TENANT=acme` in single-token mode.
  - **OAuth 2.0 JWTs (`SCIM_AUTH=jwt`)**: set `SCIM_JWT_TENANT_CLAIM` (for
    example `tid` or `org.id`) to bind each token to the tenant named in that
    claim.
  - A tenant-bound credential calling `/tenants/<other>/scim/v2/...` is rejected
    with `403 Forbidden`; calling `/scim/v2/...` automatically scopes the
    operation to the credential's bound tenant.
- **Isolation semantics**: CRUD, `GET` and `POST /.search`, `POST /Bulk`,
  `Group.members <-> User.groups` synchronisation, and `uniqueness: "server"`
  (such as `userName`) are isolated per tenant, while attributes declared
  `uniqueness: "global"` are enforced across all tenants.

## Live Conformance CLI (`goscim conformance`)

Verify any running SCIM 2.0 endpoint (GoSCIM or third-party) from the command
line or CI pipeline:

```bash
goscim conformance \
  -url http://127.0.0.1:8080/scim/v2 \
  -token demo-secret-token \
  -profile all
```

Available flags:

- `-url`: SCIM 2.0 base URL (supports `/scim/v2` and `/tenants/<id>/scim/v2`).
- `-token`: OAuth 2.0 / opaque Bearer token.
- `-basic`: HTTP Basic `username:password`.
- `-profile`: `all` (default), `rfc`, `entra`, `okta`, or `keycloak`.
- `-json`: emit a structured JSON report for CI assertions (exits `0` when all
  checks pass, `1` on any failure).
