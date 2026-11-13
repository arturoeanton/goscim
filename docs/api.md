# API reference

Base path `/scim/v2`. Payloads are `application/scim+json` (RFC 7644 §3.1);
`application/json` is accepted on requests.

Every resource endpoint requires authentication. Discovery and health probes
(`/healthz`, `/readyz`) do not.

Every response includes an `X-Request-Id` header. If the caller supplies a
valid `X-Request-Id` (1–128 ASCII characters from `[A-Za-z0-9._:-]`), it is
echoed back and attached to the structured audit event; otherwise the server
generates a UUID.

## Endpoints

Resource endpoints are generated from `config/resourceType/`. The shipped
configuration produces `/Users`, `/Groups` and `/Elements`.

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/scim/v2/{resource}` | Create |
| `GET` | `/scim/v2/{resource}/{id}` | Read one |
| `GET` | `/scim/v2/{resource}` | Search |
| `POST` | `/scim/v2/{resource}/.search` | Search via POST body (RFC 7644 §3.4.3) |
| `POST` | `/scim/v2/.search` | Search across all resource types (RFC 7644 §3.4.3) |
| `PUT` | `/scim/v2/{resource}/{id}` | Replace |
| `PATCH` | `/scim/v2/{resource}/{id}` | Modify |
| `DELETE` | `/scim/v2/{resource}/{id}` | Delete |

Discovery, unauthenticated:

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/scim/v2/ServiceProviderConfig` | Which parts of the spec this deployment implements |
| `GET` | `/scim/v2/ResourceTypes` | The resource types available |
| `GET` | `/scim/v2/ResourceTypes/{id}` | One of them |
| `GET` | `/scim/v2/Schemas` | Schemas and extensions |
| `GET` | `/scim/v2/Schemas/{urn}` | One of them |

## Create

`POST /scim/v2/Users`

```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "userName": "jane.doe@example.com",
  "emails": [{ "value": "jane.doe@example.com", "type": "work", "primary": true }],
  "active": true
}
```

Answers `201 Created` with the resource, a `Location` header holding its URI,
and an `ETag` holding its revision.

The server owns `id` and `meta`; anything the client sends for them is
discarded. Attributes the caller may not write are refused; attributes the
caller may not read are absent from the response.

## Read

`GET /scim/v2/Users/{id}` → `200`, or `404` if there is no such resource.

Send `If-None-Match` with a previously seen `ETag` to get `304 Not Modified`
instead of the body.

## Search

`GET /scim/v2/Users?filter=...&sortBy=...&sortOrder=...&startIndex=1&count=50`

