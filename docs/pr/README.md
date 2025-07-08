# GoSCIM Documentação Técnica

## Visão Geral do Projeto

**GoSCIM** é uma implementação completa de SCIM 2.0 (System for Cross-domain Identity Management) construída em Go. Ela fornece uma solução robusta e escalável para gerenciamento de identidade em ambientes distribuídos, especialmente projetada para integrar sistemas de identidade heterogêneos.

## Principais Funcionalidades

### Conformidade SCIM 2.0
- ✅ Operações CRUD completas (Create, Read, Update, Delete)
- ✅ Busca avançada com filtros SCIM
- ✅ Paginação e ordenação
- ✅ Esquemas extensíveis e personalizáveis
- ✅ Suporte a múltiplos tipos de recursos
- ✅ Operações em lote (em desenvolvimento)

### Arquitetura Técnica
- **Linguagem**: Go 1.16+
- **Framework Web**: Gin (alta performance)
- **Banco de Dados**: Couchbase (NoSQL distribuído)
- **Parser**: ANTLR v4 para filtros SCIM
- **Formato de Dados**: JSON nativo

## Arquitetura do Sistema

### Componentes Principais

#### 1. Núcleo do Servidor (`main.go`)
```go
// Inicialização do servidor
func main() {
    scim.InitDB()                    // Conexão Couchbase
    r := gin.Default()               // Roteador HTTP
    scim.ReadResourceType(config, r) // Carregamento dinâmico de esquemas
    r.Run(port)                      // Servidor HTTP
}
```

#### 2. Gerenciamento de Configuração (`scim/config.go`)
- **Carregamento dinâmico de esquemas** a partir de arquivos JSON
- **Registro automático de endpoints** baseado em tipos de recursos
- **Validação de esquemas** na inicialização do servidor

#### 3. Integração com Banco de Dados (`scim/couchbase.go`)
- **Conexão segura** com autenticação
- **Criação automática de buckets** por tipo de recurso
- **Configuração personalizável de buckets**
- **Índices primários automáticos**

#### 4. Parser de Filtros (`scim/parser/`)
- **Gramática ANTLR** para filtros SCIM
- **Conversão automática** para consultas N1QL
- **Suporte completo** para operadores SCIM

### Operações SCIM

#### Criar (POST)
```http
POST /scim/v2/Users
Content-Type: application/json

{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "userName": "user@example.com",
  "name": {
    "familyName": "Sobrenome",
    "givenName": "Nome"
  }
}
```

#### Ler (GET)
```http
GET /scim/v2/Users/12345
```

#### Atualizar (PATCH)
```http
PATCH /scim/v2/Users/12345
Content-Type: application/json

{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
  "Operations": [
    {
      "op": "replace",
      "path": "name.familyName",
      "value": "NovoSobrenome"
    }
  ]
}
```

#### Buscar (GET)
```http
GET /scim/v2/Users?filter=userName sw "admin"&sortBy=userName&sortOrder=ascending&startIndex=1&count=10
```

#### Deletar (DELETE)
```http
DELETE /scim/v2/Users/12345
```

## Configuração do Sistema

### Variáveis de Ambiente

#### Obrigatórias
```bash
export SCIM_ADMIN_USER="Administrator"     # Usuário admin do Couchbase
export SCIM_ADMIN_PASSWORD="admin123"     # Senha admin do Couchbase
```

#### Opcionais
```bash
export SCIM_COUCHBASE_URL="localhost"     # URL do servidor Couchbase
export SCIM_PORT=":8080"                  # Porta do servidor SCIM
```

### Estrutura de Configuração

```
config/
├── schemas/                    # Definições de esquemas SCIM
│   ├── urn+ietf+params+scim+schemas+core+2.0+User.json
│   ├── urn+ietf+params+scim+schemas+core+2.0+Group.json
│   └── urn+ietf+params+scim+schemas+extension+*.json
├── resourceType/              # Tipos de recursos
│   ├── User.json
│   ├── Group.json
│   └── Element.json
├── bucketSettings/            # Configuração de buckets Couchbase
│   ├── User.json
│   ├── Group.json
│   └── Element.json
└── serviceProviderConfig/     # Configuração do provedor
    └── sp_config.json
```

