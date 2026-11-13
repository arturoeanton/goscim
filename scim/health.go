package scim

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const readinessTimeout = 2 * time.Second

// Healthz is GET /healthz: a liveness probe reporting whether the HTTP process
// is alive. It deliberately does not touch the database so a transient
// Couchbase outage causes orchestrators to stop routing traffic (/readyz)
// rather than restarting healthy pods in a loop.
func Healthz(c *gin.Context) {
	c.Header("X-Content-Type-Options", "nosniff")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readyz is GET /readyz: a readiness probe verifying that a Store is installed
// and reachable within a short deadline.
func Readyz(c *gin.Context) {
	c.Header("X-Content-Type-Options", "nosniff")
	if DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), readinessTimeout)
	defer cancel()
	if err := DB.Ping(ctx); err != nil {
		log.Printf("scim readiness check failed: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
