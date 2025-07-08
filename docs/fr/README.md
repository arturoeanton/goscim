# GoSCIM Documentation Technique

## Aperçu du Projet

**GoSCIM** est une implémentation complète de SCIM 2.0 (System for Cross-domain Identity Management) construite en Go. Elle fournit une solution robuste et scalable pour la gestion d'identité dans des environnements distribués, spécialement conçue pour intégrer des systèmes d'identité hétérogènes.

## Fonctionnalités Principales

### Conformité SCIM 2.0
- ✅ Opérations CRUD complètes (Create, Read, Update, Delete)
- ✅ Recherche avancée avec filtres SCIM
- ✅ Pagination et tri
- ✅ Schémas extensibles et personnalisables
- ✅ Support de multiples types de ressources
- ✅ Opérations en lot (en développement)

### Architecture Technique
- **Langage**: Go 1.16+
- **Framework Web**: Gin (haute performance)
- **Base de données**: Couchbase (NoSQL distribué)
- **Parseur**: ANTLR v4 pour les filtres SCIM
- **Format de données**: JSON natif

## Architecture Système

### Composants Principaux

#### 1. Noyau du Serveur (`main.go`)
```go
// Initialisation du serveur
func main() {
    scim.InitDB()                    // Connexion Couchbase
    r := gin.Default()               // Routeur HTTP
    scim.ReadResourceType(config, r) // Chargement dynamique des schémas
    r.Run(port)                      // Serveur HTTP
}
```

#### 2. Gestion de Configuration (`scim/config.go`)
- **Chargement dynamique des schémas** depuis des fichiers JSON
- **Enregistrement automatique des endpoints** basé sur les types de ressources
- **Validation des schémas** au démarrage du serveur

#### 3. Intégration Base de Données (`scim/couchbase.go`)
- **Connexion sécurisée** avec authentification
- **Création automatique de buckets** par type de ressource
- **Configuration personnalisable des buckets**
- **Index primaires automatiques**

#### 4. Parseur de Filtres (`scim/parser/`)
- **Grammaire ANTLR** pour les filtres SCIM
- **Conversion automatique** vers les requêtes N1QL
- **Support complet** des opérateurs SCIM

### Opérations SCIM

#### Créer (POST)
```http
POST /scim/v2/Users
Content-Type: application/json

{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "userName": "user@example.com",
  "name": {
    "familyName": "Nom",
    "givenName": "Prénom"
  }
}
```

#### Lire (GET)
```http
GET /scim/v2/Users/12345
```

#### Mettre à jour (PATCH)
```http
PATCH /scim/v2/Users/12345
Content-Type: application/json

{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
  "Operations": [
    {
      "op": "replace",
      "path": "name.familyName",
      "value": "NouveauNom"
    }
  ]
}
```

#### Rechercher (GET)
```http
GET /scim/v2/Users?filter=userName sw "admin"&sortBy=userName&sortOrder=ascending&startIndex=1&count=10
```

#### Supprimer (DELETE)
```http
DELETE /scim/v2/Users/12345
```

## Configuration Système

### Variables d'Environnement

#### Requis
```bash
export SCIM_ADMIN_USER="Administrator"     # Utilisateur admin Couchbase
export SCIM_ADMIN_PASSWORD="admin123"     # Mot de passe admin Couchbase
```

#### Optionnel
```bash
export SCIM_COUCHBASE_URL="localhost"     # URL du serveur Couchbase
export SCIM_PORT=":8080"                  # Port du serveur SCIM
```

### Structure de Configuration

```
config/
├── schemas/                    # Définitions des schémas SCIM
│   ├── urn+ietf+params+scim+schemas+core+2.0+User.json
│   ├── urn+ietf+params+scim+schemas+core+2.0+Group.json
│   └── urn+ietf+params+scim+schemas+extension+*.json
├── resourceType/              # Types de ressources
│   ├── User.json
│   ├── Group.json
│   └── Element.json
├── bucketSettings/            # Configuration des buckets Couchbase
│   ├── User.json
│   ├── Group.json
│   └── Element.json
└── serviceProviderConfig/     # Configuration du fournisseur
    └── sp_config.json
```

## Schémas et Extensions

