# Installationsanleitung

## Schnelle Installation

### Option 1: Docker (Empfohlen)

Der schnellste Weg, GoSCIM zum Laufen zu bringen:

```bash
# Repository klonen
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Mit Docker Compose starten
docker-compose up -d

# Installation überprüfen
curl http://localhost:8080/ServiceProviderConfig
```

### Option 2: Aus dem Quellcode erstellen

```bash
# Voraussetzungen: Go 1.16+
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Erstellen und ausführen
go build -o goscim main.go
./goscim
```

## Systemanforderungen

### Mindestanforderungen
- **CPU**: 1 Kern
- **Speicher**: 512MB RAM
- **Festplatte**: 100MB verfügbarer Speicher
- **Betriebssystem**: Linux, macOS, Windows

### Empfohlene Konfiguration
- **CPU**: 2+ Kerne
- **Speicher**: 2GB+ RAM  
- **Festplatte**: 1GB+ verfügbarer Speicher (für Daten)
- **Netzwerk**: 1Gbps Ethernet

## Umgebungsvariablen

### Erforderlich
```bash
export SCIM_ADMIN_USER="Administrator"     # Couchbase-Admin-Benutzer
export SCIM_ADMIN_PASSWORD="admin123"     # Couchbase-Admin-Passwort
```

### Optional
```bash
export SCIM_COUCHBASE_URL="localhost"     # Couchbase-Server-URL
export SCIM_PORT=":8080"                  # Server-Port
export SCIM_LOG_LEVEL="info"              # Log-Level
```

## Datenbank-Setup

### Couchbase-Installation

```bash
# Couchbase in Docker ausführen
docker run -d --name couchbase-scim \
  -p 8091-8094:8091-8094 \
  -p 11210:11210 \
  couchbase:latest

# Cluster konfigurieren (über Web-UI)
# Zugriff: http://localhost:8091/ui/index.html
```

### Produktions-Deployment

```bash
# Mit Umgebungsvariablen konfigurieren
export SCIM_ADMIN_USER="your-admin-user"
export SCIM_ADMIN_PASSWORD="your-secure-password"
export SCIM_COUCHBASE_URL="your-couchbase-server.com"

# GoSCIM starten
./goscim
```

## Installation überprüfen

```bash
# Server-Status prüfen
curl http://localhost:8080/ServiceProviderConfig

# Ressourcentypen prüfen
curl http://localhost:8080/ResourceTypes

# Schemas prüfen
curl http://localhost:8080/Schemas
```

## Fehlerbehebung

### Häufige Probleme

1. **Verbindungsfehler**
   ```bash
   # Prüfen ob Couchbase läuft
   telnet localhost 8091
   ```

2. **Authentifizierungsfehler**
   ```bash
   # Umgebungsvariablen überprüfen
   echo $SCIM_ADMIN_USER
   echo $SCIM_ADMIN_PASSWORD
   ```

3. **Port-Konflikt**
   ```bash
   # Port ändern
   export SCIM_PORT=":8081"
   ```