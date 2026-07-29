package scim

import (
	"bytes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DefaultMaxBodyBytes bounds a request payload. The handlers read the whole
// body into memory before parsing it, so without a limit a single large POST
// is enough to exhaust the server.
const DefaultMaxBodyBytes int64 = 1 << 20 // 1 MiB

// LimitBody caps how much of a request body the server will read.
func LimitBody(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}

// readBody reads the request payload, answering 413 when it exceeds the limit
// installed by LimitBody. It reports whether the caller may proceed; when it
// returns false the response has already been written.
func readBody(c *gin.Context) ([]byte, bool) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(c.Request.Body); err != nil {
		// http.MaxBytesReader has already set the status, but the SCIM error
		// body is ours to write.
		MakeError(c, http.StatusRequestEntityTooLarge, "the request body is too large")
		return nil, false
	}
	return buf.Bytes(), true
}
