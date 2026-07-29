package scim

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"

	"github.com/gin-gonic/gin"
)

// FolderServiceProviderConfig holds the optional deployment-specific parts of
// the ServiceProviderConfig, currently just documentationUri.
var FolderServiceProviderConfig = "/serviceProviderConfig/"

// supported is the RFC 7643 5 "supported" complex value.
type supported struct {
	Supported bool `json:"supported"`
}

type filterSupport struct {
	Supported  bool `json:"supported"`
	MaxResults int  `json:"maxResults"`
}

// AuthenticationScheme describes one way of authenticating, for the
// ServiceProviderConfig (RFC 7643 5).
type AuthenticationScheme struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SpecURI     string `json:"specUri,omitempty"`
	Primary     bool   `json:"primary,omitempty"`
}

// ServiceProviderConfig is the document served at /ServiceProviderConfig.
type ServiceProviderConfig struct {
	Schemas               []string               `json:"schemas"`
	DocumentationURI      string                 `json:"documentationUri,omitempty"`
	Patch                 supported              `json:"patch"`
	Bulk                  supported              `json:"bulk"`
	Filter                filterSupport          `json:"filter"`
	ChangePassword        supported              `json:"changePassword"`
	Sort                  supported              `json:"sort"`
	ETag                  supported              `json:"etag"`
	AuthenticationSchemes []AuthenticationScheme `json:"authenticationSchemes"`
	Meta                  Meta                   `json:"meta"`
}

// discoveryAuthenticator is the authenticator the discovery endpoints describe.
// It is set when the router is built.
var discoveryAuthenticator Authenticator

// buildServiceProviderConfig describes what this server actually does.
//
// It is derived from the code rather than read from a file. The file that used
// to sit unread in config/ advertised bulk support, which does not exist: a
// provider config that can drift from the implementation is worse than none,
// because a client believes it.
func buildServiceProviderConfig(folderConfig string) ServiceProviderConfig {
	config := ServiceProviderConfig{
		Schemas:        []string{"urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"},
		Patch:          supported{Supported: true},
		Bulk:           supported{Supported: false}, // /Bulk is not implemented
		Filter:         filterSupport{Supported: true, MaxResults: maxSearchCount},
		ChangePassword: supported{Supported: false},
		Sort:           supported{Supported: true},
		ETag:           supported{Supported: true},
		Meta: Meta{
			ResourceType: "ServiceProviderConfig",
			Location:     PREFIX + "/ServiceProviderConfig",
		},
	}
	if discoveryAuthenticator != nil {
		config.AuthenticationSchemes = discoveryAuthenticator.AuthenticationSchemes()
	}
	if config.AuthenticationSchemes == nil {
		config.AuthenticationSchemes = []AuthenticationScheme{}
	}

	// Only the deployment-specific parts still come from the file.
	var fromFile struct {
		DocumentationURI string `json:"documentationUri"`
	}
	if raw, err := os.ReadFile(folderConfig + FolderServiceProviderConfig + "sp_config.json"); err == nil {
		if err := json.Unmarshal(raw, &fromFile); err == nil {
			config.DocumentationURI = fromFile.DocumentationURI
		}
	}
	return config
}

// DiscoveryServiceProviderConfig is GET /ServiceProviderConfig: which parts of
// the specification this deployment implements.
func DiscoveryServiceProviderConfig(folderConfig string) gin.HandlerFunc {
	config := buildServiceProviderConfig(folderConfig)
	return func(c *gin.Context) {
		writeSCIM(c, http.StatusOK, config)
	}
}

// DiscoveryResourceTypes is GET /ResourceTypes: the resource types available.
func DiscoveryResourceTypes(c *gin.Context) {
	endpoints := make([]string, 0, len(Resources))
	for endpoint := range Resources {
		endpoints = append(endpoints, endpoint)
	}
	sort.Strings(endpoints)

	items := make([]interface{}, 0, len(endpoints))
	for _, endpoint := range endpoints {
		items = append(items, describeResourceType(Resources[endpoint]))
	}
	writeSCIM(c, http.StatusOK, listOf(items))
}

