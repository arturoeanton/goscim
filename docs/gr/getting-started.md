# Erste Schritte mit GoSCIM

Willkommen bei GoSCIM! Diese Anleitung hilft Ihnen dabei, Ihren eigenen SCIM 2.0-Server in nur wenigen Minuten zum Laufen zu bringen.

## Was ist GoSCIM?

GoSCIM ist eine leichtgewichtige, schnelle und flexible Implementierung des SCIM 2.0-Protokolls, die in Go erstellt wurde. Es ist darauf ausgelegt:

- 🚀 **Identity Management zu vereinfachen** über mehrere Systeme hinweg
- 🔧 **Einfache Integration** in bestehende Infrastruktur
- 📈 **Skalierung** von kleinen Startups bis zu großen Unternehmen
- 🛡️ **Schutz** Ihrer Identitätsdaten mit branchenüblichen Best Practices

## Schnellstart (2 Minuten)

### Option 1: Docker (Empfohlen)

Der schnellste Weg, GoSCIM auszuprobieren:

```bash
# Repository klonen
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Mit Docker Compose starten
docker-compose up -d

# Warten bis Services starten (etwa 30 Sekunden)
sleep 30

# SCIM-Server testen
curl http://localhost:8080/ServiceProviderConfig
```

### Option 2: Aus dem Quellcode erstellen

Wenn Sie lieber aus dem Quellcode erstellen möchten:

```bash
# Voraussetzungen: Go 1.16+ und Couchbase
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Umgebungsvariablen setzen
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"

# Server starten
go run main.go
```

## Ihre ersten SCIM-Operationen

Sobald Ihr Server läuft, probieren Sie diese grundlegenden Operationen aus:

### 1. Server-Konfiguration prüfen
```bash
curl http://localhost:8080/ServiceProviderConfig
```

### 2. Ihren ersten Benutzer erstellen
```bash
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "jane.doe@example.com",
    "name": {
      "familyName": "Müller",
      "givenName": "Anna"
    },
    "emails": [{
      "value": "jane.doe@example.com",
      "type": "work",
      "primary": true
    }],
    "active": true
  }'
```

### 3. Nach Benutzern suchen
```bash
curl "http://localhost:8080/scim/v2/Users?filter=userName sw 'jane'"
```

### 4. Verfügbare Ressourcen auflisten
```bash
curl http://localhost:8080/ResourceTypes
```

## Die Antwort verstehen

Wenn Sie einen Benutzer erstellen, erhalten Sie eine Antwort wie diese:

```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "userName": "jane.doe@example.com",
  "name": {
    "familyName": "Müller",
    "givenName": "Anna",
    "formatted": "Anna Müller"
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

Schlüsselelemente:
- **`id`**: Vom Server generierte eindeutige Kennung
- **`meta`**: Metadaten einschließlich Erstellungszeit und Standort
- **`schemas`**: Für diese Ressource verwendete SCIM-Schemas

## Häufige Anwendungsfälle

### 1. Mitarbeiter-Onboarding
```bash
# Neuen Mitarbeiter erstellen
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": [
      "urn:ietf:params:scim:schemas:core:2.0:User",
      "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
    ],
    "userName": "john.smith@company.com",
    "name": {
      "familyName": "Schmidt",
      "givenName": "Hans"
    },
    "emails": [{
      "value": "john.smith@company.com",
      "type": "work"
    }],
    "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
      "employeeNumber": "EMP001",
      "department": "Engineering",
      "manager": {
        "value": "manager-id-here"
      }
    }
  }'
```

### 2. Erweiterte Filterung
```bash
# Aktive Benutzer in der Engineering-Abteilung finden
curl "http://localhost:8080/scim/v2/Users?filter=active eq true and urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department eq 'Engineering'"

# Benutzer mit Firmen-E-Mail finden
curl "http://localhost:8080/scim/v2/Users?filter=emails[type eq 'work' and value ew '@company.com']"

# Kürzlich geänderte Benutzer finden
curl "http://localhost:8080/scim/v2/Users?filter=meta.lastModified gt '2023-01-01T00:00:00Z'"
```

## Fehlerbehandlung

GoSCIM gibt Standard-HTTP-Statuscodes und SCIM-Fehlerantworten zurück:

```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"],
  "detail": "Attribut 'userName' ist erforderlich",
  "status": "400",
  "scimType": "invalidValue"
}
```

Häufige Fehler:
- **400 Bad Request**: Ungültige Daten oder fehlende Pflichtfelder
- **404 Not Found**: Ressource existiert nicht
- **409 Conflict**: Ressource existiert bereits (z.B. doppelter userName)
- **500 Internal Server Error**: Server-seitige Probleme

## Konfiguration Grundlagen

### Umgebungsvariablen
```bash
# Datenbankverbindung
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="your-secure-password"
export SCIM_COUCHBASE_URL="couchbase-server.example.com"

# Server-Einstellungen
export SCIM_PORT=":8080"
export SCIM_LOG_LEVEL="info"

# Sicherheit (für Produktion)
export SCIM_TLS_ENABLED="true"
export SCIM_TLS_CERT_FILE="/path/to/cert.pem"
export SCIM_TLS_KEY_FILE="/path/to/key.pem"
```

## Nächste Schritte

Jetzt, da GoSCIM läuft, sind hier einige nächste Schritte:

1. **🔐 [Authentifizierung einrichten](security.md)** - OAuth 2.0 oder JWT-Authentifizierung hinzufügen
2. **📊 [Monitoring konfigurieren](operations.md)** - Metriken und Logging einrichten
3. **🔌 [Integrationen hinzufügen](integrations.md)** - Mit Active Directory, LDAP oder SaaS-Apps verbinden
4. **🏗️ [Architektur lernen](architecture.md)** - Verstehen, wie GoSCIM intern funktioniert
5. **👩‍💻 [Beitragen](development.md)** - Helfen Sie, GoSCIM zu verbessern

## Hilfe erhalten

- 📚 **Dokumentation**: Schauen Sie sich die [vollständige Dokumentation](README.md) an
- 🐛 **Issues**: Melden Sie Bugs in [GitHub Issues](https://github.com/arturoeanton/goscim/issues)
- 💬 **Diskussionen**: Stellen Sie Fragen in [GitHub Discussions](https://github.com/arturoeanton/goscim/discussions)
- 📖 **Beispiele**: Erkunden Sie das `httpexamples/` Verzeichnis für weitere Anwendungsbeispiele

Willkommen in der GoSCIM-Community! 🎉