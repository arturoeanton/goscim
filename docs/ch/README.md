# GoSCIM 技术文档

## 项目概述

**GoSCIM** 是一个用Go语言构建的完整SCIM 2.0（跨域身份管理系统）实现。它为分布式环境中的身份管理提供了强健且可扩展的解决方案，专门设计用于集成异构身份系统。

## 主要特性

### SCIM 2.0 合规性
- ✅ 完整的CRUD操作（创建、读取、更新、删除）
- ✅ 带有SCIM过滤器的高级搜索
- ✅ 分页和排序
- ✅ 可扩展和可自定义的模式
- ✅ 多资源类型支持
- ✅ 批量操作（开发中）

### 技术架构
- **编程语言**: Go 1.16+
- **Web框架**: Gin（高性能）
- **数据库**: Couchbase（分布式NoSQL）
- **解析器**: ANTLR v4用于SCIM过滤器
- **数据格式**: 原生JSON

## 系统架构

### 核心组件

#### 1. 服务器核心 (`main.go`)
```go
// 服务器初始化
func main() {
    scim.InitDB()                    // Couchbase连接
    r := gin.Default()               // HTTP路由器
    scim.ReadResourceType(config, r) // 动态模式加载
    r.Run(port)                      // HTTP服务器
}
```

#### 2. 配置管理 (`scim/config.go`)
- **动态模式加载** 从JSON文件
- **基于资源类型的自动端点注册**
- **服务器启动时的模式验证**

#### 3. 数据库集成 (`scim/couchbase.go`)
- **安全连接** 带身份验证
- **每个资源类型的自动存储桶创建**
- **可自定义的存储桶配置**
- **自动主索引**

#### 4. 过滤器解析器 (`scim/parser/`)
- **SCIM过滤器的ANTLR语法**
- **自动转换** 为N1QL查询
- **完全支持** SCIM操作符

### SCIM 操作

#### 创建 (POST)
```http
POST /scim/v2/Users
Content-Type: application/json

{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "userName": "user@example.com",
  "name": {
    "familyName": "姓",
    "givenName": "名"
  }
}
```

#### 读取 (GET)
```http
GET /scim/v2/Users/12345
```

#### 更新 (PATCH)
```http
PATCH /scim/v2/Users/12345
Content-Type: application/json

{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
  "Operations": [
    {
      "op": "replace",
      "path": "name.familyName",
      "value": "新姓"
    }
  ]
}
```

#### 搜索 (GET)
```http
GET /scim/v2/Users?filter=userName sw "admin"&sortBy=userName&sortOrder=ascending&startIndex=1&count=10
```

#### 删除 (DELETE)
```http
DELETE /scim/v2/Users/12345
```

## 系统配置

### 环境变量

#### 必需的
```bash
export SCIM_ADMIN_USER="Administrator"     # Couchbase管理员用户
export SCIM_ADMIN_PASSWORD="admin123"     # Couchbase管理员密码
```

#### 可选的
```bash
export SCIM_COUCHBASE_URL="localhost"     # Couchbase服务器URL
export SCIM_PORT=":8080"                  # SCIM服务器端口
```

### 配置结构

```
config/
├── schemas/                    # SCIM模式定义
│   ├── urn+ietf+params+scim+schemas+core+2.0+User.json
│   ├── urn+ietf+params+scim+schemas+core+2.0+Group.json
│   └── urn+ietf+params+scim+schemas+extension+*.json
├── resourceType/              # 资源类型
│   ├── User.json
│   ├── Group.json
│   └── Element.json
├── bucketSettings/            # Couchbase存储桶配置
│   ├── User.json
│   ├── Group.json
│   └── Element.json
└── serviceProviderConfig/     # 提供者配置
    └── sp_config.json
```

## 模式和扩展

### 基础用户模式
```json
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User",
  "name": "User",
  "description": "用户账户",
  "attributes": [
    {
      "name": "userName",
      "type": "string",
      "required": true,
      "uniqueness": "server"
    },
    {
      "name": "name",
      "type": "complex",
      "subAttributes": [
        {
          "name": "familyName",
          "type": "string"
        },
        {
          "name": "givenName", 
          "type": "string"
        }
      ]
    }
  ]
}
```

