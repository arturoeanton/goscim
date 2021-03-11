package scim

import (
	"time"

	"github.com/google/uuid"
)

// GenerateMeta ...
func GenerateMeta(element map[string]interface{}, resourceType ResoruceType) Meta {
	now := time.Now()
	meta := Meta{}
	meta.ResourceType = resourceType.Name
	meta.Created = now.Format(time.RFC3339)
	meta.LastModified = meta.Created
	meta.Version = uuid.New().String()
	meta.Location = resourceType.Endpoint + "/" + element["id"].(string)
	return meta
}
