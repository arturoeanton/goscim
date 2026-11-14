package scim

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var latencyBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0}

type requestMetricKey struct {
	method       string
	action       string
	resourceType string
	status       int
}

type webhookMetricKey struct {
	eventType string
	outcome   string
}

type histogramSeries struct {
	buckets [10]uint64
	count   uint64
	sumNs   uint64
}

// MetricsRegistry collects SCIM server operational metrics and renders them in
// standard Prometheus text exposition format (0.0.4) with zero external
// dependencies.
type MetricsRegistry struct {
	mu                  sync.RWMutex
	requests            map[requestMetricKey]uint64
	durations           map[string]*histogramSeries
	uniquenessConflicts map[string]uint64
	webhooks            map[webhookMetricKey]uint64
	authFailures        atomic.Uint64
	buildVersion        atomic.Value
}

// NewMetricsRegistry returns an empty MetricsRegistry.
func NewMetricsRegistry() *MetricsRegistry {
	m := &MetricsRegistry{
		requests:            make(map[requestMetricKey]uint64),
		durations:           make(map[string]*histogramSeries),
		uniquenessConflicts: make(map[string]uint64),
		webhooks:            make(map[webhookMetricKey]uint64),
	}
	m.buildVersion.Store("dev")
	return m
}

// DefaultMetrics is the process-wide metrics collector updated by AuditMiddleware
// and the webhook dispatcher.
var DefaultMetrics = NewMetricsRegistry()

// SetBuildVersion records the binary version reported in goscim_build_info.
func (m *MetricsRegistry) SetBuildVersion(v string) {
	if strings.TrimSpace(v) != "" {
		m.buildVersion.Store(strings.TrimSpace(v))
	}
}

// Reset clears all accumulated counters and histograms (intended for tests).
func (m *MetricsRegistry) Reset() {
	m.mu.Lock()
	m.requests = make(map[requestMetricKey]uint64)
	m.durations = make(map[string]*histogramSeries)
	m.uniquenessConflicts = make(map[string]uint64)
	m.webhooks = make(map[webhookMetricKey]uint64)
	m.mu.Unlock()
	m.authFailures.Store(0)
}

// RecordRequest records one completed SCIM HTTP operation.
func (m *MetricsRegistry) RecordRequest(method, action, resourceType, scimType string, status int, duration time.Duration) {
	if action == "" {
		action = "other"
	}
	if resourceType == "" {
		resourceType = "none"
	}
	if status == http.StatusUnauthorized {
		m.authFailures.Add(1)
	}

	sec := duration.Seconds()
	ns := uint64(duration.Nanoseconds())

	m.mu.Lock()
	m.requests[requestMetricKey{
		method:       method,
		action:       action,
		resourceType: resourceType,
		status:       status,
	}]++

	h := m.durations[action]
	if h == nil {
		h = &histogramSeries{}
		m.durations[action] = h
	}
	for i, bound := range latencyBuckets {
		if sec <= bound {
			h.buckets[i]++
		}
	}
	h.count++
	h.sumNs += ns

	if status == http.StatusConflict && strings.EqualFold(scimType, "uniqueness") {
		m.uniquenessConflicts[resourceType]++
	}
	m.mu.Unlock()
}

// RecordWebhookDelivery increments goscim_webhook_deliveries_total.
func (m *MetricsRegistry) RecordWebhookDelivery(eventType, outcome string) {
	m.mu.Lock()
	m.webhooks[webhookMetricKey{eventType: eventType, outcome: outcome}]++
	m.mu.Unlock()
}

// SnapshotSummary returns aggregate counters used by the embedded HTMX dashboard.
type MetricsSummary struct {
	TotalRequests       uint64
	TotalErrors         uint64
	AuthFailures        uint64
	UniquenessConflicts uint64
	WebhookSuccess      uint64
	WebhookFailed       uint64
}

// Summary returns high-level counters from the registry.
func (m *MetricsRegistry) Summary() MetricsSummary {
	var s MetricsSummary
	s.AuthFailures = m.authFailures.Load()
	m.mu.RLock()
	for k, v := range m.requests {
		s.TotalRequests += v
		if k.status >= 400 {
			s.TotalErrors += v
		}
	}
	for _, v := range m.uniquenessConflicts {
		s.UniquenessConflicts += v
	}
	for k, v := range m.webhooks {
		if k.outcome == "success" {
			s.WebhookSuccess += v
		} else {
			s.WebhookFailed += v
		}
	}
	m.mu.RUnlock()
	return s
}

