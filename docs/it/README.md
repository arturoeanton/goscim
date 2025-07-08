# GoSCIM Documentazione Tecnica

## Panoramica del Progetto

**GoSCIM** è un'implementazione completa di SCIM 2.0 (System for Cross-domain Identity Management) costruita in Go. Fornisce una soluzione robusta e scalabile per la gestione delle identità in ambienti distribuiti, specificamente progettata per integrare sistemi di identità eterogenei.

## Caratteristiche Principali

### Conformità SCIM 2.0
- ✅ Operazioni CRUD complete (Create, Read, Update, Delete)
- ✅ Ricerca avanzata con filtri SCIM
- ✅ Paginazione e ordinamento
- ✅ Schemi estensibili e personalizzabili
- ✅ Supporto per tipi di risorse multiple
- ✅ Operazioni in blocco (in sviluppo)

### Architettura Tecnica
- **Linguaggio**: Go 1.16+
- **Framework Web**: Gin (alta performance)
- **Database**: Couchbase (NoSQL distribuito)
- **Parser**: ANTLR v4 per filtri SCIM
- **Formato Dati**: JSON nativo

## Architettura del Sistema

### Componenti Principali

#### 1. Nucleo del Server (`main.go`)
```go
// Inizializzazione del server
func main() {
    scim.InitDB()                    // Connessione Couchbase
    r := gin.Default()               // Router HTTP
    scim.ReadResourceType(config, r) // Caricamento dinamico degli schemi
    r.Run(port)                      // Server HTTP
}
```

#### 2. Gestione della Configurazione (`scim/config.go`)
- **Caricamento dinamico degli schemi** da file JSON
- **Registrazione automatica degli endpoint** basata sui tipi di risorsa
- **Validazione degli schemi** all'avvio del server

#### 3. Integrazione Database (`scim/couchbase.go`)
- **Connessione sicura** con autenticazione
- **Creazione automatica dei bucket** per tipo di risorsa
- **Configurazione personalizzabile dei bucket**
- **Indici primari automatici**

#### 4. Parser dei Filtri (`scim/parser/`)
- **Grammatica ANTLR** per filtri SCIM
- **Conversione automatica** a query N1QL
- **Supporto completo** per operatori SCIM

### Operazioni SCIM

#### Creare (POST)
```http
POST /scim/v2/Users
Content-Type: application/json

{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "userName": "user@example.com",
  "name": {
    "familyName": "Cognome",
    "givenName": "Nome"
  }
}
```

#### Leggere (GET)
```http
GET /scim/v2/Users/12345
```

#### Aggiornare (PATCH)
```http
PATCH /scim/v2/Users/12345
Content-Type: application/json

{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
  "Operations": [
    {
      "op": "replace",
      "path": "name.familyName",
      "value": "NuovoCognome"
    }
  ]
}
```

#### Cercare (GET)
```http
GET /scim/v2/Users?filter=userName sw "admin"&sortBy=userName&sortOrder=ascending&startIndex=1&count=10
```

#### Eliminare (DELETE)
```http
DELETE /scim/v2/Users/12345
```

## Configurazione del Sistema

### Variabili d'Ambiente

#### Obbligatorie
```bash
export SCIM_ADMIN_USER="Administrator"     # Utente admin Couchbase
export SCIM_ADMIN_PASSWORD="admin123"     # Password admin Couchbase
```

#### Opzionali
```bash
export SCIM_COUCHBASE_URL="localhost"     # URL server Couchbase
export SCIM_PORT=":8080"                  # Porta server SCIM
```

### Struttura di Configurazione

```
config/
├── schemas/                    # Definizioni schemi SCIM
│   ├── urn+ietf+params+scim+schemas+core+2.0+User.json
│   ├── urn+ietf+params+scim+schemas+core+2.0+Group.json
│   └── urn+ietf+params+scim+schemas+extension+*.json
├── resourceType/              # Tipi di risorse
│   ├── User.json
│   ├── Group.json
│   └── Element.json
├── bucketSettings/            # Configurazione bucket Couchbase
│   ├── User.json
│   ├── Group.json
│   └── Element.json
└── serviceProviderConfig/     # Configurazione provider
    └── sp_config.json
```

## Schemi ed Estensioni

