package meta

import (
	"time"

	"github.com/arturoeanton/goscim/scim/types"
	"github.com/google/uuid"
)

// GenerateMeta ...
func GenerateMeta(element map[string]interface{}, resourceType types.ResoruceType) types.Meta {
	now := time.Now()
	meta := types.Meta{}
	meta.ResourceType = resourceType.Name
	meta.Created = now.Format(time.RFC3339)
	meta.LastModified = meta.Created
	meta.Version = uuid.New().String()
	meta.Location = resourceType.Endpoint + "/" + element["id"].(string)
	return meta
}