// RenderExposition formats all metrics in Prometheus text exposition format.
func (m *MetricsRegistry) RenderExposition(storeUp bool) string {
	var b strings.Builder

	version, _ := m.buildVersion.Load().(string)
	if version == "" {
		version = "dev"
	}

	b.WriteString("# HELP goscim_build_info Build metadata for the running GoSCIM server.\n")
	b.WriteString("# TYPE goscim_build_info gauge\n")
	fmt.Fprintf(&b, "goscim_build_info{version=%q} 1\n", version)

	b.WriteString("# HELP goscim_store_up Whether the configured SCIM backing store answered Ping (1 = healthy, 0 = unavailable).\n")
	b.WriteString("# TYPE goscim_store_up gauge\n")
	if storeUp {
		b.WriteString("goscim_store_up 1\n")
	} else {
		b.WriteString("goscim_store_up 0\n")
	}

	b.WriteString("# HELP goscim_auth_failures_total Total unauthenticated SCIM API requests rejected with HTTP 401.\n")
	b.WriteString("# TYPE goscim_auth_failures_total counter\n")
	fmt.Fprintf(&b, "goscim_auth_failures_total %d\n", m.authFailures.Load())

	m.mu.RLock()
	reqKeys := make([]requestMetricKey, 0, len(m.requests))
	for k := range m.requests {
		reqKeys = append(reqKeys, k)
	}
	sort.Slice(reqKeys, func(i, j int) bool {
		a, c := reqKeys[i], reqKeys[j]
		if a.action != c.action {
			return a.action < c.action
		}
		if a.method != c.method {
			return a.method < c.method
		}
		if a.resourceType != c.resourceType {
			return a.resourceType < c.resourceType
		}
		return a.status < c.status
	})

	b.WriteString("# HELP goscim_http_requests_total Total SCIM HTTP requests processed.\n")
	b.WriteString("# TYPE goscim_http_requests_total counter\n")
	for _, k := range reqKeys {
		fmt.Fprintf(&b, "goscim_http_requests_total{action=%q,method=%q,resource_type=%q,status=%q} %d\n",
			k.action, k.method, k.resourceType, strconv.Itoa(k.status), m.requests[k])
	}

	actions := make([]string, 0, len(m.durations))
	for act := range m.durations {
		actions = append(actions, act)
	}
	sort.Strings(actions)

	b.WriteString("# HELP goscim_http_request_duration_seconds SCIM HTTP request latency distributions by action.\n")
	b.WriteString("# TYPE goscim_http_request_duration_seconds histogram\n")
	for _, act := range actions {
		h := m.durations[act]
		for i, bound := range latencyBuckets {
			fmt.Fprintf(&b, "goscim_http_request_duration_seconds_bucket{action=%q,le=%q} %d\n",
				act, strconv.FormatFloat(bound, 'f', -1, 64), h.buckets[i])
		}
		fmt.Fprintf(&b, "goscim_http_request_duration_seconds_bucket{action=%q,le=\"+Inf\"} %d\n", act, h.count)
		fmt.Fprintf(&b, "goscim_http_request_duration_seconds_sum{action=%q} %.6f\n", act, float64(h.sumNs)/1e9)
		fmt.Fprintf(&b, "goscim_http_request_duration_seconds_count{action=%q} %d\n", act, h.count)
	}

	conflictTypes := make([]string, 0, len(m.uniquenessConflicts))
	for rt := range m.uniquenessConflicts {
		conflictTypes = append(conflictTypes, rt)
	}
	sort.Strings(conflictTypes)

	b.WriteString("# HELP goscim_uniqueness_conflicts_total Total HTTP 409 uniqueness conflicts by resource type.\n")
	b.WriteString("# TYPE goscim_uniqueness_conflicts_total counter\n")
	for _, rt := range conflictTypes {
		fmt.Fprintf(&b, "goscim_uniqueness_conflicts_total{resource_type=%q} %d\n", rt, m.uniquenessConflicts[rt])
	}

	whKeys := make([]webhookMetricKey, 0, len(m.webhooks))
	for k := range m.webhooks {
		whKeys = append(whKeys, k)
	}
	sort.Slice(whKeys, func(i, j int) bool {
		if whKeys[i].eventType != whKeys[j].eventType {
			return whKeys[i].eventType < whKeys[j].eventType
		}
		return whKeys[i].outcome < whKeys[j].outcome
	})

	b.WriteString("# HELP goscim_webhook_deliveries_total Total outbound SCIM webhook delivery outcomes.\n")
	b.WriteString("# TYPE goscim_webhook_deliveries_total counter\n")
	for _, k := range whKeys {
		fmt.Fprintf(&b, "goscim_webhook_deliveries_total{event_type=%q,outcome=%q} %d\n",
			k.eventType, k.outcome, m.webhooks[k])
	}
	m.mu.RUnlock()

	return b.String()
}

// MetricsEnabledFromEnv reports whether GET /metrics should be mounted
// (enabled by default unless SCIM_METRICS=false or off).
func MetricsEnabledFromEnv() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("SCIM_METRICS")))
	return v != "false" && v != "off" && v != "0"
}

// MetricsHandler serves GET /metrics in Prometheus exposition format, optionally
// guarded by SCIM_METRICS_TOKEN.
func MetricsHandler(reg *MetricsRegistry) gin.HandlerFunc {
	if reg == nil {
		reg = DefaultMetrics
	}
	return func(c *gin.Context) {
		if expected := strings.TrimSpace(os.Getenv("SCIM_METRICS_TOKEN")); expected != "" {
			if !bearerHeaderMatches(c.GetHeader("Authorization"), expected) {
				c.Header("WWW-Authenticate", `Bearer realm="metrics"`)
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
		}
		storeUp := false
		if DB != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			storeUp = DB.Ping(ctx) == nil
			cancel()
		}
		c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(reg.RenderExposition(storeUp)))
	}
}

func bearerHeaderMatches(header, expected string) bool {
	const prefix = "bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return false
	}
	presented := strings.TrimSpace(header[len(prefix):])
	a := sha256.Sum256([]byte(presented))
	b := sha256.Sum256([]byte(expected))
	return subtle.ConstantTimeCompare(a[:], b[:]) == 1
}
