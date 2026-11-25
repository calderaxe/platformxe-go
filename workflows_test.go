// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-12): Initial (caldera-platformxe Task 4 fix round 3, I-1,
// NO-DRIFT). Pins the /api/v1/workflows/evaluate result as the route answers
// it: each matched trigger's actionsExecuted with every step status (SUCCESS,
// FAILED, SKIPPED, DEFERRED), and an error item with no executionId (no
// execution could be recorded). Uses newTestServer from documents_test.go.
// CHANGELOG (2026-11-12): Task 4 fix round 4 (N3-I1): a recorded step is in
// either spelling; WorkflowActionResultStatus.Normalize upper-cases it.
// Evaluate sends entityData (the route never read payload); EvaluateWithOptions
// sends the route's other fields; Message and ActionsWouldExecute are decoded.
// CHANGELOG (2026-11-20/21, v1.13.0 pending): List's query (re-review R5); more
// than one List option is an error (review N1); "." / ".." ids refused (L8).
// CHANGELOG (2026-11-22, v1.14.1 pending): entityData, the custom-parameters
// object: a nil or empty map is not sent by Evaluate / EvaluateWithOptions
// (omitempty; the route reads none); Ingest sends an empty map as {} and omits
// a nil one.
// CHANGELOG (2026-11-22, review round 1): the Evaluate test asserts each
// request's body (both calls), and the test servers close with defer.
// =============================================================================

package platformxe

import (
	"errors"
	"net/http"
	"reflect"
	"testing"
)

