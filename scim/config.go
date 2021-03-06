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

// ReadResourceType read all file in resourceType
func ReadResourceType(folderConfig string, r *gin.Engine) {
	Resources = make(map[string]ResoruceType)
	Schemas = make(map[string]Schema)

	files, err := ioutil.ReadDir(folderConfig + FolderSchema)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			addSchema(folderConfig, f.Name())
		}
	}

	files, err = ioutil.ReadDir(folderConfig + FolderResoruceType)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			file, err := ioutil.ReadFile(folderConfig + FolderResoruceType + f.Name())
			if err != nil {
				log.Fatal(err.Error())
			}
			resourceType := ResoruceType{}
			err = json.Unmarshal([]byte(file), &resourceType)
			if err != nil {
				log.Fatal(err.Error())
			}
			Resources[resourceType.Endpoint] = resourceType
			CreateBucket(resourceType.Name)

			r.POST(PREFIX+resourceType.Endpoint, Create(resourceType.Endpoint))          // Create:  	POST https://example.com/{v}/{resource}
			r.GET(PREFIX+resourceType.Endpoint+"/:id", Read(resourceType.Endpoint))      // Read: 	GET https://example.com/{v}/{resource}/{id}
			r.PUT(PREFIX+resourceType.Endpoint+"/:id", Replace(resourceType.Endpoint))   // Replace: 	PUT https://example.com/{v}/{resource}/{id}
			r.DELETE(PREFIX+resourceType.Endpoint+"/:id", Delete(resourceType.Endpoint)) // Delete: 	DELETE https://example.com/{v}/{resource}/{id}
			r.PATCH(PREFIX+resourceType.Endpoint+"/:id", Update(resourceType.Endpoint))  // Update: 	PATCH https://example.com/{v}/{resource}/{id}
			r.GET(PREFIX+resourceType.Endpoint, Search(resourceType.Endpoint))           // Search: 	GET https://example.com/{v}/{resource}????lter={attribute}{op}{value}&sortBy={attributeName}&sortOrder={ascending|descending}
		}

	}
}

// add schema if no exist in schemas
func addSchema(folderConfig string, schemaName string) {
	file, err := ioutil.ReadFile(folderConfig + FolderSchema + schemaName)
	if err != nil {
		log.Fatal(err)
	}
	schema := Schema{}
	err = json.Unmarshal([]byte(file), &schema)
	if err != nil {
		log.Fatal(err.Error())
	}
	Schemas[schema.ID] = schema
	log.Println(schema.ID)
}
