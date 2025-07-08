# 安装指南

## 快速安装

### 选项 1：Docker（推荐）

最快的 GoSCIM 运行方式：

```bash
# 克隆仓库
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# 使用 Docker Compose 启动
docker-compose up -d

# 验证安装
curl http://localhost:8080/ServiceProviderConfig
```

### 选项 2：从源代码构建

```bash
# 先决条件：Go 1.16+
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# 构建并运行
go build -o goscim main.go
./goscim
```

## 系统要求

### 最低要求
- **CPU**: 1 核心
- **内存**: 512MB RAM
- **磁盘**: 100MB 可用空间
- **操作系统**: Linux, macOS, Windows

### 推荐配置
- **CPU**: 2+ 核心
- **内存**: 2GB+ RAM  
- **磁盘**: 1GB+ 可用空间（用于数据）
- **网络**: 1Gbps 以太网

## 环境变量

### 必需的
```bash
export SCIM_ADMIN_USER="Administrator"     # Couchbase 管理员用户
export SCIM_ADMIN_PASSWORD="admin123"     # Couchbase 管理员密码
```

### 可选的
```bash
export SCIM_COUCHBASE_URL="localhost"     # Couchbase 服务器 URL
export SCIM_PORT=":8080"                  # 服务器端口
export SCIM_LOG_LEVEL="info"              # 日志级别
```

## 数据库设置

### Couchbase 安装

```bash
# Docker 中运行 Couchbase
docker run -d --name couchbase-scim \
  -p 8091-8094:8091-8094 \
  -p 11210:11210 \
  couchbase:latest

# 配置集群（通过 Web UI）
# 访问: http://localhost:8091/ui/index.html
```

### 生产部署

```bash
# 使用环境变量配置
export SCIM_ADMIN_USER="your-admin-user"
export SCIM_ADMIN_PASSWORD="your-secure-password"
export SCIM_COUCHBASE_URL="your-couchbase-server.com"

# 启动 GoSCIM
./goscim
```

## 验证安装

```bash
# 检查服务器状态
curl http://localhost:8080/ServiceProviderConfig

# 检查资源类型
curl http://localhost:8080/ResourceTypes

# 检查模式
curl http://localhost:8080/Schemas
```

## 故障排除

### 常见问题

1. **连接失败**
   ```bash
   # 检查 Couchbase 是否运行
   telnet localhost 8091
   ```

2. **认证错误**
   ```bash
   # 验证环境变量
   echo $SCIM_ADMIN_USER
   echo $SCIM_ADMIN_PASSWORD
   ```

3. **端口冲突**
   ```bash
   # 更改端口
   export SCIM_PORT=":8081"
   ```