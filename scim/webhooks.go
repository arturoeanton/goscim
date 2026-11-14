package scim

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	defaultWebhookQueueSize  = 512
	defaultWebhookHistoryCap = 100
	defaultWebhookMaxRetries = 3
)

// CloudEvent is the CloudEvents 1.0 JSON envelope emitted for SCIM lifecycle
// mutations.
type CloudEvent struct {
	SpecVersion     string                 `json:"specversion"`
	ID              string                 `json:"id"`
	Source          string                 `json:"source"`
	Type            string                 `json:"type"`
	DataContentType string                 `json:"datacontenttype"`
	Time            string                 `json:"time"`
	Subject         string                 `json:"subject,omitempty"`
	Tenant          string                 `json:"tenant,omitempty"`
	RequestID       string                 `json:"requestid,omitempty"`
	Actor           string                 `json:"actor,omitempty"`
	Data            map[string]interface{} `json:"data,omitempty"`
}

// WebhookDelivery records one outbound delivery attempt/outcome for inspection.
type WebhookDelivery struct {
	ID         string    `json:"id"`
	EventType  string    `json:"eventType"`
	TargetURL  string    `json:"targetUrl"`
	Subject    string    `json:"subject"`
	Tenant     string    `json:"tenant,omitempty"`
	StatusCode int       `json:"statusCode"`
	Attempts   int       `json:"attempts"`
	Outcome    string    `json:"outcome"`
	Error      string    `json:"error,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Payload    string    `json:"payload"`
}

// WebhookConfig configures outbound SCIM event webhooks.
type WebhookConfig struct {
	URLs        []string
	Secret      string
	Events      []string
	MaxRetries  int
	BaseBackoff time.Duration
	Client      *http.Client
}

// WebhookDispatcher asynchronously delivers signed CloudEvents for SCIM
// resource mutations, sharding queues by Subject so lifecycle events targeting
// the same resource ID preserve causal FIFO order across retries.
type WebhookDispatcher struct {
	cfg     WebhookConfig
	enqMu   sync.RWMutex
	shards  []chan CloudEvent
	wg      sync.WaitGroup
	closed  bool
	mu      sync.RWMutex
	history []WebhookDelivery
}

var activeWebhookDispatcher atomic.Pointer[WebhookDispatcher]

// SetWebhookDispatcher installs d as the active webhook dispatcher.
func SetWebhookDispatcher(d *WebhookDispatcher) {
	activeWebhookDispatcher.Store(d)
}

// ActiveWebhookDispatcher returns the currently configured dispatcher, or nil.
func ActiveWebhookDispatcher() *WebhookDispatcher {
	return activeWebhookDispatcher.Load()
}

// ConfigureWebhooksFromEnv reads SCIM_WEBHOOK_URLS, SCIM_WEBHOOK_SECRET,
// SCIM_WEBHOOK_EVENTS and SCIM_WEBHOOK_MAX_RETRIES, installing the dispatcher if
// at least one URL is configured.
func ConfigureWebhooksFromEnv() (*WebhookDispatcher, error) {
	urls := splitAndTrim(os.Getenv("SCIM_WEBHOOK_URLS"))
	if len(urls) == 0 {
		SetWebhookDispatcher(nil)
		return nil, nil
	}
	maxRetries := defaultWebhookMaxRetries
	if raw := strings.TrimSpace(os.Getenv("SCIM_WEBHOOK_MAX_RETRIES")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 || parsed > 10 {
			return nil, fmt.Errorf("SCIM_WEBHOOK_MAX_RETRIES=%q must be an integer between 0 and 10", raw)
		}
		maxRetries = parsed
	}
	d := NewWebhookDispatcher(WebhookConfig{
		URLs:        urls,
		Secret:      os.Getenv("SCIM_WEBHOOK_SECRET"),
		Events:      splitAndTrim(os.Getenv("SCIM_WEBHOOK_EVENTS")),
		MaxRetries:  maxRetries,
		BaseBackoff: 150 * time.Millisecond,
	})
	SetWebhookDispatcher(d)
	return d, nil
}

// NewWebhookDispatcher starts background delivery workers for cfg.
func NewWebhookDispatcher(cfg WebhookConfig) *WebhookDispatcher {
	if cfg.Client == nil {
		cfg.Client = &http.Client{Timeout: 5 * time.Second}
	}
	if cfg.BaseBackoff <= 0 {
		cfg.BaseBackoff = 100 * time.Millisecond
	}
	const workers = 4
	shards := make([]chan CloudEvent, workers)
	for i := range shards {
		shards[i] = make(chan CloudEvent, defaultWebhookQueueSize/workers)
	}
	d := &WebhookDispatcher{
		cfg:     cfg,
		shards:  shards,
		history: make([]WebhookDelivery, 0, defaultWebhookHistoryCap),
	}
	for i := 0; i < workers; i++ {
		d.wg.Add(1)
		go d.worker(shards[i])
	}
	return d
}

// URLs returns a copy of the configured destination URLs.
func (d *WebhookDispatcher) URLs() []string {
	if d == nil {
		return nil
	}
	out := make([]string, len(d.cfg.URLs))
	copy(out, d.cfg.URLs)
	return out
}

func (d *WebhookDispatcher) shardFor(ev CloudEvent) chan CloudEvent {
	key := ev.Tenant + ":" + ev.Subject
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return d.shards[int(h.Sum32())%len(d.shards)]
}

// Enqueue schedules ev for delivery if it matches the configured event filter.
func (d *WebhookDispatcher) Enqueue(ev CloudEvent) {
	if d == nil || len(d.cfg.URLs) == 0 {
		return
	}
	if !d.matchesEvent(ev.Type) {
		return
	}
	d.enqMu.RLock()
	defer d.enqMu.RUnlock()
	if d.closed {
		return
	}
	select {
	case d.shardFor(ev) <- ev:
	default:
		DefaultMetrics.RecordWebhookDelivery(ev.Type, "dropped")
		log.Printf("scim webhook queue full, dropping event %s (%s)", ev.ID, ev.Type)
	}
}

// Close stops accepting new events and waits up to ctx's deadline for queued
// deliveries to finish.
func (d *WebhookDispatcher) Close(ctx context.Context) error {
	if d == nil {
		return nil
	}
	d.enqMu.Lock()
	if !d.closed {
		d.closed = true
		for _, ch := range d.shards {
			close(ch)
		}
	}
	d.enqMu.Unlock()
	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// RecentDeliveries returns up to limit recent deliveries in newest-first order.
func (d *WebhookDispatcher) RecentDeliveries(limit int) []WebhookDelivery {
	if d == nil {
		return nil
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	n := len(d.history)
	if limit <= 0 || limit > n {
		limit = n
	}
	out := make([]WebhookDelivery, 0, limit)
	for i := n - 1; i >= n-limit; i-- {
		out = append(out, d.history[i])
	}
	return out
}

func (d *WebhookDispatcher) matchesEvent(eventType string) bool {
	if len(d.cfg.Events) == 0 {
		return true
	}
	short := strings.TrimPrefix(eventType, "io.goscim.")
	for _, pat := range d.cfg.Events {
		pat = strings.ToLower(strings.TrimSpace(pat))
		if pat == "" || pat == "*" {
			return true
		}
		target := strings.ToLower(eventType)
		shortTarget := strings.ToLower(short)
		if pat == target || pat == shortTarget {
			return true
		}
		if strings.HasSuffix(pat, ".*") {
			prefix := strings.TrimSuffix(pat, "*")
			if strings.HasPrefix(target, prefix) || strings.HasPrefix(shortTarget, prefix) {
				return true
			}
		}
	}
	return false
}

func (d *WebhookDispatcher) worker(ch <-chan CloudEvent) {
	defer d.wg.Done()
	for ev := range ch {
		body, err := json.Marshal(ev)
		if err != nil {
			continue
		}
		for _, targetURL := range d.cfg.URLs {
			d.deliverOne(targetURL, ev, body)
		}
	}
}

func (d *WebhookDispatcher) deliverOne(targetURL string, ev CloudEvent, body []byte) {
	maxAttempts := d.cfg.MaxRetries + 1
	var lastStatus int
	var lastErr string

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			backoff := d.cfg.BaseBackoff * time.Duration(1<<(attempt-2))
			time.Sleep(backoff)
		}
		ts := strconv.FormatInt(time.Now().UTC().Unix(), 10)
		req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(body))
		if err != nil {
			lastErr = err.Error()
			break
		}
		req.Header.Set("Content-Type", "application/cloudevents+json; charset=utf-8")
		req.Header.Set("User-Agent", "GoSCIM-Webhook/1.0")
		req.Header.Set("X-SCIM-Event", ev.Type)
		req.Header.Set("X-SCIM-Delivery", ev.ID)
		req.Header.Set("X-SCIM-Timestamp", ts)
		if ev.RequestID != "" {
			req.Header.Set(RequestIDHeader, ev.RequestID)
		}
		if d.cfg.Secret != "" {
			req.Header.Set("X-SCIM-Signature-256", ComputeWebhookSignature(d.cfg.Secret, ts, body))
		}

		resp, err := d.cfg.Client.Do(req)
		if err != nil {
			lastStatus = 0
			lastErr = err.Error()
			continue
		}
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
		lastStatus = resp.StatusCode

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			DefaultMetrics.RecordWebhookDelivery(ev.Type, "success")
			d.recordHistory(WebhookDelivery{
				ID:         ev.ID,
				EventType:  ev.Type,
				TargetURL:  targetURL,
				Subject:    ev.Subject,
				Tenant:     ev.Tenant,
				StatusCode: resp.StatusCode,
				Attempts:   attempt,
				Outcome:    "success",
				Timestamp:  time.Now().UTC(),
				Payload:    string(body),
			})
			return
		}
		lastErr = fmt.Sprintf("HTTP %d", resp.StatusCode)
		// Retry only on 429 or 5xx; permanent 4xx errors stop immediately.
		if resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			break
		}
	}

	DefaultMetrics.RecordWebhookDelivery(ev.Type, "failure")
	d.recordHistory(WebhookDelivery{
		ID:         ev.ID,
		EventType:  ev.Type,
		TargetURL:  targetURL,
		Subject:    ev.Subject,
		Tenant:     ev.Tenant,
		StatusCode: lastStatus,
		Attempts:   maxAttempts,
		Outcome:    "failure",
		Error:      lastErr,
		Timestamp:  time.Now().UTC(),
		Payload:    string(body),
	})
}

func (d *WebhookDispatcher) recordHistory(entry WebhookDelivery) {
	d.mu.Lock()
	if len(d.history) >= defaultWebhookHistoryCap {
		copy(d.history, d.history[1:])
		d.history[len(d.history)-1] = entry
	} else {
		d.history = append(d.history, entry)
	}
	d.mu.Unlock()
}

// ComputeWebhookSignature returns "sha256=<hex>" over "<timestamp>.<body>".
func ComputeWebhookSignature(secret, timestamp string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func emitResourceWebhookFromContext(c *gin.Context, action string, rt ResourceType, id string, previous, current map[string]interface{}) {
	if ActiveWebhookDispatcher() == nil {
		return
	}
	var actor string
	if p := PrincipalOf(c); p != nil {
		actor = p.Subject
	}
	emitResourceWebhook(action, rt, id, TenantOf(c), RequestIDOf(c), actor, EffectivePrefix(c)+rt.Endpoint, previous, current)
}

func emitResourceWebhook(action string, rt ResourceType, id, tenant, reqID, actor, source string, previous, current map[string]interface{}) {
	d := ActiveWebhookDispatcher()
	if d == nil {
		return
	}
	resKey := strings.ToLower(rt.Name)
	eventVerb := action
	switch action {
	case "replace", "patch":
		if isUserResourceType(rt) && wasDeactivated(previous, current) {
			eventVerb = "deactivated"
		} else if isGroupResourceType(rt) {
			eventVerb = "updated"
		} else if action == "replace" {
			eventVerb = "replaced"
		} else {
			eventVerb = "patched"
		}
	case "create":
		eventVerb = "created"
	case "delete":
		eventVerb = "deleted"
	}

	var sanitized map[string]interface{}
	if current != nil {
		sanitized = sanitizeWebhookData(rt, current)
	} else if previous != nil {
		sanitized = map[string]interface{}{
			"id": id,
		}
		if uName, ok := previous["userName"].(string); ok && uName != "" {
			sanitized["userName"] = uName
		}
		if dName, ok := previous["displayName"].(string); ok && dName != "" {
			sanitized["displayName"] = dName
		}
	}

	d.Enqueue(CloudEvent{
		SpecVersion:     "1.0",
		ID:              uuid.NewString(),
		Source:          source,
		Type:            "io.goscim." + resKey + "." + eventVerb,
		DataContentType: scimContentType,
		Time:            time.Now().UTC().Format(time.RFC3339),
		Subject:         id,
		Tenant:          tenant,
		RequestID:       reqID,
		Actor:           actor,
		Data:            sanitized,
	})
}

func wasDeactivated(previous, current map[string]interface{}) bool {
	if current == nil {
		return false
	}
	curActive, ok := current["active"].(bool)
	if !ok || curActive {
		return false
	}
	if previous == nil {
		return true
	}
	prevActive, hasPrev := previous["active"].(bool)
	return !hasPrev || prevActive
}

func sanitizeWebhookData(rt ResourceType, doc map[string]interface{}) map[string]interface{} {
	cloned, err := cloneDoc(doc)
	if err != nil {
		return nil
	}
	delete(cloned, "_scim_tenant")
	if core, ok := LookupSchema(rt.Schema); ok {
		stripNeverReturned(cloned, core.Attributes)
	}
	for _, ext := range rt.SchemaExtensions {
		if extSchema, ok := LookupSchema(ext.Schema); ok {
			if rawExt, present := lookupValueFold(cloned, ext.Schema); present {
				if extMap, isMap := rawExt.(map[string]interface{}); isMap {
					stripNeverReturned(extMap, extSchema.Attributes)
				}
			}
		}
	}
	return cloned
}

func stripNeverReturned(obj map[string]interface{}, attrs []Attribute) {
	for _, attr := range attrs {
		if !isReturnable(attr) {
			deleteKeyFold(obj, attr.Name)
			continue
		}
		if len(attr.SubAttributes) > 0 {
			if val, present := lookupValueFold(obj, attr.Name); present {
				switch typed := val.(type) {
				case map[string]interface{}:
					stripNeverReturned(typed, attr.SubAttributes)
				case []interface{}:
					for _, item := range typed {
						if m, ok := item.(map[string]interface{}); ok {
							stripNeverReturned(m, attr.SubAttributes)
						}
					}
				}
			}
		}
	}
}
