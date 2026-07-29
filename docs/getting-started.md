# Getting started

A running server and your first user. Ten minutes, assuming Docker and Go 1.25
or later.

## 1. Start Couchbase

```bash
docker run -d --name goscim-db \
  -p 8091-8094:8091-8094 -p 11210:11210 \
  couchbase:community-7.1.1
```

Open <http://localhost:8091>, choose **Setup New Cluster**, and set the
administrator credentials. This guide assumes `Administrator` / `password`.
Accept the defaults for services; GoSCIM needs the data, query and index
services, which are all on by default.

Community Edition is enough. If you use Enterprise you can additionally set
`compression_mode` in the bucket settings — see
[configuration](configuration.md#bucket-settings).

## 2. Run the server

```bash
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# How clients authenticate to the SCIM API.
export SCIM_AUTH=basic
export SCIM_BASIC_USERS='admin:secret:admin,role1'

# How the server authenticates to Couchbase. Different thing.
export SCIM_ADMIN_USER=Administrator
export SCIM_ADMIN_PASSWORD=password
export SCIM_COUCHBASE_TLS=false   # a local cluster with no certificate

go run .
```

On the first run the server creates one bucket per resource type — `User`,
`Group`, `Element` — and a primary index on each, then listens on `:8080`.

`SCIM_AUTH` has no default. The server refuses to start until you choose
`jwt`, `basic`, or `none`. An earlier version served every endpoint
unauthenticated, and that was not something to leave as an accident.

## 3. Ask the server what it supports

Discovery does not require credentials, which is deliberate: a client needs it
to find out how to authenticate.

```bash
curl -s http://localhost:8080/scim/v2/ServiceProviderConfig | jq
```

```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"],
  "patch": { "supported": true },
  "bulk": { "supported": false },
  "filter": { "supported": true, "maxResults": 200 },
  "changePassword": { "supported": false },
  "sort": { "supported": true },
  "etag": { "supported": true },
  "authenticationSchemes": [
    { "type": "httpbasic", "name": "HTTP Basic", "primary": true }
  ]
}
```

This document is derived from the code, not from a file, so it cannot claim a
capability the server does not have.

## 4. Create a user

```bash
curl -s -u admin:secret -X POST http://localhost:8080/scim/v2/Users \
  -H 'Content-Type: application/scim+json' \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "jane.doe@example.com",
    "name": { "givenName": "Jane", "familyName": "Doe" },
    "emails": [
      { "value": "jane.doe@example.com", "type": "work", "primary": true }
    ],
    "active": true
  }' -i
```

The response is `201 Created`, with the resource in the body, its URI in
`Location`, and its revision in `ETag`:

```
HTTP/1.1 201 Created
Content-Type: application/scim+json
Location: http://localhost:8080/scim/v2/Users/6f1c...-...
ETag: W/"9a20d07a-bb4b-4c06-af43-75c77e5f8c2c"
```

Creating the same `userName` again answers `409` with
`"scimType": "uniqueness"`: the core schema declares `userName` unique, which
is what stops a retried provisioning request from producing two users.

## 5. Read, search, update

```bash
ID=<the id from the response above>

# Read one
curl -s -u admin:secret http://localhost:8080/scim/v2/Users/$ID | jq

# Search, with a filter and sorting
curl -s -u admin:secret --get http://localhost:8080/scim/v2/Users \
  --data-urlencode 'filter=userName sw "jane" and active eq true' \
  --data-urlencode 'sortBy=userName' | jq

# Patch one attribute
curl -s -u admin:secret -X PATCH http://localhost:8080/scim/v2/Users/$ID \
  -H 'Content-Type: application/scim+json' \
  -d '{
    "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
    "Operations": [{ "op": "replace", "path": "active", "value": false }]
  }' | jq

# Delete
curl -s -u admin:secret -X DELETE http://localhost:8080/scim/v2/Users/$ID -i
```

Attribute names are case-insensitive, so `USERNAME` works too; the resource is
stored under the spelling the schema declares.

## 6. Add a resource type without writing code

Drop two files in `config/` and restart.

`config/schemas/urn+ietf+params+scim+schemas+custom+2.0+Device.json`:

```json
{
  "id": "urn:ietf:params:scim:schemas:custom:2.0:Device",
  "name": "Device",
  "description": "A managed device",
  "attributes": [
    {
      "name": "serialNumber",
      "type": "string",
      "required": true,
      "uniqueness": "server",
      "mutability": "immutable"
    },
    {
      "name": "assetTag",
      "type": "string",
      "mutability": "readWrite"
    }
  ]
}
```

`config/resourceType/Device.json`:

```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:ResourceType"],
  "id": "Device",
  "name": "Device",
  "endpoint": "/Devices",
  "description": "A managed device",
  "schema": "urn:ietf:params:scim:schemas:custom:2.0:Device",
  "schemaExtensions": []
}
```

Restart, and `/scim/v2/Devices` exists with all six verbs, a `Device` bucket,
`serialNumber` enforced unique and refused on update because it is
`immutable`. The file name of a schema replaces `:` with `+`, since a colon is
awkward in a path.

See [schemas and permissions](schemas.md) for the full attribute vocabulary,
including the `$reader` and `$writer` role lists.

## Where next

- [Configuration](configuration.md) — every environment variable
- [API reference](api.md) — status codes, filters, concurrency control
- [Security](security.md) — before exposing this to anything real