func TestEvaluateSendsEntityData(t *testing.T) {
	var body map[string]interface{}
	srv := newTestServer(t, func(r *http.Request) {
		body = decodeJSONBody(t, r)
	}, map[string]interface{}{"matched": 0, "total": 0, "results": []interface{}{}, "message": "No active triggers for event"})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Workflows.Evaluate("BOOKING_CREATED", map[string]interface{}{"guestEmail": "a@b.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]interface{}{"eventType": "BOOKING_CREATED", "entityData": map[string]interface{}{"guestEmail": "a@b.com"}}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("expected body %v, got %v", want, body)
	}
	if result.Message == "" {
		t.Errorf("the message is decoded")
	}
}

func TestEvaluateWithOptionsSendsEveryFieldAndDecodesADryRun(t *testing.T) {
	var body map[string]interface{}
	srv := newTestServer(t, func(r *http.Request) {
		body = decodeJSONBody(t, r)
	}, map[string]interface{}{
		"matched": 1, "total": 1,
		"results": []interface{}{map[string]interface{}{
			"triggerId": "trg_1", "triggerCode": "A", "triggerName": "A", "matched": true, "dryRun": true,
			"actionsWouldExecute": []interface{}{map[string]interface{}{"type": "SEND_EMAIL", "config": map[string]interface{}{}}},
		}},
	})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Workflows.EvaluateWithOptions(WorkflowEvaluateInput{
		EventType: "BOOKING_CREATED", EntityType: "BOOKING", EntityID: "bk_1",
		EntityData: map[string]interface{}{"x": 1.0}, SourceApp: "lettings",
		Metadata: map[string]interface{}{"m": 1.0}, DryRun: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]interface{}{
		"eventType": "BOOKING_CREATED", "entityType": "BOOKING", "entityId": "bk_1",
		"entityData": map[string]interface{}{"x": 1.0}, "sourceApp": "lettings",
		"metadata": map[string]interface{}{"m": 1.0}, "dryRun": true,
	}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("expected body %v, got %v", want, body)
	}
	d := result.Results[0]
	if !d.DryRun || len(d.ActionsWouldExecute) != 1 || d.ActionsWouldExecute[0]["type"] != "SEND_EMAIL" {
		t.Errorf("unexpected dry-run item: %+v", d)
	}
}

func TestEvaluateDecodesEveryStepStatusAndAnErrorItem(t *testing.T) {
	srv := newTestServer(t, func(r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/workflows/evaluate" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}, map[string]interface{}{
		"matched": 2,
		"total":   2,
		"results": []interface{}{
			map[string]interface{}{
				"triggerId": "trg_1", "triggerCode": "A", "triggerName": "A", "matched": true,
				"executionId": "exec_1", "durationMs": 12,
				"actionsExecuted": []interface{}{
					map[string]interface{}{"type": "SEND_EMAIL", "config": map[string]interface{}{}, "status": "DEFERRED", "error": "Not sent: the sending limit for this recipient was reached."},
					map[string]interface{}{"type": "UNKNOWN", "config": map[string]interface{}{}, "status": "SKIPPED"},
					map[string]interface{}{"type": "CREATE_TASK", "config": map[string]interface{}{}, "status": "SUCCESS"},
					map[string]interface{}{"type": "WEBHOOK", "config": map[string]interface{}{}, "status": "FAILED", "error": "timeout"},
				},
			},
			map[string]interface{}{
				"triggerId": "trg_2", "triggerCode": "B", "triggerName": "B", "matched": true,
				"error": "invalid input value for enum entity_type",
			},
		},
	})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Workflows.Evaluate("BOOKING_CREATED", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	steps := result.Results[0].ActionsExecuted
	want := []WorkflowExecutedActionStatus{
		WorkflowActionDeferred, WorkflowActionSkipped, WorkflowActionSucceeded, WorkflowActionFailed,
	}
	if len(steps) != len(want) {
		t.Fatalf("expected %d steps, got %d", len(want), len(steps))
	}
	for i, s := range steps {
		if s.Status != want[i] {
			t.Errorf("step %d: expected %s, got %s", i, want[i], s.Status)
		}
	}
	if steps[0].Error == "" {
		t.Errorf("a deferred step carries its reason")
	}
	if result.Results[0].ExecutionID != "exec_1" || result.Results[0].DurationMs != 12 {
		t.Errorf("unexpected first item: %+v", result.Results[0])
	}
	if result.Results[1].ExecutionID != "" || result.Results[1].Error == "" {
		t.Errorf("an error item has no executionId and an error: %+v", result.Results[1])
	}
}

func TestWorkflowStepStatusValues(t *testing.T) {
	cases := map[WorkflowExecutedActionStatus]string{
		WorkflowActionSucceeded: "SUCCESS", WorkflowActionFailed: "FAILED",
		WorkflowActionSkipped: "SKIPPED", WorkflowActionDeferred: "DEFERRED",
	}
	for got, want := range cases {
		if string(got) != want {
			t.Errorf("expected %s, got %s", want, got)
		}
	}
	steps := map[WorkflowActionResultStatus]string{
		WorkflowStepSucceeded: "success", WorkflowStepFailed: "failed",
		WorkflowStepSkipped: "skipped", WorkflowStepDeferred: "deferred",
	}
	for got, want := range steps {
		if string(got) != want {
			t.Errorf("expected %s, got %s", want, got)
		}
	}
}

func TestWorkflowStepStatusNormalizeAcceptsBothSpellings(t *testing.T) {
	cases := map[WorkflowActionResultStatus]WorkflowExecutedActionStatus{
		WorkflowStepSucceeded: WorkflowActionSucceeded, "SUCCESS": WorkflowActionSucceeded,
		WorkflowStepFailed: WorkflowActionFailed, "FAILED": WorkflowActionFailed,
		WorkflowStepSkipped: WorkflowActionSkipped, "SKIPPED": WorkflowActionSkipped,
		WorkflowStepDeferred: WorkflowActionDeferred, "DEFERRED": WorkflowActionDeferred,
	}
	for in, want := range cases {
		if got := in.Normalize(); got != want {
			t.Errorf("%s: expected %s, got %s", in, want, got)
		}
	}
}

// TestWorkflowListQuery pins List's query string (re-review R5): no options,
// a nil option and a zero option send none; every field is the route's name.
func TestWorkflowListQuery(t *testing.T) {
	active, inactive := true, false
	answer := map[string]interface{}{
		"triggers":   []interface{}{map[string]interface{}{"id": "trg_2", "name": "Two", "code": "TWO", "triggerType": "EVENT"}},
		"pagination": map[string]interface{}{"page": 2, "limit": 1, "total": 3},
	}
	cases := []struct {
		name string
		call func(*WorkflowsService) (*WorkflowListResponse, error)
		want string
	}{
		{"no options", func(s *WorkflowsService) (*WorkflowListResponse, error) { return s.List() }, ""},
		{"nil options", func(s *WorkflowsService) (*WorkflowListResponse, error) { return s.List(nil) }, ""},
		{"zero options", func(s *WorkflowsService) (*WorkflowListResponse, error) { return s.List(&WorkflowListOptions{}) }, ""},
		{"page and limit", func(s *WorkflowsService) (*WorkflowListResponse, error) {
			return s.List(&WorkflowListOptions{Page: 2, Limit: 1})
		}, "limit=1&page=2"},
		{"every filter", func(s *WorkflowsService) (*WorkflowListResponse, error) {
			return s.List(&WorkflowListOptions{Page: 3, Limit: 25, IsActive: &active, TriggerType: "EVENT", EventType: "BOOKING_CREATED"})
		}, "eventType=BOOKING_CREATED&isActive=true&limit=25&page=3&triggerType=EVENT"},
		{"isActive false is sent", func(s *WorkflowsService) (*WorkflowListResponse, error) {
			return s.List(&WorkflowListOptions{IsActive: &inactive})
		}, "isActive=false"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var query string
			var path string
			srv := newTestServer(t, func(r *http.Request) {
				path = r.URL.Path
				query = r.URL.Query().Encode()
			}, answer)
			defer srv.Close()
			c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
			got, err := tc.call(c.Workflows)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if path != "/api/v1/workflows" || query != tc.want {
				t.Errorf("expected GET /api/v1/workflows?%s, got %s?%s", tc.want, path, query)
			}
			if len(got.Triggers) != 1 || got.Triggers[0].ID != "trg_2" || got.Pagination == nil || got.Pagination.Total != 3 || got.Pagination.Page != 2 {
				t.Errorf("unexpected decode: %+v", got)
			}
			if len(got.Workflows) != 1 {
				t.Errorf("the deprecated Workflows is filled from Triggers: %+v", got.Workflows)
			}
		})
	}
}

// Review N1: more than one option is an error, never a silent drop.
func TestWorkflowListRefusesMoreThanOneOption(t *testing.T) {
	calls := 0
	srv := newTestServer(t, func(*http.Request) { calls++ }, map[string]interface{}{"triggers": []interface{}{}})
	defer srv.Close()
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Workflows.List(&WorkflowListOptions{Page: 1}, &WorkflowListOptions{Page: 2}); err == nil {
		t.Fatal("want an error for two options")
	}
	if calls != 0 {
		t.Errorf("no request is sent, got %d", calls)
	}
}

// Review L8: an id of "." or ".." never reaches the server (it would address
// another route once normalised), on every service that fills a path.
func TestDotSegmentIDsAreRefusedBeforeAnyRequest(t *testing.T) {
	calls := 0
	srv := newTestServer(t, func(*http.Request) { calls++ }, map[string]interface{}{})
	defer srv.Close()
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	for _, id := range []string{".", ".."} {
		if _, err := c.Workflows.Get(id); !errors.Is(err, ErrDotSegmentPath) {
			t.Errorf("Workflows.Get(%q): want ErrDotSegmentPath, got %v", id, err)
		}
		if _, err := c.Workflows.DryRunEntityData(id, nil); !errors.Is(err, ErrDotSegmentPath) {
			t.Errorf("Workflows.DryRunEntityData(%q): want ErrDotSegmentPath, got %v", id, err)
		}
		if _, err := c.Events.GetSubscription(id); !errors.Is(err, ErrDotSegmentPath) {
			t.Errorf("Events.GetSubscription(%q): want ErrDotSegmentPath, got %v", id, err)
		}
	}
	if calls != 0 {
		t.Errorf("no request is sent, got %d", calls)
	}
	// A dot inside an id is fine.
	if hasDotSegment("/api/v1/workflows/a.b") || hasDotSegment("/api/v1/workflows/x?next=..") {
		t.Error("only a whole . or .. segment before the query is refused")
	}
	if !hasDotSegment("/api/v1/workflows/..") || !hasDotSegment("/api/v1/a/./b") {
		t.Error("dot segments are found")
	}
}

func TestEvaluateEntityDataNilOrEmptyIsNotSent(t *testing.T) {
	want := map[string]interface{}{"eventType": "SERVICE_REQUESTED"}
	for _, data := range []map[string]interface{}{nil, {}} {
		var bodies []map[string]interface{}
		srv := newTestServer(t, func(r *http.Request) {
			bodies = append(bodies, decodeJSONBody(t, r))
		}, map[string]interface{}{"matched": 0, "total": 0, "results": []interface{}{}})
		defer srv.Close()
		c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
		if _, err := c.Workflows.Evaluate("SERVICE_REQUESTED", data); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, err := c.Workflows.EvaluateWithOptions(WorkflowEvaluateInput{EventType: "SERVICE_REQUESTED", EntityData: data}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(bodies) != 2 {
			t.Fatalf("entityData %#v: expected 2 requests, got %d", data, len(bodies))
		}
		for i, body := range bodies {
			if !reflect.DeepEqual(body, want) {
				t.Errorf("entityData %#v, request %d: expected body %v, got %v", data, i+1, want, body)
			}
		}
	}
}

func TestIngestEntityDataEmptyIsSentAndNilIsNot(t *testing.T) {
	cases := []struct {
		input map[string]interface{}
		sent  map[string]interface{}
	}{
		{
			map[string]interface{}{"eventType": "PAYMENT_RECEIVED", "entityData": map[string]interface{}{}},
			map[string]interface{}{"eventType": "PAYMENT_RECEIVED", "entityData": map[string]interface{}{}},
		},
		{
			map[string]interface{}{"eventType": "PAYMENT_RECEIVED", "entityData": nil},
			map[string]interface{}{"eventType": "PAYMENT_RECEIVED"},
		},
	}
	for _, tc := range cases {
		var body map[string]interface{}
		srv := newTestServer(t, func(r *http.Request) {
			body = decodeJSONBody(t, r)
		}, map[string]interface{}{"eventType": "PAYMENT_RECEIVED", "triggersEvaluated": 0, "triggersMatched": 0, "triggersExecuted": 0, "results": []interface{}{}})
		defer srv.Close()
		c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
		if _, err := c.Events.Ingest(tc.input); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(body, tc.sent) {
			t.Errorf("input %v: expected body %v, got %v", tc.input, tc.sent, body)
		}
	}
}
