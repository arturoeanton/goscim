# Guida all'Installazione

## Installazione Rapida

### Opzione 1: Docker (Raccomandato)

Il modo più veloce per far funzionare GoSCIM:

```bash
# Clonare il repository
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Avviare con Docker Compose
docker-compose up -d

# Verificare l'installazione
curl http://localhost:8080/ServiceProviderConfig
```

### Opzione 2: Compilare dal codice sorgente

```bash
# Prerequisiti: Go 1.16+
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Compilare ed eseguire
go build -o goscim main.go
./goscim
```

## Requisiti di Sistema

### Requisiti Minimi
- **CPU**: 1 core
- **Memoria**: 512MB RAM
- **Disco**: 100MB di spazio disponibile
- **SO**: Linux, macOS, Windows

### Configurazione Raccomandata
- **CPU**: 2+ core
- **Memoria**: 2GB+ RAM  
- **Disco**: 1GB+ di spazio disponibile (per i dati)
- **Rete**: Ethernet 1Gbps

## Variabili d'Ambiente

### Obbligatorie
```bash
export SCIM_ADMIN_USER="Administrator"     # Utente admin Couchbase
export SCIM_ADMIN_PASSWORD="admin123"     # Password admin Couchbase
```

### Opzionali
```bash
export SCIM_COUCHBASE_URL="localhost"     # URL server Couchbase
export SCIM_PORT=":8080"                  # Porta del server
export SCIM_LOG_LEVEL="info"              # Livello di log
```

## Configurazione Database

### Installazione Couchbase

```bash
# Eseguire Couchbase in Docker
docker run -d --name couchbase-scim \
  -p 8091-8094:8091-8094 \
  -p 11210:11210 \
  couchbase:latest

# Configurare cluster (tramite interfaccia web)
# Accesso: http://localhost:8091/ui/index.html
```

### Distribuzione in Produzione

```bash
# Configurare con variabili d'ambiente
export SCIM_ADMIN_USER="your-admin-user"
export SCIM_ADMIN_PASSWORD="your-secure-password"
export SCIM_COUCHBASE_URL="your-couchbase-server.com"

# Avviare GoSCIM
./goscim
```

## Verificare l'Installazione

```bash
# Verificare stato del server
curl http://localhost:8080/ServiceProviderConfig

# Verificare tipi di risorse
curl http://localhost:8080/ResourceTypes

# Verificare schemi
curl http://localhost:8080/Schemas
```

## Risoluzione Problemi

### Problemi Comuni

1. **Fallimento connessione**
   ```bash
   # Verificare se Couchbase è in esecuzione
   telnet localhost 8091
   ```

2. **Errore di autenticazione**
   ```bash
   # Verificare variabili d'ambiente
   echo $SCIM_ADMIN_USER
   echo $SCIM_ADMIN_PASSWORD
   ```

3. **Conflitto porta**
   ```bash
   # Cambiare porta
   export SCIM_PORT=":8081"
   ```