# Iniziare con GoSCIM

Benvenuto in GoSCIM! Questa guida ti aiuterà a far funzionare il tuo server SCIM 2.0 in pochi minuti.

## Cos'è GoSCIM?

GoSCIM è un'implementazione leggera, veloce e flessibile del protocollo SCIM 2.0 costruita in Go. È progettata per:

- 🚀 **Semplificare la gestione delle identità** attraverso sistemi multipli
- 🔧 **Integrarsi facilmente** con l'infrastruttura esistente
- 📈 **Scalare** da piccole startup a grandi aziende
- 🛡️ **Proteggere** i tuoi dati di identità con le migliori pratiche del settore

## Avvio Rapido (2 Minuti)

### Opzione 1: Docker (Raccomandato)

Il modo più veloce per provare GoSCIM:

```bash
# Clonare il repository
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Avviare con Docker Compose
docker-compose up -d

# Attendere l'avvio dei servizi (circa 30 secondi)
sleep 30

# Testare il server SCIM
curl http://localhost:8080/ServiceProviderConfig
```

### Opzione 2: Compilare dal codice sorgente

Se preferisci compilare dal codice sorgente:

```bash
# Prerequisiti: Go 1.16+ e Couchbase
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Impostare variabili d'ambiente
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"

# Eseguire il server
go run main.go
```

## Le Tue Prime Operazioni SCIM

Una volta che il server è in esecuzione, prova queste operazioni di base:

### 1. Verificare la configurazione del server
```bash
curl http://localhost:8080/ServiceProviderConfig
```

### 2. Creare il tuo primo utente
```bash
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "jane.doe@example.com",
    "name": {
      "familyName": "Rossi",
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

### 3. Cercare utenti
```bash
curl "http://localhost:8080/scim/v2/Users?filter=userName sw 'jane'"
```

### 4. Elencare le risorse disponibili
```bash
curl http://localhost:8080/ResourceTypes
```

## Comprendere la Risposta

Quando crei un utente, otterrai una risposta come questa:

```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "userName": "jane.doe@example.com",
  "name": {
    "familyName": "Rossi",
    "givenName": "Maria",
    "formatted": "Maria Rossi"
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

Elementi chiave:
- **`id`**: Identificatore unico generato dal server
- **`meta`**: Metadati inclusi tempo di creazione e posizione
- **`schemas`**: Schemi SCIM utilizzati per questa risorsa

## Casi d'Uso Comuni

### 1. Onboarding dei Dipendenti
```bash
# Creare nuovo dipendente
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": [
      "urn:ietf:params:scim:schemas:core:2.0:User",
      "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
    ],
    "userName": "john.smith@company.com",
    "name": {
      "familyName": "Bianchi",
      "givenName": "Marco"
    },
    "emails": [{
      "value": "john.smith@company.com",
      "type": "work"
    }],
    "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
      "employeeNumber": "EMP001",
      "department": "Ingegneria",
      "manager": {
        "value": "manager-id-here"
      }
    }
  }'
```

### 2. Filtraggio Avanzato
```bash
# Trovare utenti attivi nel dipartimento ingegneria
curl "http://localhost:8080/scim/v2/Users?filter=active eq true and urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department eq 'Ingegneria'"

# Trovare utenti con email aziendale
curl "http://localhost:8080/scim/v2/Users?filter=emails[type eq 'work' and value ew '@company.com']"

# Trovare utenti modificati di recente
curl "http://localhost:8080/scim/v2/Users?filter=meta.lastModified gt '2023-01-01T00:00:00Z'"
```

## Gestione degli Errori

GoSCIM restituisce codici di stato HTTP standard e risposte di errore SCIM:

```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"],
  "detail": "L'attributo 'userName' è obbligatorio",
  "status": "400",
  "scimType": "invalidValue"
}
```

Errori comuni:
- **400 Bad Request**: Dati non validi o campi obbligatori mancanti
- **404 Not Found**: La risorsa non esiste
- **409 Conflict**: La risorsa esiste già (es. userName duplicato)
- **500 Internal Server Error**: Problemi lato server

## Configurazione di Base

### Variabili d'Ambiente
```bash
# Connessione database
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="your-secure-password"
export SCIM_COUCHBASE_URL="couchbase-server.example.com"

# Impostazioni server
export SCIM_PORT=":8080"
export SCIM_LOG_LEVEL="info"

# Sicurezza (per produzione)
export SCIM_TLS_ENABLED="true"
export SCIM_TLS_CERT_FILE="/path/to/cert.pem"
export SCIM_TLS_KEY_FILE="/path/to/key.pem"
```

## Prossimi Passi

Ora che GoSCIM è in esecuzione, ecco alcuni passi successivi:

1. **🔐 [Configurare l'autenticazione](security.md)** - Aggiungere autenticazione OAuth 2.0 o JWT
2. **📊 [Configurare il monitoraggio](operations.md)** - Impostare metriche e logging
3. **🔌 [Aggiungere integrazioni](integrations.md)** - Connettere ad Active Directory, LDAP o app SaaS
4. **🏗️ [Imparare l'architettura](architecture.md)** - Capire come funziona internamente GoSCIM
5. **👩‍💻 [Contribuire](development.md)** - Aiutare a migliorare GoSCIM

## Ottenere Aiuto

- 📚 **Documentazione**: Consulta la [documentazione completa](README.md)
- 🐛 **Issues**: Segnala bug su [GitHub Issues](https://github.com/arturoeanton/goscim/issues)
- 💬 **Discussioni**: Fai domande in [GitHub Discussions](https://github.com/arturoeanton/goscim/discussions)
- 📖 **Esempi**: Esplora la directory `httpexamples/` per altri esempi di utilizzo

Benvenuto nella comunità GoSCIM! 🎉