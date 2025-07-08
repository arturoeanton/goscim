# Começando com GoSCIM

Bem-vindo ao GoSCIM! Este guia irá ajudá-lo a colocar seu próprio servidor SCIM 2.0 em funcionamento em apenas alguns minutos.

## O que é GoSCIM?

GoSCIM é uma implementação leve, rápida e flexível do protocolo SCIM 2.0 construída em Go. É projetado para:

- 🚀 **Simplificar o gerenciamento de identidade** através de múltiplos sistemas
- 🔧 **Integrar facilmente** com infraestrutura existente
- 📈 **Escalar** de pequenas startups a grandes empresas
- 🛡️ **Proteger** seus dados de identidade com as melhores práticas da indústria

## Início Rápido (2 Minutos)

### Opção 1: Docker (Recomendado)

A maneira mais rápida de experimentar o GoSCIM:

```bash
# Clonar o repositório
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Iniciar com Docker Compose
docker-compose up -d

# Aguardar os serviços iniciarem (cerca de 30 segundos)
sleep 30

# Testar seu servidor SCIM
curl http://localhost:8080/ServiceProviderConfig
```

### Opção 2: Construir a partir do código fonte

Se você preferir construir a partir do código fonte:

```bash
# Pré-requisitos: Go 1.16+ e Couchbase
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Definir variáveis de ambiente
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"

# Executar o servidor
go run main.go
```

## Suas Primeiras Operações SCIM

Uma vez que seu servidor esteja rodando, experimente estas operações básicas:

### 1. Verificar configuração do servidor
```bash
curl http://localhost:8080/ServiceProviderConfig
```

### 2. Criar seu primeiro usuário
```bash
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "jane.doe@example.com",
    "name": {
      "familyName": "Silva",
      "givenName": "Maria"
    },
    "emails": [{
      "value": "jane.doe@example.com",
      "type": "work",
      "primary": true
    }],
    "active": true
  }'
```

### 3. Buscar usuários
```bash
curl "http://localhost:8080/scim/v2/Users?filter=userName sw 'jane'"
```

### 4. Listar recursos disponíveis
```bash
curl http://localhost:8080/ResourceTypes
```

## Entendendo a Resposta

Quando você criar um usuário, receberá uma resposta como esta:

```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "userName": "jane.doe@example.com",
  "name": {
    "familyName": "Silva",
    "givenName": "Maria",
    "formatted": "Maria Silva"
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

Elementos chave:
- **`id`**: Identificador único gerado pelo servidor
- **`meta`**: Metadados incluindo tempo de criação e localização
- **`schemas`**: Esquemas SCIM usados para este recurso

## Casos de Uso Comuns

### 1. Integração de Funcionários
```bash
# Criar novo funcionário
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": [
      "urn:ietf:params:scim:schemas:core:2.0:User",
      "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
    ],
    "userName": "john.smith@company.com",
    "name": {
      "familyName": "Santos",
      "givenName": "João"
    },
    "emails": [{
      "value": "john.smith@company.com",
      "type": "work"
    }],
    "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
      "employeeNumber": "EMP001",
      "department": "Engenharia",
      "manager": {
        "value": "manager-id-here"
      }
    }
  }'
```

### 2. Filtragem Avançada
```bash
# Encontrar usuários ativos no departamento de engenharia
curl "http://localhost:8080/scim/v2/Users?filter=active eq true and urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department eq 'Engenharia'"

# Encontrar usuários com email da empresa
curl "http://localhost:8080/scim/v2/Users?filter=emails[type eq 'work' and value ew '@company.com']"

# Encontrar usuários modificados recentemente
curl "http://localhost:8080/scim/v2/Users?filter=meta.lastModified gt '2023-01-01T00:00:00Z'"
```

## Tratamento de Erros

GoSCIM retorna códigos de status HTTP padrão e respostas de erro SCIM:

```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"],
  "detail": "Atributo 'userName' é obrigatório",
  "status": "400",
  "scimType": "invalidValue"
}
```

Erros comuns:
- **400 Bad Request**: Dados inválidos ou campos obrigatórios ausentes
- **404 Not Found**: Recurso não existe
- **409 Conflict**: Recurso já existe (ex: userName duplicado)
- **500 Internal Server Error**: Problemas do lado do servidor

## Configuração Básica

### Variáveis de Ambiente
```bash
# Conexão com banco de dados
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="your-secure-password"
export SCIM_COUCHBASE_URL="couchbase-server.example.com"

# Configurações do servidor
export SCIM_PORT=":8080"
export SCIM_LOG_LEVEL="info"

# Segurança (para produção)
export SCIM_TLS_ENABLED="true"
export SCIM_TLS_CERT_FILE="/path/to/cert.pem"
export SCIM_TLS_KEY_FILE="/path/to/key.pem"
```

## Próximos Passos

Agora que o GoSCIM está rodando, aqui estão alguns próximos passos:

1. **🔐 [Configurar autenticação](security.md)** - Adicionar autenticação OAuth 2.0 ou JWT
2. **📊 [Configurar monitoramento](operations.md)** - Configurar métricas e logging
3. **🔌 [Adicionar integrações](integrations.md)** - Conectar ao Active Directory, LDAP ou apps SaaS
4. **🏗️ [Aprender a arquitetura](architecture.md)** - Entender como o GoSCIM funciona internamente
5. **👩‍💻 [Contribuir](development.md)** - Ajudar a melhorar o GoSCIM

## Obtendo Ajuda

- 📚 **Documentação**: Confira a [documentação completa](README.md)
- 🐛 **Issues**: Reporte bugs no [GitHub Issues](https://github.com/arturoeanton/goscim/issues)
- 💬 **Discussões**: Faça perguntas no [GitHub Discussions](https://github.com/arturoeanton/goscim/discussions)
- 📖 **Exemplos**: Explore o diretório `httpexamples/` para mais exemplos de uso

Bem-vindo à comunidade GoSCIM! 🎉