// DiscoveryResourceType is GET /ResourceTypes/{id}.
func DiscoveryResourceType(c *gin.Context) {
	id := c.Param("id")
	for _, resourceType := range Resources {
		if resourceType.ID == id || resourceType.Name == id {
			writeSCIM(c, http.StatusOK, describeResourceType(resourceType))
			return
		}
	}
	MakeError(c, http.StatusNotFound, "no resource type "+id)
}

// DiscoverySchemas is GET /Schemas: the schemas and attribute extensions.
func DiscoverySchemas(c *gin.Context) {
	ids := make([]string, 0, len(Schemas))
	for id := range Schemas {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	items := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		items = append(items, describeSchema(Schemas[id]))
	}
	writeSCIM(c, http.StatusOK, listOf(items))
}

// DiscoverySchema is GET /Schemas/{id}.
func DiscoverySchema(c *gin.Context) {
	schema, ok := Schemas[c.Param("id")]
	if !ok {
		MakeError(c, http.StatusNotFound, "no schema "+c.Param("id"))
		return
	}
	writeSCIM(c, http.StatusOK, describeSchema(schema))
}

// listOf wraps discovery results in a SCIM ListResponse.
func listOf(items []interface{}) ListResponse {
	return ListResponse{
		Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		TotalResults: len(items),
		ItemsPerPage: len(items),
		StartIndex:   1,
		Resources:    items,
	}
}

func describeResourceType(resourceType ResourceType) ResourceType {
	described := resourceType
	described.Schemas = []string{"urn:ietf:params:scim:schemas:core:2.0:ResourceType"}
	described.Meta = Meta{
		ResourceType: "ResourceType",
		Location:     PREFIX + "/ResourceTypes/" + resourceType.Name,
	}
	return described
}

// describedSchema mirrors Schema without the $reader and $writer lists: they
// are this project's own authorization extension, not part of the schema a
// client is meant to consume.
type describedSchema struct {
	Schemas     []string             `json:"schemas"`
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	Attributes  []describedAttribute `json:"attributes"`
	Meta        Meta                 `json:"meta"`
}

type describedAttribute struct {
	Name          string               `json:"name"`
	Type          string               `json:"type"`
	MultiValued   bool                 `json:"multiValued"`
	Description   string               `json:"description,omitempty"`
	Required      bool                 `json:"required"`
	CaseExact     bool                 `json:"caseExact"`
	Mutability    string               `json:"mutability,omitempty"`
	Returned      string               `json:"returned,omitempty"`
	Uniqueness    string               `json:"uniqueness,omitempty"`
	SubAttributes []describedAttribute `json:"subAttributes,omitempty"`
}

func describeSchema(schema Schema) describedSchema {
	return describedSchema{
		Schemas:     []string{"urn:ietf:params:scim:schemas:core:2.0:Schema"},
		ID:          schema.ID,
		Name:        schema.Name,
		Description: schema.Description,
		Attributes:  describeAttributes(schema.Attributes),
		Meta: Meta{
			ResourceType: "Schema",
			Location:     PREFIX + "/Schemas/" + schema.ID,
		},
	}
}

func describeAttributes(attributes []Attribute) []describedAttribute {
	described := make([]describedAttribute, 0, len(attributes))
	for _, attribute := range attributes {
		described = append(described, describedAttribute{
			Name:          attribute.Name,
			Type:          attribute.Type,
			MultiValued:   attribute.MultiValued,
			Description:   attribute.Description,
			Required:      attribute.Required,
			CaseExact:     attribute.CaseExact,
			Mutability:    attribute.Mutability,
			Returned:      attribute.Returned,
			Uniqueness:    attribute.Uniqueness,
			SubAttributes: describeAttributes(attribute.SubAttributes),
		})
	}
	return described
}
