# Configuration

Environment variables and CLI flags control deployment concerns; JSON files
define the SCIM model.

## CLI subcommands and flags

```bash
goscim [-demo] [-ui] [-store memory|sqlite|couchbase] [-healthcheck]
goscim conformance -url <base-scim-url> [-token <bearer> | -basic <user:pass>] [-profile all|rfc|entra|okta|keycloak] [-json]
goscim version
```

| Flag / Subcommand | Meaning |
|---|---|
| `-demo` | Turnkey evaluation mode: sets defaults `SCIM_STORE=memory`, `SCIM_AUTH=bearer` (`demo-secret-token`), `SCIM_UI=true`, seeds sample Users and nested Groups, and prints ready-to-paste `curl` commands |
| `-ui` | Mount the embedded HTMX Web Console and Live SCIM Inspector at `/ui` (equivalent to `SCIM_UI=true`) |
| `-store <backend>` | Override `SCIM_STORE` (`couchbase`, `sqlite`, or `memory`) |
| `-healthcheck` | Probe `GET /readyz` on `SCIM_PORT` and exit `0` or `1` (used by distroless container `HEALTHCHECK`) |
| `conformance` | Run the live RFC 7643/7644 + Microsoft Entra ID + Okta + Keycloak conformance suite against any SCIM v2 endpoint |
| `version` | Print the binary version (`1.4.0`) |

## Environment variables

### Authentication and multi-tenancy

`SCIM_AUTH` has no default (unless `-demo` is passed). The server exits with a
message rather than starting unauthenticated by omission.

| Variable | Default | Meaning |
|---|---|---|
| `SCIM_AUTH` | *(required)* | `jwt`, `bearer`, `basic`, or `none` |
| `SCIM_JWT_JWKS_URL` | — | Required for `jwt`. Where the issuer publishes its signing keys (RSA `RS256`/`384`/`512`, EC `ES256`/`384`/`512`, OKP `EdDSA`) |
| `SCIM_JWT_ISSUER` | — | Required for `jwt`. The `iss` a token must carry |
| `SCIM_JWT_AUDIENCE` | — | Required for `jwt`. The `aud` a token must carry |
| `SCIM_JWT_ROLES_CLAIM` | — | Optional extra claim (or comma-separated list of claims, including dotted nested paths such as `resource_access.goscim-service.roles`) to read roles from alongside `roles`, `groups`, `entitlements`, `scope`, `scp`, and `realm_access.roles` |
| `SCIM_JWT_TENANT_CLAIM` | — | Optional JWT claim (supports dotted nested paths, e.g. `tid` or `org.id`) binding the caller to a specific tenant (`[A-Za-z0-9._-]{1,64}`) |
| `SCIM_BEARER_TOKEN` | — | Single opaque secret token for `bearer` (Microsoft Entra ID Secret Token / Okta API Token) |
| `SCIM_BEARER_SUBJECT` | `bearer-client` | Audit principal `subject` paired with `SCIM_BEARER_TOKEN` |
| `SCIM_BEARER_ROLES` | *(empty)* | Comma-separated roles paired with `SCIM_BEARER_TOKEN` |
| `SCIM_BEARER_TENANT` | *(empty)* | Optional tenant bound to `SCIM_BEARER_TOKEN` |
| `SCIM_BEARER_TOKENS` | — | Multi-token list for `bearer`: `subject:token:role1,role2[@tenant];other:token2:role3[@tenant2]` |
| `SCIM_BASIC_USERS` | — | Required for `basic`. `user:password:role1,role2;other:password:role3` |
| `SCIM_ANONYMOUS_ROLES` | *(empty)* | Roles given to every caller under `none` |

Issuer and audience are required rather than optional. A token is only an
authorization decision if the server checks it was minted by the issuer it
trusts and intended for this service; skipping either turns any token the
client can obtain anywhere into a valid credential here.

