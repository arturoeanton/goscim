package scim

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// scimContentType is the media type RFC 7644 3.1 defines for SCIM payloads.
const scimContentType = "application/scim+json"

// writeSCIM renders a SCIM body with the SCIM media type. gin only sets
// Content-Type when the header is not already present, so setting it up front
// wins over the application/json it would otherwise write.
func writeSCIM(c *gin.Context, status int, body interface{}) {
	c.Header("Content-Type", scimContentType)
	c.JSON(status, body)
}

// entityTag formats a resource version as a weak ETag. The version is a random
// identifier rather than a digest of the content, so the comparison it supports
// is "is this the same revision", which is exactly what weak means.
func entityTag(version string) string {
	if version == "" {
		return ""
	}
	return `W/"` + version + `"`
}

// versionOf reads meta.version out of a stored resource.
func versionOf(element map[string]interface{}) string {
	meta, ok := element["meta"].(map[string]interface{})
	if !ok {
		return ""
	}
	version, _ := meta["version"].(string)
	return version
}

// etagMatches reports whether an If-Match or If-None-Match header covers
// version. The bare version is accepted alongside its quoted and weak forms so
// that a client echoing meta.version verbatim interoperates.
func etagMatches(header, version string) bool {
	if version == "" {
		return false
	}
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" {
			return true
		}
		candidate = strings.TrimPrefix(candidate, "W/")
		if strings.Trim(candidate, `"`) == version {
			return true
		}
	}
	return false
}

// checkPrecondition enforces If-Match on a write. It reports whether the caller
// may proceed; when it returns false the response has already been written.
//
// Without this two concurrent updates silently overwrite each other: every
// write stamps a fresh meta.version but nothing ever compared it.
func checkPrecondition(c *gin.Context, stored map[string]interface{}) bool {
	ifMatch := c.GetHeader("If-Match")
	if ifMatch == "" {
		return true
	}
	if etagMatches(ifMatch, versionOf(stored)) {
		return true
	}
	MakeError(c, http.StatusPreconditionFailed,
		"the resource has been modified since the version given in If-Match")
	return false
}

// setVersionHeader publishes the resource's revision as an ETag.
func setVersionHeader(c *gin.Context, meta Meta) {
	if tag := entityTag(meta.Version); tag != "" {
		c.Header("ETag", tag)
	}
}

// absoluteLocation builds the URI of a resource for the Location header.
//
// The scheme is taken from the connection, so a deployment behind a
// TLS-terminating proxy needs the public base URL to be configurable; that is
// still open.
func absoluteLocation(c *gin.Context, resourceType ResourceType, id string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host + resourcePath(resourceType, id)
}

// resourcePath is the server-relative path of a resource, which is what
// meta.location carries.
func resourcePath(resourceType ResourceType, id string) string {
	return PREFIX + resourceType.Endpoint + "/" + id
}
