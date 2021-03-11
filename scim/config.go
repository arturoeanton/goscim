package scim

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
)

const (

	// FolderSchema is ..
	FolderSchema = "/schemas/"
	// FolderResoruceType is ..
	FolderResoruceType = "/resourceType/"
)

var (
	// Resources ...
	Resources map[string]ResoruceType
	// Schemas ...
	Schemas map[string]Schema
)

// ReadResourceType read all file in resourceType
func ReadResourceType(folderConfig string) {
	Resources = make(map[string]ResoruceType)
	Schemas = make(map[string]Schema)
	files, err := ioutil.ReadDir(folderConfig + FolderResoruceType)
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
			addSchema(folderConfig, resourceType.Schema)
			CreateBucket(resourceType.Name)
			for _, schemaExtension := range resourceType.SchemaExtensions {
				addSchema(folderConfig, schemaExtension.Schema)
			}
		}
	}
}

// add schema if no exist in schemas
func addSchema(folderConfig string, schemaName string) {
	if _, ok := Schemas[schemaName]; ok {
		return
	}
	file, err := ioutil.ReadFile(folderConfig + FolderSchema + schemaName + ".json")
	if err != nil {
		log.Fatal(err)
	}
	schema := Schema{}
	err = json.Unmarshal([]byte(file), &schema)
	if err != nil {
		log.Fatal(err.Error())
	}
	Schemas[schemaName] = schema
}
