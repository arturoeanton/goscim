# Débuter avec GoSCIM

Bienvenue dans GoSCIM ! Ce guide vous permettra de faire fonctionner votre propre serveur SCIM 2.0 en quelques minutes.

## Qu'est-ce que GoSCIM ?

GoSCIM est une implémentation légère, rapide et flexible du protocole SCIM 2.0 construite en Go. Il est conçu pour :

- 🚀 **Simplifier la gestion d'identité** à travers plusieurs systèmes
- 🔧 **S'intégrer facilement** avec l'infrastructure existante
- 📈 **Évoluer** des petites startups aux grandes entreprises
- 🛡️ **Sécuriser** vos données d'identité avec les meilleures pratiques de l'industrie

## Démarrage Rapide (2 Minutes)

### Option 1 : Docker (Recommandé)

Le moyen le plus rapide d'essayer GoSCIM :

```bash
# Cloner le dépôt
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Démarrer avec Docker Compose
docker-compose up -d

# Attendre que les services démarrent (environ 30 secondes)
sleep 30

# Tester votre serveur SCIM
curl http://localhost:8080/ServiceProviderConfig
```

### Option 2 : Construire depuis les sources

Si vous préférez construire depuis les sources :

```bash
# Prérequis : Go 1.16+ et Couchbase
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Définir les variables d'environnement
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"

# Lancer le serveur
go run main.go
```

## Vos Premières Opérations SCIM

Une fois votre serveur en cours d'exécution, essayez ces opérations de base :

### 1. Vérifier la configuration du serveur
```bash
curl http://localhost:8080/ServiceProviderConfig
```

### 2. Créer votre premier utilisateur
```bash
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "jane.doe@example.com",
    "name": {
      "familyName": "Martin",
      "givenName": "Marie"
    },
    "emails": [{
      "value": "jane.doe@example.com",
      "type": "work",
      "primary": true
    }],
    "active": true
  }'
```

### 3. Rechercher des utilisateurs
```bash
curl "http://localhost:8080/scim/v2/Users?filter=userName sw 'jane'"
```

### 4. Lister les ressources disponibles
```bash
curl http://localhost:8080/ResourceTypes
```

## Comprendre la Réponse

Lorsque vous créez un utilisateur, vous obtiendrez une réponse comme celle-ci :

```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "userName": "jane.doe@example.com",
  "name": {
    "familyName": "Martin",
    "givenName": "Marie",
    "formatted": "Marie Martin"
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

Éléments clés :
- **`id`** : Identifiant unique généré par le serveur
- **`meta`** : Métadonnées incluant l'heure de création et l'emplacement
- **`schemas`** : Schémas SCIM utilisés pour cette ressource

## Cas d'Usage Courants

### 1. Intégration d'employés
```bash
# Créer un nouvel employé
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": [
      "urn:ietf:params:scim:schemas:core:2.0:User",
      "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
    ],
    "userName": "john.smith@company.com",
    "name": {
      "familyName": "Dupont",
      "givenName": "Jean"
    },
    "emails": [{
      "value": "john.smith@company.com",
      "type": "work"
    }],
    "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
      "employeeNumber": "EMP001",
      "department": "Ingénierie",
      "manager": {
        "value": "manager-id-here"
      }
    }
  }'
```

### 2. Filtrage Avancé
```bash
# Trouver les utilisateurs actifs du département ingénierie
curl "http://localhost:8080/scim/v2/Users?filter=active eq true and urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department eq 'Ingénierie'"

# Trouver les utilisateurs avec un email d'entreprise
curl "http://localhost:8080/scim/v2/Users?filter=emails[type eq 'work' and value ew '@company.com']"

# Trouver les utilisateurs récemment modifiés
curl "http://localhost:8080/scim/v2/Users?filter=meta.lastModified gt '2023-01-01T00:00:00Z'"
```

## Gestion des Erreurs

GoSCIM retourne des codes de statut HTTP standard et des réponses d'erreur SCIM :

```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"],
  "detail": "L'attribut 'userName' est requis",
  "status": "400",
  "scimType": "invalidValue"
}
```

Erreurs courantes :
- **400 Bad Request** : Données invalides ou champs requis manquants
- **404 Not Found** : La ressource n'existe pas
- **409 Conflict** : La ressource existe déjà (ex: userName en double)
- **500 Internal Server Error** : Problèmes côté serveur

## Configuration de Base

### Variables d'Environnement
```bash
# Connexion à la base de données
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="your-secure-password"
export SCIM_COUCHBASE_URL="couchbase-server.example.com"

# Paramètres du serveur
export SCIM_PORT=":8080"
export SCIM_LOG_LEVEL="info"

# Sécurité (pour la production)
export SCIM_TLS_ENABLED="true"
export SCIM_TLS_CERT_FILE="/path/to/cert.pem"
export SCIM_TLS_KEY_FILE="/path/to/key.pem"
```

## Étapes Suivantes

Maintenant que GoSCIM fonctionne, voici quelques étapes suivantes :

1. **🔐 [Configurer l'authentification](security.md)** - Ajouter l'authentification OAuth 2.0 ou JWT
2. **📊 [Configurer la surveillance](operations.md)** - Mettre en place les métriques et la journalisation
3. **🔌 [Ajouter des intégrations](integrations.md)** - Se connecter à Active Directory, LDAP ou des applications SaaS
4. **🏗️ [Apprendre l'architecture](architecture.md)** - Comprendre le fonctionnement interne de GoSCIM
5. **👩‍💻 [Contribuer](development.md)** - Aider à améliorer GoSCIM

## Obtenir de l'Aide

- 📚 **Documentation** : Consultez la [documentation complète](README.md)
- 🐛 **Issues** : Signalez les bugs sur [GitHub Issues](https://github.com/arturoeanton/goscim/issues)
- 💬 **Discussions** : Posez des questions dans [GitHub Discussions](https://github.com/arturoeanton/goscim/discussions)
- 📖 **Exemples** : Explorez le répertoire `httpexamples/` pour plus d'exemples d'utilisation

Bienvenue dans la communauté GoSCIM ! 🎉