When a credential carries a tenant (via `@<tenant>` in `SCIM_BEARER_TOKENS`,
`SCIM_BEARER_TENANT`, or `SCIM_JWT_TENANT_CLAIM`), requests to `/scim/v2/...`
automatically resolve inside that tenant's partition, and attempts to call
`/tenants/<other>/scim/v2/...` are rejected with HTTP `403`. Global credentials
(no tenant pin) may address either `/scim/v2/...` (the default partition) or
any `/tenants/:tenant/scim/v2/...` namespace.

Under `SCIM_AUTH=none` with no `SCIM_ANONYMOUS_ROLES`, callers hold no roles,
so every attribute with a `$reader` list is filtered out of responses.

### Server, Web Console, Metrics, and Webhooks

| Variable | Default | Meaning |
|---|---|---|
| `SCIM_PORT` | `:8080` | Listen address, in `host:port` form |
| `SCIM_CONFIG_DIR` | `config` | Directory holding `schemas/`, `resourceType/`, `bucketSettings/`, `serviceProviderConfig/` (falls back to embedded defaults when absent) |
| `SCIM_TRUSTED_PROXIES` | `127.0.0.1` | Comma-separated addresses whose `X-Forwarded-For` is believed, or `none` to trust nobody |
| `SCIM_PUBLIC_URL` | *(empty)* | Externally reachable base URL (e.g. `https://scim.example.com`) used to build `Location` headers |
| `SCIM_PASSWORD_MODE` | `argon2id` | How `password` attributes are stored: `argon2id`, `bcrypt`, `reject` (`400 invalidValue`), or `plain` (logs a warning) |
| `SCIM_AUDIT_LOG` | `json` | Structured `log/slog` audit stream on SCIM endpoints: `json`, `text`, or `off` (`none`) |
| `SCIM_UI` | `false` | Mount the embedded HTMX Web Console & Live SCIM Inspector at `/ui` |
| `SCIM_UI_TOKEN` | *(empty)* | Shared secret protecting `/ui` (accepted via login form cookie or `Authorization: Bearer`) |
| `SCIM_METRICS` | `true` | Expose Prometheus metrics on `GET /metrics` (`false` / `off` to disable) |
| `SCIM_METRICS_TOKEN` | *(empty)* | Optional bearer token required to scrape `GET /metrics` |
| `SCIM_WEBHOOK_URLS` | *(empty)* | Comma-separated outbound HTTP/HTTPS destinations receiving CloudEvents 1.0 JSON lifecycle notifications |
| `SCIM_WEBHOOK_SECRET` | *(empty)* | Optional HMAC-SHA256 key used to sign outbound webhook bodies in `X-SCIM-Signature-256: sha256=<hex>` |
| `SCIM_WEBHOOK_EVENTS` | *(all)* | Optional comma-separated filter of event types (e.g. `io.goscim.user.created,io.goscim.user.deactivated`) |
| `SCIM_WEBHOOK_MAX_RETRIES` | `3` | Retry attempts with exponential backoff on network errors, `429`, or `5xx` |

Fixed in code rather than configurable, because no deployment has asked yet:
read header timeout 10s, read and write timeouts 30s, idle timeout 120s,
shutdown drain 20s, max header 1 MiB, max body 1 MiB, max search page 200, max
bulk operations 1000.

### Storage (`SCIM_STORE`), SQLite, and Couchbase

| Variable | Default | Meaning |
|---|---|---|
| `SCIM_STORE` | `couchbase` | Storage engine: `couchbase`, `sqlite`, or `memory` |
| `SCIM_SQLITE_PATH` | `goscim.db` | Database file path when `SCIM_STORE=sqlite` (runs in WAL mode with `busy_timeout=5000`) |
| `SCIM_COUCHBASE_URL` | `localhost` | Host, optionally `host:port` |
| `SCIM_ADMIN_USER` | `Administrator` | Cluster user. **Not** the SCIM API credential |
| `SCIM_ADMIN_PASSWORD` | — | Cluster password |
| `SCIM_COUCHBASE_TLS` | `true` | `false` connects over plain `couchbase://` |
| `SCIM_COUCHBASE_CA_CERT` | — | Path to the PEM that signed the cluster certificate |
| `SCIM_COUCHBASE_TLS_SKIP_VERIFY` | `false` | Accept any certificate. Logs a warning |
| `SCIM_QUERY_CONSISTENCY` | `request_plus` | `not_bounded` trades correctness for latency |
| `SCIM_COUCHBASE_AUTO_INDEX` | `true` | Derive and build secondary GSI N1QL indexes (`idx_scim_*`) automatically from configured JSON schemas (`false` to skip) |