## Esquemas e Extensões

### Esquema Base de Usuário
```json
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User",
  "name": "User",
  "description": "Conta de usuário",
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

### Extensões Personalizadas
```json
{
  "id": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
  "name": "Usuário Empresarial",
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

## Controle de Acesso

### Papéis e Permissões
```json
{
  "name": "sensitiveAttribute",
  "type": "string",
  "$reader": ["admin", "hr"],      # Papéis que podem ler
  "$writer": ["admin"],            # Papéis que podem escrever
  "returned": "default"
}
```

### Validação de Papéis
```go
// Validação automática em buscas
roles := []string{"user", "admin", "superadmin"}
element := ValidateReadRole(roles, resourceType, item)
```

## Filtros SCIM

### Sintaxe Suportada
```
# Comparações básicas
userName eq "admin"
name.familyName co "Silva"
userName sw "admin"
active pr

# Comparações temporais
meta.lastModified gt "2023-01-01T00:00:00Z"
meta.lastModified ge "2023-01-01T00:00:00Z"

# Operadores lógicos
title pr and userType eq "Employee"
title pr or userType eq "Intern"
userType eq "Employee" and (emails co "company.com" or emails co "company.org")
```

### Conversão N1QL
```go
// Exemplo de conversão
query, _ := parser.FilterToN1QL("User", 
    "userName eq \"admin\" and active eq true")
// Resultado: SELECT * FROM `User` WHERE `userName` = "admin" AND `active` = true
```

## Instalação e Implantação

### Requisitos do Sistema

#### Ambiente de Desenvolvimento
- Go 1.16 ou superior
- Couchbase Server 6.0+
- ANTLR 4.7 (para regeneração do parser)

#### Ambiente de Produção
- CPU: 2+ núcleos
- RAM: 4GB+ (dependendo do volume)
- Armazenamento: SSD recomendado
- Rede: 1Gbps+ para alta concorrência

### Instalação Local

#### 1. Clonar o Repositório
```bash
git clone https://github.com/arturoeanton/goscim.git
cd goscim
```

#### 2. Instalar Dependências
```bash
go mod download
```

#### 3. Configurar Couchbase
```bash
# Executar Couchbase no Docker
docker run -d --name couchbase-scim \
  -p 8091-8094:8091-8094 \
  -p 11210:11210 \
  couchbase:latest

# Configurar cluster via interface web
# http://localhost:8091/ui/index.html
```

#### 4. Configurar Variáveis de Ambiente
```bash
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"
export SCIM_COUCHBASE_URL="localhost"
export SCIM_PORT=":8080"
```

#### 5. Executar o Servidor
```bash
go run main.go
```

## Monitoramento e Operações

### Logs do Sistema
```bash
# Configurar logs estruturados
export SCIM_LOG_LEVEL=debug
export SCIM_LOG_FORMAT=json

# Exemplo de log
{"level":"info","timestamp":"2023-12-01T10:00:00Z","message":"GoScim v0.1 iniciado"}
{"level":"debug","timestamp":"2023-12-01T10:00:01Z","message":"Bucket criado","bucket":"User"}
```

### Métricas Recomendadas
- Requisições por segundo (RPS)
- Percentis de tempo de resposta
- Taxa de erro por endpoint
- Conexões ativas do Couchbase
- Uso de memória e CPU

### Verificações de Saúde
```http
GET /health
{
  "status": "healthy",
  "database": "connected",
  "version": "v0.1"
}
```

## Segurança

### Considerações de Segurança

#### Autenticação
- Implementar OAuth 2.0 / OpenID Connect
- Suporte a tokens JWT
- Validação de tokens em cada requisição

#### Autorização
- Controle granular baseado em papéis
- Permissões de recursos e operações
- Logs de auditoria de acesso

#### Comunicação
- TLS 1.3 obrigatório em produção
- Certificados válidos
- Headers de segurança HTTP

## Solução de Problemas

### Problemas Comuns

#### Conexão Couchbase
```bash
# Verificar conectividade
telnet localhost 8091

# Verificar credenciais
curl -u Administrator:admin123 http://localhost:8091/pools
```

#### Erros de Esquema
```bash
# Validar JSON do esquema
jq . config/schemas/schema.json

# Verificar sintaxe
go run main.go --validate-config
```

### Implantação em Produção

#### Docker Compose
```yaml
version: '3.8'
services:
  couchbase:
    image: couchbase:latest
    ports:
      - "8091-8094:8091-8094"
      - "11210:11210"
    environment:
      - CLUSTER_NAME=scim-cluster
    volumes:
      - couchbase-data:/opt/couchbase/var

  goscim:
    build: .
    ports:
      - "8080:8080"
    environment:
      - SCIM_ADMIN_USER=Administrator
      - SCIM_ADMIN_PASSWORD=admin123
      - SCIM_COUCHBASE_URL=couchbase
    depends_on:
      - couchbase
    restart: unless-stopped

volumes:
  couchbase-data:
```

#### Kubernetes
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: goscim
spec:
  replicas: 3
  selector:
    matchLabels:
      app: goscim
  template:
    metadata:
      labels:
        app: goscim
    spec:
      containers:
      - name: goscim
        image: goscim:latest
        ports:
        - containerPort: 8080
        env:
        - name: SCIM_COUCHBASE_URL
          value: "couchbase-service"
        - name: SCIM_ADMIN_USER
          valueFrom:
            secretKeyRef:
              name: couchbase-secret
              key: username
        - name: SCIM_ADMIN_PASSWORD
          valueFrom:
            secretKeyRef:
              name: couchbase-secret
              key: password
```

## Testes e Desenvolvimento

### Executar Testes
```bash
# Testes unitários
go test ./...

# Testes específicos
go test ./scim/parser -v

# Testes com cobertura
go test -cover ./...
```

### Exemplos de Uso
```bash
# Criar usuário
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "testuser",
    "name": {
      "familyName": "Sobrenome",
      "givenName": "Nome"
    }
  }'

