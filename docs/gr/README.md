# GoSCIM Technische Dokumentation

## Projektübersicht

**GoSCIM** ist eine vollständige SCIM 2.0 (System for Cross-domain Identity Management) Implementierung in Go. Es bietet eine robuste und skalierbare Lösung für Identity Management in verteilten Umgebungen, speziell entwickelt für die Integration heterogener Identitätssysteme.

## Hauptfunktionen

### SCIM 2.0 Konformität
- ✅ Vollständige CRUD-Operationen (Create, Read, Update, Delete)
- ✅ Erweiterte Suche mit SCIM-Filtern
- ✅ Paginierung und Sortierung
- ✅ Erweiterbare und anpassbare Schemas
- ✅ Unterstützung für mehrere Ressourcentypen
- ✅ Bulk-Operationen (in Entwicklung)

### Technische Architektur
- **Sprache**: Go 1.16+
- **Web-Framework**: Gin (hohe Performance)
- **Datenbank**: Couchbase (verteilte NoSQL)
- **Parser**: ANTLR v4 für SCIM-Filter
- **Datenformat**: Native JSON

## Systemarchitektur

### Kernkomponenten

#### 1. Server-Kern (`main.go`)
```go
// Server-Initialisierung
func main() {
    scim.InitDB()                    // Couchbase-Verbindung
    r := gin.Default()               // HTTP-Router
    scim.ReadResourceType(config, r) // Dynamisches Schema-Laden
    r.Run(port)                      // HTTP-Server
}
```

#### 2. Konfigurationsverwaltung (`scim/config.go`)
- **Dynamisches Schema-Laden** aus JSON-Dateien
- **Automatische Endpunkt-Registrierung** basierend auf Ressourcentypen
- **Schema-Validierung** beim Server-Start

#### 3. Datenbankintegration (`scim/couchbase.go`)
- **Sichere Verbindung** mit Authentifizierung
- **Automatische Bucket-Erstellung** pro Ressourcentyp
- **Anpassbare Bucket-Konfiguration**
- **Automatische Primärindizes**

#### 4. Filter-Parser (`scim/parser/`)
- **ANTLR-Grammatik** für SCIM-Filter
- **Automatische Konvertierung** zu N1QL-Abfragen
- **Vollständige Unterstützung** für SCIM-Operatoren

### SCIM-Operationen

#### Erstellen (POST)
```http
POST /scim/v2/Users
Content-Type: application/json

{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "userName": "user@example.com",
  "name": {
    "familyName": "Nachname",
    "givenName": "Vorname"
  }
}
```

#### Lesen (GET)
```http
GET /scim/v2/Users/12345
```

#### Aktualisieren (PATCH)
```http
PATCH /scim/v2/Users/12345
Content-Type: application/json

{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
  "Operations": [
    {
      "op": "replace",
      "path": "name.familyName",
      "value": "NeuerNachname"
    }
  ]
}
```

#### Suchen (GET)
```http
GET /scim/v2/Users?filter=userName sw "admin"&sortBy=userName&sortOrder=ascending&startIndex=1&count=10
```

#### Löschen (DELETE)
```http
DELETE /scim/v2/Users/12345
```

## Systemkonfiguration

### Umgebungsvariablen

#### Erforderlich
```bash
export SCIM_ADMIN_USER="Administrator"     # Couchbase-Admin-Benutzer
export SCIM_ADMIN_PASSWORD="admin123"     # Couchbase-Admin-Passwort
```

#### Optional
```bash
export SCIM_COUCHBASE_URL="localhost"     # Couchbase-Server-URL
export SCIM_PORT=":8080"                  # SCIM-Server-Port
```

### Konfigurationsstruktur

