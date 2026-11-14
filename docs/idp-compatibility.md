# Identity provider compatibility

GoSCIM 1.3+ implements the wire behaviours, authentication schemes, filter
dialects and `PATCH` / `PUT` conventions required by **Microsoft Entra ID**,
**Okta**, and **Keycloak**. Every flow documented on this page is exercised
both in the unit conformance suite (`scim/compat_*_test.go`) and end to end
against a live Couchbase 7 cluster (`test/integration/idp_compat_test.go`).

## Compatibility matrix

| Capability | Microsoft Entra ID | Okta | Keycloak | GoSCIM support |
|---|---|---|---|---|
| Static opaque `Bearer` secret token | Required (Secret Token mode) | Required (API Token mode) | Optional | `SCIM_AUTH=bearer` (constant-time SHA-256 comparison) |
| OAuth 2.0 JWT against JWKS | Entra v2.0 (`RS256`, `scp`, `roles`, `oid`/`azp`) | Custom Auth Server (`RS256` / `ES256`, `scp`, `cid`) | Realm OIDC certs (`RS256` / `ES256` / `EdDSA` + `enc` keys ignored) | `SCIM_AUTH=jwt` (`RS256`/`384`/`512`, `ES256`/`384`/`512`, `EdDSA`) |
| Nested / custom JWT role claims | `scp`, `roles` | `scp`, `groups` | `realm_access.roles`, `resource_access.<client>.roles` | Built-in (`scp`, `realm_access.roles`) + dotted `SCIM_JWT_ROLES_CLAIM` |
| Common `externalId` (`caseExact: true`) | Object ID correlation | Okta `00u...` / `00g...` correlation | Keycloak user / group UUID | Validated, projected, sortable, case-exact in N1QL & memory |
| Case-insensitive string filters | `userName eq "ADELEV@..."` | `userName eq "USER@..."` | `userName eq "..."` | `LOWER(...)` in N1QL and folded matching when `caseExact: false` |
| Colon-delimited URN paths in filters & `PATCH` | `urn:...:User:employeeNumber`, `...:manager` | `urn:...:User:department` | `urn:...:User:userName` | Accepted in `filter`, `sortBy`, `PATCH`, `attributes`, `excludedAttributes` |
| Dotted multi-valued filters | `emails.value eq "..."` | `emails.value eq "..."` | `members.value eq "..."` | Quantified `ANY _v0 IN ... SATISFIES ... END` in N1QL & memory |
| Bracketed array value filters in `PATCH` | `emails[type eq "work"].value`, `members[value eq "..."]` | `emails[type eq "work"].value` | `emails[type eq "work"].value` | Supported on `add` (with entry synthesis), `replace`, `remove` |
| Pathless `PATCH` with dotted / URN keys | `{"op":"Add","value":{"urn:...:department":"..."}}` | `{"op":"replace","value":{"active":false}}` | Standard pathless `PATCH` | Expanded and schema-validated |
| Legacy `PATCH Remove` on `members` with `value` array | Sent when `?aadOptscim062020` is omitted | Sent by Group Push connectors | Optional | Removes only matching member elements, keeps remaining members |
| Legacy stringified booleans & manager ID | `"active": "False"`, `"manager": "<id>"` | Native JSON types | Native JSON types | Normalised to JSON `boolean` and `{"value": "<id>"}` |
| Explicit JSON `null` clearing on `POST` / `PUT` | Optional attributes | Full `PUT` User replacement & `POST` | Optional attributes | Clears non-required attributes and optional schema extensions |
| `addresses.primary` sub-attribute | Sent in default gallery mapping | Sent in User profile push | Standard RFC 7643 §4.1.2 | Declared in `urn:ietf:params:scim:schemas:core:2.0:User` |
| `/Bulk` & `ETag` / `If-Match` | Not used by Entra provisioning | Optional | Used by outbound realm sync plugins | `POST /scim/v2/Bulk` (`bulkId:`), weak `ETag`, `412` on stale `If-Match` |

