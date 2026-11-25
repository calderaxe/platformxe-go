// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-20, v1.13.0 pending): Initial (fidelity register fixes).
// Pins the method, path and body of the two calls the register found off the
// API: Workflows.DryRun (POST /api/v1/workflows/{id} with {entityData}; it
// posted to a /dry-run path that does not exist) and Storage.ReorderFiles
// (PATCH; it sent POST).
// CHANGELOG (2026-11-20, fix round): the handlers record what they saw and
// never call t.Fatalf on the server goroutine (N2); the deprecated dry-run
// fields are filled (M1); the context rules match Python (L1); the id is
// path-escaped and an empty id never reaches the server (L2); List, Create,
// Get, Update and Delete decode the shapes the routes answer (O1).
// CHANGELOG (2026-11-20, re-review R2): the standard trigger context is unwrapped.
// CHANGELOG (2026-11-20, v1.13.0 pending, fidelity W2): Events.Ingest sends a
// deprecated "payload" as entityData (the route never read payload).
// CHANGELOG (2026-11-20, fidelity W2 review round 1): a typed-nil payload or
// entityData is never sent as null (L1); the caller's map is compared by
// content, not length (L5).
// =============================================================================

package platformxe

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// seenRequest is what a test server saw; decodeErr is checked on the test goroutine.
type seenRequest struct {
	method, path, rawPath string
	body                  map[string]interface{}
	decodeErr             error
	calls                 int
}

// recordingServer answers every request with {success: true, data} and
// records the last request. It never calls t.Fatalf on the handler goroutine.
func recordingServer(t *testing.T, data interface{}) (*Client, *seenRequest) {
	t.Helper()
	seen := &seenRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen.calls++
		seen.method, seen.path, seen.rawPath = r.Method, r.URL.Path, r.URL.EscapedPath()
		seen.body, seen.decodeErr = nil, nil
		if r.Body != nil && r.ContentLength != 0 {
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				seen.decodeErr = err
			}
			seen.body = body
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": data})
	}))
	t.Cleanup(srv.Close)
	return NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL}), seen
}

func (s *seenRequest) check(t *testing.T, method, rawPath string, body map[string]interface{}) {
	t.Helper()
	if s.decodeErr != nil {
		t.Fatalf("the request body did not decode: %v", s.decodeErr)
	}
	if s.method != method || s.rawPath != rawPath {
		t.Errorf("expected %s %s, got %s %s", method, rawPath, s.method, s.rawPath)
	}
	if body != nil && !reflect.DeepEqual(s.body, body) {
		t.Errorf("expected body %v, got %v", body, s.body)
	}
}

var dryRunAnswer = map[string]interface{}{"dryRun": map[string]interface{}{
	"wouldMatch":          true,
	"conditionsEvaluated": map[string]interface{}{"amount": true},
	"actionsWouldExecute": []interface{}{map[string]interface{}{"type": "SEND_EMAIL", "config": map[string]interface{}{}}},
	"actionsPreview": []interface{}{map[string]interface{}{
		"type": "SEND_EMAIL", "config": map[string]interface{}{}, "interpolatedConfig": map[string]interface{}{"to": "a@b.com"},
	}},
	"validation": map[string]interface{}{"ok": true, "errors": []interface{}{}},
}}

