// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-20, v1.14.0 pending, fidelity W3): Initial. Pins what
// Qr.GenerateBatchItems and Qr.GenerateBatch send (POST /api/v1/qr/batch:
// body, x-idempotency-key) and how the answer decodes (generated and failed
// items, the zip).
// =============================================================================

package platformxe

import (
	"net/http"
	"testing"
)

var qrBatchAnswer = map[string]interface{}{
	"items": []interface{}{
		map[string]interface{}{
			"index": 0, "status": "generated", "entityType": "PROPERTY", "entityId": "prop_001",
			"targetUrl": "https://calderasuite.com/properties/prop_001", "format": "png",
			"fileId": "f_1", "signedUrl": "https://x?sp=r&sig=s", "expiresAt": "2026-11-21T14:30:00.000Z",
			"size": 3812,
		},
		map[string]interface{}{
			"index": 1, "status": "failed", "entityType": nil, "entityId": nil,
			"error": map[string]interface{}{"code": "VALIDATION_ERROR", "message": "entityType is invalid"},
		},
	},
	"totalGenerated": 1, "totalFailed": 1,
	"zip": map[string]interface{}{
		"fileId": "f_z", "signedUrl": "https://z", "expiresAt": "2026-11-21T14:30:00.000Z", "size": 4021, "entries": 1,
	},
	"generatedAt": "2026-11-20T14:30:00.000Z",
}

func TestQrBatchGenerateBatchItemsSendsBodyAndKey(t *testing.T) {
	srv := newTestServer(t, func(r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/qr/batch" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("x-idempotency-key"); got != "idem-q" {
			t.Errorf("x-idempotency-key = %q", got)
		}
		body := decodeJSONBody(t, r)
		if body["format"] != "svg" || body["zip"] != true || body["urlTtlSeconds"] != float64(600) {
			t.Errorf("unexpected body: %#v", body)
		}
		items, _ := body["items"].([]interface{})
		if len(items) != 2 {
			t.Fatalf("items = %#v", body["items"])
		}
		first, _ := items[0].(map[string]interface{})
		if first["entityType"] != "PROPERTY" || first["entityId"] != "prop_001" {
			t.Errorf("item 0 = %#v", first)
		}
		for _, k := range []string{"accessCode", "baseUrl", "options"} {
			if _, ok := first[k]; ok {
				t.Errorf("item 0: %s must not be sent", k)
			}
		}
		second, _ := items[1].(map[string]interface{})
		opts, _ := second["options"].(map[string]interface{})
		utm, _ := opts["utmParams"].(map[string]interface{})
		if second["accessCode"] != "A1B2" || opts["size"] != "print" || opts["margin"] != float64(0) || utm["source"] != "qr" {
			t.Errorf("item 1 = %#v", second)
		}
	}, qrBatchAnswer)
	defer srv.Close()

	zero := 0
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Qr.GenerateBatchItems(QrBatchInput{
		Items: []QrBatchItem{
			{EntityType: "PROPERTY", EntityID: "prop_001"},
			{EntityType: "PRINCIPAL", AccessCode: "A1B2", Options: &QrBatchItemOptions{
				Size: "print", Margin: &zero, UtmParams: &QrUtmParams{Source: "qr"},
			}},
		},
		Format: "svg", Zip: true, URLTTLSeconds: 600,
	}, &QrBatchCallOptions{IdempotencyKey: "idem-q"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalGenerated != 1 || result.TotalFailed != 1 || len(result.Items) != 2 || result.GeneratedAt == "" {
		t.Fatalf("unexpected result: %#v", result)
	}
	gen, failed := result.Items[0], result.Items[1]
	if gen.Status != "generated" || gen.EntityType == nil || *gen.EntityType != "PROPERTY" || gen.FileID != "f_1" || gen.Size != 3812 {
		t.Errorf("generated item = %#v", gen)
	}
	if failed.Status != "failed" || failed.EntityType != nil || failed.EntityID != nil || failed.Error == nil || failed.Error.Code != ErrCodeValidationError {
		t.Errorf("failed item = %#v", failed)
	}
	if result.Zip == nil || result.Zip.Entries != 1 || result.Zip.FileID != "f_z" {
		t.Errorf("zip = %#v", result.Zip)
	}
}

func TestQrBatchGenerateBatchItemsWithoutOptionsSendsNoKey(t *testing.T) {
	srv := newTestServer(t, func(r *http.Request) {
		if got := r.Header.Get("x-idempotency-key"); got != "" {
			t.Errorf("no idempotency key expected, got %q", got)
		}
		body := decodeJSONBody(t, r)
		for _, k := range []string{"format", "zip", "urlTtlSeconds"} {
			if _, ok := body[k]; ok {
				t.Errorf("%s must not be sent", k)
			}
		}
	}, map[string]interface{}{"items": []interface{}{}, "totalGenerated": 0, "totalFailed": 0, "zip": nil, "generatedAt": "x"})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Qr.GenerateBatchItems(QrBatchInput{Items: []QrBatchItem{{EntityType: "AGENT"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Zip != nil {
		t.Errorf("zip = %#v", result.Zip)
	}
}

func TestQrBatchGenerateBatchMapPostsUnchanged(t *testing.T) {
	srv := newTestServer(t, func(r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/qr/batch" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body := decodeJSONBody(t, r)
		if body["zip"] != true {
			t.Errorf("unexpected body: %#v", body)
		}
	}, qrBatchAnswer)
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Qr.GenerateBatch(map[string]interface{}{
		"items": []map[string]interface{}{{"entityType": "PROPERTY", "entityId": "prop_001"}},
		"zip":   true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 || result.Zip == nil {
		t.Errorf("unexpected result: %#v", result)
	}
}

func TestQrBatchServiceDisabledIsAnAPIError(t *testing.T) {
	srv, attempts := cannedServer(t, cannedResponse{403, `{"success":false,"error":{"code":"SERVICE_DISABLED","message":"QR generation is disabled"}}`})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	_, err := c.Qr.GenerateBatchItems(QrBatchInput{Items: []QrBatchItem{{EntityType: "AGENT"}}})
	if apiErr := asAPIError(t, err); apiErr.Code != ErrCodeServiceDisabled || apiErr.StatusCode != 403 {
		t.Errorf("unexpected APIError: %#v", apiErr)
	}
	if *attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", *attempts)
	}
}
