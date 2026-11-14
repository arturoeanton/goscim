# Installation

## Prebuilt binary (1-line installer)

Prebuilt static binaries (`CGO_ENABLED=0`) are published for Linux and macOS (`amd64` and `arm64`) and Windows (`amd64`):

```bash
curl -fsSL https://raw.githubusercontent.com/arturoeanton/goscim/main/scripts/install.sh | sh
```

Pin a release tag or custom destination directory when needed:

```bash
GOSCIM_VERSION=v1.4.0 GOSCIM_INSTALL_DIR=/usr/local/bin \
  curl -fsSL https://raw.githubusercontent.com/arturoeanton/goscim/main/scripts/install.sh | sh
```

Because the binary embeds the default `config/` directory (`User`, `Group`, `Element`, and `EnterpriseUser`), `goscim -demo` and `SCIM_STORE=sqlite` work from any working directory without external files. When a local `./config` directory exists (or `SCIM_CONFIG_DIR` is set explicitly), disk files take precedence.

## Container image (`ghcr.io`)

Multi-arch images (`linux/amd64`, `linux/arm64`) built onto `gcr.io/distroless/static-debian12:nonroot` with a native `-healthcheck` probe:

```bash
# Instant demo mode with HTMX Web Console on http://localhost:8080/ui
docker run --rm -p 8080:8080 ghcr.io/arturoeanton/goscim:latest -demo

# Persistent single-node SQLite mode
docker run -d --name goscim -p 8080:8080 \
  -v goscim-data:/data \
  -e SCIM_STORE=sqlite \
  -e SCIM_SQLITE_PATH=/data/goscim.db \
  -e SCIM_AUTH=bearer \
  -e SCIM_BEARER_TOKEN=change-me-32-bytes \
  -e SCIM_BEARER_ROLES=admin \
  -e SCIM_UI=true \
  -e SCIM_UI_TOKEN=change-me-ui-token \
  ghcr.io/arturoeanton/goscim:latest
```

## Storage backends and requirements

| Backend | `SCIM_STORE` | External dependencies | Best for |
|---|---|---|---|
| Couchbase 7.x | `couchbase` *(default)* | Couchbase Data, Query, Index services (200 MB quota per resource type by default) | Horizontally scaled, multi-replica production clusters |
| Embedded SQLite | `sqlite` | None (`modernc.org/sqlite`, pure Go, WAL mode) | Single-node appliances, VMs, edge deployments, persistent staging |
| In-memory | `memory` | None | `-demo`, local sandboxes, ephemeral CI conformance jobs |

Building from source requires Go 1.26 or later (1.27.4+ recommended).

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

GoSCIM creates its buckets, `_default.scim_unique` collections, primary indexes,
and schema-derived secondary N1QL indexes (`idx_scim_*`) on first start, which
needs cluster-level privilege. Once they exist, that privilege is no longer
needed.

1. Start once with an administrator account so the buckets, `_default.scim_unique`
   collections, and GSI indexes are created.
2. Create a user with, per GoSCIM bucket: **Application Access** (or **Data
   Reader** and **Data Writer** across `_default`), **Query Select**, **Query
   Insert**, **Query Update**, **Query Delete**, and **Query Manage Index** (or
   set `SCIM_COUCHBASE_AUTO_INDEX=false` after initial provisioning).
3. Switch `SCIM_ADMIN_USER` and `SCIM_ADMIN_PASSWORD` to that user.

`EnsureBucket` still runs on every start, but finds the buckets and indexes
present (`CREATE INDEX IF NOT EXISTS` is a no-op).

## Building from source

```bash
git clone https://github.com/arturoeanton/goscim.git
cd goscim
CGO_ENABLED=0 go build -o goscim .
```

Or run the checks first:

```bash
make check   # build, vet, and the unit suite with the race detector
```

### Docker Compose stacks

Two Compose stacks ship at the repository root:

```bash
# Couchbase 7.1.1 + cluster init + GoSCIM
docker compose up -d --build

# Keycloak 26 OIDC/JWKS realm + SQLite-backed GoSCIM + HTMX Web Console (/ui)
docker compose -f docker-compose.keycloak.yml up -d --build
```

In the Keycloak lab, Keycloak listens on <http://localhost:8081> (`admin` /
`admin`, pre-imported realm `goscim-lab` with OIDC client `goscim-provisioner`),
and GoSCIM listens on <http://localhost:8080> (`/ui` enabled, `SCIM_AUTH=bearer`
with `keycloak-lab-scim-token` by default, and `SCIM_JWT_*` pre-wired so
`SCIM_AUTH=jwt docker compose -f docker-compose.keycloak.yml up` validates OIDC
tokens against Keycloak's JWKS directly).

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

### Kubernetes and Helm

A Helm chart ships under `charts/goscim/` with support for `couchbase`,
`sqlite` (backed by a `PersistentVolumeClaim`), and `memory` stores, Kubernetes
`Secret` generation or `existingSecret` binding, `/healthz` and `/readyz`
probes, `terminationGracePeriodSeconds: 30`, and optional Prometheus Operator
`ServiceMonitor`:

```bash
# Single-replica SQLite deployment with bearer auth
helm upgrade --install goscim ./charts/goscim \
  --set store.backend=sqlite \
  --set auth.mode=bearer \
  --set auth.bearerToken="$(openssl rand -hex 32)" \
  --set auth.bearerRoles="admin"

# Multi-replica Couchbase deployment with JWT auth and ServiceMonitor
helm upgrade --install goscim ./charts/goscim \
  --set store.backend=couchbase \
  --set replicaCount=3 \
  --set store.couchbase.url=couchbase.db.svc.cluster.local \
  --set store.couchbase.adminPassword="..." \
  --set auth.mode=jwt \
  --set auth.jwt.jwksURL=https://idp.example.com/.well-known/jwks.json \
  --set auth.jwt.issuer=https://idp.example.com \
  --set auth.jwt.audience=scim-api \
  --set metrics.serviceMonitor.enabled=true
```

## First start

The first run against Couchbase creates one bucket per resource type, its
primary index, and schema-derived secondary indexes (`idx_scim_*`). Expect log
lines like:

```
GoSCIM 1.4.0 starting (store=couchbase)
urn:ietf:params:scim:schemas:core:2.0:User
...
Create Bucket -> User
Ready Bucket -> User
scim indexes on User: 18 active
listening on :8080
```

Index creation retries while the query service catches up with a newly created
bucket, which takes a few seconds on a fresh cluster.

## Verifying

```bash
curl -s http://localhost:8080/scim/v2/ServiceProviderConfig | jq .
curl -s http://localhost:8080/scim/v2/ResourceTypes | jq '.Resources[].endpoint'

# Live end-to-end conformance suite (RFC 7643/7644 + Entra ID + Okta + Keycloak)
goscim conformance -url http://localhost:8080/scim/v2 -token "$SCIM_BEARER_TOKEN"
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

**`SCIM_AUTH is not set`** — deliberate. Choose `jwt`, `bearer`, `basic`, or `none`.

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