func TestDryRunPostsToTheTriggerRouteAndDecodesTheResult(t *testing.T) {
	c, seen := recordingServer(t, dryRunAnswer)

	result, err := c.Workflows.DryRunEntityData("trg_1", map[string]interface{}{"amount": 5.0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	seen.check(t, "POST", "/api/v1/workflows/trg_1", map[string]interface{}{"entityData": map[string]interface{}{"amount": 5.0}})
	if !result.WouldMatch || !result.ConditionsEvaluated["amount"] || len(result.ActionsWouldExecute) != 1 {
		t.Errorf("unexpected result: %+v", result)
	}
	if len(result.ActionsPreview) != 1 || result.ActionsPreview[0].InterpolatedConfig["to"] != "a@b.com" {
		t.Errorf("unexpected preview: %+v", result.ActionsPreview)
	}
	if result.Validation == nil || !result.Validation.OK {
		t.Errorf("unexpected validation: %+v", result.Validation)
	}
	// The deprecated fields, for code written to the old result (M1).
	if !result.Matched || result.TriggerID != "trg_1" || len(result.Actions) != 1 || result.Actions[0]["type"] != "SEND_EMAIL" {
		t.Errorf("deprecated fields not filled: %+v", result)
	}
}

type entityPayload struct {
	ID string `json:"id"`
}

func TestDryRunReadsTheContextAsPythonDoes(t *testing.T) {
	cases := []struct {
		name    string
		context map[string]interface{}
		sent    map[string]interface{}
	}{
		{"lone entityData", map[string]interface{}{"entityData": map[string]interface{}{"a": 1.0}}, map[string]interface{}{"a": 1.0}},
		{"eventType and entityData", map[string]interface{}{"eventType": "INVOICE_PAID", "entityData": map[string]interface{}{"a": 1.0}}, map[string]interface{}{"a": 1.0}},
		{"old eventType and payload", map[string]interface{}{"eventType": "INVOICE_PAID", "payload": map[string]interface{}{"b": 2.0}}, map[string]interface{}{"b": 2.0}},
		{"standard trigger context", map[string]interface{}{
			"eventType": "INVOICE_PAID", "entityType": "INVOICE", "entityId": "inv_1", "entityData": map[string]interface{}{"amount": 250000.0},
		}, map[string]interface{}{"amount": 250000.0}},
		{"context keys with payload", map[string]interface{}{
			"eventType": "X", "sourceApp": "lettings", "metadata": map[string]interface{}{"m": 1.0}, "payload": map[string]interface{}{"b": 2.0},
		}, map[string]interface{}{"b": 2.0}},
		{"old shape with a struct payload", map[string]interface{}{"eventType": "INVOICE_PAID", "payload": entityPayload{ID: "x"}}, map[string]interface{}{"id": "x"}},
		{"old shape with a string map payload", map[string]interface{}{"eventType": "INVOICE_PAID", "payload": map[string]string{"id": "x"}}, map[string]interface{}{"id": "x"}},
		{"raw entity data", map[string]interface{}{"c": 3.0}, map[string]interface{}{"c": 3.0}},
		{"raw entity data with a payload field", map[string]interface{}{"status": "FAILED", "payload": map[string]interface{}{"id": "x"}},
			map[string]interface{}{"status": "FAILED", "payload": map[string]interface{}{"id": "x"}}},
		{"raw entity data with an entityData field", map[string]interface{}{"status": "FAILED", "entityData": map[string]interface{}{"id": "x"}},
			map[string]interface{}{"status": "FAILED", "entityData": map[string]interface{}{"id": "x"}}},
		{"payload that is not an object", map[string]interface{}{"eventType": "X", "payload": "s"}, map[string]interface{}{"eventType": "X", "payload": "s"}},
		{"nil", nil, map[string]interface{}{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, seen := recordingServer(t, dryRunAnswer)
			if _, err := c.Workflows.DryRun("trg_1", tc.context); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			seen.check(t, "POST", "/api/v1/workflows/trg_1", map[string]interface{}{"entityData": tc.sent})
		})
	}
}

func TestWorkflowIDsAreEscapedAndAnEmptyIDNeverReachesTheServer(t *testing.T) {
	c, seen := recordingServer(t, map[string]interface{}{"trigger": map[string]interface{}{"id": "a/b?c"}})

	if _, err := c.Workflows.Get("a/b?c"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	seen.check(t, "GET", "/api/v1/workflows/a%2Fb%3Fc", nil)

	calls := seen.calls
	_, errGet := c.Workflows.Get("")
	_, errUpdate := c.Workflows.Update("", map[string]interface{}{"name": "n"})
	_, errDelete := c.Workflows.Delete("")
	_, errDryRun := c.Workflows.DryRun("", nil)
	for _, err := range []error{errGet, errUpdate, errDelete, errDryRun} {
		if !errors.Is(err, ErrWorkflowIDRequired) {
			t.Errorf("expected ErrWorkflowIDRequired, got %v", err)
		}
	}
	if seen.calls != calls {
		t.Errorf("an empty id reached the server")
	}
}

var triggerAnswer = map[string]interface{}{
	"id": "trg_1", "name": "Welcome", "code": "WELCOME", "triggerType": "EVENT", "eventType": "BOOKING_CREATED",
	"actions": []interface{}{map[string]interface{}{"type": "SEND_EMAIL"}}, "isActive": true, "priority": 1,
}

func TestWorkflowListDecodesTriggersAndPagination(t *testing.T) {
	c, seen := recordingServer(t, map[string]interface{}{
		"triggers":   []interface{}{triggerAnswer},
		"pagination": map[string]interface{}{"page": 1, "limit": 20, "total": 1},
	})
	result, err := c.Workflows.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	seen.check(t, "GET", "/api/v1/workflows", nil)
	if len(result.Triggers) != 1 || result.Triggers[0].ID != "trg_1" || result.Triggers[0].Name != "Welcome" {
		t.Errorf("triggers not decoded: %+v", result.Triggers)
	}
	if result.Pagination == nil || result.Pagination.Total != 1 || result.Pagination.Limit != 20 {
		t.Errorf("pagination not decoded: %+v", result.Pagination)
	}
	if len(result.Workflows) != 1 || result.Workflows[0].ID != "trg_1" {
		t.Errorf("the deprecated Workflows field is the triggers: %+v", result.Workflows)
	}
}

func TestWorkflowSingleTriggerCallsDecodeTheNestedTrigger(t *testing.T) {
	answer := map[string]interface{}{"trigger": triggerAnswer, "recentExecutions": []interface{}{}}
	input := map[string]interface{}{"name": "Welcome"}

	calls := []struct {
		name, method, path string
		run                func(c *Client) (*WorkflowTrigger, error)
	}{
		{"Create", "POST", "/api/v1/workflows", func(c *Client) (*WorkflowTrigger, error) { return c.Workflows.Create(input) }},
		{"Get", "GET", "/api/v1/workflows/trg_1", func(c *Client) (*WorkflowTrigger, error) { return c.Workflows.Get("trg_1") }},
		{"Update", "PATCH", "/api/v1/workflows/trg_1", func(c *Client) (*WorkflowTrigger, error) { return c.Workflows.Update("trg_1", input) }},
	}
	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			c, seen := recordingServer(t, answer)
			trigger, err := tc.run(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			seen.check(t, tc.method, tc.path, nil)
			if trigger.ID != "trg_1" || trigger.Code != "WELCOME" || len(trigger.Actions) != 1 || !trigger.IsActive {
				t.Errorf("trigger not decoded: %+v", trigger)
			}
		})
	}

	t.Run("Delete", func(t *testing.T) {
		c, seen := recordingServer(t, map[string]interface{}{"trigger": map[string]interface{}{"id": "trg_1", "isActive": false}})
		result, err := c.Workflows.Delete("trg_1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		seen.check(t, "DELETE", "/api/v1/workflows/trg_1", nil)
		if !result.Deleted {
			t.Errorf("a switched-off trigger is Deleted true")
		}
	})
}

func TestReorderFilesSendsPatch(t *testing.T) {
	c, seen := recordingServer(t, map[string]interface{}{"reordered": true})
	result, err := c.Storage.ReorderFiles([]string{"f1", "f2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	seen.check(t, "PATCH", "/api/v1/storage/media/files/reorder", map[string]interface{}{"fileIds": []interface{}{"f1", "f2"}})
	if !result.Reordered {
		t.Errorf("reordered is decoded")
	}
}

func TestIngestSendsPayloadAsEntityData(t *testing.T) {
	cases := []struct {
		input map[string]interface{}
		sent  map[string]interface{}
	}{
		{
			map[string]interface{}{"eventType": "BOOKING_CONFIRMED", "entityId": "bk_1", "payload": map[string]interface{}{"a": 1.0}},
			map[string]interface{}{"eventType": "BOOKING_CONFIRMED", "entityId": "bk_1", "entityData": map[string]interface{}{"a": 1.0}},
		},
		{
			map[string]interface{}{"eventType": "BOOKING_CONFIRMED", "entityData": map[string]interface{}{"b": 2.0}, "payload": map[string]interface{}{"a": 1.0}},
			map[string]interface{}{"eventType": "BOOKING_CONFIRMED", "entityData": map[string]interface{}{"b": 2.0}},
		},
		{
			map[string]interface{}{"eventType": "BOOKING_CONFIRMED", "entityData": map[string]interface{}{"b": 2.0}, "sourceApp": "lettings"},
			map[string]interface{}{"eventType": "BOOKING_CONFIRMED", "entityData": map[string]interface{}{"b": 2.0}, "sourceApp": "lettings"},
		},
		// L1: a typed-nil payload map is omitted, never sent as entityData: null.
		{
			map[string]interface{}{"eventType": "BOOKING_CONFIRMED", "payload": map[string]interface{}(nil)},
			map[string]interface{}{"eventType": "BOOKING_CONFIRMED"},
		},
		{
			map[string]interface{}{"eventType": "BOOKING_CONFIRMED", "payload": nil},
			map[string]interface{}{"eventType": "BOOKING_CONFIRMED"},
		},
		// A typed-nil entityData gives way to a payload, and is omitted alone.
		{
			map[string]interface{}{"eventType": "BOOKING_CONFIRMED", "entityData": map[string]interface{}(nil), "payload": map[string]interface{}{"a": 1.0}},
			map[string]interface{}{"eventType": "BOOKING_CONFIRMED", "entityData": map[string]interface{}{"a": 1.0}},
		},
		{
			map[string]interface{}{"eventType": "BOOKING_CONFIRMED", "entityData": map[string]interface{}(nil)},
			map[string]interface{}{"eventType": "BOOKING_CONFIRMED"},
		},
	}
	for _, tc := range cases {
		var method, path string
		var body map[string]interface{}
		srv := newTestServer(t, func(r *http.Request) {
			method, path = r.Method, r.URL.Path
			body = decodeJSONBody(t, r)
		}, map[string]interface{}{
			"eventType": "BOOKING_CONFIRMED", "entityType": "BOOKING", "entityId": "bk_1", "sourceApp": "lettings",
			"triggersEvaluated": 1, "triggersMatched": 1, "triggersExecuted": 1,
			"results": []interface{}{map[string]interface{}{"triggerId": "trg_1", "triggerName": "T", "matched": true, "executed": true}},
		})
		c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
		original := copyIngestInput(tc.input)
		result, err := c.Events.Ingest(tc.input)
		srv.Close()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if method != "POST" || path != "/api/v1/events" {
			t.Errorf("expected POST /api/v1/events, got %s %s", method, path)
		}
		if !reflect.DeepEqual(body, tc.sent) {
			t.Errorf("input %v: expected body %v, got %v", tc.input, tc.sent, body)
		}
		if !reflect.DeepEqual(tc.input, original) {
			t.Errorf("the caller's map was modified: %v, was %v", tc.input, original)
		}
		if result.EntityType != "BOOKING" || result.EntityID != "bk_1" || result.TriggersExecuted != 1 || len(result.Results) != 1 {
			t.Errorf("unexpected result: %+v", result)
		}
	}
}

// copyIngestInput deep-copies the nested maps of an Ingest input, keeping a
// typed nil as it is, so a change anywhere in the caller's map is seen.
func copyIngestInput(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		if m, ok := v.(map[string]interface{}); ok && m != nil {
			out[k] = copyIngestInput(m)
		} else {
			out[k] = v
		}
	}
	return out
}