### 自定义扩展
```json
{
  "id": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
  "name": "企业用户",
  "attributes": [
    {
      "name": "employeeNumber",
      "type": "string",
      "uniqueness": "server"
    },
    {
      "name": "department",
      "type": "string"
    }
  ]
}
```

## 访问控制

### 角色和权限
```json
{
  "name": "sensitiveAttribute",
  "type": "string",
  "$reader": ["admin", "hr"],      # 可读取的角色
  "$writer": ["admin"],            # 可写入的角色
  "returned": "default"
}
```

### 角色验证
```go
// 搜索时的自动验证
roles := []string{"user", "admin", "superadmin"}
element := ValidateReadRole(roles, resourceType, item)
```

## SCIM 过滤器

### 支持的语法
```
# 基本比较
userName eq "admin"
name.familyName co "张"
userName sw "admin"
active pr

# 时间比较
meta.lastModified gt "2023-01-01T00:00:00Z"
meta.lastModified ge "2023-01-01T00:00:00Z"

# 逻辑操作符
title pr and userType eq "Employee"
title pr or userType eq "Intern"
userType eq "Employee" and (emails co "company.com" or emails co "company.org")
```

### N1QL 转换
```go
// 转换示例
query, _ := parser.FilterToN1QL("User", 
    "userName eq \"admin\" and active eq true")
// 结果: SELECT * FROM `User` WHERE `userName` = "admin" AND `active` = true
```

## 安装和部署

### 系统要求

#### 开发环境
- Go 1.16或更高版本
- Couchbase Server 6.0+
- ANTLR 4.7（用于解析器重新生成）

#### 生产环境
- CPU: 2+核心
- 内存: 4GB+（取决于数据量）
- 存储: 推荐SSD
- 网络: 1Gbps+（高并发）

### 本地安装

#### 1. 克隆仓库
```bash
git clone https://github.com/arturoeanton/goscim.git
cd goscim
```

#### 2. 安装依赖
```bash
go mod download
```

#### 3. 配置Couchbase
```bash
# 在Docker中运行Couchbase
docker run -d --name couchbase-scim \
  -p 8091-8094:8091-8094 \
  -p 11210:11210 \
  couchbase:latest

# 通过Web界面配置集群
# http://localhost:8091/ui/index.html
```

#### 4. 配置环境变量
```bash
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"
export SCIM_COUCHBASE_URL="localhost"
export SCIM_PORT=":8080"
```

#### 5. 运行服务器
```bash
go run main.go
```

## 监控和操作

### 系统日志
```bash
# 配置结构化日志
export SCIM_LOG_LEVEL=debug
export SCIM_LOG_FORMAT=json

# 日志示例
{"level":"info","timestamp":"2023-12-01T10:00:00Z","message":"GoScim v0.1 已启动"}
{"level":"debug","timestamp":"2023-12-01T10:00:01Z","message":"存储桶已创建","bucket":"User"}
```

### 推荐指标
- 每秒请求数（RPS）
- 响应时间百分位数
- 每个端点的错误率
- 活跃Couchbase连接数
- 内存和CPU使用率

### 健康检查
```http
GET /health
{
  "status": "healthy",
  "database": "connected",
  "version": "v0.1"
}
```

## 安全性

### 安全考虑

#### 身份验证
- 实施OAuth 2.0 / OpenID Connect
- JWT令牌支持
- 每个请求的令牌验证

#### 授权
- 细粒度基于角色的控制
- 资源和操作权限
- 访问审计日志

#### 通信
- 生产环境中强制TLS 1.3
- 有效证书
- HTTP安全标头

## 故障排除

### 常见问题

#### Couchbase连接
```bash
# 验证连接
telnet localhost 8091

# 验证凭据
curl -u Administrator:admin123 http://localhost:8091/pools
```

#### 模式错误
```bash
# 验证模式JSON
jq . config/schemas/schema.json

# 验证语法
go run main.go --validate-config
```

## 社区和支持

技术支持、错误报告或功能请求：
- **问题**: [GitHub Issues](https://github.com/arturoeanton/goscim/issues)
- **讨论**: [GitHub Discussions](https://github.com/arturoeanton/goscim/discussions)
- **文档**: [项目Wiki](https://github.com/arturoeanton/goscim/wiki)
- **示例**: `httpexamples/` 目录