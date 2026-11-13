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
// =============================================================================

package platformxe

import (
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
