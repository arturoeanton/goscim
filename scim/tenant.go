package scim

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	tenantContextKey    = "scim.tenant"
	urlTenantContextKey = "scim.url_tenant"
	maxTenantIDLen      = 64
	scopedBucketPrefix  = "t:"
)

// IsValidTenantID reports whether id is a safe tenant identifier
// (`[A-Za-z0-9._-]{1,64}`).
func IsValidTenantID(id string) bool {
	if len(id) == 0 || len(id) > maxTenantIDLen {
		return false
	}
	for i := 0; i < len(id); i++ {
		b := id[i]
		isAlphaNum := (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
		isAllowed := b == '-' || b == '_' || b == '.'
		if !isAlphaNum && !isAllowed {
			return false
		}
	}
	return true
}

// TenantMiddleware extracts and validates the `:tenant` path parameter for
// `/tenants/:tenant/scim/v2/...` routes before authentication runs.
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenant := strings.TrimSpace(c.Param("tenant"))
		if !IsValidTenantID(tenant) {
			MakeTypedError(c, http.StatusBadRequest, "invalidValue", "invalid tenant identifier")
			c.Abort()
			return
		}
		c.Set(urlTenantContextKey, tenant)
		c.Set(tenantContextKey, tenant)
		c.Next()
	}
}

// TenantOf returns the active tenant identifier for c (either from the URL
// `/tenants/:tenant/scim/v2` or from the authenticated Principal's bound
// tenant), or "" for the default single-tenant namespace.
func TenantOf(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Get(tenantContextKey); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	if p := PrincipalOf(c); p != nil && p.Tenant != "" {
		return p.Tenant
	}
	return ""
}

// URLTenantOf returns the tenant segment explicitly present in the request URL,
// or "" when the request used the root `/scim/v2` prefix.
func URLTenantOf(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if v, ok := c.Get(urlTenantContextKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// EffectivePrefix returns `/tenants/<tenant>/scim/v2` when the request arrived
// on a tenant-prefixed URL, or `/scim/v2` otherwise.
func EffectivePrefix(c *gin.Context) string {
	if urlTenant := URLTenantOf(c); urlTenant != "" {
		return "/tenants/" + urlTenant + PREFIX
	}
	return PREFIX
}

// ScopedBucket encodes an optional tenant and resource bucket name into a
// single Store bucket key. For the default tenant ("") it returns bucket
// verbatim so single-tenant callers and tests behave identically.
func ScopedBucket(tenant, bucket string) string {
	tenant = strings.TrimSpace(tenant)
	if tenant == "" || strings.HasPrefix(bucket, scopedBucketPrefix) {
		return bucket
	}
	return scopedBucketPrefix + tenant + ":" + bucket
}

// SplitScopedBucket decodes a bucket key produced by ScopedBucket into its
// tenant and physical resource bucket name.
func SplitScopedBucket(scoped string) (tenant, bucket string) {
	if !strings.HasPrefix(scoped, scopedBucketPrefix) {
		return "", scoped
	}
	rest := strings.TrimPrefix(scoped, scopedBucketPrefix)
	idx := strings.IndexByte(rest, ':')
	if idx <= 0 || idx == len(rest)-1 {
		return "", scoped
	}
	return rest[:idx], rest[idx+1:]
}

func storeBucketForContext(c *gin.Context, bucketName string) string {
	return ScopedBucket(TenantOf(c), bucketName)
}
