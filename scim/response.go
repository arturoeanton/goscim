package scim

import (
	"net/http"
	"os"
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
	c.Header("X-Content-Type-Options", "nosniff")
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
// When SCIM_PUBLIC_URL is set (for deployments behind a TLS-terminating proxy),
// it supplies the scheme and authority; otherwise they are derived from the
// request after rejecting host values carrying path, userinfo or control
// characters.
func absoluteLocation(c *gin.Context, resourceType ResourceType, id string) string {
	path := resourcePath(resourceType, id)
	if base := strings.TrimRight(strings.TrimSpace(os.Getenv("SCIM_PUBLIC_URL")), "/"); base != "" {
		return base + path
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := safeHost(c.Request.Host)
	if host == "" {
		return path
	}
	return scheme + "://" + host + path
}

func safeHost(host string) string {
	if host == "" || strings.ContainsAny(host, "/\\@ \t\r\n") {
		return ""
	}
	for i := 0; i < len(host); i++ {
		if host[i] < 0x20 || host[i] == 0x7f {
			return ""
		}
	}
	return host
}

// resourcePath is the server-relative path of a resource, which is what
// meta.location carries.
func resourcePath(resourceType ResourceType, id string) string {
	return PREFIX + resourceType.Endpoint + "/" + id
}
