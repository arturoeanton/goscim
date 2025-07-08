# Guide d'Installation

## Installation Rapide

### Option 1 : Docker (Recommandé)

Le moyen le plus rapide de faire fonctionner GoSCIM :

```bash
# Cloner le dépôt
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Démarrer avec Docker Compose
docker-compose up -d

# Vérifier l'installation
curl http://localhost:8080/ServiceProviderConfig
```

### Option 2 : Construire depuis les sources

```bash
# Prérequis : Go 1.16+
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Construire et exécuter
go build -o goscim main.go
./goscim
```

## Configuration Système

### Exigences Minimales
- **CPU** : 1 cœur
- **Mémoire** : 512MB RAM
- **Disque** : 100MB d'espace disponible
- **OS** : Linux, macOS, Windows

### Configuration Recommandée
- **CPU** : 2+ cœurs
- **Mémoire** : 2GB+ RAM  
- **Disque** : 1GB+ d'espace disponible (pour les données)
- **Réseau** : Ethernet 1Gbps

## Variables d'Environnement

### Requises
```bash
export SCIM_ADMIN_USER="Administrator"     # Utilisateur admin Couchbase
export SCIM_ADMIN_PASSWORD="admin123"     # Mot de passe admin Couchbase
```

### Optionnelles
```bash
export SCIM_COUCHBASE_URL="localhost"     # URL du serveur Couchbase
export SCIM_PORT=":8080"                  # Port du serveur
export SCIM_LOG_LEVEL="info"              # Niveau de log
```

## Configuration de la Base de Données

### Installation Couchbase

```bash
# Exécuter Couchbase dans Docker
docker run -d --name couchbase-scim \
  -p 8091-8094:8091-8094 \
  -p 11210:11210 \
  couchbase:latest

# Configurer le cluster (via l'interface web)
# Accès : http://localhost:8091/ui/index.html
```

### Déploiement en Production

```bash
# Configurer avec les variables d'environnement
export SCIM_ADMIN_USER="your-admin-user"
export SCIM_ADMIN_PASSWORD="your-secure-password"
export SCIM_COUCHBASE_URL="your-couchbase-server.com"

# Démarrer GoSCIM
./goscim
```

## Vérifier l'Installation

```bash
# Vérifier l'état du serveur
curl http://localhost:8080/ServiceProviderConfig

# Vérifier les types de ressources
curl http://localhost:8080/ResourceTypes

# Vérifier les schémas
curl http://localhost:8080/Schemas
```

## Dépannage

### Problèmes Courants

1. **Échec de connexion**
   ```bash
   # Vérifier si Couchbase fonctionne
   telnet localhost 8091
   ```

2. **Erreur d'authentification**
   ```bash
   # Vérifier les variables d'environnement
   echo $SCIM_ADMIN_USER
   echo $SCIM_ADMIN_PASSWORD
   ```

3. **Conflit de port**
   ```bash
   # Changer le port
   export SCIM_PORT=":8081"
   ```