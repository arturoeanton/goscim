package scim

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWebhookCloudEventsSignatureRetryAndRedaction(t *testing.T) {
	var (
		mu       sync.Mutex
		events   []CloudEvent
		sigOK    atomic.Bool
		attempts atomic.Int32
	)

	const secret = "whsec_test_signing_secret"

	dispatcher := NewWebhookDispatcher(WebhookConfig{
		URLs:        []string{"http://webhook.receiver.internal/scim-events"},
		Secret:      secret,
		Events:      []string{"user.*"},
		MaxRetries:  2,
		BaseBackoff: 5 * time.Millisecond,
		Client: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				n := attempts.Add(1)
				raw, _ := io.ReadAll(req.Body)
				// Fail the very first attempt with 503 to verify exponential retry.
				if n == 1 {
					rec := httptest.NewRecorder()
					rec.WriteHeader(http.StatusServiceUnavailable)
					return rec.Result(), nil
				}
				ts := req.Header.Get("X-SCIM-Timestamp")
				gotSig := req.Header.Get("X-SCIM-Signature-256")
				if gotSig != "" && gotSig == ComputeWebhookSignature(secret, ts, raw) {
					sigOK.Store(true)
				}
				var ev CloudEvent
				if err := json.Unmarshal(raw, &ev); err == nil {
					mu.Lock()
					events = append(events, ev)
					mu.Unlock()
				}
				rec := httptest.NewRecorder()
				rec.WriteHeader(http.StatusAccepted)
				return rec.Result(), nil
			}),
		},
	})
	SetWebhookDispatcher(dispatcher)
	t.Cleanup(func() {
		SetWebhookDispatcher(nil)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = dispatcher.Close(ctx)
	})

	r, _ := newTestServer(t)

	// 1. Create User with password -> emits io.goscim.user.created without password.
	wCreate := do(t, r, http.MethodPost, "/scim/v2/Users",
		`{"schemas":["urn:ietf:params:scim:schemas:core:2.0:User"],"userName":"wh.user@example.com","password":"SuperSecretPassword!","active":true}`)
	requireStatus(t, wCreate, http.StatusCreated)
	userID := decode(t, wCreate)["id"].(string)

	// 2. Create Group -> filtered out by Events: ["user.*"].
	wGrp := do(t, r, http.MethodPost, "/scim/v2/Groups",
		`{"schemas":["urn:ietf:params:scim:schemas:core:2.0:Group"],"displayName":"FilteredOutGroup"}`)
	requireStatus(t, wGrp, http.StatusCreated)

	// 3. Deactivate User -> emits io.goscim.user.deactivated.
	wDeact := do(t, r, http.MethodPatch, "/scim/v2/Users/"+userID,
		`{"schemas":["urn:ietf:params:scim:api:messages:2.0:PatchOp"],"Operations":[{"op":"replace","path":"active","value":false}]}`)
	requireStatus(t, wDeact, http.StatusOK)

	// 4. Delete User -> emits io.goscim.user.deleted.
	wDel := do(t, r, http.MethodDelete, "/scim/v2/Users/"+userID, "")
	requireStatus(t, wDel, http.StatusNoContent)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := dispatcher.Close(ctx); err != nil {
		t.Fatalf("dispatcher.Close: %v", err)
	}

	if !sigOK.Load() {
		t.Fatalf("expected valid HMAC-SHA256 X-SCIM-Signature-256 header")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(events) != 3 {
		t.Fatalf("received %d events, want 3: %+v", len(events), events)
	}
	wantTypes := []string{"io.goscim.user.created", "io.goscim.user.deactivated", "io.goscim.user.deleted"}
	for i, want := range wantTypes {
		if events[i].Type != want {
			t.Errorf("events[%d].Type = %q, want %q", i, events[i].Type, want)
		}
		if _, leaked := events[i].Data["password"]; leaked {
			t.Errorf("events[%d] leaked write-only password field", i)
		}
	}

	deliveries := dispatcher.RecentDeliveries(10)
	if len(deliveries) != 3 {
		t.Fatalf("RecentDeliveries len = %d, want 3", len(deliveries))
	}
}
