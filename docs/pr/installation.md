# Guia de Instalação

## Instalação Rápida

### Opção 1: Docker (Recomendado)

A maneira mais rápida de colocar o GoSCIM em funcionamento:

```bash
# Clonar o repositório
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Iniciar com Docker Compose
docker-compose up -d

# Verificar instalação
curl http://localhost:8080/ServiceProviderConfig
```

### Opção 2: Construir a partir do código fonte

```bash
# Pré-requisitos: Go 1.16+
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Construir e executar
go build -o goscim main.go
./goscim
```

## Requisitos do Sistema

### Requisitos Mínimos
- **CPU**: 1 núcleo
- **Memória**: 512MB RAM
- **Disco**: 100MB de espaço disponível
- **SO**: Linux, macOS, Windows

### Configuração Recomendada
- **CPU**: 2+ núcleos
- **Memória**: 2GB+ RAM  
- **Disco**: 1GB+ de espaço disponível (para dados)
- **Rede**: Ethernet 1Gbps

## Variáveis de Ambiente

### Obrigatórias
```bash
export SCIM_ADMIN_USER="Administrator"     # Usuário admin Couchbase
export SCIM_ADMIN_PASSWORD="admin123"     # Senha admin Couchbase
```

### Opcionais
```bash
export SCIM_COUCHBASE_URL="localhost"     # URL do servidor Couchbase
export SCIM_PORT=":8080"                  # Porta do servidor
export SCIM_LOG_LEVEL="info"              # Nível de log
```

## Configuração do Banco de Dados

### Instalação do Couchbase

```bash
# Executar Couchbase no Docker
docker run -d --name couchbase-scim \
  -p 8091-8094:8091-8094 \
  -p 11210:11210 \
  couchbase:latest

# Configurar cluster (via interface web)
# Acesso: http://localhost:8091/ui/index.html
```

### Implantação em Produção

```bash
# Configurar com variáveis de ambiente
export SCIM_ADMIN_USER="your-admin-user"
export SCIM_ADMIN_PASSWORD="your-secure-password"
export SCIM_COUCHBASE_URL="your-couchbase-server.com"

# Iniciar GoSCIM
./goscim
```

## Verificar Instalação

```bash
# Verificar status do servidor
curl http://localhost:8080/ServiceProviderConfig

# Verificar tipos de recursos
curl http://localhost:8080/ResourceTypes

# Verificar esquemas
curl http://localhost:8080/Schemas
```

## Solução de Problemas

### Problemas Comuns

1. **Falha de conexão**
   ```bash
   # Verificar se Couchbase está rodando
   telnet localhost 8091
   ```

2. **Erro de autenticação**
   ```bash
   # Verificar variáveis de ambiente
   echo $SCIM_ADMIN_USER
   echo $SCIM_ADMIN_PASSWORD
   ```

3. **Conflito de porta**
   ```bash
   # Mudar porta
   export SCIM_PORT=":8081"
   ```