---

## Microsoft Entra ID

Microsoft Entra ID's provisioning service connects to SCIM 2.0 endpoints either
with a long-lived **Secret Token** (`Authorization: Bearer <token>`) or with an
OAuth 2.0 bearer JWT issued by Entra ID.

### 1. Configure GoSCIM

#### Option A: Secret Token (custom non-gallery Enterprise Application)

```bash
export SCIM_AUTH=bearer
export SCIM_BEARER_TOKEN='<long-random-secret-token>'
export SCIM_BEARER_SUBJECT='entra-provisioning'
export SCIM_BEARER_ROLES='admin'
export SCIM_PUBLIC_URL='https://scim.example.com'
```

Multiple rotation tokens can be active simultaneously via
`SCIM_BEARER_TOKENS='entra-current:<token1>:admin;entra-next:<token2>:admin'`.

#### Option B: Entra ID v2.0 OAuth 2.0 JWT

```bash
export SCIM_AUTH=jwt
export SCIM_JWT_JWKS_URL='https://login.microsoftonline.com/<tenant-id>/discovery/v2.0/keys'
export SCIM_JWT_ISSUER='https://login.microsoftonline.com/<tenant-id>/v2.0'
export SCIM_JWT_AUDIENCE='api://<app-registration-client-id>'
export SCIM_PUBLIC_URL='https://scim.example.com'
```

GoSCIM reads Entra ID App Roles from `roles`, delegated scopes from `scp`, and
resolves the caller subject from `sub`, `azp`, `oid`, or `appid`.

### 2. Configure Entra ID Provisioning

1. In **Microsoft Entra admin center** → **Enterprise applications** → **New
   application** → **Create your own application** (Non-gallery).
2. Open **Provisioning** → **Provisioning Mode: Automatic**.
3. Under **Admin Credentials**:
   - **Tenant URL**: `https://scim.example.com/scim/v2?aadOptscim062020`
     (GoSCIM also supports the legacy Tenant URL without `?aadOptscim062020`).
   - **Secret Token**: paste the value of `SCIM_BEARER_TOKEN`.
4. Click **Test Connection** (Entra queries
   `GET /scim/v2/Users?filter=userName eq "<random-guid>"` and expects `200 OK`
   with `totalResults: 0`).
5. Enable **Provision Azure Active Directory Users** and **Provision Azure
   Active Directory Groups**.

### Entra ID dialect details handled automatically

- **Case-insensitive `userName` + case-exact `externalId`**: Entra matches users
  by `userName eq "..."` (matched case-insensitively in N1QL) and `externalId eq
  "..."` (matched case-sensitively per RFC 7643 §3.1).
- **Pathless `PATCH` with colon-URN and dotted keys**: operations such as
  `{"op": "Add", "value": {"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department": "Engineering", "name.givenName": "Adele"}}`
  expand into their target attributes and attach the `EnterpriseUser` schema URI
  when missing.
- **Stringified booleans and `manager`**: legacy Entra payloads sending
  `"active": "True"` / `"False"` or `"manager": "<userId>"` are normalised to
  `true` / `false` and `{"value": "<userId>"}` before storage so subsequent
  N1QL filters (`active eq false`, `...:manager.value eq "<userId>"`) match.
- **Group member removal**: both `?aadOptscim062020` (`"path":
  "members[value eq \"<id>\"]"`) and legacy Entra removal (`"path": "members"`,
  `"value": [{"value": "<id>"}]`) remove only the targeted member and update
  `User.groups` accordingly.

Ready-to-run requests: `httpexamples/idp_entra.http`.

---

## Okta

Okta's SCIM 2.0 Lifecycle Management and Group Push connectors support **HTTP
Header (`Bearer` API Token)**, **OAuth 2.0**, and **HTTP Basic**.

