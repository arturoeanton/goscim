# Guía de Integraciones GoSCIM

## Introducción a las Integraciones

GoSCIM está diseñado para actuar como un hub central de identidades, facilitando la sincronización bidireccional entre diferentes sistemas de gestión de identidades y aplicaciones empresariales. Esta guía cubre las integraciones más comunes y proporciona ejemplos prácticos de implementación.

## Arquitectura de Integración

### Modelo Hub-and-Spoke

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Active        │    │      LDAP       │    │    Okta         │
│   Directory     │◄──►│   Directory     │◄──►│   Identity      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         ▲                        ▲                        ▲
         │                        │                        │
         ▼                        ▼                        ▼
┌─────────────────────────────────────────────────────────────────┐
│                        GoSCIM Hub                              │
│                   (Identity Provider)                          │
└─────────────────────────────────────────────────────────────────┘
         ▲                        ▲                        ▲
         │                        │                        │
         ▼                        ▼                        ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Salesforce    │    │     Slack       │    │    Jira         │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Patrones de Integración

#### 1. Pull Pattern (GoSCIM como Consumer)
```go
// GoSCIM sincroniza datos desde sistema externo
func syncFromExternalSystem() {
    users := externalSystem.GetUsers()
    for _, user := range users {
        scimUser := convertToSCIM(user)
        goscim.CreateOrUpdateUser(scimUser)
    }
}
```

#### 2. Push Pattern (GoSCIM como Provider)
```go
// Sistema externo consume cambios de GoSCIM
func handleUserCreated(user SCIMUser) {
    externalUser := convertFromSCIM(user)
    externalSystem.CreateUser(externalUser)
}
```

#### 3. Webhook Pattern (Notificaciones en Tiempo Real)
```go
// GoSCIM notifica cambios via webhooks
func notifySubscribers(event UserEvent) {
    for _, webhook := range registeredWebhooks {
        webhook.Send(event)
    }
}
```

## Integración con Active Directory

### Configuración del Conector AD

#### 1. Instalar Active Directory Connector
```bash
# Crear directorio para el conector
mkdir -p /opt/goscim-connectors/ad
cd /opt/goscim-connectors/ad

# Descargar conector (ejemplo)
wget https://releases.goscim.com/connectors/ad-connector-v1.0.tar.gz
tar -xzf ad-connector-v1.0.tar.gz
```

#### 2. Configuración del Conector
```yaml
# config/ad-connector.yaml
active_directory:
  server: "dc01.empresa.local"
  port: 389
  use_tls: true
  bind_dn: "CN=SCIM Service,OU=Service Accounts,DC=empresa,DC=local"
  bind_password: "${AD_BIND_PASSWORD}"
  base_dn: "OU=Users,DC=empresa,DC=local"
  
scim_endpoint:
  url: "https://scim.empresa.com/scim/v2"
  auth_token: "${SCIM_AUTH_TOKEN}"
  
sync_settings:
  interval: "15m"
  batch_size: 100
  attributes:
    - sAMAccountName -> userName
    - givenName -> name.givenName
    - sn -> name.familyName
    - mail -> emails[type eq "work"].value
    - telephoneNumber -> phoneNumbers[type eq "work"].value
    - department -> urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department
    - employeeID -> urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:employeeNumber
```

