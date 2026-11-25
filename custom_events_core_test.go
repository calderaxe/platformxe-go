// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// Core custom-events request paths and bodies (core gap review, 2026-11-17):
// Update / Publish (PATCH), DryRunEmit, ListPage; the {id} template is
// URL-escaped.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type coreCall struct {
	Method string
	Path   string
	Query  string
	Body   map[string]interface{}
}

func coreRecordingClient(t *testing.T) (*Client, *[]coreCall) {
	t.Helper()
	calls := []coreCall{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]interface{}
		_ = json.Unmarshal(raw, &body)
		calls = append(calls, coreCall{Method: r.Method, Path: r.URL.EscapedPath(), Query: r.URL.RawQuery, Body: body})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{}}`))
	}))
	t.Cleanup(srv.Close)
	return NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, DisableRetries: true}), &calls
}

func TestCustomEventsCorePaths(t *testing.T) {
	c, calls := coreRecordingClient(t)
	ce := c.Events.Custom
	if _, err := ce.Get("cer 1"); err != nil {
		t.Fatal(err)
	}
	desc := "Fired when a booking is cancelled"
	if _, err := ce.Update("cer_1", UpdateCustomEventInput{Description: &desc}); err != nil {
		t.Fatal(err)
	}
	if _, err := ce.Update("cer_1", UpdateCustomEventInput{ClearDescription: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := ce.Publish("cer_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := ce.ListPage(ListCustomEventsParams{Limit: 20, Offset: 40}); err != nil {
		t.Fatal(err)
	}
	v := "1.0.0"
	if _, err := ce.DryRunEmit(EmitCustomEventInput{Name: "lettings.booking.cancelled", Version: &v, Payload: map[string]interface{}{"bookingId": "b1"}}); err != nil {
		t.Fatal(err)
	}

	got := *calls
	if len(got) != 6 {
		t.Fatalf("calls = %d", len(got))
	}
	if got[0].Method != "GET" || got[0].Path != "/api/v1/events/custom/cer%201" {
		t.Errorf("get: %+v", got[0])
	}
	if got[1].Method != "PATCH" || got[1].Body["description"] != desc {
		t.Errorf("update: %+v", got[1])
	}
	if v, ok := got[2].Body["description"]; !ok || v != nil {
		t.Errorf("clear description: %+v", got[2])
	}
	if got[3].Body["status"] != "published" {
		t.Errorf("publish: %+v", got[3])
	}
	if got[4].Query != "limit=20&offset=40" {
		t.Errorf("list page query: %q", got[4].Query)
	}
	if got[5].Path != "/api/v1/events/custom/dry-run" || got[5].Body["mode"] != "emit" || got[5].Body["version"] != "1.0.0" {
		t.Errorf("dry-run emit: %+v", got[5])
	}
}
