# Operations

## Lifecycle

The process is stateless: everything lives in Couchbase, so instances can be
added, removed or replaced freely.

**Startup** stops at the first failure, in this order: authenticator
configuration, Couchbase connection, config directory, bucket and index
creation. A failure at any of them exits non-zero rather than serving a
half-configured server.

**Shutdown** on `SIGINT` or `SIGTERM` stops accepting connections and drains
in-flight requests for up to 20 seconds. Give the supervisor a stop timeout
above that — `TimeoutStopSec=30` for systemd,
`terminationGracePeriodSeconds: 30` for Kubernetes — or it will kill the
process mid-request.

## Timeouts

Fixed in code. Change them there if a deployment needs different values.

| | Value |
|---|---|
| Read header | 10 s |
| Read | 30 s |
| Write | 30 s |
| Idle | 120 s |
| Shutdown drain | 20 s |
| Max header | 1 MiB |
| Max body | 1 MiB |

Long-running clients should expect connections to be closed after 120 seconds
idle.

## Search consistency

The single most consequential runtime setting.

`SCIM_QUERY_CONSISTENCY=request_plus`, the default, makes every search wait for
the index to include every mutation accepted before the query was issued. That
costs latency and buys the guarantee that a resource created a moment ago is
found. Identity providers create-then-search constantly, so this is the right
default.

`not_bounded` skips the wait. Searches get faster and a freshly written
resource may be missing from the results for a short while. Choose it only
after measuring, and knowing that a provisioning run may then create duplicates
it believed did not exist.

Reads by id are unaffected: they go straight to the key-value service and are
always current.

## Capacity

Each resource type is a bucket, quota set at creation from
`config/bucketSettings/`, 200 MB by default. Three shipped resource types means
600 MB of cluster quota for an untouched install.

Changing a bucket settings file afterwards does nothing — the settings apply at
creation. Change the bucket in Couchbase.

Searches are capped at 200 resources per page, which is also what
`/ServiceProviderConfig` advertises as `filter.maxResults`. A client asking for
more gets 200.

## Backups

Everything is in Couchbase; the server holds nothing. Use `cbbackupmgr` or
whatever your operator provides, per bucket.

Configuration lives in the config directory and in environment variables. It is
worth keeping under version control, since it defines the SCIM model.

## Logging

Requests go through gin's default logger, to stdout. Startup and shutdown are
logged with UTC timestamps.

What is **not** logged: who changed what. There is no audit trail. If you need
one, it is an [open item](../RELEASE-1.0.md) and would go in the write path
alongside the permission checks.

Warnings worth alerting on:

```
WARNING: SCIM_AUTH=none - every request is served unauthenticated
WARNING: SCIM_COUCHBASE_TLS_SKIP_VERIFY is set - the Couchbase certificate is not verified
```

Both mean a safety mechanism was deliberately disabled. Neither should appear
in production.

## Monitoring

There are no metrics, health or readiness endpoints. This is an
[open item](../RELEASE-1.0.md), and is the main thing missing for a
well-instrumented deployment.

Available today:

- **Liveness**: a TCP check on the listen port.
- **Readiness**: `GET /scim/v2/ServiceProviderConfig` answers 200 without
  credentials and only after the config loaded, so it is a reasonable proxy —
  but it does not prove Couchbase is reachable.
- **Couchbase health**: from the cluster's own monitoring.

## Troubleshooting

### 401 on every request

The credential is not reaching the server, or is not valid. Check
`WWW-Authenticate` in the response for the scheme in force. Under `jwt`, the
usual causes are an `iss` or `aud` that does not match, an expired token, or a
signing key absent from the JWKS.

The server deliberately answers the same way for every authentication failure,
so the response will not say which. The startup log records which scheme is
configured.

### 403 on a write

The caller's roles do not include one listed in the attribute's `$writer`. The
`detail` names the attribute. Check the roles in the token against the schema.

### An attribute is missing from a response

Three possible reasons, in order of likelihood:

1. The caller's roles do not match its `$reader` list — it is omitted, not
   blanked.
2. It is `returned: never` or `mutability: writeOnly`, like `password`.
3. It was never stored: a `readOnly` attribute sent by a client is dropped.

### 409 uniqueness on a retry

Working as intended: the schema declares that attribute unique. If the earlier
attempt actually failed, delete the resource it created, or reuse it.

### Searches return nothing after a create

With the default consistency this should not happen. Check
`SCIM_QUERY_CONSISTENCY`. If it is `not_bounded`, that is the cause.

Otherwise, check that the attribute in the filter is declared in the schema —
an undeclared attribute is not rejected in a filter, it simply matches nothing.

### 413

The body is over 1 MiB. That limit is in `scim/limits.go`.

### Startup fails creating a bucket

See [installation troubleshooting](installation.md#troubleshooting). The
common causes are Enterprise-only bucket settings against Community Edition,
insufficient cluster quota, and an account without permission to create
buckets.

## Upgrading

1. Read the [CHANGELOG](../CHANGELOG.md) for breaking changes.
2. Run `make check` against the new version.
3. Roll instances one at a time; the process is stateless and drains cleanly.

Stored documents have needed no migration so far. If that changes, the
CHANGELOG will say so.