#### 3. Script de Sincronización
```go
// ad-connector/main.go
package main

import (
    "log"
    "time"
    
    "github.com/go-ldap/ldap/v3"
    "github.com/arturoeanton/goscim/client"
)

type ADConnector struct {
    ldapConn   *ldap.Conn
    scimClient *client.SCIMClient
    config     *Config
}

func (c *ADConnector) SyncUsers() error {
    // Buscar usuarios en AD
    searchRequest := ldap.NewSearchRequest(
        c.config.BaseDN,
        ldap.ScopeWholeSubtree,
        ldap.NeverDerefAliases,
        0, 0, false,
        "(&(objectClass=user)(!(userAccountControl:1.2.840.113556.1.4.803:=2)))",
        []string{"sAMAccountName", "givenName", "sn", "mail", "telephoneNumber", "department", "employeeID"},
        nil,
    )
    
    searchResult, err := c.ldapConn.Search(searchRequest)
    if err != nil {
        return err
    }
    
    // Convertir y sincronizar usuarios
    for _, entry := range searchResult.Entries {
        scimUser := c.convertADUserToSCIM(entry)
        
        // Verificar si el usuario ya existe
        existingUser, err := c.scimClient.GetUserByUserName(scimUser.UserName)
        if err != nil {
            // Usuario no existe, crear nuevo
            _, err = c.scimClient.CreateUser(scimUser)
            if err != nil {
                log.Printf("Error creating user %s: %v", scimUser.UserName, err)
            }
        } else {
            // Usuario existe, actualizar
            scimUser.ID = existingUser.ID
            _, err = c.scimClient.UpdateUser(scimUser)
            if err != nil {
                log.Printf("Error updating user %s: %v", scimUser.UserName, err)
            }
        }
    }
    
    return nil
}

func (c *ADConnector) convertADUserToSCIM(entry *ldap.Entry) *client.User {
    return &client.User{
        Schemas:  []string{"urn:ietf:params:scim:schemas:core:2.0:User", "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"},
        UserName: entry.GetAttributeValue("sAMAccountName"),
        Name: &client.Name{
            GivenName:  entry.GetAttributeValue("givenName"),
            FamilyName: entry.GetAttributeValue("sn"),
        },
        Emails: []client.Email{
            {
                Value:   entry.GetAttributeValue("mail"),
                Type:    "work",
                Primary: true,
            },
        },
        PhoneNumbers: []client.PhoneNumber{
            {
                Value:   entry.GetAttributeValue("telephoneNumber"),
                Type:    "work",
                Primary: true,
            },
        },
        Enterprise: &client.EnterpriseUser{
            Department:     entry.GetAttributeValue("department"),
            EmployeeNumber: entry.GetAttributeValue("employeeID"),
        },
        Active: true,
    }
}
```

#### 4. Servicio Systemd
```ini
# /etc/systemd/system/goscim-ad-connector.service
[Unit]
Description=GoSCIM Active Directory Connector
After=network.target

[Service]
Type=simple
User=goscim
Group=goscim
WorkingDirectory=/opt/goscim-connectors/ad
ExecStart=/opt/goscim-connectors/ad/ad-connector
Restart=always
RestartSec=10
Environment=AD_BIND_PASSWORD=SecurePassword123
Environment=SCIM_AUTH_TOKEN=bearer-token-here

[Install]
WantedBy=multi-user.target
```

## Integración con LDAP

### Cliente LDAP Generic

```go
// ldap-connector/ldap_client.go
package main

import (
    "crypto/tls"
    "fmt"
    
    "github.com/go-ldap/ldap/v3"
)

type LDAPClient struct {
    conn   *ldap.Conn
    config *LDAPConfig
}

type LDAPConfig struct {
    Server     string `yaml:"server"`
    Port       int    `yaml:"port"`
    UseTLS     bool   `yaml:"use_tls"`
    BindDN     string `yaml:"bind_dn"`
    BindPass   string `yaml:"bind_password"`
    BaseDN     string `yaml:"base_dn"`
    UserFilter string `yaml:"user_filter"`
    GroupFilter string `yaml:"group_filter"`
}

func NewLDAPClient(config *LDAPConfig) (*LDAPClient, error) {
    var conn *ldap.Conn
    var err error
    
    if config.UseTLS {
        conn, err = ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", config.Server, config.Port), &tls.Config{})
    } else {
        conn, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", config.Server, config.Port))
    }
    
    if err != nil {
        return nil, err
    }
    
    err = conn.Bind(config.BindDN, config.BindPass)
    if err != nil {
        return nil, err
    }
    
    return &LDAPClient{conn: conn, config: config}, nil
}

func (c *LDAPClient) GetUsers() ([]*LDAPUser, error) {
    searchRequest := ldap.NewSearchRequest(
        c.config.BaseDN,
        ldap.ScopeWholeSubtree,
        ldap.NeverDerefAliases,
        0, 0, false,
        c.config.UserFilter,
        []string{"uid", "cn", "givenName", "sn", "mail", "telephoneNumber", "ou"},
        nil,
    )
    
    result, err := c.conn.Search(searchRequest)
    if err != nil {
        return nil, err
    }
    
    var users []*LDAPUser
    for _, entry := range result.Entries {
        user := &LDAPUser{
            DN:           entry.DN,
            UID:          entry.GetAttributeValue("uid"),
            CN:           entry.GetAttributeValue("cn"),
            GivenName:    entry.GetAttributeValue("givenName"),
            Surname:      entry.GetAttributeValue("sn"),
            Email:        entry.GetAttributeValue("mail"),
            Phone:        entry.GetAttributeValue("telephoneNumber"),
            Organization: entry.GetAttributeValue("ou"),
        }
        users = append(users, user)
    }
    
    return users, nil
}
```

## Integración con Aplicaciones SaaS

### Salesforce

