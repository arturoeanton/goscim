# Installation

## Requirements

| | |
|---|---|
| Go | 1.25 or later, to build |
| Couchbase | 7.x, Community or Enterprise |
| Services | Data, Query, Index |
| Memory | 200 MB of cluster quota per resource type, by default |

The shipped configuration defines three resource types, so a default install
asks for 600 MB of bucket quota. Adjust in
[bucket settings](configuration.md#bucket-settings) if that is too much.

## Couchbase

### Development

```bash
docker run -d --name goscim-db \
  -p 8091-8094:8091-8094 -p 11210:11210 \
  couchbase:community-7.1.1
```

Open <http://localhost:8091> and set up a cluster. Leave the data, query and
index services on.

### Production

Use a real cluster, and prefer a dedicated account over the administrator.

#### A least-privilege Couchbase user

GoSCIM creates its buckets and primary indexes on first start, which needs
cluster-level privilege. Once they exist, that privilege is no longer needed.

1. Start once with an administrator account so the buckets are created.
2. Create a user with, per GoSCIM bucket: **Application Access**, **Query
   Select**, **Query Insert**, **Query Update**, **Query Delete**.
3. Switch `SCIM_ADMIN_USER` and `SCIM_ADMIN_PASSWORD` to that user.

`EnsureBucket` still runs on every start, but finds the buckets present and
only asks for the primary index, which `IGNORE IF EXISTS` makes a no-op.

## Building

```bash
git clone https://github.com/arturoeanton/goscim.git
cd goscim
go build -o goscim .
```

Or run the checks first:

```bash
make check   # build, vet, and the unit suite with the race detector
```

The binary needs the config directory at runtime. Either run it from a
directory containing `config/`, or point `SCIM_CONFIG_DIR` at one.

### Container

There is no Dockerfile in the repository yet — it is an
[open item](../RELEASE-1.0.md). A minimal one:

```dockerfile
FROM golang:1.25 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /goscim .

FROM gcr.io/distroless/static-debian12
COPY --from=build /goscim /goscim
COPY config /config
ENV SCIM_CONFIG_DIR=/config
EXPOSE 8080
ENTRYPOINT ["/goscim"]
```

## Running

```bash
export SCIM_AUTH=jwt
export SCIM_JWT_JWKS_URL=https://issuer.example/.well-known/jwks.json
export SCIM_JWT_ISSUER=https://issuer.example
export SCIM_JWT_AUDIENCE=scim-api

export SCIM_COUCHBASE_URL=couchbase.internal
export SCIM_ADMIN_USER=goscim
export SCIM_ADMIN_PASSWORD=...
export SCIM_COUCHBASE_CA_CERT=/etc/ssl/couchbase-ca.pem

export SCIM_PORT=:8080
export SCIM_TRUSTED_PROXIES=10.0.0.5
export SCIM_CONFIG_DIR=/etc/goscim/config

./goscim
```

Every variable is documented in [configuration](configuration.md).

### systemd

```ini
[Unit]
Description=GoSCIM
After=network-online.target

[Service]
Type=simple
User=goscim
EnvironmentFile=/etc/goscim/env
ExecStart=/usr/local/bin/goscim
Restart=on-failure
KillSignal=SIGTERM
TimeoutStopSec=30

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadOnlyPaths=/etc/goscim

[Install]
WantedBy=multi-user.target
```

`TimeoutStopSec` should exceed the 20-second shutdown drain, so systemd does
not kill the process while it is finishing requests.

### Kubernetes

Nothing special is required: the process is stateless and handles `SIGTERM`.

- `terminationGracePeriodSeconds: 30`, again above the 20-second drain.
- No readiness probe endpoint exists yet. A TCP check on the port is the
  closest available; HTTP health endpoints are an
  [open item](../RELEASE-1.0.md).
- Secrets belong in a `Secret`, not in the manifest.

## First start

The first run creates one bucket per resource type and a primary index on
each. Expect log lines like:

```
GoSCIM 1.0.0 starting
urn:ietf:params:scim:schemas:core:2.0:User
...
Create Bucket -> User
Ready Bucket -> User
listening on :8080
```

Index creation retries while the query service catches up with a newly created
bucket, which takes a few seconds on a fresh cluster.

## Verifying

```bash
curl -s http://localhost:8080/scim/v2/ServiceProviderConfig | jq .
curl -s http://localhost:8080/scim/v2/ResourceTypes | jq '.Resources[].endpoint'
```

Discovery needs no credentials. If those work, the config loaded and the
buckets are reachable.

## Upgrading

Read the [CHANGELOG](../CHANGELOG.md) first. 1.0.0 changes the wire contract in
ways that break clients written against earlier code: authentication is
required, creates answer `201`, responses are `application/scim+json`, and
attributes the caller may not read are omitted rather than blanked.

Stored documents need no migration.

## Troubleshooting

**`SCIM_AUTH is not set`** — deliberate. Choose `jwt`, `basic`, or `none`.

**`Compression mode is supported in enterprise edition only`** — a bucket
settings file sets `compression_mode` against Community Edition. Remove the
key.

**`creating the primary index on "User" after N attempts`** — the query
service never became ready. Check that it is enabled on the cluster.

**`unambiguous timeout ... WaitUntilReady`** — the server cannot reach
Couchbase. Check `SCIM_COUCHBASE_URL`, credentials, and whether the cluster
speaks TLS on the port you are using; a development cluster usually needs
`SCIM_COUCHBASE_TLS=false`.

**A resource that was just created is not found by a search** — should not
happen with the default `SCIM_QUERY_CONSISTENCY=request_plus`. If it is set to
`not_bounded`, that is the cause.

More in [operations](operations.md).
