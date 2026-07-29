package scim

import (
	"time"

	"github.com/google/uuid"
)

func generateMeta(element map[string]interface{}, resourceType ResoruceType) Meta {
	now := time.Now()
	meta := Meta{}
	meta.ResourceType = resourceType.Name
	meta.Created = now.Format(time.RFC3339)
	meta.LastModified = meta.Created
	meta.Version = uuid.New().String()
	meta.Location = resourceLocation(resourceType, element)
	return meta
}

// updateMeta stamps a new version and lastModified while carrying the original
// created forward. metaOld is whatever the stored resource had, which may be
// missing or malformed on a document written by an older version, so it is read
// defensively rather than asserted.
func updateMeta(metaOld map[string]interface{}, element map[string]interface{}, resourceType ResoruceType) Meta {
	now := time.Now()
	meta := Meta{}
	meta.LastModified = now.Format(time.RFC3339)
	meta.Version = uuid.New().String()
	meta.ResourceType = resourceType.Name
	meta.Location = resourceLocation(resourceType, element)

	created, _ := metaOld["created"].(string)
	if created == "" {
		created = meta.LastModified
	}
	meta.Created = created

	return meta
}

func resourceLocation(resourceType ResoruceType, element map[string]interface{}) string {
	id, _ := element["id"].(string)
	return resourceType.Endpoint + "/" + id
}