#### 1. Configuración OAuth 2.0
```go
// salesforce-connector/auth.go
package main

import (
    "encoding/json"
    "net/http"
    "net/url"
    "strings"
)

type SalesforceAuth struct {
    ClientID     string
    ClientSecret string
    Username     string
    Password     string
    SecurityToken string
    LoginURL     string
}

type SalesforceToken struct {
    AccessToken string `json:"access_token"`
    InstanceURL string `json:"instance_url"`
    TokenType   string `json:"token_type"`
}

func (auth *SalesforceAuth) GetAccessToken() (*SalesforceToken, error) {
    data := url.Values{}
    data.Set("grant_type", "password")
    data.Set("client_id", auth.ClientID)
    data.Set("client_secret", auth.ClientSecret)
    data.Set("username", auth.Username)
    data.Set("password", auth.Password+auth.SecurityToken)
    
    resp, err := http.Post(
        auth.LoginURL+"/services/oauth2/token",
        "application/x-www-form-urlencoded",
        strings.NewReader(data.Encode()),
    )
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var token SalesforceToken
    err = json.NewDecoder(resp.Body).Decode(&token)
    return &token, err
}
```

#### 2. Sincronización de Usuarios
```go
// salesforce-connector/user_sync.go
func (sf *SalesforceConnector) SyncUserToSalesforce(scimUser *SCIMUser) error {
    token, err := sf.auth.GetAccessToken()
    if err != nil {
        return err
    }
    
    // Convertir usuario SCIM a formato Salesforce
    sfUser := map[string]interface{}{
        "Username":   scimUser.UserName,
        "Email":      scimUser.Emails[0].Value,
        "FirstName":  scimUser.Name.GivenName,
        "LastName":   scimUser.Name.FamilyName,
        "IsActive":   scimUser.Active,
        "ProfileId":  sf.getProfileId(scimUser),
    }
    
    // Verificar si usuario existe
    existingUser, err := sf.findUserByEmail(scimUser.Emails[0].Value, token)
    if err != nil {
        return err
    }
    
    if existingUser != nil {
        // Actualizar usuario existente
        return sf.updateUser(existingUser.ID, sfUser, token)
    } else {
        // Crear nuevo usuario
        return sf.createUser(sfUser, token)
    }
}

func (sf *SalesforceConnector) createUser(user map[string]interface{}, token *SalesforceToken) error {
    jsonData, _ := json.Marshal(user)
    
    req, _ := http.NewRequest("POST", 
        token.InstanceURL+"/services/data/v52.0/sobjects/User",
        bytes.NewBuffer(jsonData))
    
    req.Header.Set("Authorization", "Bearer "+token.AccessToken)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("failed to create user in Salesforce: %d", resp.StatusCode)
    }
    
    return nil
}
```

### Slack

#### 1. Integración con Slack API
```go
// slack-connector/slack_client.go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type SlackClient struct {
    Token   string
    BaseURL string
}

type SlackUser struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    RealName string `json:"real_name"`
    IsActive bool   `json:"deleted"`
}

func NewSlackClient(token string) *SlackClient {
    return &SlackClient{
        Token:   token,
        BaseURL: "https://slack.com/api",
    }
}

func (c *SlackClient) CreateUser(user *SlackUser) error {
    payload := map[string]interface{}{
        "user": map[string]interface{}{
            "email":      user.Email,
            "first_name": extractFirstName(user.RealName),
            "last_name":  extractLastName(user.RealName),
        },
    }
    
    return c.makeAPICall("POST", "/admin.users.invite", payload)
}

func (c *SlackClient) UpdateUser(userID string, user *SlackUser) error {
    payload := map[string]interface{}{
        "user_id": userID,
        "user": map[string]interface{}{
            "email":     user.Email,
            "real_name": user.RealName,
        },
    }
    
    return c.makeAPICall("POST", "/admin.users.set", payload)
}

func (c *SlackClient) DeactivateUser(userID string) error {
    payload := map[string]interface{}{
        "user_id": userID,
    }
    
    return c.makeAPICall("POST", "/admin.users.remove", payload)
}

func (c *SlackClient) makeAPICall(method, endpoint string, payload map[string]interface{}) error {
    jsonData, _ := json.Marshal(payload)
    
    req, _ := http.NewRequest(method, c.BaseURL+endpoint, bytes.NewBuffer(jsonData))
    req.Header.Set("Authorization", "Bearer "+c.Token)
    req.Header.Set("Content-Type", "application/json")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        return fmt.Errorf("Slack API error: %d", resp.StatusCode)
    }
    
    return nil
}
```

## Webhooks y Notificaciones

### Sistema de Webhooks