# Buscar usuários
curl "http://localhost:8080/scim/v2/Users?filter=userName sw \"test\""

# Obter configuração do provedor
curl http://localhost:8080/ServiceProviderConfig
```

#### Problemas de Performance
```bash
# Verificar índices
curl -u admin:pass http://localhost:8091/query/service \
  -d 'statement=SELECT * FROM system:indexes WHERE keyspace_id="User"'
```

## Roteiro de Desenvolvimento

### Fase 1: Estabilização
- Implementar autenticação robusta
- Suíte completa de testes
- Melhorar logging e monitoramento

### Fase 2: Escalabilidade
- Suporte a cluster
- Cache distribuído
- Otimizações de performance

### Fase 3: Funcionalidades Avançadas
- Operações em lote completas
- Webhooks e notificações
- Dashboard de administração

## Integração de Sistemas Externos

### Provedores de Identidade
- Active Directory
- LDAP
- Provedores OAuth 2.0
- SAML 2.0

### Sistemas de Destino
- Aplicações SaaS
- Bancos de dados de usuários
- Sistemas de diretório
- APIs de terceiros

## Contribuições

### Adicionar Novos Recursos
1. Criar esquema JSON em `config/schemas/`
2. Definir tipo de recurso em `config/resourceType/`
3. Configurar bucket em `config/bucketSettings/`
4. Reiniciar servidor para carregar mudanças

### Regeneração do Parser
```bash
# Instalar ANTLR
wget http://www.antlr.org/download/antlr-4.7-complete.jar
alias antlr='java -jar $PWD/antlr-4.7-complete.jar'

# Regenerar parser
antlr -Dlanguage=Go -o scim/parser ScimFilter.g4
```

## Comunidade e Suporte

Para suporte técnico, relatórios de bugs ou solicitações de funcionalidades:
- **Issues**: [GitHub Issues](https://github.com/arturoeanton/goscim/issues)
- **Discussões**: [GitHub Discussions](https://github.com/arturoeanton/goscim/discussions)
- **Documentação**: [Wiki do Projeto](https://github.com/arturoeanton/goscim/wiki)
- **Exemplos**: Diretório `httpexamples/`