### Schema Utente Base
```json
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User",
  "name": "User",
  "description": "Account utente",
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

### Estensioni Personalizzate
```json
{
  "id": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
  "name": "Utente Aziendale",
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

## Controllo degli Accessi

### Ruoli e Permessi
```json
{
  "name": "sensitiveAttribute",
  "type": "string",
  "$reader": ["admin", "hr"],      # Ruoli che possono leggere
  "$writer": ["admin"],            # Ruoli che possono scrivere
  "returned": "default"
}
```

### Validazione dei Ruoli
```go
// Validazione automatica nelle ricerche
roles := []string{"user", "admin", "superadmin"}
element := ValidateReadRole(roles, resourceType, item)
```

## Filtri SCIM

### Sintassi Supportata
```
# Confronti di base
userName eq "admin"
name.familyName co "Rossi"
userName sw "admin"
active pr

# Confronti temporali
meta.lastModified gt "2023-01-01T00:00:00Z"
meta.lastModified ge "2023-01-01T00:00:00Z"

# Operatori logici
title pr and userType eq "Employee"
title pr or userType eq "Intern"
userType eq "Employee" and (emails co "company.com" or emails co "company.org")
```

### Conversione N1QL
```go
// Esempio di conversione
query, _ := parser.FilterToN1QL("User", 
    "userName eq \"admin\" and active eq true")
// Risultato: SELECT * FROM `User` WHERE `userName` = "admin" AND `active` = true
```

## Installazione e Distribuzione

### Requisiti di Sistema

#### Ambiente di Sviluppo
- Go 1.16 o superiore
- Couchbase Server 6.0+
- ANTLR 4.7 (per rigenerazione parser)

#### Ambiente di Produzione
- CPU: 2+ core
- RAM: 4GB+ (a seconda del volume)
- Storage: SSD raccomandato
- Rete: 1Gbps+ per alta concorrenza

### Installazione Locale

#### 1. Clonare il Repository
```bash
git clone https://github.com/arturoeanton/goscim.git
cd goscim
```

#### 2. Installare le Dipendenze
```bash
go mod download
```

#### 3. Configurare Couchbase
```bash
# Eseguire Couchbase in Docker
docker run -d --name couchbase-scim \
  -p 8091-8094:8091-8094 \
  -p 11210:11210 \
  couchbase:latest

# Configurare cluster tramite interfaccia web
# http://localhost:8091/ui/index.html
```

#### 4. Configurare Variabili d'Ambiente
```bash
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"
export SCIM_COUCHBASE_URL="localhost"
export SCIM_PORT=":8080"
```

#### 5. Avviare il Server
```bash
go run main.go
```

### Distribuzione in Produzione

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

## Test e Sviluppo

### Eseguire i Test
```bash
# Test unitari
go test ./...

# Test specifici
go test ./scim/parser -v

# Test con copertura
go test -cover ./...
```

### Esempi di Utilizzo
```bash
# Creare utente
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "testuser",
    "name": {
      "familyName": "Cognome",
      "givenName": "Nome"
    }
  }'

# Cercare utenti
curl "http://localhost:8080/scim/v2/Users?filter=userName sw \"test\""

# Ottenere configurazione provider
curl http://localhost:8080/ServiceProviderConfig
```

## Monitoraggio e Operazioni

### Log di Sistema
```bash
# Configurare logging strutturato
export SCIM_LOG_LEVEL=debug
export SCIM_LOG_FORMAT=json

# Esempio di log
{"level":"info","timestamp":"2023-12-01T10:00:00Z","message":"GoScim v0.1 avviato"}
{"level":"debug","timestamp":"2023-12-01T10:00:01Z","message":"Bucket creato","bucket":"User"}
```

### Metriche Raccomandate
- Richieste per secondo (RPS)
- Percentili dei tempi di risposta
- Tassi di errore per endpoint
- Connessioni Couchbase attive
- Utilizzo memoria e CPU

### Controlli di Salute
```http
GET /health
{
  "status": "healthy",
  "database": "connected",
  "version": "v0.1"
}
```

## Sicurezza

### Considerazioni di Sicurezza

#### Autenticazione
- Implementare OAuth 2.0 / OpenID Connect
- Supporto token JWT
- Validazione token ad ogni richiesta

#### Autorizzazione
- Controllo granulare basato sui ruoli
- Permessi su risorse e operazioni
- Log di audit degli accessi

#### Comunicazione
- TLS 1.3 obbligatorio in produzione
- Certificati validi
- Header di sicurezza HTTP

#### Validazione
- Sanitizzazione input
- Validazione rigorosa degli schemi
- Rate limiting per IP/utente

## Integrazione Sistemi Esterni

### Provider di Identità
- Active Directory
- LDAP
- Provider OAuth 2.0
- SAML 2.0

### Sistemi Target
- Applicazioni SaaS
- Database utenti
- Sistemi di directory
- API di terze parti

## Risoluzione Problemi

### Problemi Comuni

#### Connessione Couchbase
```bash
# Verificare connettività
telnet localhost 8091

# Verificare credenziali
curl -u Administrator:admin123 http://localhost:8091/pools
```

#### Errori di Schema
```bash
# Validare JSON dello schema
jq . config/schemas/schema.json

# Verificare sintassi
go run main.go --validate-config
```

#### Problemi di Performance
```bash
# Verificare indici
curl -u admin:pass http://localhost:8091/query/service \
  -d 'statement=SELECT * FROM system:indexes WHERE keyspace_id="User"'
```

## Roadmap di Sviluppo

### Fase 1: Stabilizzazione
- Implementare autenticazione robusta
- Suite di test completa
- Migliorare logging e monitoraggio

### Fase 2: Scalabilità
- Supporto clustering
- Cache distribuita
- Ottimizzazioni performance

### Fase 3: Funzionalità Avanzate
- Operazioni Bulk complete
- Webhook e notifiche
- Dashboard di amministrazione

## Contribuire

### Aggiungere Nuove Risorse
1. Creare schema JSON in `config/schemas/`
2. Definire tipo di risorsa in `config/resourceType/`
3. Configurare bucket in `config/bucketSettings/`
4. Riavviare server per caricare modifiche

### Rigenerazione Parser
```bash
# Installare ANTLR
wget http://www.antlr.org/download/antlr-4.7-complete.jar
alias antlr='java -jar $PWD/antlr-4.7-complete.jar'

# Rigenerare parser
antlr -Dlanguage=Go -o scim/parser ScimFilter.g4
```

## Comunità e Supporto

Per supporto tecnico, segnalazioni di bug o richieste di funzionalità:
- **Issues**: [GitHub Issues](https://github.com/arturoeanton/goscim/issues)
- **Discussioni**: [GitHub Discussions](https://github.com/arturoeanton/goscim/discussions)
- **Documentazione**: [Wiki del Progetto](https://github.com/arturoeanton/goscim/wiki)
- **Esempi**: Directory `httpexamples/`