# Configuration

Two kinds: environment variables for deployment concerns, and JSON files for
the SCIM model.

## Environment variables

### Authentication

`SCIM_AUTH` has no default. The server exits with a message rather than
starting unauthenticated by omission.

| Variable | Default | Meaning |
|---|---|---|
| `SCIM_AUTH` | *(required)* | `jwt`, `basic`, or `none` |
| `SCIM_JWT_JWKS_URL` | — | Required for `jwt`. Where the issuer publishes its signing keys |
| `SCIM_JWT_ISSUER` | — | Required for `jwt`. The `iss` a token must carry |
| `SCIM_JWT_AUDIENCE` | — | Required for `jwt`. The `aud` a token must carry |
| `SCIM_JWT_ROLES_CLAIM` | — | An extra claim to read roles from, for issuers that use their own |
| `SCIM_BASIC_USERS` | — | Required for `basic`. `user:password:role1,role2;other:password:role3` |
| `SCIM_ANONYMOUS_ROLES` | *(empty)* | Roles given to every caller under `none` |

Issuer and audience are required rather than optional. A token is only an
authorization decision if the server checks it was minted by the issuer it
trusts and intended for this service; skipping either turns any token the
client can obtain anywhere into a valid credential here.

Under `SCIM_AUTH=none` with no `SCIM_ANONYMOUS_ROLES`, callers hold no roles,
so every attribute with a `$reader` list is filtered out of responses.

### Server

| Variable | Default | Meaning |
|---|---|---|
| `SCIM_PORT` | `:8080` | Listen address, in `host:port` form |
| `SCIM_CONFIG_DIR` | `config` | Directory holding `schemas/`, `resourceType/`, `bucketSettings/`, `serviceProviderConfig/` |
| `SCIM_TRUSTED_PROXIES` | `127.0.0.1` | Comma-separated addresses whose `X-Forwarded-For` is believed, or `none` to trust nobody |

Fixed in code rather than configurable, because no deployment has asked yet:
read header timeout 10s, read and write timeouts 30s, idle timeout 120s,
shutdown drain 20s, max header 1 MiB, max body 1 MiB, max search page 200.

### Couchbase

| Variable | Default | Meaning |
|---|---|---|
| `SCIM_COUCHBASE_URL` | `localhost` | Host, optionally `host:port` |
| `SCIM_ADMIN_USER` | `Administrator` | Cluster user. **Not** the SCIM API credential |
| `SCIM_ADMIN_PASSWORD` | — | Cluster password |
| `SCIM_COUCHBASE_TLS` | `true` | `false` connects over plain `couchbase://` |
| `SCIM_COUCHBASE_CA_CERT` | — | Path to the PEM that signed the cluster certificate |
| `SCIM_COUCHBASE_TLS_SKIP_VERIFY` | `false` | Accept any certificate. Logs a warning |
| `SCIM_QUERY_CONSISTENCY` | `request_plus` | `not_bounded` trades correctness for latency |

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
`filter.maxResults`, the authentication schemes — is derived from the running
code. It used to be served from this file, which advertised bulk support that
does not exist. A capability document that can drift from the implementation
is worse than none, because a client believes it.
