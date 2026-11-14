package scim

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	defaultInspectorCapacity = 200
	maxCapturedBodyBytes     = 4096
)

// InspectedExchange represents one captured SCIM HTTP request/response pair for
// the embedded live inspector console. Credentials and password fields are
// stripped prior to buffering.
type InspectedExchange struct {
	Seq             uint64    `json:"seq"`
	Timestamp       time.Time `json:"timestamp"`
	RequestID       string    `json:"requestId"`
	Tenant          string    `json:"tenant,omitempty"`
	Method          string    `json:"method"`
	Path            string    `json:"path"`
	Query           string    `json:"query,omitempty"`
	Action          string    `json:"action"`
	ResourceType    string    `json:"resourceType,omitempty"`
	ResourceID      string    `json:"resourceId,omitempty"`
	Subject         string    `json:"subject,omitempty"`
	Status          int       `json:"status"`
	ScimType        string    `json:"scimType,omitempty"`
	DurationMs      float64   `json:"durationMs"`
	RequestPreview  string    `json:"requestPreview,omitempty"`
	ResponsePreview string    `json:"responsePreview,omitempty"`
}

// InspectorBuffer is a bounded ring buffer of recent SCIM API exchanges.
type InspectorBuffer struct {
	enabled  atomic.Bool
	seq      atomic.Uint64
	mu       sync.RWMutex
	capacity int
	entries  []InspectedExchange
}

// NewInspectorBuffer creates a ring buffer holding up to capacity exchanges.
func NewInspectorBuffer(capacity int) *InspectorBuffer {
	if capacity <= 0 {
		capacity = defaultInspectorCapacity
	}
	return &InspectorBuffer{
		capacity: capacity,
		entries:  make([]InspectedExchange, 0, capacity),
	}
}

// DefaultInspector is the process-wide live inspector buffer.
var DefaultInspector = NewInspectorBuffer(defaultInspectorCapacity)

// SetEnabled toggles request/response capture for the live inspector.
func (b *InspectorBuffer) SetEnabled(on bool) {
	b.enabled.Store(on)
}

// Enabled reports whether capture is active.
func (b *InspectorBuffer) Enabled() bool {
	return b.enabled.Load()
}

// Clear removes all buffered exchanges.
func (b *InspectorBuffer) Clear() {
	b.mu.Lock()
	b.entries = b.entries[:0]
	b.mu.Unlock()
}

// Record appends ex to the ring buffer, evicting the oldest entry when full.
func (b *InspectorBuffer) Record(ex InspectedExchange) {
	ex.Seq = b.seq.Add(1)
	b.mu.Lock()
	if len(b.entries) >= b.capacity {
		copy(b.entries, b.entries[1:])
		b.entries[len(b.entries)-1] = ex
	} else {
		b.entries = append(b.entries, ex)
	}
	b.mu.Unlock()
}

// Recent returns up to limit exchanges in newest-first order.
func (b *InspectorBuffer) Recent(limit int) []InspectedExchange {
	b.mu.RLock()
	defer b.mu.RUnlock()
	n := len(b.entries)
	if limit <= 0 || limit > n {
		limit = n
	}
	out := make([]InspectedExchange, 0, limit)
	for i := n - 1; i >= n-limit; i-- {
		out = append(out, b.entries[i])
	}
	return out
}

// Lookup finds a buffered exchange by sequence number.
func (b *InspectorBuffer) Lookup(seq uint64) (InspectedExchange, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for i := len(b.entries) - 1; i >= 0; i-- {
		if b.entries[i].Seq == seq {
			return b.entries[i], true
		}
	}
	return InspectedExchange{}, false
}

type bodyCaptureWriter struct {
	gin.ResponseWriter
	buf bytes.Buffer
}

func (w *bodyCaptureWriter) Write(p []byte) (int, error) {
	if rem := maxCapturedBodyBytes - w.buf.Len(); rem > 0 {
		if len(p) <= rem {
			w.buf.Write(p)
		} else {
			w.buf.Write(p[:rem])
		}
	}
	return w.ResponseWriter.Write(p)
}

func captureRequestPreview(c *gin.Context) string {
	if c.Request == nil || c.Request.Body == nil {
		return ""
	}
	raw, err := io.ReadAll(io.LimitReader(c.Request.Body, DefaultMaxBodyBytes+1))
	if err != nil {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(raw))
	if len(raw) == 0 {
		return ""
	}
	return redactJSONPreview(raw)
}

func redactJSONPreview(raw []byte) string {
	trimmed := raw
	truncated := false
	if len(trimmed) > maxCapturedBodyBytes {
		trimmed = trimmed[:maxCapturedBodyBytes]
		truncated = true
	}
	var parsed interface{}
	if !truncated && json.Unmarshal(trimmed, &parsed) == nil {
		redactSecretsInValue(parsed)
		if pretty, err := json.MarshalIndent(parsed, "", "  "); err == nil {
			if len(pretty) > maxCapturedBodyBytes {
				return string(pretty[:maxCapturedBodyBytes]) + "\n... (truncated)"
			}
			return string(pretty)
		}
	}
	// Fallback when payload is truncated or non-JSON: scrub obvious password literals.
	text := string(trimmed)
	if strings.Contains(strings.ToLower(text), "password") {
		return "[JSON preview redacted because payload contains password field]"
	}
	if truncated {
		return text + "... (truncated)"
	}
	return text
}

func redactSecretsInValue(v interface{}) {
	switch typed := v.(type) {
	case map[string]interface{}:
		for k, val := range typed {
			if strings.EqualFold(k, "password") || strings.EqualFold(k, "secret") || strings.EqualFold(k, "token") {
				typed[k] = "[REDACTED]"
				continue
			}
			redactSecretsInValue(val)
		}
	case []interface{}:
		for _, item := range typed {
			redactSecretsInValue(item)
		}
	}
}