```
config/
├── schemas/                    # SCIM-Schema-Definitionen
│   ├── urn+ietf+params+scim+schemas+core+2.0+User.json
│   ├── urn+ietf+params+scim+schemas+core+2.0+Group.json
│   └── urn+ietf+params+scim+schemas+extension+*.json
├── resourceType/              # Ressourcentypen
│   ├── User.json
│   ├── Group.json
│   └── Element.json
├── bucketSettings/            # Couchbase-Bucket-Konfiguration
│   ├── User.json
│   ├── Group.json
│   └── Element.json
└── serviceProviderConfig/     # Anbieter-Konfiguration
    └── sp_config.json
```

## Schemas und Erweiterungen

### Basis-Benutzer-Schema
```json
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User",
  "name": "User",
  "description": "Benutzerkonto",
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

### Benutzerdefinierte Erweiterungen
```json
{
  "id": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
  "name": "Unternehmensbenutzer",
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

## Zugriffskontrolle

### Rollen und Berechtigungen
```json
{
  "name": "sensitiveAttribute",
  "type": "string",
  "$reader": ["admin", "hr"],      # Rollen mit Leseberechtigung
  "$writer": ["admin"],            # Rollen mit Schreibberechtigung
  "returned": "default"
}
```

### Rollenvalidierung
```go
// Automatische Validierung bei Suchen
roles := []string{"user", "admin", "superadmin"}
element := ValidateReadRole(roles, resourceType, item)
```

## SCIM-Filter

### Unterstützte Syntax
```
# Grundlegende Vergleiche
userName eq "admin"
name.familyName co "Müller"
userName sw "admin"
active pr

# Zeitvergleiche
meta.lastModified gt "2023-01-01T00:00:00Z"
meta.lastModified ge "2023-01-01T00:00:00Z"

# Logische Operatoren
title pr and userType eq "Employee"
title pr or userType eq "Intern"
userType eq "Employee" and (emails co "company.com" or emails co "company.org")
```

### N1QL-Konvertierung
```go
// Konvertierungsbeispiel
query, _ := parser.FilterToN1QL("User", 
    "userName eq \"admin\" and active eq true")
// Ergebnis: SELECT * FROM `User` WHERE `userName` = "admin" AND `active` = true
```

## Installation und Deployment

### Systemanforderungen

#### Entwicklungsumgebung
- Go 1.16 oder höher
- Couchbase Server 6.0+
- ANTLR 4.7 (für Parser-Regenerierung)

#### Produktionsumgebung
- CPU: 2+ Kerne
- RAM: 4GB+ (abhängig vom Volumen)
- Speicher: SSD empfohlen
- Netzwerk: 1Gbps+ für hohe Parallelität

### Lokale Installation

#### 1. Repository klonen
```bash
git clone https://github.com/arturoeanton/goscim.git
cd goscim
```

#### 2. Abhängigkeiten installieren
```bash
go mod download
```

#### 3. Couchbase konfigurieren
```bash
# Couchbase in Docker ausführen
docker run -d --name couchbase-scim \
  -p 8091-8094:8091-8094 \
  -p 11210:11210 \
  couchbase:latest

# Cluster über Web-UI konfigurieren
# http://localhost:8091/ui/index.html
```

#### 4. Umgebungsvariablen konfigurieren
```bash
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"
export SCIM_COUCHBASE_URL="localhost"
export SCIM_PORT=":8080"
```

#### 5. Server starten
```bash
go run main.go
```

## Überwachung und Betrieb

### Systemlogs
```bash
# Strukturierte Logs konfigurieren
export SCIM_LOG_LEVEL=debug
export SCIM_LOG_FORMAT=json

# Log-Beispiel
{"level":"info","timestamp":"2023-12-01T10:00:00Z","message":"GoScim v0.1 gestartet"}
{"level":"debug","timestamp":"2023-12-01T10:00:01Z","message":"Bucket erstellt","bucket":"User"}
```

### Empfohlene Metriken
- Anfragen pro Sekunde (RPS)
- Antwortzeit-Perzentile
- Fehlerrate pro Endpunkt
- Aktive Couchbase-Verbindungen
- Speicher- und CPU-Nutzung

### Gesundheitsprüfungen
```http
GET /health
{
  "status": "healthy",
  "database": "connected",
  "version": "v0.1"
}
```

## Sicherheit

### Sicherheitsüberlegungen

#### Authentifizierung
- OAuth 2.0 / OpenID Connect implementieren
- JWT-Token-Unterstützung
- Token-Validierung bei jeder Anfrage

#### Autorisierung
- Granulare rollenbasierte Kontrolle
- Ressourcen- und Operationsberechtigungen
- Zugriffs-Audit-Logs

#### Kommunikation
- Verpflichtende TLS 1.3 in Produktion
- Gültige Zertifikate
- HTTP-Sicherheitsheader

## Fehlerbehebung

### Häufige Probleme

#### Couchbase-Verbindung
```bash
# Konnektivität prüfen
telnet localhost 8091

# Anmeldedaten prüfen
curl -u Administrator:admin123 http://localhost:8091/pools
```

#### Schema-Fehler
```bash
# Schema-JSON validieren
jq . config/schemas/schema.json

# Syntax prüfen
go run main.go --validate-config
```

### Produktions-Deployment

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

## Tests und Entwicklung

### Tests ausführen
```bash
# Unit-Tests
go test ./...

# Spezifische Tests
go test ./scim/parser -v

# Tests mit Coverage
go test -cover ./...
```

### Verwendungsbeispiele
```bash
# Benutzer erstellen
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "testuser",
    "name": {
      "familyName": "Nachname",
      "givenName": "Vorname"
    }
  }'

# Benutzer suchen
curl "http://localhost:8080/scim/v2/Users?filter=userName sw \"test\""

# Provider-Konfiguration abrufen
curl http://localhost:8080/ServiceProviderConfig
```

#### Performance-Probleme
```bash
# Indizes überprüfen
curl -u admin:pass http://localhost:8091/query/service \
  -d 'statement=SELECT * FROM system:indexes WHERE keyspace_id="User"'
```

## Entwicklungs-Roadmap

### Phase 1: Stabilisierung
- Robuste Authentifizierung implementieren
- Vollständige Test-Suite
- Verbesserung von Logging und Monitoring

### Phase 2: Skalierbarkeit
- Cluster-Unterstützung
- Verteilter Cache
- Performance-Optimierungen

### Phase 3: Erweiterte Features
- Vollständige Bulk-Operationen
- Webhooks und Benachrichtigungen
- Administrations-Dashboard

## Externe System-Integration

### Identity Provider
- Active Directory
- LDAP
- OAuth 2.0 Provider
- SAML 2.0

### Zielsysteme
- SaaS-Anwendungen
- Benutzer-Datenbanken
- Verzeichnissysteme
- Drittanbieter-APIs

## Mitwirken

### Neue Ressourcen hinzufügen
1. JSON-Schema in `config/schemas/` erstellen
2. Ressourcentyp in `config/resourceType/` definieren
3. Bucket in `config/bucketSettings/` konfigurieren
4. Server neu starten, um Änderungen zu laden

### Parser-Regenerierung
```bash
# ANTLR installieren
wget http://www.antlr.org/download/antlr-4.7-complete.jar
alias antlr='java -jar $PWD/antlr-4.7-complete.jar'

# Parser regenerieren
antlr -Dlanguage=Go -o scim/parser ScimFilter.g4
```

## Community und Support

Für technischen Support, Fehlermeldungen oder Feature-Requests:
- **Issues**: [GitHub Issues](https://github.com/arturoeanton/goscim/issues)
- **Diskussionen**: [GitHub Discussions](https://github.com/arturoeanton/goscim/discussions)
- **Dokumentation**: [Projekt-Wiki](https://github.com/arturoeanton/goscim/wiki)
- **Beispiele**: `httpexamples/` Verzeichnis