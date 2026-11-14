package scim

import (
	"time"

	"github.com/google/uuid"
)

func generateMeta(element map[string]interface{}, resourceType ResourceType) Meta {
	return generateMetaForPrefix(PREFIX, element, resourceType)
}

func generateMetaForPrefix(prefix string, element map[string]interface{}, resourceType ResourceType) Meta {
	now := time.Now()
	meta := Meta{}
	meta.ResourceType = resourceType.Name
	meta.Created = now.Format(time.RFC3339)
	meta.LastModified = meta.Created
	meta.Version = uuid.New().String()
	id, _ := element["id"].(string)
	meta.Location = resourcePathWithPrefix(prefix, resourceType, id)
	return meta
}

// updateMeta stamps a new version and lastModified while carrying the original
// created forward. metaOld is whatever the stored resource had, which may be
// missing or malformed on a document written by an older version, so it is read
// defensively rather than asserted.
func updateMeta(metaOld map[string]interface{}, element map[string]interface{}, resourceType ResourceType) Meta {
	return updateMetaForPrefix(PREFIX, metaOld, element, resourceType)
}

func updateMetaForPrefix(prefix string, metaOld map[string]interface{}, element map[string]interface{}, resourceType ResourceType) Meta {
	now := time.Now()
	meta := Meta{}
	meta.LastModified = now.Format(time.RFC3339)
	meta.Version = uuid.New().String()
	meta.ResourceType = resourceType.Name
	id, _ := element["id"].(string)
	meta.Location = resourcePathWithPrefix(prefix, resourceType, id)

	created, _ := metaOld["created"].(string)
	if created == "" {
		created = meta.LastModified
	}
	meta.Created = created

	return meta
}

func resourceLocation(resourceType ResourceType, element map[string]interface{}) string {
	id, _ := element["id"].(string)
	return resourcePath(resourceType, id)
}