#### 1. Estructura de Webhook
```go
// webhooks/webhook.go
package webhooks

import (
    "bytes"
    "encoding/json"
    "net/http"
    "time"
)

type WebhookEvent struct {
    EventType   string                 `json:"eventType"`
    Timestamp   time.Time              `json:"timestamp"`
    Resource    string                 `json:"resource"`
    Action      string                 `json:"action"`
    ResourceID  string                 `json:"resourceId"`
    Data        map[string]interface{} `json:"data"`
    Changes     []Change               `json:"changes,omitempty"`
}

type Change struct {
    Field    string      `json:"field"`
    OldValue interface{} `json:"oldValue"`
    NewValue interface{} `json:"newValue"`
}

type WebhookSubscription struct {
    ID          string   `json:"id"`
    URL         string   `json:"url"`
    Events      []string `json:"events"`
    Secret      string   `json:"secret"`
    Active      bool     `json:"active"`
    CreatedAt   time.Time `json:"createdAt"`
}

type WebhookManager struct {
    subscriptions map[string]*WebhookSubscription
}

func (wm *WebhookManager) SendEvent(event *WebhookEvent) {
    for _, subscription := range wm.subscriptions {
        if subscription.Active && wm.shouldSendEvent(subscription, event) {
            go wm.sendWebhook(subscription, event)
        }
    }
}

func (wm *WebhookManager) sendWebhook(subscription *WebhookSubscription, event *WebhookEvent) {
    jsonData, _ := json.Marshal(event)
    
    req, _ := http.NewRequest("POST", subscription.URL, bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-SCIM-Event", event.EventType)
    req.Header.Set("X-SCIM-Signature", wm.generateSignature(jsonData, subscription.Secret))
    
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Failed to send webhook to %s: %v", subscription.URL, err)
        return
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        log.Printf("Webhook failed with status %d for %s", resp.StatusCode, subscription.URL)
    }
}
```

#### 2. Integración en Operaciones SCIM
```go
// scim/op_create.go - agregar notificación
func Create(resource string) func(c *gin.Context) {
    return func(c *gin.Context) {
        // ... lógica de creación existente ...
        
        // Notificar via webhook
        event := &webhooks.WebhookEvent{
            EventType:  "user.created",
            Timestamp:  time.Now(),
            Resource:   "User",
            Action:     "create",
            ResourceID: newUser.ID,
            Data:       newUser,
        }
        
        webhookManager.SendEvent(event)
        
        c.JSON(http.StatusCreated, newUser)
    }
}
```

### Configuración de Webhooks

#### API para Gestión de Webhooks
```go
// webhooks/api.go
func RegisterWebhookRoutes(r *gin.Engine, wm *WebhookManager) {
    webhooks := r.Group("/webhooks")
    
    webhooks.POST("/subscriptions", createSubscription(wm))
    webhooks.GET("/subscriptions", listSubscriptions(wm))
    webhooks.GET("/subscriptions/:id", getSubscription(wm))
    webhooks.PUT("/subscriptions/:id", updateSubscription(wm))
    webhooks.DELETE("/subscriptions/:id", deleteSubscription(wm))
}

func createSubscription(wm *WebhookManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        var subscription WebhookSubscription
        if err := c.ShouldBindJSON(&subscription); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        subscription.ID = generateID()
        subscription.CreatedAt = time.Now()
        subscription.Secret = generateSecret()
        
        wm.AddSubscription(&subscription)
        
        c.JSON(201, subscription)
    }
}
```

## Sincronización Bidireccional