The account needs enough privilege to create buckets and primary indexes on
first start. See [installation](installation.md#a-least-privilege-couchbase-user)
for narrowing it afterwards.

`SCIM_QUERY_CONSISTENCY` deserves a note. N1QL defaults to `not_bounded`: a
query runs against whatever the index currently holds, so a resource created a
moment ago is not found by a search. For a provisioning API that is the
ordinary flow rather than an edge case, so the default here is `request_plus`,
which waits for every mutation accepted before the query was issued. Set
`not_bounded` only if you have measured that the latency matters and can live
with the staleness.

## Configuration files

Under `SCIM_CONFIG_DIR`, default `config/`:

```
config/
  schemas/                 attribute definitions, one file per schema
  resourceType/            what is exposed at which endpoint
  bucketSettings/          optional per-bucket Couchbase settings
  serviceProviderConfig/   documentationUri only
```

Everything is read once at startup. Changing a file means restarting.

### Resource types

One file per resource type. The name of the file does not matter; the fields
do.

```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:ResourceType"],
  "id": "User",
  "name": "User",
  "endpoint": "/Users",
  "description": "User Account",
  "schema": "urn:ietf:params:scim:schemas:core:2.0:User",
  "schemaExtensions": [
    {
      "schema": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
      "required": false
    }
  ]
}
```

`name` and `endpoint` are not interchangeable, and confusing them is the
easiest mistake to make here:

- **`name`** names the Couchbase bucket and appears as `meta.resourceType`.
- **`endpoint`** is the URL, mounted under `/scim/v2`.

So this file yields `/scim/v2/Users` backed by a bucket called `User`.

`schemaExtensions` with `"required": true` means a request must carry that
extension's object; `false` means it may.

### Schemas

One file per schema, named after the URN with `:` replaced by `+`. See
[schemas and permissions](schemas.md) for the attribute vocabulary.

### Bucket settings

Optional. A file named after the resource type's `name` — `User.json` for the
`User` bucket — applied when the bucket is created. Missing files mean
defaults: 200 MB quota, one node, no replicas.

```json
{
  "flush_enabled": true,
  "replica_index_disabled": true,
  "ram_quota_mb": 200,
  "num_replicas": 0,
  "bucket_type": "couchbase",
  "MaxExpiry": "0s",
  "EvictionPolicy": "",
  "conflict_resolution_type": "seqno"
}
```

| Key | Values |
|---|---|
| `bucket_type` | `couchbase`, `ephemeral`, `memcached` |
| `compression_mode` | `off`, `passive`, `active` — **Enterprise Edition only** |
| `EvictionPolicy` | `fullEviction`, `valueOnly`, `nruEviction`, `noEviction` |
| `conflict_resolution_type` | `seqno`, `lww` |
| `MaxExpiry` | A Go duration, `"0s"` for none |

`compression_mode` is not in the shipped files on purpose. Community Edition
rejects the field outright, and sending it made the server fail to start
against the free edition — which is the one the quick start uses.

Settings apply at creation. Changing the file later does nothing to a bucket
that already exists; change it in Couchbase.

### Service provider config

`serviceProviderConfig/sp_config.json` contributes exactly one field:

```json
{ "documentationUri": "https://example.com/help/scim.html" }
```

Everything else in `/ServiceProviderConfig` — which operations are supported,
`bulk.maxOperations`, `bulk.maxPayloadSize`, `filter.maxResults`, the
authentication schemes — is derived from the running code so the capability
document cannot drift from the implementation.
