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
	meta.Location = resourceType.Endpoint + "/" + element["id"].(string)
	return meta
}

func updateMeta(metaOld map[string]interface{}, element map[string]interface{}, resourceType ResoruceType) Meta {
	now := time.Now()
	meta := Meta{}
	meta.LastModified = now.Format(time.RFC3339)
	meta.Version = uuid.New().String()
	meta.ResourceType = resourceType.Name
	meta.Created = metaOld["created"].(string)
	meta.Location = resourceType.Endpoint + "/" + element["id"].(string)

	return meta
}
