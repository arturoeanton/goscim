# GoSCIM 快速入门

欢迎使用 GoSCIM！本指南将帮助您在几分钟内建立并运行自己的 SCIM 2.0 服务器。

## 什么是 GoSCIM？

GoSCIM 是一个轻量级、快速且灵活的 SCIM 2.0 协议实现，用 Go 语言构建。它设计用于：

- 🚀 **简化身份管理** 跨多个系统
- 🔧 **轻松集成** 现有基础设施
- 📈 **扩展** 从小型初创公司到大型企业
- 🛡️ **保护** 您的身份数据，采用行业最佳实践

## 快速开始（2分钟）

### 选项 1：Docker（推荐）

试用 GoSCIM 的最快方式：

```bash
# 克隆仓库
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# 使用 Docker Compose 启动
docker-compose up -d

# 等待服务启动（约30秒）
sleep 30

# 测试您的 SCIM 服务器
curl http://localhost:8080/ServiceProviderConfig
```

### 选项 2：从源代码构建

如果您更喜欢从源代码构建：

```bash
# 先决条件：Go 1.16+ 和 Couchbase
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# 设置环境变量
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"

# 运行服务器
go run main.go
```

## 您的第一个 SCIM 操作

一旦您的服务器运行，尝试这些基本操作：

### 1. 检查服务器配置
```bash
curl http://localhost:8080/ServiceProviderConfig
```

### 2. 创建您的第一个用户
```bash
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "jane.doe@example.com",
    "name": {
      "familyName": "张",
      "givenName": "三"
    },
    "emails": [{
      "value": "jane.doe@example.com",
      "type": "work",
      "primary": true
    }],
    "active": true
  }'
```

### 3. 搜索用户
```bash
curl "http://localhost:8080/scim/v2/Users?filter=userName sw 'jane'"
```

### 4. 列出可用资源
```bash
curl http://localhost:8080/ResourceTypes
```

## 理解响应

当您创建用户时，您会得到这样的响应：

```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "userName": "jane.doe@example.com",
  "name": {
    "familyName": "张",
    "givenName": "三",
    "formatted": "张三"
  },
  "emails": [{
    "value": "jane.doe@example.com",
    "type": "work",
    "primary": true
  }],
  "active": true,
  "meta": {
    "resourceType": "User",
    "created": "2023-12-01T10:30:00Z",
    "lastModified": "2023-12-01T10:30:00Z",
    "version": "W/\"a330bc54f0671c9\"",
    "location": "/scim/v2/Users/550e8400-e29b-41d4-a716-446655440000"
  }
}
```

关键元素：
- **`id`**: 服务器生成的唯一标识符
- **`meta`**: 包括创建时间和位置的元数据
- **`schemas`**: 此资源使用的 SCIM 模式

## 常见用例

### 1. 员工入职
```bash
# 创建新员工
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": [
      "urn:ietf:params:scim:schemas:core:2.0:User",
      "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
    ],
    "userName": "john.smith@company.com",
    "name": {
      "familyName": "史密斯",
      "givenName": "约翰"
    },
    "emails": [{
      "value": "john.smith@company.com",
      "type": "work"
    }],
    "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
      "employeeNumber": "EMP001",
      "department": "工程部",
      "manager": {
        "value": "manager-id-here"
      }
    }
  }'
```

### 2. 高级过滤
```bash
# 查找工程部的活跃用户
curl "http://localhost:8080/scim/v2/Users?filter=active eq true and urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department eq '工程部'"

# 查找有公司邮箱的用户
curl "http://localhost:8080/scim/v2/Users?filter=emails[type eq 'work' and value ew '@company.com']"

# 查找最近修改的用户
curl "http://localhost:8080/scim/v2/Users?filter=meta.lastModified gt '2023-01-01T00:00:00Z'"
```

## 错误处理

GoSCIM 返回标准 HTTP 状态码和 SCIM 错误响应：

```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"],
  "detail": "属性 'userName' 是必需的",
  "status": "400",
  "scimType": "invalidValue"
}
```

常见错误：
- **400 Bad Request**: 无效数据或缺少必需字段
- **404 Not Found**: 资源不存在
- **409 Conflict**: 资源已存在（如重复的 userName）
- **500 Internal Server Error**: 服务器端问题

## 配置基础

### 环境变量
```bash
# 数据库连接
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="your-secure-password"
export SCIM_COUCHBASE_URL="couchbase-server.example.com"

# 服务器设置
export SCIM_PORT=":8080"
export SCIM_LOG_LEVEL="info"

# 安全性（生产环境）
export SCIM_TLS_ENABLED="true"
export SCIM_TLS_CERT_FILE="/path/to/cert.pem"
export SCIM_TLS_KEY_FILE="/path/to/key.pem"
```

## 下一步

现在您已经运行了 GoSCIM，以下是一些后续步骤：

1. **🔐 [设置身份验证](security.md)** - 添加 OAuth 2.0 或 JWT 身份验证
2. **📊 [配置监控](operations.md)** - 设置指标和日志记录
3. **🔌 [添加集成](integrations.md)** - 连接到 Active Directory、LDAP 或 SaaS 应用
4. **🏗️ [学习架构](architecture.md)** - 了解 GoSCIM 内部工作原理
5. **👩‍💻 [贡献](development.md)** - 帮助改进 GoSCIM

## 获取帮助

- 📚 **文档**: 查看[完整文档](README.md)
- 🐛 **问题**: 在 [GitHub Issues](https://github.com/arturoeanton/goscim/issues) 报告错误
- 💬 **讨论**: 在 [GitHub Discussions](https://github.com/arturoeanton/goscim/discussions) 提问
- 📖 **示例**: 浏览 `httpexamples/` 目录查看更多使用示例

欢迎加入 GoSCIM 社区！🎉