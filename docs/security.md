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
  unencrypted, `password` included.

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
  id triggers one refetch, rate limited so a stream of tokens naming keys that
  do not exist cannot be turned into load on the issuer.
- **Algorithm**, restricted to `RS256`, `RS384`, `RS512`. Without an
  allow-list, a token signed with an algorithm that happens to verify against
  the same RSA key is accepted; `HS256` with the public key as the secret is
  the classic version of that attack.
- **Issuer and audience**, both required. Skipping either turns any token the
  client can obtain anywhere into a valid credential here.
- **Expiry**, required to be present. A token with no `exp` is refused.

Only RSA keys are understood. An EC or OKP key in the JWKS is skipped.

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

Roles drive the `$reader` and `$writer` checks and come from the token.
RFC 9068 §2.2.3.1 reuses SCIM's own `roles`, `groups` and `entitlements`
claims from RFC 7643 §4.1.2, so those are read, along with `scope`.
`SCIM_JWT_ROLES_CLAIM` names an additional claim for issuers that use their
own.

Each is accepted as a list of strings, a list of `{"value": "..."}` objects, or
a space-delimited string, because all three occur in practice.

A request that never went through the middleware holds no roles, so the
failure mode of a routing mistake is that restricted attributes stay hidden.

## Authorization

Per attribute, not per endpoint. See
[schemas and permissions](schemas.md#role-based-permissions).

Enforced on the way in by `$writer` and `mutability`, and on the way out by
`$reader`, `returned: never` and `writeOnly`. Both walk the schema, so complex
values, the elements of multi-valued attributes, and extension objects are all
covered.

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
| Request body | 1 MiB | Payloads are read into memory before parsing |
| Request headers | 1 MiB | |
| Read header timeout | 10 s | A connection dribbling a request holds a goroutine |
| Read / write timeout | 30 s | |
| Idle timeout | 120 s | |
| Search page | 200 | Also advertised as `filter.maxResults` |

There is **no rate limiting**. Put it in the proxy.

## Not handled

Be explicit about these before deploying:

- **Passwords are stored as sent.** No hashing. Treat the `User` bucket as
  holding secrets, or keep passwords out of SCIM.
- **No audit log.** Requests are logged by gin's default logger; who changed
  what is not recorded.
- **No rate limiting or lockout.**
- **No encryption at rest.** Use Couchbase's own, if you need it.
- **The uniqueness check is read-then-write.** Two concurrent creates of the
  same `userName` can both pass. Closing that needs a unique index or a
  sentinel document.
- **`meta.location` is a server-relative path**, not an absolute URI. Deriving
  the host from the connection is wrong behind a TLS-terminating proxy, so a
  configurable public base URL is still needed.

## Checklist

Before exposing this to anything real:

- [ ] `SCIM_AUTH=jwt`, with issuer and audience set to your own values
- [ ] TLS terminated in front, HTTP not reachable directly
- [ ] `SCIM_TRUSTED_PROXIES` naming your proxy, not the default
- [ ] `SCIM_COUCHBASE_TLS_SKIP_VERIFY` unset, `SCIM_COUCHBASE_CA_CERT` set
- [ ] A Couchbase account scoped to the buckets GoSCIM uses
- [ ] Rate limiting in the proxy
- [ ] The `Element` resource type removed, unless you want the example endpoint
- [ ] Decided what happens to `password`, given it is stored in the clear