### 1. Configure GoSCIM

#### Option A: API Token (recommended for Okta SCIM 2.0 integrations)

```bash
export SCIM_AUTH=bearer
export SCIM_BEARER_TOKEN='<okta-scim-api-token>'
export SCIM_BEARER_SUBJECT='okta-lifecycle-mgmt'
export SCIM_BEARER_ROLES='admin'
export SCIM_PUBLIC_URL='https://scim.example.com'
```

#### Option B: Okta Custom Authorization Server (OAuth 2.0 `RS256` / `ES256`)

```bash
export SCIM_AUTH=jwt
export SCIM_JWT_JWKS_URL='https://<org>.okta.com/oauth2/<authServerId>/v1/keys'
export SCIM_JWT_ISSUER='https://<org>.okta.com/oauth2/<authServerId>'
export SCIM_JWT_AUDIENCE='https://scim.example.com/scim/v2'
export SCIM_PUBLIC_URL='https://scim.example.com'
```

Okta client-credentials tokens carry scopes in `"scp": [...]` and the client
identifier in `"cid"` (with no `"sub"` claim on some custom authorisation
servers); GoSCIM extracts both natively.

### 2. Configure Okta SCIM Integration

1. In **Okta Admin Console** → **Applications** → **Create App Integration**
   (or add **SCIM 2.0 Test App (Header Auth / OAuth Bearer Token)**).
2. Under **Provisioning** → **Integration**:
   - **SCIM connector base URL**: `https://scim.example.com/scim/v2`
   - **Unique identifier field for users**: `userName`
   - **Supported provisioning actions**: Push New Users, Push Profile Updates,
     Push Groups, Import New Users and Profile Updates, Push Password Updates
     (optional; GoSCIM hashes passwords with Argon2id by default).
   - **Authentication Mode**: **HTTP Header** → `Authorization: Bearer <token>`
     (or **OAuth 2.0**).

### Okta lifecycle details handled automatically

- **Full `PUT` profile updates**: Okta replaces user profiles via `PUT
  /scim/v2/Users/{id}`, echoing server-owned `readOnly` fields (`groups`, `meta`),
  omitting `writeOnly` `password`, and setting cleared optional fields or
  extensions to `null` (`"title": null`,
  `"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": null`). GoSCIM
  unassigns `null` optional attributes/extensions, keeps computed `User.groups`,
  and preserves the stored Argon2id password hash.
- **Dotted multi-valued search**: `GET /scim/v2/Users?filter=emails.value eq
  "..."` translates to a quantified N1QL `ANY ... SATISFIES ... END` query with
  case-insensitive matching.
- **Deactivation**: Okta's pathless deactivation patch
  (`{"Operations":[{"op":"replace","value":{"active":false}}]}`) sets
  `active: false` without disturbing other attributes or group memberships.
- **Case-insensitive schema discovery**: `GET /scim/v2/Schemas` and
  `GET /scim/v2/Schemas/{urn}` match schema URNs case-insensitively and expose
  `canonicalValues` and `referenceTypes`.

Ready-to-run requests: `httpexamples/idp_okta.http`.

---

## Keycloak

Keycloak integrates with GoSCIM in two complementary ways:

1. **As an OIDC / OAuth 2.0 Identity Provider (`SCIM_AUTH=jwt`)** issuing
   service-account or user access tokens verified against the realm JWKS.
2. **As an outbound SCIM 2.0 provisioning source** (via Keycloak SCIM
   provisioning extensions such as `scim-for-keycloak` or `SCIM-SDK`) pushing
   realm users and groups into GoSCIM.

### 1. Configure Keycloak Realm Client for `SCIM_AUTH=jwt`

1. In **Keycloak Admin Console**, create a confidential client (for example,
   `goscim-provisioner`) with **Service accounts roles** enabled.
