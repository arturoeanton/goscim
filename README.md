# 🚀 GoSCIM - Lightning Fast Identity Management



[![Go Report Card](https://goreportcard.com/badge/github.com/arturoeanton/goscim)](https://goreportcard.com/report/github.com/arturoeanton/goscim)
[![MIT License](https://img.shields.io/badge/License-MIT-green.svg)](https://choosealicense.com/licenses/mit/)
[![GitHub contributors](https://img.shields.io/github/contributors/arturoeanton/goscim.svg)](https://GitHub.com/arturoeanton/goscim/graphs/contributors/)
[![GitHub issues](https://img.shields.io/github/issues/arturoeanton/goscim.svg)](https://GitHub.com/arturoeanton/goscim/issues/)
[![GitHub stars](https://img.shields.io/github/stars/arturoeanton/goscim.svg?style=social&label=Star&maxAge=2592000)](https://GitHub.com/arturoeanton/goscim/stargazers/)

> 💡 **Note:** If you need to use PostgreSQL backend instead of Couchbase, please check out our sister project: **[go-vorpal-scim](https://github.com/arturoeanton/go-vorpal-scim)**.

---

> **A blazingly fast, lightweight SCIM 2.0 server built in Go that makes identity management simple and scalable** 🔥

GoSCIM is a small, config-driven implementation of the SCIM 2.0 protocol backed by Couchbase. Resource types, schemas and per-attribute permissions are JSON files, so adding one takes no code.

## ✨ Why GoSCIM?

- 🔧 **Dynamic schemas**: add a resource type by dropping in two JSON files
- 🛡️ **Authentication**: OAuth 2.0 access tokens (RFC 9068) or HTTP Basic
- 🔐 **Per-attribute authorization**: role lists plus SCIM's own `mutability`
- 🎯 **RFC 7643/7644**: CRUD, filtering, sorting, pagination, ETags
- 🔍 **Filter parser**: ANTLR grammar translated to N1QL with bound parameters
- 🧪 **Tested**: unit suite plus an integration suite against a real Couchbase

See [what is not implemented](CHANGELOG.md#not-implemented) before choosing it.

## 🎯 Perfect For

- **Startups** building their first identity system
- **Enterprises** replacing expensive identity solutions
- **DevOps Teams** automating user provisioning
- **SaaS Companies** needing multi-tenant identity management
- **Developers** learning SCIM protocol implementation

## ⚡ Quick Start

Get GoSCIM running in under 2 minutes:

```bash
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# A Couchbase to talk to
docker run -d --name db -p 8091-8094:8091-8094 -p 11210:11210 couchbase:community-7.1.1
# then create the cluster at http://localhost:8091 with the credentials below

export SCIM_AUTH=basic
export SCIM_BASIC_USERS='admin:secret:admin'
export SCIM_ADMIN_USER=Administrator      # Couchbase, not the SCIM API
export SCIM_ADMIN_PASSWORD='password'
export SCIM_COUCHBASE_TLS=false           # a local cluster without certificates

go run .
```

Then `curl -u admin:secret http://localhost:8080/scim/v2/ServiceProviderConfig`.

`SCIM_AUTH` has no default: the server refuses to start until you choose
`jwt`, `basic`, or `none` for an unauthenticated deployment.

## 🌟 Features That Make Developers Happy

### 🏗️ **Dynamic Schema System**
No code changes needed - just drop JSON schemas and restart:

```json
{
  "id": "urn:ietf:params:scim:schemas:custom:2.0:Employee",
  "name": "Employee",
  "attributes": [
    {
      "name": "employeeId",
      "type": "string",
      "required": true,
      "uniqueness": "server"
    }
  ]
}
```

### 🔍 **Powerful Query Engine**
Advanced filtering with natural syntax:

```http
GET /Users?filter=name.familyName co "Garcia" and active eq true
GET /Users?filter=emails[type eq "work" and value ew "@company.com"]
```

### 🔌 **Easy Integrations**
Connect to any system with our flexible connector architecture:

```go
// Custom connector in just a few lines
func (c *CustomConnector) SyncUsers() error {
    users := c.externalSystem.GetUsers()
    for _, user := range users {
        scimUser := convertToSCIM(user)
        c.scimClient.CreateOrUpdateUser(scimUser)
    }
    return nil
}
```

## 🛠️ Real-World Examples

### Create a User

Answers `201 Created` with the new resource in the body and its URI in
`Location`.

```bash
curl -X POST https://your-scim-server.com/scim/v2/Users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "jane.doe@company.com",
    "name": {
      "familyName": "Doe",
      "givenName": "Jane"
    },
    "emails": [{
      "value": "jane.doe@company.com",
      "type": "work",
      "primary": true
    }],
    "active": true
  }'
```

### Search with Filters
```bash
curl "https://your-scim-server.com/scim/v2/Users?filter=userName sw 'admin'&sortBy=name.familyName"
```

### Update User
```bash
curl -X PATCH https://your-scim-server.com/scim/v2/Users/123 \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
    "Operations": [{
      "op": "replace",
      "path": "active",
      "value": false
    }]
  }'
```

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Your Apps     │    │   Identity      │    │   External      │
│   (Consumers)   │◄──►│   Hub (GoSCIM)  │◄──►│   Systems       │
│                 │    │                 │    │   (Providers)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

- **Stateless**: no per-process state, so instances scale horizontally
- **Couchbase backend**: one bucket per resource type, N1QL for search

## 📈 Scale

No benchmarks have been published for this project, so this section used to
contain numbers that were not measured. If you run any, a pull request adding
them with the method alongside would be welcome.

## 🤝 Join Our Amazing Community

We're building something special, and we'd love your help! 

### 🌟 **Star us** if you find GoSCIM useful!

### 🛠️ **Ways to Contribute**
- 🐛 [Report bugs](https://github.com/arturoeanton/goscim/issues/new?template=bug_report.md)
- 💡 [Suggest features](https://github.com/arturoeanton/goscim/issues/new?template=feature_request.md)
- 📝 [Improve documentation](https://github.com/arturoeanton/goscim/tree/main/docs)
- 🔧 [Submit pull requests](https://github.com/arturoeanton/goscim/pulls)
- 💬 [Join discussions](https://github.com/arturoeanton/goscim/discussions)

### 🎯 **Quick Contribution Ideas**
- Implement `/Bulk` (RFC 7644 3.7)
- Implement value paths in filters and patch paths
- Implement `attributes` / `excludedAttributes`
- Add a Dockerfile and a compose file
- Add health, readiness and metrics endpoints
- Complete the translations under `docs/`

`RELEASE-1.0.md` lists what is open and why.

### 🏆 **Hall of Fame**
Special thanks to all our contributors! Every contribution matters, from bug reports to major features.

## 📚 Documentation

| Topic | Link |
|-------|------|
| 🚀 **Getting Started** | [docs/en/getting-started.md](docs/en/getting-started.md) |
| 🔧 **Installation Guide** | [docs/en/installation.md](docs/en/installation.md) |
| 📋 **Changelog** | [CHANGELOG.md](CHANGELOG.md) |
| 🔎 **1.0 audit and open items** | [RELEASE-1.0.md](RELEASE-1.0.md) |
| 🛡️ **Security Guide** | [docs/en/security.md](docs/en/security.md) |
| 👩‍💻 **Working on the code** | [CLAUDE.md](CLAUDE.md) |

### 🌍 **Multi-Language Docs**
- 🇺🇸 [English](docs/en/)
- 🇪🇸 [Español](docs/es/)

## ⚙️ Configuration

| Variable | Default | Meaning |
|---|---|---|
| `SCIM_AUTH` | *(none — required)* | `jwt`, `basic` or `none` |
| `SCIM_PORT` | `:8080` | Listen address |
| `SCIM_CONFIG_DIR` | `config` | Where schemas and resource types live |
| `SCIM_TRUSTED_PROXIES` | `127.0.0.1` | Comma-separated, or `none` |
| `SCIM_JWT_JWKS_URL` | — | Required for `SCIM_AUTH=jwt` |
| `SCIM_JWT_ISSUER` | — | Required for `jwt`; the token's `iss` |
| `SCIM_JWT_AUDIENCE` | — | Required for `jwt`; the token's `aud` |
| `SCIM_JWT_ROLES_CLAIM` | — | An extra claim to read roles from |
| `SCIM_BASIC_USERS` | — | `user:password:role1,role2;...` |
| `SCIM_ANONYMOUS_ROLES` | *(empty)* | Roles for `SCIM_AUTH=none` |
| `SCIM_ADMIN_USER` | `Administrator` | Couchbase user |
| `SCIM_ADMIN_PASSWORD` | — | Couchbase password |
| `SCIM_COUCHBASE_URL` | `localhost` | Couchbase host, `host:port` |
| `SCIM_COUCHBASE_TLS` | `true` | `false` for a plain local cluster |
| `SCIM_COUCHBASE_CA_CERT` | — | PEM that signed the cluster certificate |
| `SCIM_COUCHBASE_TLS_SKIP_VERIFY` | `false` | Accept any certificate |
| `SCIM_QUERY_CONSISTENCY` | `request_plus` | `not_bounded` trades freshness for latency |

Roles for `jwt` come from the token: RFC 9068 §2.2.3.1 reuses SCIM's own
`roles`, `groups` and `entitlements` claims, and `scope` is read too.

## 🛠️ Working on it

```bash
make check        # build, vet, and the unit suite with the race detector
make integration  # the suite against a real Couchbase, in a container
make cover        # coverage
make generate     # regenerate the filter parser (needs Docker)
```

## 🔧 Tech Stack

- **Language**: Go 1.25+
- **Database**: Couchbase 7.x (Community or Enterprise)
- **Web framework**: Gin
- **Query parser**: ANTLR 4.13
- **Auth**: OAuth 2.0 access tokens (RFC 9068) via JWKS, or HTTP Basic

## 📊 Project Status

- ✅ **Core operations**: create, read, replace, patch, delete, search
- ✅ **Discovery**: ServiceProviderConfig, ResourceTypes, Schemas
- ✅ **Filtering**: all RFC 7644 operators, with bound parameters
- ✅ **Schema extensions**: custom attributes and resource types
- ✅ **Authorization**: per-attribute roles, `mutability`, `returned`
- ✅ **Concurrency**: ETag with If-Match / If-None-Match
- ❌ **Bulk operations**: not implemented
- ❌ **Value paths** (`emails[type eq "work"]`): not implemented
- ❌ **`attributes` / `excludedAttributes`**: not implemented
- ❌ **Health, readiness, metrics endpoints**: not implemented

## 💡 Use Cases

### Identity Automation
```
Employee Onboarding → GoSCIM → Automatic provisioning in:
├── Active Directory
├── Salesforce
├── Slack
├── Jira
└── Custom Apps
```

### Multi-Tenant SaaS
```
Customer Signup → GoSCIM → Isolated tenant with:
├── Custom schemas
├── Role-based access
├── Branded experience
└── API access
```

### Compliance & Audit
```
User Changes → GoSCIM → Automated:
├── Audit logging
├── Compliance reports
├── Access reviews
└── Webhook notifications
```

## 🌟 Why Open Source?

We believe identity management should be **accessible**, **transparent**, and **community-driven**. By open-sourcing GoSCIM, we're empowering developers worldwide to build better identity solutions.

**Join us in democratizing identity management!** 🚀

## 📄 License

GoSCIM is released under the [MIT License](LICENSE). Feel free to use it in your projects, contribute back, and help us make identity management better for everyone!

**Commercial use** is welcomed, but we'd appreciate:
- 🌟 A star on GitHub
- 📢 Attribution in your project
- 🤝 Contributing improvements back to the community

## 🙏 Acknowledgments

- Built with ❤️ by [Arturo Anton](https://github.com/arturoeanton) and the community
- Inspired by the SCIM protocol and the need for simple, scalable identity management
- Special thanks to all contributors and early adopters!

---

<div align="center">

**[⭐ Star us on GitHub](https://github.com/arturoeanton/goscim)** • **[🐛 Report Issues](https://github.com/arturoeanton/goscim/issues)** • **[💬 Join Discussions](https://github.com/arturoeanton/goscim/discussions)**

Made with ❤️ for the developer community

</div>
