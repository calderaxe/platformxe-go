// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-10-20, pending release):
// - Created (webhooks v2, NO-DRIFT train). Pins the method, path, query,
//   headers and body of every WebhooksService call against a real
//   net/http/httptest server and decodes the API-shaped responses into the
//   typed results. Secrets in the fixtures are obvious placeholders.
// - Review fix wave 2: List() with no argument still compiles and sends no
//   query; WebhookCreated's promoted fields (got.ID, got.EngineVersion) read
//   the embedded endpoint and its JSON keeps the "webhook" / "secret" keys;
//   Dispatch sends a nil Data as {}.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"net/http"
	"testing"
)

func webhookTestClient(t *testing.T, check func(r *http.Request), data interface{}) *Client {
	t.Helper()
	srv := newTestServer(t, check, data)
	t.Cleanup(srv.Close)
	return NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
}

func TestWebhooksCreateV2(t *testing.T) {
	c := webhookTestClient(t, func(r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/webhooks" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		body := decodeJSONBody(t, r)
		if body["engineVersion"] != float64(2) || body["recipientRef"] != "host:1" {
			t.Errorf("unexpected body: %#v", body)
		}
	}, map[string]interface{}{
		"webhook": map[string]interface{}{"id": "wh_1", "status": "ACTIVE", "engineVersion": 2, "recipientRef": "host:1"},
		"secret":  "placeholder-secret-not-real",
	})
	got, err := c.Webhooks.Create(map[string]interface{}{
		"name": "Host hook", "url": "https://example.com/hook", "events": []string{"custom.x"},
		"engineVersion": 2, "recipientRef": "host:1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Secret == "" || got.Webhook.EngineVersion != 2 || got.Webhook.RecipientRef == nil || *got.Webhook.RecipientRef != "host:1" {
		t.Errorf("unexpected result: %#v", got)
	}
	// Promoted from the embedded Webhook: v1 code reading wh.ID still compiles.
	if got.ID != "wh_1" || got.EngineVersion != 2 || got.Status != "ACTIVE" {
		t.Errorf("promoted fields: %#v", got)
	}
	if got.Secret != "placeholder-secret-not-real" {
		t.Errorf("Secret must be the one-time secret, got %q", got.Secret)
	}
}

func TestWebhookCreatedJSONKeepsWebhookKey(t *testing.T) {
	raw, err := json.Marshal(WebhookCreated{Webhook: Webhook{ID: "wh_1"}, Secret: "placeholder-secret-not-real"})
	if err != nil {
		t.Fatal(err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatal(err)
	}
	if len(top) != 2 || top["webhook"] == nil || top["secret"] == nil {
		t.Errorf("expected exactly the webhook and secret keys, got %s", raw)
	}
	var inner map[string]interface{}
	if err := json.Unmarshal(top["webhook"], &inner); err != nil {
		t.Fatal(err)
	}
	if inner["id"] != "wh_1" {
		t.Errorf("unexpected webhook object: %s", top["webhook"])
	}
	if _, has := inner["secret"]; has {
		t.Errorf("the deprecated Webhook.Secret must stay omitted: %s", top["webhook"])
	}
}

func TestWebhooksListFiltersByRecipient(t *testing.T) {
	c := webhookTestClient(t, func(r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/webhooks" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("recipientRef") != "host:1" || q.Get("limit") != "50" || q.Get("page") != "2" {
			t.Errorf("unexpected query: %v", q)
		}
	}, map[string]interface{}{"items": []interface{}{map[string]interface{}{"id": "wh_1"}}, "total": 1, "page": 2, "limit": 50})
	got, err := c.Webhooks.List(&ListWebhooksOptions{RecipientRef: "host:1", Limit: 50, Page: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 || got.Total != 1 || got.Page != 2 {
		t.Errorf("unexpected result: %#v", got)
	}
}

func TestWebhooksListNilOptionsSendsNoQuery(t *testing.T) {
	c := webhookTestClient(t, func(r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query, got %q", r.URL.RawQuery)
		}
	}, map[string]interface{}{"items": []interface{}{}, "total": 0, "page": 1, "limit": 20})
	if _, err := c.Webhooks.List(nil); err != nil {
		t.Fatal(err)
	}
}

func TestWebhooksListWithoutArgumentsSendsNoQuery(t *testing.T) {
	c := webhookTestClient(t, func(r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/webhooks" || r.URL.RawQuery != "" {
			t.Errorf("unexpected request %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
	}, map[string]interface{}{"items": []interface{}{map[string]interface{}{"id": "wh_1"}}, "total": 1, "page": 1, "limit": 20})
	got, err := c.Webhooks.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 || got.Items[0].ID != "wh_1" {
		t.Errorf("unexpected result: %#v", got)
	}
}

func TestWebhooksRotateUsesRotatePath(t *testing.T) {
	c := webhookTestClient(t, func(r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/webhooks/wh_1/rotate" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}, map[string]interface{}{"webhook": map[string]interface{}{"id": "wh_1"}, "secret": "placeholder-secret-not-real"})
	got, err := c.Webhooks.RotateSecret("wh_1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Secret == "" || got.ID != "wh_1" {
		t.Errorf("expected the endpoint and the one-time secret, got %#v", got)
	}
}

func TestWebhooksDispatchSendsNilDataAsEmptyObject(t *testing.T) {
	c := webhookTestClient(t, func(r *http.Request) {
		body := decodeJSONBody(t, r)
		data, ok := body["data"].(map[string]interface{})
		if !ok || len(data) != 0 {
			t.Errorf(`expected "data":{}, got %#v`, body["data"])
		}
	}, map[string]interface{}{"targeted": 0, "created": 0})
	input := DispatchWebhookInput{RecipientRef: "host:1", EventType: "booking.created", EventID: "evt_1"}
	if _, err := c.Webhooks.Dispatch(input, nil); err != nil {
		t.Fatal(err)
	}
	if input.Data != nil {
		t.Error("the caller's input must not be changed")
	}
}

func TestWebhooksDispatchSendsBodyAndIdempotencyKey(t *testing.T) {
	c := webhookTestClient(t, func(r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/webhooks/dispatch" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("x-idempotency-key") != "idem-1" {
			t.Errorf("missing idempotency header: %v", r.Header)
		}
		body := decodeJSONBody(t, r)
		if body["recipientRef"] != "host:1" || body["eventType"] != "custom.x" || body["eventId"] != "evt_1" {
			t.Errorf("unexpected body: %#v", body)
		}
	}, map[string]interface{}{"targeted": 1, "created": 1})
	got, err := c.Webhooks.Dispatch(
		DispatchWebhookInput{RecipientRef: "host:1", EventType: "custom.x", EventID: "evt_1", Data: map[string]interface{}{"a": 1}},
		&DispatchWebhookOptions{IdempotencyKey: "idem-1"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got.Targeted != 1 || got.Created != 1 {
		t.Errorf("unexpected result: %#v", got)
	}
}

func TestWebhooksDeliveries(t *testing.T) {
	c := webhookTestClient(t, func(r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/webhooks/wh_1/deliveries" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "10" || r.URL.Query().Get("page") != "3" {
			t.Errorf("unexpected query: %v", r.URL.Query())
		}
	}, map[string]interface{}{
		"items": []interface{}{map[string]interface{}{"id": "dlv_1", "status": "FAILED", "attemptCount": 5, "responseSnippet": nil}},
		"total": 1, "page": 3, "limit": 10,
	})
	got, err := c.Webhooks.Deliveries("wh_1", 10, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 1 || got.Items[0].Status != "FAILED" {
		t.Errorf("unexpected result: %#v", got)
	}
}

func TestWebhooksRedeliver(t *testing.T) {
	c := webhookTestClient(t, func(r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/webhooks/deliveries/dlv_9/redeliver" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}, map[string]interface{}{"deliveryId": "dlv_9", "ok": false, "status": nil, "responseTimeMs": 12, "reason": "URL_REFUSED"})
	got, err := c.Webhooks.Redeliver("dlv_9")
	if err != nil {
		t.Fatal(err)
	}
	if got.OK || got.Reason == nil || *got.Reason != "URL_REFUSED" {
		t.Errorf("unexpected result: %#v", got)
	}
}
