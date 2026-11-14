package scim

import (
	"encoding/json"
	"io/fs"
	"log"
	"os"
	"path"
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

	fallbackConfigFS fs.FS
)

// SetFallbackConfigFS registers an embedded filesystem used when the default
// "config" directory is not present in the working directory.
func SetFallbackConfigFS(fsys fs.FS) {
	fallbackConfigFS = fsys
}

func useFallbackConfig(folderConfig string, diskErr error) bool {
	return fallbackConfigFS != nil && os.IsNotExist(diskErr) &&
		(folderConfig == "" || folderConfig == "config" || folderConfig == "./config")
}

func readConfigDirEntries(folderConfig, subDir string) ([]fs.DirEntry, error) {
	entries, err := os.ReadDir(folderConfig + subDir)
	if err == nil || !useFallbackConfig(folderConfig, err) {
		return entries, err
	}
	return fs.ReadDir(fallbackConfigFS, strings.Trim(subDir, "/"))
}

func readConfigBytes(folderConfig, subDir, fileName string) ([]byte, error) {
	raw, err := os.ReadFile(folderConfig + subDir + fileName)
	if err == nil || !useFallbackConfig(folderConfig, err) {
		return raw, err
	}
	return fs.ReadFile(fallbackConfigFS, path.Join(strings.Trim(subDir, "/"), fileName))
}

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

	files, err := readConfigDirEntries(folderConfig, FolderResourceType)
	if err != nil {
		return err
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		file, err := readConfigBytes(folderConfig, FolderResourceType, f.Name())
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
		if cb, ok := DB.(*CouchbaseStore); ok {
			if err := cb.EnsureResourceIndexes(resourceType); err != nil {
				return err
			}
		}
	}
	if r != nil {
		mountResourceRoutes(r)
	}
	return nil
}

func mountResourceRoutes(r gin.IRouter) {
	for _, resourceType := range Resources {
		r.POST(resourceType.Endpoint, Create(resourceType.Endpoint))                // Create:  	POST https://example.com/{v}/{resource}
		r.GET(resourceType.Endpoint+"/:id", Read(resourceType.Endpoint))            // Read: 	GET https://example.com/{v}/{resource}/{id}
		r.PUT(resourceType.Endpoint+"/:id", Replace(resourceType.Endpoint))         // Replace: 	PUT https://example.com/{v}/{resource}/{id}
		r.DELETE(resourceType.Endpoint+"/:id", Delete(resourceType.Endpoint))       // Delete: 	DELETE https://example.com/{v}/{resource}/{id}
		r.PATCH(resourceType.Endpoint+"/:id", Update(resourceType.Endpoint))        // Update: 	PATCH https://example.com/{v}/{resource}/{id}
		r.GET(resourceType.Endpoint, Search(resourceType.Endpoint))                 // Search: 	GET https://example.com/{v}/{resource}?filter=...
		r.POST(resourceType.Endpoint+"/.search", PostSearch(resourceType.Endpoint)) // Search:	POST https://example.com/{v}/{resource}/.search
	}
}

// ReadSchemas fills the Schemas map from folderConfig/schemas.
func ReadSchemas(folderConfig string) error {
	if Schemas == nil {
		Schemas = make(map[string]Schema)
	}
	files, err := readConfigDirEntries(folderConfig, FolderSchema)
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
	file, err := readConfigBytes(folderConfig, FolderSchema, schemaName)
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