| Parameter | Default | Notes |
|---|---|---|
| `filter` | *(none)* | See [filters](#filters) |
| `sortBy` | `id` | Must name an attribute the schema declares |
| `sortOrder` | `ascending` | `descending` for the other direction |
| `startIndex` | `1` | 1-based, per RFC 7644 |
| `count` | `100` | Capped at 200 |
| `attributes` | *(none)* | Comma-separated attributes to include (mutually exclusive with `excludedAttributes`) |
| `excludedAttributes` | *(none)* | Comma-separated attributes to omit (mutually exclusive with `attributes`) |

Clients that prefer not to put filters in URLs may `POST` an RFC 7644 §3.4.3
`SearchRequest` to `/scim/v2/{resource}/.search` (single resource type) or
`/scim/v2/.search` (all resource types):

```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:SearchRequest"],
  "attributes": ["userName", "emails.value"],
  "filter": "userName sw \"jane\"",
  "sortBy": "userName",
  "sortOrder": "ascending",
  "startIndex": 1,
  "count": 50
}
```

The response is a `ListResponse`:

```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:ListResponse"],
  "totalResults": 137,
  "itemsPerPage": 50,
  "startIndex": 1,
  "Resources": [ { "id": "...", "userName": "..." } ]
}
```

`totalResults` counts the whole match, not the page. `itemsPerPage` is how many
resources this page carries. All four fields are always present, including on
an empty result.

`sortBy` is checked against the schema and refused with `400` if it names
something undeclared. That is a whitelist rather than sanitising, because the
value reaches the query as an identifier rather than as a bound parameter.

## Replace and modify

`PUT` replaces the resource. The client need not send `meta`; the server takes
the previous one from storage, since `meta.created` is not the client's to set.

`PATCH` takes an RFC 7644 §3.5.2 operation list:

```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
  "Operations": [
    { "op": "replace", "path": "active", "value": false },
    { "op": "replace", "path": "emails[type eq \"work\"].value", "value": "jane.work@example.com" },
    { "op": "add", "path": "phoneNumbers[type eq \"mobile\"].value", "value": "+1-555-0100" },
    { "op": "remove", "path": "nickName" },
    { "op": "replace", "value": { "title": "Engineer" } }
  ]
}
```

- `add` appends when the target is an array, and sets otherwise.
- An operation with no `path` applies the members of its object value to the
  resource. Several identity providers send this form.
- `remove` requires a path.
- **Value-selection paths** (`emails[type eq "work"]`,
  `emails[type eq "work"].value`, `members[value eq "..."]`) are supported on
  `add`, `replace` and `remove`:
  - `replace` and `remove` return `400` (`scimType: "noTarget"`) when no array
    element matches the filter.
  - Removing all remaining elements of a multi-valued attribute unassigns the
    attribute per RFC 7644 §3.5.2.2.
  - `add` targeting `attr[filter].subAttr` updates matching elements when
    present, or synthesises a new array entry seeded from conjunctive `eq`
    predicates in `filter` when none match (matching Microsoft Entra ID's
    provisioning requests).
- Both `urn:...:Schema:attr` (RFC 7644 §3.10) and `urn:...:Schema.attr` are
  accepted in `PATCH` paths; patching a declared extension attribute attaches
  its schema URI to `"schemas"` automatically if absent.

Both funnel through the same validation as a create: the result is checked
against the schema, so a patch cannot leave a resource that violates it.

## Delete

`DELETE /scim/v2/Users/{id}` → `204 No Content`, or `404`.

## Concurrency control

Every write stamps a new `meta.version`, published as a weak `ETag`.

```bash
# Read, and keep the ETag
curl -i -u admin:secret http://localhost:8080/scim/v2/Users/$ID

# Write only if nobody else has since
curl -X PUT -u admin:secret http://localhost:8080/scim/v2/Users/$ID \
  -H 'If-Match: W/"9a20d07a-bb4b-4c06-af43-75c77e5f8c2c"' \
  -H 'Content-Type: application/scim+json' -d @user.json
```

A stale `If-Match` answers `412 Precondition Failed` and changes nothing.
`If-Match` is honoured on `PUT`, `PATCH` and `DELETE`; `If-None-Match` on
`GET`. Both accept the bare `meta.version`, its quoted form, its weak form,
and `*`.

Without `If-Match` the last write wins silently.

## Filters

RFC 7644 §3.4.2.2 operators, translated to N1QL with the values bound rather
than interpolated.

| Operator | Meaning |
|---|---|
| `eq` `ne` | Equal, not equal |
| `co` `sw` `ew` | Contains, starts with, ends with |
| `gt` `ge` `lt` `le` | Greater, greater or equal, less, less or equal |
| `pr` | Present, i.e. not null |
| `and` `or` `not` | Logical, with `( )` for grouping |

```
userName eq "bjensen"
name.familyName co "O"
title pr and userType eq "Employee"
emails[type eq "work" and primary eq true]
emails[type eq "work"].value ew "@example.com"
not emails[type eq "home"]
userType eq "Employee" and (emails co "example.com" or emails co "example.org")
meta.lastModified gt "2011-05-13T04:42:34Z"
urn:ietf:params:scim:schemas:extension:enterprise:2.0:User.department eq "Sales"
```

String values are double-quoted; numbers and booleans are bare. Attribute names
and schema URNs are case-insensitive.

Complex array value filters (`attr[subFilter]` and `attr[subFilter].subAttr op value`
or `attr[subFilter].subAttr pr`) translate to N1QL quantified `ANY ... IN ...
SATISFIES ... END` predicates with bound positional parameters.

`%` and `_` inside a `co`, `sw` or `ew` value are literal characters, not
wildcards.

A filter that does not parse is refused with `400` and
`"scimType": "invalidFilter"` rather than translated, so no part of a
malformed expression reaches the query.

## Errors

RFC 7644 §3.12:

```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"],
  "scimType": "uniqueness",
  "detail": "a User with this userName already exists",
  "status": "409"
}
```

| Status | When |
|---|---|
| `400` | Malformed payload, schema violation, bad filter, bad `sortBy` |
| `401` | Missing or invalid credentials. Carries `WWW-Authenticate` |
| `403` | The caller's roles do not permit writing that attribute |
| `404` | No such resource |
| `409` | A unique attribute already holds that value |
| `412` | `If-Match` did not match the current revision |
| `413` | Body over 1 MiB |
| `500` | Storage or internal failure |

`scimType` values in use: `invalidFilter`, `invalidValue`, `invalidSyntax`,
`invalidPath`, `noTarget`, `mutability`, `uniqueness`.

## Attribute visibility and projection

What comes back depends on the schema, the caller's roles, and optional
`attributes` / `excludedAttributes` parameters (accepted on `GET`, `POST`, `PUT`,
`PATCH` and `POST /.search`, per RFC 7644 §3.4.2.5 and §3.9):

- `returned: never` and `mutability: writeOnly` are never returned, even if
  requested in `attributes`. The shipped `password` is both.
- An attribute whose `$reader` list does not include one of the caller's roles
  is **omitted**, not blanked, even if requested in `attributes`.
- `returned: always` (along with `id` and `schemas`) is always returned when the
  caller may read it, even when `attributes` omits it or `excludedAttributes`
  names it.
- `returned: request` is omitted by default and included only when named in
  `attributes`.
- Both `urn:...:Schema:attr` (RFC 7644 §3.10) and `urn:...:Schema.attr` are
  accepted for extension and core URN-prefixed paths.

See [schemas and permissions](schemas.md).
