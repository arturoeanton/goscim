package scim

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

const (

	// FolderSchema is ..
	FolderSchema = "/schemas/"
	// FolderResoruceType is ..
	FolderResoruceType = "/resourceType/"
	PREFIX             = "/scim/v2"
)

var (
	// Resources ...
	Resources map[string]ResoruceType
	// Schemas ...
	Schemas map[string]Schema
)

func GetAttribute(attributes []Attribute, path string) Attribute {
	fields := strings.Split(path, ".")
	finalName := fields[len(fields)-1]
	for i := 0; i < len(fields)-1; i++ {
		name := fields[i]
		for _, attribute := range attributes {
			if attribute.Name == name {
				attributes = attribute.SubAttributes
				break
			}
		}
	}
	for _, attribute := range attributes {
		if attribute.Name == finalName {
			return attribute
		}
	}
	return Attribute{}
}

// ReadResourceType loads every schema and resource type under folderConfig,
// provisions one bucket per resource type and registers its SCIM routes.
func ReadResourceType(folderConfig string, r *gin.Engine) error {
	Resources = make(map[string]ResoruceType)
	Schemas = make(map[string]Schema)

	if err := ReadSchemas(folderConfig); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(folderConfig + FolderResoruceType)
	if err != nil {
		return err
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		file, err := ioutil.ReadFile(folderConfig + FolderResoruceType + f.Name())
		if err != nil {
			return err
		}
		resourceType := ResoruceType{}
		if err := json.Unmarshal(file, &resourceType); err != nil {
			return err
		}
		Resources[resourceType.Endpoint] = resourceType
		if err := DB.EnsureBucket(resourceType.Name); err != nil {
			return err
		}

		r.POST(PREFIX+resourceType.Endpoint, Create(resourceType.Endpoint))          // Create:  	POST https://example.com/{v}/{resource}
		r.GET(PREFIX+resourceType.Endpoint+"/:id", Read(resourceType.Endpoint))      // Read: 	GET https://example.com/{v}/{resource}/{id}
		r.PUT(PREFIX+resourceType.Endpoint+"/:id", Replace(resourceType.Endpoint))   // Replace: 	PUT https://example.com/{v}/{resource}/{id}
		r.DELETE(PREFIX+resourceType.Endpoint+"/:id", Delete(resourceType.Endpoint)) // Delete: 	DELETE https://example.com/{v}/{resource}/{id}
		r.PATCH(PREFIX+resourceType.Endpoint+"/:id", Update(resourceType.Endpoint))  // Update: 	PATCH https://example.com/{v}/{resource}/{id}
		r.GET(PREFIX+resourceType.Endpoint, Search(resourceType.Endpoint))           // Search: 	GET https://example.com/{v}/{resource}?ﬁlter={attribute}{op}{value}&sortBy={attributeName}&sortOrder={ascending|descending}
	}
	return nil
}

// ReadSchemas fills the Schemas map from folderConfig/schemas.
func ReadSchemas(folderConfig string) error {
	if Schemas == nil {
		Schemas = make(map[string]Schema)
	}
	files, err := ioutil.ReadDir(folderConfig + FolderSchema)
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
	file, err := ioutil.ReadFile(folderConfig + FolderSchema + schemaName)
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