2. Add an **Audience** protocol mapper (`Included Client Audience:
   goscim-service`, `Add to access token: On`) so minted access tokens carry
   `"aud": "goscim-service"`.
3. Assign the required role (`admin` for the shipped schemas) either as a
   **Realm role** (emitted in `realm_access.roles`) or as a **Client role** on
   `goscim-service` (emitted in `resource_access.goscim-service.roles`).

### 2. Configure GoSCIM

```bash
export SCIM_AUTH=jwt
export SCIM_JWT_JWKS_URL='https://sso.example.com/realms/<realm>/protocol/openid-connect/certs'
export SCIM_JWT_ISSUER='https://sso.example.com/realms/<realm>'
export SCIM_JWT_AUDIENCE='goscim-service'
# Optional when using Keycloak client roles in addition to realm_access.roles:
export SCIM_JWT_ROLES_CLAIM='resource_access.goscim-service.roles'
export SCIM_PUBLIC_URL='https://scim.example.com'
```

### Keycloak details handled automatically

- **Mixed realm JWKS (`sig` + `enc` keys across `RS256`, `ES256`, `EdDSA`)**:
  Keycloak's `/protocol/openid-connect/certs` endpoint publishes an `RSA-OAEP`
  encryption key (`"use": "enc"`) alongside active `RS256`, `ES256` or `EdDSA`
  (`Ed25519`) signing keys. GoSCIM ignores non-signing keys when loading the
  JWKS and refuses any token referencing an encryption key ID.
- **Nested role claims and `azp` subject fallback**: `realm_access.roles` is
  inspected automatically, `SCIM_JWT_ROLES_CLAIM` accepts dotted nested paths
  such as `resource_access.goscim-service.roles` (or comma-separated lists of
  paths), and tokens omitting `sub` fall back cleanly to `client_id` / `azp`.
- **Realm bulk export and reconciliation**: `POST /scim/v2/Bulk` (with `bulkId:`
  references and `externalId` holding Keycloak UUIDs), `ETag` / `If-Match`
  optimistic concurrency (`412 Precondition Failed`), core URN-prefixed filters
  (`urn:ietf:params:scim:schemas:core:2.0:User:userName eq "..."`), and
  `GET /scim/v2/Groups?filter=members.value eq "<userId>"` all work out of the
  box.

Ready-to-run requests: `httpexamples/idp_keycloak.http`.

---

## Verification and automated evidence

Compatibility is continuously verified at three levels:

1. **N1QL filter translation unit tests** (`scim/parser/value_filter_test.go` —
   `TestFilterToN1QLWithOptionsIdPPatterns`): verifies generated N1QL and bound
   parameters for case-insensitive `userName`, case-exact `externalId`,
   colon-delimited `EnterpriseUser` URNs, core URN prefixes, and dotted
   multi-valued `emails.value` / `members.value` expressions.
2. **Router + `MemoryStore` conformance suites** (`scim/compat_entra_test.go`,
   `scim/compat_okta_test.go`, `scim/compat_keycloak_test.go`): exercises
   authentication, validation, `PATCH` / `PUT` / `/Bulk` semantics, RBAC and
   group membership synchronisation with the Go race detector enabled.
3. **End-to-end Couchbase integration suite**
   (`test/integration/idp_compat_test.go`): starts `couchbase:community-7.1.1`
   in Docker and runs the complete Microsoft Entra ID, Okta Runscope, and
   Keycloak OIDC + SCIM lifecycles against real Couchbase KV, `_default.scim_unique`
   atomic locks, and N1QL queries.

Run both suites locally:

```bash
# Unit conformance tests (Entra ID, Okta, Keycloak) with the race detector
go test -race -v -run "TestEntraID|TestOkta|TestKeycloak" ./scim
go test -race -v -run "TestFilterToN1QLWithOptionsIdPPatterns" ./scim/parser

# Full unit + vet + race suite
make check

# End-to-end integration suite against Couchbase in Docker
make integration
```
