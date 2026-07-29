# API reference

Base path `/scim/v2`. Payloads are `application/scim+json` (RFC 7644 §3.1);
`application/json` is accepted on requests.

Every resource endpoint requires authentication. Discovery does not.

## Endpoints

Resource endpoints are generated from `config/resourceType/`. The shipped
configuration produces `/Users`, `/Groups` and `/Elements`.

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/scim/v2/{resource}` | Create |
| `GET` | `/scim/v2/{resource}/{id}` | Read one |
| `GET` | `/scim/v2/{resource}` | Search |
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
    { "op": "add", "path": "emails", "value": [{ "value": "j@example.org" }] },
    { "op": "remove", "path": "nickName" },
    { "op": "replace", "value": { "title": "Engineer" } }
  ]
}
```

- `add` appends when the target is an array, and sets otherwise.
- An operation with no `path` applies the members of its object value to the
  resource. Several identity providers send this form.
- `remove` requires a path.
- Value paths — `emails[type eq "work"].value` — are **not implemented**, and
  are refused with `400` rather than misinterpreted as a literal attribute
  name.

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
userType eq "Employee" and (emails co "example.com" or emails co "example.org")
meta.lastModified gt "2011-05-13T04:42:34Z"
urn:ietf:params:scim:schemas:extension:enterprise:2.0:User.department eq "Sales"
```

String values are double-quoted; numbers and booleans are bare. Attribute names
and schema URNs are case-insensitive.

`%` and `_` inside a `co`, `sw` or `ew` value are literal characters, not
wildcards.

A filter that does not parse is refused with `400` and
`"scimType": "invalidFilter"` rather than translated, so no part of a
malformed expression reaches the query.

**Not implemented:** value paths, `emails[type eq "work"]`. They parse as
grammar but are not translated.

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

## Attribute visibility

What comes back depends on the schema and on the caller:

- `returned: never` and `mutability: writeOnly` are never returned. The shipped
  `password` is both.
- An attribute whose `$reader` list does not include one of the caller's roles
  is **omitted**, not blanked. An empty string cannot be told apart from a
  legitimately empty value.

This applies to every response carrying a resource, including create, replace
and patch — not only read and search.

See [schemas and permissions](schemas.md).
