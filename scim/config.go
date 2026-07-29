package scim

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const (

	// FolderSchema is ..
	FolderSchema = "/schemas/"
	// FolderResourceType is ..
	FolderResourceType = "/resourceType/"
	PREFIX             = "/scim/v2"
)

var (
	// Resources ...
	Resources map[string]ResourceType
	// Schemas ...
	Schemas map[string]Schema
)

// GetAttribute resolves a dotted attribute path against a schema's attributes.
// Names are matched ignoring case, as RFC 7643 2.1 requires.
func GetAttribute(attributes []Attribute, path string) Attribute {
	fields := strings.Split(path, ".")
	finalName := fields[len(fields)-1]
	for i := 0; i < len(fields)-1; i++ {
		if attribute, ok := FindAttribute(attributes, fields[i]); ok {
			attributes = attribute.SubAttributes
		}
	}
	attribute, _ := FindAttribute(attributes, finalName)
	return attribute
}

// ReadResourceType loads every schema and resource type under folderConfig,
// provisions one bucket per resource type and registers its SCIM routes on the
// given router, which is expected to already carry the authentication
// middleware and the /scim/v2 prefix.
func ReadResourceType(folderConfig string, r gin.IRouter) error {
	Resources = make(map[string]ResourceType)
	Schemas = make(map[string]Schema)

	if err := ReadSchemas(folderConfig); err != nil {
		return err
	}

	files, err := os.ReadDir(folderConfig + FolderResourceType)
	if err != nil {
		return err
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		file, err := os.ReadFile(folderConfig + FolderResourceType + f.Name())
		if err != nil {
			return err
		}
		resourceType := ResourceType{}
		if err := json.Unmarshal(file, &resourceType); err != nil {
			return err
		}
		Resources[resourceType.Endpoint] = resourceType
		if err := DB.EnsureBucket(resourceType.Name); err != nil {
			return err
		}

		r.POST(resourceType.Endpoint, Create(resourceType.Endpoint))          // Create:  	POST https://example.com/{v}/{resource}
		r.GET(resourceType.Endpoint+"/:id", Read(resourceType.Endpoint))      // Read: 	GET https://example.com/{v}/{resource}/{id}
		r.PUT(resourceType.Endpoint+"/:id", Replace(resourceType.Endpoint))   // Replace: 	PUT https://example.com/{v}/{resource}/{id}
		r.DELETE(resourceType.Endpoint+"/:id", Delete(resourceType.Endpoint)) // Delete: 	DELETE https://example.com/{v}/{resource}/{id}
		r.PATCH(resourceType.Endpoint+"/:id", Update(resourceType.Endpoint))  // Update: 	PATCH https://example.com/{v}/{resource}/{id}
		r.GET(resourceType.Endpoint, Search(resourceType.Endpoint))           // Search: 	GET https://example.com/{v}/{resource}?ﬁlter={attribute}{op}{value}&sortBy={attributeName}&sortOrder={ascending|descending}
	}
	return nil
}

// ReadSchemas fills the Schemas map from folderConfig/schemas.
func ReadSchemas(folderConfig string) error {
	if Schemas == nil {
		Schemas = make(map[string]Schema)
	}
	files, err := os.ReadDir(folderConfig + FolderSchema)
	if err != nil {
		return err
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		if err := addSchema(folderConfig, f.Name()); err != nil {
			return err
		}
	}
	return nil
}

// add schema if no exist in schemas
func addSchema(folderConfig string, schemaName string) error {
	file, err := os.ReadFile(folderConfig + FolderSchema + schemaName)
	if err != nil {
		return err
	}
	schema := Schema{}
	if err := json.Unmarshal(file, &schema); err != nil {
		return err
	}
	Schemas[schema.ID] = schema
	log.Println(schema.ID)
	return nil
}
