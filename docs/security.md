# Security

A SCIM server holds identity data and is, by design, a write path into every
system it provisions. This page describes what is enforced, what is not, and
what a deployment has to do itself.

To report a vulnerability, see [SECURITY.md](../SECURITY.md).

## Threat model

What this design assumes:

- **The network between clients and the server is hostile.** Terminate TLS in
  front of GoSCIM; it does not serve HTTPS itself.
- **The network between the server and Couchbase is hostile.** TLS with
  certificate verification is on by default.
- **Clients are not trusted.** Every payload is validated against the schema,
  filter values are bound rather than interpolated, and attribute-level
  permissions are enforced on both read and write.
- **Anyone with cluster credentials has everything.** GoSCIM stores documents
  unencrypted (`password` is hashed with Argon2id by default unless
  `SCIM_PASSWORD_MODE=plain` is set).

What it does not defend against: a compromised token within its lifetime,
denial of service beyond basic limits, or anyone with access to Couchbase.

## Authentication

`SCIM_AUTH` is required. The server exits rather than starting unauthenticated
because someone forgot to configure it.

### JWT — production

OAuth 2.0 access tokens per RFC 9068, verified against the issuer's JWKS.

```bash
export SCIM_AUTH=jwt
export SCIM_JWT_JWKS_URL=https://issuer.example/.well-known/jwks.json
export SCIM_JWT_ISSUER=https://issuer.example
export SCIM_JWT_AUDIENCE=scim-api
```

What is checked, and why each matters:

- **Signature**, against a key fetched from the JWKS and cached. An unknown key
  id triggers one refetch, serialized and rate limited so a stream of tokens
  naming keys that do not exist cannot be turned into load on the issuer.
- **Key parameters**, across three asymmetric key families whose `use` (if set)
  is `"sig"` and `key_ops` (if set) includes `"verify"`:
  - **RSA** (`"kty": "RSA"`): modulus of at least 2048 bits (`RFC 7518 §3.3`)
    and valid odd exponent `>= 3`.
  - **EC** (`"kty": "EC"`): NIST `P-256`, `P-384`, or `P-521` with exact
    coordinate byte lengths (`32`, `48`, `66` per `RFC 7518 §6.2.1.2`) and
    on-curve / non-infinity verification via `crypto/ecdh`.
  - **OKP** (`"kty": "OKP"`): `Ed25519` 32-byte public keys (`RFC 8037`);
    encryption curves such as `X25519` are rejected.
- **Algorithm binding**, restricted to `RS256`, `RS384`, `RS512`, `ES256`,
  `ES384`, `ES512`, and `EdDSA`, with strict key-type and curve matching (`P-256`
  only verifies `ES256`, `P-384` only `ES384`, `P-521` only `ES512`, `Ed25519`
  only `EdDSA`, RSA only `RS*`) and enforcement of any JWK `"alg"` pin. Symmetric
  algorithms (`HS*`) and `"none"` are never accepted.
- **Issuer and audience**, both required. Skipping either turns any token the
  client can obtain anywhere into a valid credential here.
- **Expiry**, required to be present. A token with no `exp` is refused.

### Static Bearer token — IdP secret-token connectors

For identity providers whose SCIM 2.0 provisioning client authenticates with a
pre-shared opaque bearer secret rather than OAuth 2.0 JWTs (such as Microsoft
Entra ID's **Secret Token** mode and Okta's **API Token** mode):

```bash
export SCIM_AUTH=bearer
export SCIM_BEARER_TOKEN='<high-entropy-secret>'
export SCIM_BEARER_SUBJECT='entra-provisioning'
export SCIM_BEARER_ROLES='admin'
```

Or with multiple active tokens during secret rotation:

```bash
export SCIM_AUTH=bearer
export SCIM_BEARER_TOKENS='entra-2026q4:<token1>:admin;entra-2026q3:<token0>:admin'
```

Every configured token is hashed with SHA-256 and compared in constant time
without early exit, so response latency discloses neither prefix matches nor
how many tokens are configured. Because a static token has no `exp` of its own,
use high-entropy secrets (at least 32 random bytes), restrict source egress IP
ranges at your proxy where possible, and rotate periodically via
`SCIM_BEARER_TOKENS`.

### Basic — development

```bash
export SCIM_AUTH=basic
export SCIM_BASIC_USERS='alice:secret:admin,role1;bob:hunter2:user'
```

Credentials sit in an environment variable in the clear, and there is no
lockout or throttling. Passwords are compared in constant time, and an unknown
user takes the same path as a wrong password so timing does not disclose which
usernames exist — but this is for development and for smoke tests, not for
production.

### None

```bash
export SCIM_AUTH=none
export SCIM_ANONYMOUS_ROLES=reader
```

Every request is served. The server logs a warning at startup. With no
`SCIM_ANONYMOUS_ROLES`, callers hold no roles, so every attribute carrying a
`$reader` list is filtered out of responses.

### Roles

