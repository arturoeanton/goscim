# Schemas and permissions

A schema is a JSON file listing attributes. It decides what a resource may
contain, what type each value has, who may read it and who may write it.
Nothing about a resource type is compiled in.

Files live in `config/schemas/`, named after the URN with `:` replaced by `+`:

```
config/schemas/urn+ietf+params+scim+schemas+core+2.0+User.json
```

## Shape

```json
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User",
  "name": "User",
  "description": "User Account",
  "attributes": [
    {
      "name": "userName",
      "type": "string",
      "multiValued": false,
      "description": "A unique identifier for the user.",
      "required": true,
      "caseExact": false,
      "mutability": "readWrite",
      "returned": "default",
      "uniqueness": "server"
    }
  ]
}
```

## Attribute keywords

Standard, from RFC 7643 §7.

| Keyword | Values | Effect |
|---|---|---|
| `type` | `string`, `boolean`, `decimal`, `integer`, `dateTime`, `binary`, `reference`, `complex` | Validated on write |
| `multiValued` | `true`, `false` | `true` requires a JSON array; each element is validated against `type` |
| `required` | `true`, `false` | Absent on a write is `400` |
| `subAttributes` | array | For `complex`; validated recursively |
| `mutability` | `readWrite`, `readOnly`, `immutable`, `writeOnly` | See below |
| `returned` | `default`, `always`, `never`, `request` | `never` is never in a response |
| `uniqueness` | `none`, `server`, `global` | `server` and `global` are enforced; a duplicate is `409` |
| `caseExact` | `true`, `false` | Recorded, not yet acted on |

`binary` and `reference` are accepted as JSON strings without further checking.
`returned: always` and `returned: request` behave like `default`, since
`attributes` and `excludedAttributes` are not implemented.

### mutability

| Value | On create | On update |
|---|---|---|
| `readWrite` | Accepted | Accepted |
| `readOnly` | Ignored | Ignored; the stored value is kept |
| `immutable` | Accepted | `400` with `scimType: mutability` if it differs |
| `writeOnly` | Accepted | Accepted; never returned |

A `readOnly` value from a client is dropped rather than refused. Read-modify-
write clients echo the whole resource back on a `PUT`, including attributes
that are not theirs, and refusing would make an ordinary update impossible.

### Multi-valued attributes

An attribute with `multiValued: true` carries an array, and each element is
validated with the attribute's own rules. Where the sub-attributes include
`primary`, at most one element may set it to `true` (RFC 7643 §2.4).

```json
{
  "name": "emails",
  "type": "complex",
  "multiValued": true,
  "subAttributes": [
    { "name": "value", "type": "string" },
    { "name": "type", "type": "string" },
    { "name": "primary", "type": "boolean" }
  ]
}
```

## Role-based permissions

Two non-standard keywords, `$reader` and `$writer`, hold role names. They are
this project's own extension and are omitted from `/Schemas`, since they are
not part of a schema a client should consume.

```json
{
  "name": "salary",
  "type": "integer",
  "$reader": ["hr", "payroll"],
  "$writer": ["payroll"]
}
```

- **No list** means unrestricted.
- **`["*"]`** means any role.
- Otherwise the caller needs at least one of the listed roles.

An attribute the caller cannot read is **omitted** from the response, not
returned empty: an empty string is indistinguishable from a legitimately empty
value, and is not a valid value for a non-string attribute at all. Writing one
without the role is `403`.

Both apply at every depth — inside complex values and inside the elements of
multi-valued ones — and inside schema extensions, which are resolved against
their own schema rather than the core one.

Roles come from the authenticated caller. See [security](security.md#roles).

## Extensions

A resource type may declare extensions:

```json
"schemaExtensions": [
  {
    "schema": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
    "required": false
  }
]
```

Their attributes live in an object keyed by the URN:

```json
{
  "schemas": [
    "urn:ietf:params:scim:schemas:core:2.0:User",
    "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
  ],
  "userName": "jane.doe",
  "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
    "department": "Engineering"
  }
}
```

`required: false` means the object may be absent. When it is present it is
validated like any other schema.

Filters, `sortBy` and PATCH paths reach extension attributes by prefixing the
URN:

```
urn:ietf:params:scim:schemas:extension:enterprise:2.0:User.department eq "Sales"
```

## Validation is a whitelist

Any key not declared in the schema is `400`. There is no passthrough for
unknown attributes; if you want one, declare it.

Attribute names are matched case-insensitively, and the stored resource uses
the spelling the schema declares. Accepting `USERNAME` and storing it that way
would leave a document that no filter can find, because N1QL identifiers are
case-sensitive even though SCIM attribute names are not.

## Shipped schemas

| Schema | Notes |
|---|---|
| `core:2.0:User` | RFC 7643 user. `password` is `writeOnly` / `returned: never`; `groups` is `readOnly` |
| `core:2.0:Group` | RFC 7643 group, with multi-valued `members` |
| `extension:enterprise:2.0:User` | Optional enterprise extension |
| `core:2.0:Element`, `extension:one:2.0:Element` | Example pair exercising `$reader`, `$writer`, extensions and every attribute type |

`Element` exists to demonstrate the model and to give the test suite something
to work against. Remove its resource type if you do not want the endpoint.

## Passwords

`password` is stored exactly as received. There is no hashing, and
`changePassword` is advertised as unsupported. Treat the `User` bucket as
holding secrets, or do not send passwords through SCIM.
