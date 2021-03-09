package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"

	"github.com/arturoeanton/goscim/scim/types"
)

const (

	// FolderSchema is ..
	FolderSchema = "/schemas/"
	// FolderResoruceType is ..
	FolderResoruceType = "/resourceType/"
)

var (
	resources map[string]types.ResoruceType
	schemas   map[string]types.Schema
)

// ReadResourceType read all file in resourceType
func ReadResourceType(folderConfig string) {
	resources = make(map[string]types.ResoruceType)
	schemas = make(map[string]types.Schema)
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
			resourceType := types.ResoruceType{}
			err = json.Unmarshal([]byte(file), &resourceType)
			if err != nil {
				log.Fatal(err.Error())
			}
			resources[resourceType.Endpoint] = resourceType
			addSchema(folderConfig, resourceType.Schema)
			for _, schemaExtension := range resourceType.SchemaExtensions {
				addSchema(folderConfig, schemaExtension.Schema)
			}
		}
	}
}

// add schema if no exist in schemas
func addSchema(folderConfig string, schemaName string) {
	if _, ok := schemas[schemaName]; ok {
		return
	}
	file, err := ioutil.ReadFile(folderConfig + FolderSchema + schemaName + ".json")
	if err != nil {
		log.Fatal(err)
	}
	schema := types.Schema{}
	err = json.Unmarshal([]byte(file), &schema)
	if err != nil {
		log.Fatal(err.Error())
	}
	schemas[schemaName] = schema
}