Roles drive the `$reader` and `$writer` checks and come from the token (or the
configured `SCIM_BEARER_ROLES` / `SCIM_BEARER_TOKENS` entry under
`SCIM_AUTH=bearer`). RFC 9068 §2.2.3.1 reuses SCIM's own `roles`, `groups` and
`entitlements` claims from RFC 7643 §4.1.2, so those are read alongside `scope`,
`scp` (Microsoft Entra ID v2.0 and Okta), and `realm_access.roles` (Keycloak).
`SCIM_JWT_ROLES_CLAIM` names an additional claim — or comma-separated list of
claims, supporting dotted nested paths such as
`resource_access.goscim-service.roles` — for issuers that use their own. The
audit principal `subject` is read from `sub`, falling back through `client_id`,
`cid`, `azp`, `oid`, and `appid` for client-credentials tokens that omit `sub`.

Each is accepted as a list of strings, a list of `{"value": "..."}` objects, or
a space-delimited string, because all three occur in practice.

A request that never went through the middleware holds no roles, so the
failure mode of a routing mistake is that restricted attributes stay hidden.

## Authorization

Per attribute, not per endpoint. See
[schemas and permissions](schemas.md#role-based-permissions).

Enforced on the way in by `$writer` and `mutability` (including `DELETE`,
case-insensitive attribute keys, bracketed `PATCH` value-selection paths such
as `emails[type eq "work"].value`, and removal attempts on `PUT` or `PATCH`),
and on the way out by `$reader`, `returned: never` and `writeOnly`. Search
`filter`, `PATCH` value-selection filters, and `sortBy` also refuse `writeOnly` /
`returned: never` attributes (and filters refuse attributes outside the
caller's `$reader` roles, including sub-attributes inside `[...]`) so secrets
cannot be probed via blind queries or selective patches. Both walk the schema,
so complex values, the elements of multi-valued attributes, and extension
objects are all covered.

Denied reads **omit** the attribute rather than blanking it: an empty string
cannot be told apart from a legitimately empty value.

## Transport

### To clients

GoSCIM serves plain HTTP. Put a TLS-terminating proxy in front, and set
`SCIM_TRUSTED_PROXIES` to its address so `X-Forwarded-For` is believed from it
and nowhere else. `SCIM_TRUSTED_PROXIES=none` trusts no proxy at all.

### To Couchbase

TLS with verification, by default.

```bash
export SCIM_COUCHBASE_CA_CERT=/etc/ssl/couchbase-ca.pem
```

`SCIM_COUCHBASE_TLS_SKIP_VERIFY=true` accepts any certificate and logs a
warning. It defeats the point of TLS — an interceptor can present any
certificate — so it is for a local cluster and nothing else. Verification used
to be disabled unconditionally, which is what made this an option rather than
a default.

`SCIM_COUCHBASE_TLS=false` drops to plain `couchbase://`, for a development
cluster with no certificate.

## Injection

The filter parser translates SCIM filters into N1QL, so it is the obvious
place to attack.

- **Filter values are bound parameters.** They never enter the query text. A
  property test asserts that for every operator, including values carrying
  backslashes, backticks, semicolons, comment markers and LIKE
  metacharacters.
- **Filters that do not parse are refused.** ANTLR recovers from syntax errors
  and keeps walking, so an unchecked parse tree emits whatever the client sent.
- **`sortBy` is restricted to declared attributes.** It reaches the query as an
  identifier, which cannot be bound, so it is a whitelist rather than an
  escaping exercise.
- **Identifiers double their backticks**, as a second line of defence behind
  that whitelist.

The 1.0 audit found three separate injection paths here, each caused by
concatenating a client value into SQL. Binding the values removes the class
rather than the instances.

## Limits

| Limit | Value | Purpose |
|---|---|---|
| Request body | 1 MiB | Payloads are read into memory before parsing (also advertised as `bulk.maxPayloadSize`) |
| Bulk operations | 1000 | Max operations per `POST /scim/v2/Bulk` request (`bulk.maxOperations`) |
| Request headers | 1 MiB | |
| Read header timeout | 10 s | A connection dribbling a request holds a goroutine |
| Read / write timeout | 30 s | |
| Idle timeout | 120 s | |
| Search page | 200 | Also advertised as `filter.maxResults` |

There is **no rate limiting**. Put it in the proxy.

## Audit logging

Every SCIM resource request (including `401` authentication failures, `403`
role denials, `409` uniqueness conflicts, `413` body-limit rejections, and
successful reads/writes) emits a structured `log/slog` event (`msg:
"scim.audit"`), controlled by `SCIM_AUDIT_LOG` (`json` default, `text`, or
`off`):

- Records `request_id`, `action`, `subject`, `roles`, `resource_type`,
  `resource_id`, `method`, `path`, `status`, `outcome`, optional `scim_type`,
  `client_ip`, and `duration_ms`.
- Never logs `Authorization` headers, `password` values, request/response
  bodies, or raw query strings (which can carry PII inside `?filter=`).
- Validates incoming `X-Request-Id` headers against `^[A-Za-z0-9._:-]{1,128}$`
  before echoing or logging them, replacing malformed or oversized values with a
  fresh UUID to prevent header splitting and log injection.
- On `POST /scim/v2/Bulk`, emits one `scim.audit` event per executed
  sub-operation (`create`, `replace`, `patch`, `delete`) plus one outer
  `"action": "bulk"` record, all correlated under the same `request_id` and
  authenticated `Principal`.

## Multi-tenant isolation

When multi-tenant endpoints (`/tenants/:tenant/scim/v2/...`) are used:

- Tenant identifiers are validated against `^[A-Za-z0-9._-]{1,64}$` before
  reaching storage keys or N1QL parameters.
- Credentials bound to a tenant (`subject:token:roles@tenant` in
  `SCIM_BEARER_TOKENS`, `SCIM_BEARER_TENANT`, or `SCIM_JWT_TENANT_CLAIM`) are
  locked to that tenant: calling `/tenants/<other>/scim/v2/...` returns HTTP
  `403`, and calling `/scim/v2/...` resolves inside the credential's tenant
  partition rather than the root namespace.
- Resource reads, searches, `/Bulk` sub-operations, `Group.members` ↔
  `User.groups` graph synchronisation, and `uniqueness: server` locks are
  isolated per tenant.

## Web Console, Metrics, Inspector, and Webhooks

- **`/ui` (`SCIM_UI=true`)**: disabled by default (`SCIM_UI=false`, except under
  `-demo`). When enabled outside localhost, set `SCIM_UI_TOKEN` to a
  high-entropy secret. Token verification uses constant-time `subtle.ConstantTimeCompare`
  and stores an `HttpOnly`, `SameSite=Strict` session cookie (`Secure` when
  served behind HTTPS).
- **Live Inspector redaction**: the 200-entry in-memory traffic inspector
  automatically redacts `Authorization` and `Cookie` headers plus `"password"`
  fields in captured JSON payloads before storing or rendering them.
- **`GET /metrics` (`SCIM_METRICS=true`)**: set `SCIM_METRICS_TOKEN` (or block
  `/metrics` at your reverse proxy) if Prometheus scrape endpoints should not be
  world-readable on your network.
- **Outbound Webhooks (`SCIM_WEBHOOK_URLS`)**: set `SCIM_WEBHOOK_SECRET` and
  verify `X-SCIM-Signature-256` (`sha256=<hex>`) with `hmac.Equal` in every
  downstream receiver. Attributes declared `returned: never` or
  `mutability: writeOnly` (such as `password`) are stripped before webhook
  serialization.

## Atomic uniqueness

Attributes with `uniqueness: server` or `uniqueness: global` are claimed
atomically — in Couchbase's `_default.scim_unique` collection via KV `Insert`
and CAS, or inside serialised SQLite / in-memory transactions — closing the
read-then-write TOCTOU race across concurrent requests.

## Not handled

Be explicit about these before deploying:

- **`SCIM_PASSWORD_MODE=plain` disables password hashing.** The default is
  `argon2id`; `bcrypt` and `reject` are also supported.
- **`SCIM_AUDIT_LOG=off` disables structured audit events.** Leave it at `json`
  (or `text`) in production.
- **No rate limiting or lockout.**
- **No encryption at rest.** Use Couchbase's own, if you need it.
- **`meta.location` is a server-relative path** unless mapped upstream; set
  `SCIM_PUBLIC_URL` so HTTP `Location` headers use your public origin rather
  than client-supplied `Host` headers.

## Checklist

Before exposing this to anything real:

- [ ] `SCIM_AUTH=jwt` (with issuer and audience set to your own values) or `SCIM_AUTH=bearer` (with high-entropy secrets)
- [ ] TLS terminated in front, HTTP not reachable directly
- [ ] `SCIM_TRUSTED_PROXIES` naming your proxy, not the default
- [ ] `SCIM_COUCHBASE_TLS_SKIP_VERIFY` unset, `SCIM_COUCHBASE_CA_CERT` set
- [ ] A Couchbase account scoped to the buckets GoSCIM uses
- [ ] Rate limiting in the proxy
- [ ] The `Element` resource type removed, unless you want the example endpoint
- [ ] `SCIM_PASSWORD_MODE` left at `argon2id` (or set to `bcrypt` / `reject`, never `plain`)
- [ ] `SCIM_AUDIT_LOG` left at `json` and shipped to central log storage
- [ ] `SCIM_UI=false` in production (or protected behind VPN + `SCIM_UI_TOKEN`)
- [ ] `SCIM_METRICS_TOKEN` set (or `/metrics` restricted to internal Prometheus scrapers at the proxy)
- [ ] `SCIM_WEBHOOK_SECRET` configured whenever `SCIM_WEBHOOK_URLS` is enabled