### Gestor de Sincronización
```go
// sync/manager.go
package sync

import (
    "time"
    "sync"
)

type SyncManager struct {
    connectors map[string]Connector
    conflicts  *ConflictResolver
    scheduler  *Scheduler
    mutex      sync.RWMutex
}

type Connector interface {
    GetUsers() ([]User, error)
    CreateUser(user User) error
    UpdateUser(user User) error
    DeleteUser(userID string) error
    GetLastSync() time.Time
    SetLastSync(timestamp time.Time) error
}

type ConflictResolver struct {
    strategy ConflictStrategy
}

type ConflictStrategy int

const (
    LastWriterWins ConflictStrategy = iota
    ManualResolution
    SourcePriority
)

func (sm *SyncManager) SyncAll() error {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()
    
    for name, connector := range sm.connectors {
        log.Printf("Starting sync for connector: %s", name)
        
        if err := sm.syncConnector(connector); err != nil {
            log.Printf("Sync failed for %s: %v", name, err)
            continue
        }
        
        log.Printf("Sync completed for connector: %s", name)
    }
    
    return nil
}

func (sm *SyncManager) syncConnector(connector Connector) error {
    lastSync := connector.GetLastSync()
    
    // Obtener cambios desde la última sincronización
    users, err := connector.GetUsers()
    if err != nil {
        return err
    }
    
    // Procesar cada usuario
    for _, user := range users {
        if user.ModifiedAt.After(lastSync) {
            if err := sm.processUserChange(user, connector); err != nil {
                log.Printf("Failed to process user %s: %v", user.ID, err)
            }
        }
    }
    
    // Actualizar timestamp de última sincronización
    return connector.SetLastSync(time.Now())
}

func (sm *SyncManager) processUserChange(user User, source Connector) error {
    // Verificar conflictos con otras fuentes
    conflicts := sm.detectConflicts(user)
    if len(conflicts) > 0 {
        return sm.conflicts.Resolve(user, conflicts)
    }
    
    // Sincronizar a todas las otras fuentes
    for name, connector := range sm.connectors {
        if connector == source {
            continue // Skip source connector
        }
        
        if err := connector.UpdateUser(user); err != nil {
            log.Printf("Failed to sync user %s to %s: %v", user.ID, name, err)
        }
    }
    
    return nil
}
```

## Configuración de Integraciones

### Archivo de Configuración Principal
```yaml
# config/integrations.yaml
integrations:
  active_directory:
    enabled: true
    connector_type: "ldap"
    config:
      server: "dc01.empresa.local"
      port: 636
      use_tls: true
      bind_dn: "CN=SCIM Service,OU=Service Accounts,DC=empresa,DC=local"
      bind_password_env: "AD_BIND_PASSWORD"
      base_dn: "OU=Users,DC=empresa,DC=local"
      sync_interval: "15m"
    
  salesforce:
    enabled: true
    connector_type: "salesforce"
    config:
      client_id: "${SALESFORCE_CLIENT_ID}"
      client_secret: "${SALESFORCE_CLIENT_SECRET}"
      username: "integration@empresa.com"
      password_env: "SALESFORCE_PASSWORD"
      security_token_env: "SALESFORCE_TOKEN"
      sandbox: false
      sync_interval: "30m"
    
  slack:
    enabled: true
    connector_type: "slack"
    config:
      token_env: "SLACK_BOT_TOKEN"
      workspace: "empresa"
      sync_interval: "1h"
    
  okta:
    enabled: false
    connector_type: "okta"
    config:
      domain: "empresa.okta.com"
      api_token_env: "OKTA_API_TOKEN"
      sync_interval: "20m"

webhooks:
  enabled: true
  subscriptions:
    - url: "https://app1.empresa.com/webhooks/scim"
      events: ["user.created", "user.updated", "user.deleted"]
      secret_env: "APP1_WEBHOOK_SECRET"
    
    - url: "https://app2.empresa.com/api/identity/webhook"
      events: ["user.created", "group.updated"]
      secret_env: "APP2_WEBHOOK_SECRET"

sync_settings:
  conflict_resolution: "last_writer_wins"
  batch_size: 100
  max_retries: 3
  retry_delay: "5s"
  enable_audit_log: true
```

### Cargador de Configuración
```go
// config/loader.go
package config

import (
    "io/ioutil"
    "os"
    "strings"
    
    "gopkg.in/yaml.v2"
)

type IntegrationConfig struct {
    Integrations map[string]Integration `yaml:"integrations"`
    Webhooks     WebhookConfig          `yaml:"webhooks"`
    SyncSettings SyncSettings           `yaml:"sync_settings"`
}

func LoadIntegrationConfig(path string) (*IntegrationConfig, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    // Expandir variables de entorno
    content := os.ExpandEnv(string(data))
    
    var config IntegrationConfig
    err = yaml.Unmarshal([]byte(content), &config)
    if err != nil {
        return nil, err
    }
    
    // Cargar secretos desde variables de entorno
    config.loadSecrets()
    
    return &config, nil
}

func (c *IntegrationConfig) loadSecrets() {
    for name, integration := range c.Integrations {
        config := integration.Config
        
        // Cargar passwords desde variables de entorno
        for key, value := range config {
            if strings.HasSuffix(key, "_env") {
                envVar := value.(string)
                secretKey := strings.TrimSuffix(key, "_env")
                config[secretKey] = os.Getenv(envVar)
                delete(config, key)
            }
        }
        
        integration.Config = config
        c.Integrations[name] = integration
    }
}
```

Esta guía de integraciones proporciona una base sólida para conectar GoSCIM con los sistemas de identidad más comunes en el ámbito empresarial, facilitando la automatización y sincronización de identidades across multiple platforms.