### Schéma Utilisateur de Base
```json
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User",
  "name": "User",
  "description": "Compte utilisateur",
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

### Extensions Personnalisées
```json
{
  "id": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
  "name": "Utilisateur Entreprise",
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

## Contrôle d'Accès

### Rôles et Permissions
```json
{
  "name": "sensitiveAttribute",
  "type": "string",
  "$reader": ["admin", "hr"],      # Rôles pouvant lire
  "$writer": ["admin"],            # Rôles pouvant écrire
  "returned": "default"
}
```

### Validation des Rôles
```go
// Validation automatique lors des recherches
roles := []string{"user", "admin", "superadmin"}
element := ValidateReadRole(roles, resourceType, item)
```

## Filtres SCIM

### Syntaxe Supportée
```
# Comparaisons de base
userName eq "admin"
name.familyName co "Martin"
userName sw "admin"
active pr

# Comparaisons temporelles
meta.lastModified gt "2023-01-01T00:00:00Z"
meta.lastModified ge "2023-01-01T00:00:00Z"

# Opérateurs logiques
title pr and userType eq "Employee"
title pr or userType eq "Intern"
userType eq "Employee" and (emails co "company.com" or emails co "company.org")
```

### Conversion N1QL
```go
// Exemple de conversion
query, _ := parser.FilterToN1QL("User", 
    "userName eq \"admin\" and active eq true")
// Résultat: SELECT * FROM `User` WHERE `userName` = "admin" AND `active` = true
```

## Installation et Déploiement

### Prérequis Système

#### Environnement de Développement
- Go 1.16 ou supérieur
- Couchbase Server 6.0+
- ANTLR 4.7 (pour la régénération du parseur)

#### Environnement de Production
- CPU: 2+ cœurs
- RAM: 4GB+ (selon le volume)
- Stockage: SSD recommandé
- Réseau: 1Gbps+ pour haute concurrence

### Installation Locale

#### 1. Cloner le Dépôt
```bash
git clone https://github.com/arturoeanton/goscim.git
cd goscim
```

#### 2. Installer les Dépendances
```bash
go mod download
```

#### 3. Configurer Couchbase
```bash
# Exécuter Couchbase dans Docker
docker run -d --name couchbase-scim \
  -p 8091-8094:8091-8094 \
  -p 11210:11210 \
  couchbase:latest

# Configurer le cluster via l'interface web
# http://localhost:8091/ui/index.html
```

#### 4. Configurer les Variables d'Environnement
```bash
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"
export SCIM_COUCHBASE_URL="localhost"
export SCIM_PORT=":8080"
```

#### 5. Démarrer le Serveur
```bash
go run main.go
```

## Surveillance et Opérations

### Logs Système
```bash
# Configurer les logs structurés
export SCIM_LOG_LEVEL=debug
export SCIM_LOG_FORMAT=json

# Exemple de log
{"level":"info","timestamp":"2023-12-01T10:00:00Z","message":"GoScim v0.1 démarré"}
{"level":"debug","timestamp":"2023-12-01T10:00:01Z","message":"Bucket créé","bucket":"User"}
```

### Métriques Recommandées
- Requêtes par seconde (RPS)
- Percentiles des temps de réponse
- Taux d'erreur par endpoint
- Connexions Couchbase actives
- Utilisation mémoire et CPU

### Vérifications de Santé
```http
GET /health
{
  "status": "healthy",
  "database": "connected",
  "version": "v0.1"
}
```

## Sécurité

### Considérations de Sécurité

#### Authentification
- Implémentation OAuth 2.0 / OpenID Connect
- Support des tokens JWT
- Validation des tokens à chaque requête

#### Autorisation
- Contrôle granulaire basé sur les rôles
- Permissions sur les ressources et opérations
- Logs d'audit d'accès

#### Communication
- TLS 1.3 obligatoire en production
- Certificats valides
- Headers de sécurité HTTP

## Dépannage

### Problèmes Courants

#### Connexion Couchbase
```bash
# Vérifier la connectivité
telnet localhost 8091

# Vérifier les identifiants
curl -u Administrator:admin123 http://localhost:8091/pools
```

#### Erreurs de Schéma
```bash
# Valider le JSON du schéma
jq . config/schemas/schema.json

# Vérifier la syntaxe
go run main.go --validate-config
```

### Déploiement en Production

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

## Tests et Développement

### Exécuter les Tests
```bash
# Tests unitaires
go test ./...

# Tests spécifiques
go test ./scim/parser -v

# Tests avec couverture
go test -cover ./...
```

### Exemples d'Utilisation
```bash
# Créer un utilisateur
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "testuser",
    "name": {
      "familyName": "Nom",
      "givenName": "Prénom"
    }
  }'

# Rechercher des utilisateurs
curl "http://localhost:8080/scim/v2/Users?filter=userName sw \"test\""

# Obtenir la configuration du fournisseur
curl http://localhost:8080/ServiceProviderConfig
```

#### Problèmes de Performance
```bash
# Vérifier les index
curl -u admin:pass http://localhost:8091/query/service \
  -d 'statement=SELECT * FROM system:indexes WHERE keyspace_id="User"'
```

## Feuille de Route de Développement

### Phase 1: Stabilisation
- Implémenter une authentification robuste
- Suite de tests complète
- Améliorer la journalisation et la surveillance

### Phase 2: Scalabilité
- Support de clustering
- Cache distribué
- Optimisations de performance

### Phase 3: Fonctionnalités Avancées
- Opérations bulk complètes
- Webhooks et notifications
- Tableau de bord d'administration

## Intégration de Systèmes Externes

### Fournisseurs d'Identité
- Active Directory
- LDAP
- Fournisseurs OAuth 2.0
- SAML 2.0

### Systèmes Cibles
- Applications SaaS
- Bases de données utilisateurs
- Systèmes d'annuaire
- APIs tierces

## Contribuer

### Ajouter de Nouvelles Ressources
1. Créer un schéma JSON dans `config/schemas/`
2. Définir le type de ressource dans `config/resourceType/`
3. Configurer le bucket dans `config/bucketSettings/`
4. Redémarrer le serveur pour charger les modifications

### Régénération du Parseur
```bash
# Installer ANTLR
wget http://www.antlr.org/download/antlr-4.7-complete.jar
alias antlr='java -jar $PWD/antlr-4.7-complete.jar'

# Régénérer le parseur
antlr -Dlanguage=Go -o scim/parser ScimFilter.g4
```

## Communauté et Support

Pour le support technique, les rapports de bugs ou les demandes de fonctionnalités:
- **Issues**: [GitHub Issues](https://github.com/arturoeanton/goscim/issues)
- **Discussions**: [GitHub Discussions](https://github.com/arturoeanton/goscim/discussions)
- **Documentation**: [Wiki du Projet](https://github.com/arturoeanton/goscim/wiki)
- **Exemples**: Répertoire `httpexamples/`