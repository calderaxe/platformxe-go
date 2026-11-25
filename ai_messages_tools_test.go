// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-22, v1.16.0 pending, AI gateway T4):
// - Created (NO-DRIFT). Contract tests for Ai.Messages, AiTools.List and
//   Toolsets on Ai.Complete against testdata/ai_vision.json (byte-identical
//   copy in packages/sdk-python/tests/fixtures/): method, path, the exact JSON
//   body, the typed answer, and the new error envelopes with their retry
//   rule (AI_BUDGET_EXCEEDED and AI_TOOL_FAILED are final; the 4xx codes are
//   never retried).
// - T5: AI_BUDGET_EXCEEDED with Details["reason"] AiBudgetReasonUnavailable
//   and the tool-loop refusal without a period.
// CHANGELOG (2026-11-23, v1.18.0 pending, AI gateway P6):
// - Actors is sent under its wire names (an entry's nil Permissions as []);
//   Actor with Actors is refused before any request.
// CHANGELOG (2026-11-23, v1.17.0 pending, AI gateway P5):
// - Actor is sent under its wire names (omitted when nil); list items decode
//   SourceApp and RequiredPermission.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
)

type aiGatewayFixtures struct {
	MessagesRequest             json.RawMessage      `json:"messagesRequest"`
	MessagesResult              json.RawMessage      `json:"messagesResult"`
	CompleteWithToolsetsRequest json.RawMessage      `json:"completeWithToolsetsRequest"`
	CompleteWithToolsetsResult  json.RawMessage      `json:"completeWithToolsetsResult"`
	ToolList                    json.RawMessage      `json:"toolList"`
	ProfileView                 json.RawMessage      `json:"profileView"`
	BudgetExceededError         aiVisionErrorFixture `json:"budgetExceededError"`
	RefusedError                aiVisionErrorFixture `json:"refusedError"`
	ToolsetDeniedError          aiVisionErrorFixture `json:"toolsetDeniedError"`
	ToolFailedError             aiVisionErrorFixture `json:"toolFailedError"`
	LoopLimitError              aiVisionErrorFixture `json:"loopLimitError"`
}

func loadAiGatewayFixtures(t *testing.T) aiGatewayFixtures {
	t.Helper()
	raw, err := os.ReadFile("testdata/ai_vision.json")
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	var f aiGatewayFixtures
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("parse fixtures: %v", err)
	}
	return f
}

func TestAiMessagesMatchesFixture(t *testing.T) {
	f := loadAiGatewayFixtures(t)
	srv := fixtureServer(t, f.MessagesResult, nil, func(r *http.Request, body []byte) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/ai/messages" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		assertSameJSON(t, "messages body", body, f.MessagesRequest)
	})
	var input AiMessagesInput
	decodeInto(t, f.MessagesRequest, &input)

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Ai.Messages(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var want AiMessagesResult
	decodeInto(t, f.MessagesResult, &want)
	if !reflect.DeepEqual(*result, want) {
		t.Errorf("result mismatch: %#v", result)
	}
	if result.Usage.CacheReadTokens != 1200 || result.Usage.CostCredits != 14 {
		t.Errorf("usage not decoded: %#v", result.Usage)
	}
	if len(result.ToolRuns) != 1 || result.ToolRuns[0].Status != "OK" || result.ToolRuns[0].DenyReason != nil {
		t.Errorf("toolRuns not decoded: %#v", result.ToolRuns)
	}
	if result.Message.StopReason == nil || *result.Message.StopReason != "end_turn" || result.Message.StopDetails != nil {
		t.Errorf("stop fields not decoded: %#v", result.Message)
	}
}

func TestAiMessagesOmitsUnsetFields(t *testing.T) {
	srv := fixtureServer(t, json.RawMessage(`{"message":{"id":"m","role":"assistant","content":[],"stop_reason":null,"stop_details":null},"model":"claude-sonnet-5-5","usage":{}}`), nil, func(_ *http.Request, body []byte) {
		var m map[string]interface{}
		_ = json.Unmarshal(body, &m)
		for _, k := range []string{"model", "system", "effort", "outputSchema", "toolsets", "tools", "toolChoice", "maxTokens", "actor"} {
			if _, ok := m[k]; ok {
				t.Errorf("%s should be omitted when unset", k)
			}
		}
		if m["featureKey"] != "ops-assistant" {
			t.Errorf("featureKey = %v", m["featureKey"])
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Ai.Messages(AiMessagesInput{
		FeatureKey: "ops-assistant",
		Messages:   []AiNativeMessage{{Role: "user", Content: "hi"}},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAiMessagesSendsTheActor(t *testing.T) {
	srv := fixtureServer(t, json.RawMessage(`{"message":{"id":"m","role":"assistant","content":[],"stop_reason":null,"stop_details":null},"model":"claude-sonnet-5-5","usage":{}}`), nil, func(_ *http.Request, body []byte) {
		var m map[string]interface{}
		_ = json.Unmarshal(body, &m)
		actor, ok := m["actor"].(map[string]interface{})
		if !ok {
			t.Fatalf("actor missing: %v", m)
		}
		if actor["ref"] != "adm_123" || actor["grant"] != "Z3JhbnQ" {
			t.Errorf("actor = %v", actor)
		}
		if p, _ := actor["permissions"].([]interface{}); len(p) != 1 || p[0] != "CN:ISSUES:VIEW" {
			t.Errorf("permissions = %v", actor["permissions"])
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Ai.Messages(AiMessagesInput{
		FeatureKey: "desk-brief",
		Messages:   []AiNativeMessage{{Role: "user", Content: "hi"}},
		Toolsets:   []string{"concierge.desk-read"},
		Actor:      &AiToolActor{Ref: "adm_123", Permissions: []string{"CN:ISSUES:VIEW"}, Grant: "Z3JhbnQ"},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAiMessagesSendsActors(t *testing.T) {
	srv := fixtureServer(t, json.RawMessage(`{"message":{"id":"m","role":"assistant","content":[],"stop_reason":null,"stop_details":null},"model":"claude-sonnet-5-5","usage":{}}`), nil, func(_ *http.Request, body []byte) {
		var m map[string]interface{}
		_ = json.Unmarshal(body, &m)
		if _, has := m["actor"]; has {
			t.Errorf("actor must be omitted: %v", m)
		}
		actors, ok := m["actors"].(map[string]interface{})
		if !ok || len(actors) != 2 {
			t.Fatalf("actors = %v", m["actors"])
		}
		l := actors["lettings"].(map[string]interface{})
		if l["ref"] != "adm_l1" || l["grant"] != "bGV0" {
			t.Errorf("lettings = %v", l)
		}
		if p, _ := l["permissions"].([]interface{}); len(p) != 1 || p[0] != "LT:BOOKINGS:VIEW" {
			t.Errorf("permissions = %v", l["permissions"])
		}
		if p, ok := actors["xadmin"].(map[string]interface{})["permissions"].([]interface{}); !ok || len(p) != 0 {
			t.Errorf("a nil Permissions must be [] : %v", actors["xadmin"])
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Ai.Messages(AiMessagesInput{
		FeatureKey: "ai-canvas",
		Messages:   []AiNativeMessage{{Role: "user", Content: "hi"}},
		Toolsets:   []string{"lettings.desk-read"},
		Actors: map[string]AiToolActorsEntry{
			"lettings": {Ref: "adm_l1", Permissions: []string{"LT:BOOKINGS:VIEW"}, Grant: "bGV0"},
			"xadmin":   {Ref: "usr_1", Grant: "eGE"},
		},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAiMessagesRefusesActorWithActors(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: "http://127.0.0.1:1"})
	_, err := c.Ai.Messages(AiMessagesInput{
		FeatureKey: "ai-canvas",
		Actor:      &AiToolActor{Ref: "a"},
		Actors:     map[string]AiToolActorsEntry{"chats": {Ref: "b", Grant: "g"}},
	})
	if !errors.Is(err, ErrAiActorAndActors) {
		t.Fatalf("err = %v, want ErrAiActorAndActors", err)
	}
	raw, _ := json.Marshal(AiMessagesInput{FeatureKey: "f"})
	if strings.Contains(string(raw), "actors") {
		t.Errorf("empty Actors must be omitted: %s", raw)
	}
}

func TestAiToolActorNilPermissionsIsAnEmptyArrayNeverNull(t *testing.T) {
	raw, err := json.Marshal(AiMessagesInput{FeatureKey: "desk-brief", Actor: &AiToolActor{Ref: "adm_1"}})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	_ = json.Unmarshal(raw, &m)
	actor := m["actor"].(map[string]interface{})
	p, ok := actor["permissions"].([]interface{})
	if !ok || len(p) != 0 {
		t.Errorf("permissions = %#v, want []", actor["permissions"])
	}
	if _, has := actor["grant"]; has {
		t.Errorf("an empty grant must be omitted: %v", actor)
	}
	same, _ := json.Marshal(AiToolActor{Ref: "adm_1", Permissions: []string{"CN:ISSUES:VIEW"}})
	if string(same) != `{"ref":"adm_1","permissions":["CN:ISSUES:VIEW"]}` {
		t.Errorf("actor = %s", same)
	}
}

func TestAiToolListItemDecodesSourceAndPermission(t *testing.T) {
	var item AiToolListItem
	raw := `{"name":"concierge_payment_state","version":1,"description":"d","inputSchema":{},"dataClasses":["MONEY"],"toolset":"concierge.desk-read","sourceApp":"concierge","requiredPermission":"CN:PAYMENTS:VIEW"}`
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		t.Fatal(err)
	}
	if item.SourceApp != "concierge" || item.RequiredPermission == nil || *item.RequiredPermission != "CN:PAYMENTS:VIEW" {
		t.Errorf("item = %+v", item)
	}
	var platform AiToolListItem
	_ = json.Unmarshal([]byte(`{"name":"platformxe_x","sourceApp":"platformx","requiredPermission":null}`), &platform)
	if platform.RequiredPermission != nil {
		t.Errorf("a null permission must decode to nil")
	}
}

func TestAiMessagesSendsToolsUnderTheirWireNames(t *testing.T) {
	srv := fixtureServer(t, json.RawMessage(`{"message":{"id":"m","role":"assistant","content":[],"stop_reason":null,"stop_details":null},"model":"claude-sonnet-5-5","usage":{}}`), nil, func(_ *http.Request, body []byte) {
		var m map[string]interface{}
		_ = json.Unmarshal(body, &m)
		tools, _ := m["tools"].([]interface{})
		if len(tools) != 1 {
			t.Fatalf("tools = %v", m["tools"])
		}
		if _, ok := tools[0].(map[string]interface{})["input_schema"]; !ok {
			t.Errorf("tool must carry input_schema: %v", tools[0])
		}
		if _, ok := m["outputSchema"]; !ok {
			t.Errorf("outputSchema missing")
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Ai.Messages(AiMessagesInput{
		FeatureKey:   "ops-assistant",
		Messages:     []AiNativeMessage{{Role: "user", Content: "hi"}},
		Tools:        []AiNativeTool{{Name: "lookup", InputSchema: map[string]interface{}{"type": "object"}}},
		OutputSchema: map[string]interface{}{"type": "object"},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAiNewClosedCodesAreReturnedAfterOneAttempt(t *testing.T) {
	f := loadAiGatewayFixtures(t)
	cases := []struct {
		name    string
		fixture aiVisionErrorFixture
		code    string
		status  int
		details map[string]interface{}
	}{
		{"refused", f.RefusedError, ErrCodeAiRefused, 422, map[string]interface{}{"category": "cyber"}},
		{"loop limit", f.LoopLimitError, ErrCodeAiLoopLimit, 422, map[string]interface{}{"reason": "TURNS"}},
		{"toolset denied", f.ToolsetDeniedError, ErrCodeAiToolsetDenied, 403, map[string]interface{}{"reason": "NOT_ALLOWED"}},
		{"tool failed", f.ToolFailedError, ErrCodeAiToolFailed, 502, nil},
		{"budget exceeded", f.BudgetExceededError, ErrCodeAiBudgetExceeded, 429, map[string]interface{}{"period": "DAY", "retryAfterSeconds": float64(52200)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, attempts := errorFixtureServer(t, tc.fixture)
			c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 2})
			_, err := c.Ai.Messages(AiMessagesInput{FeatureKey: "ops-assistant", Messages: []AiNativeMessage{{Role: "user", Content: "hi"}}})
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("want *APIError, got %v", err)
			}
			if apiErr.Code != tc.code || apiErr.StatusCode != tc.status {
				t.Errorf("got %s %d", apiErr.Code, apiErr.StatusCode)
			}
			if tc.details != nil && !reflect.DeepEqual(apiErr.Details, tc.details) {
				t.Errorf("details = %#v", apiErr.Details)
			}
			if *attempts != 1 {
				t.Errorf("attempts = %d, want 1", *attempts)
			}
		})
	}
}

func TestAiBudgetExceededCarriesRetryAfterAndIsFinalOnAGet(t *testing.T) {
	f := loadAiGatewayFixtures(t)
	if !finalErrorCodes[ErrCodeAiBudgetExceeded] || !finalErrorCodes[ErrCodeAiToolFailed] {
		t.Fatalf("AI_BUDGET_EXCEEDED and AI_TOOL_FAILED must be final")
	}
	srv, attempts := errorFixtureServer(t, f.BudgetExceededError)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 2})
	_, err := c.Ai.Profile("ops-assistant")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want *APIError, got %v", err)
	}
	if apiErr.RetryAfter != 52200 {
		t.Errorf("RetryAfter = %v", apiErr.RetryAfter)
	}
	if *attempts != 1 {
		t.Errorf("attempts = %d, want 1", *attempts)
	}
}

func TestAiBudgetUnavailableAndLoopRefusalDetails(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		wantReason interface{}
		wantPeriod interface{}
	}{
		{"unchecked", `{"success":false,"error":{"code":"AI_BUDGET_EXCEEDED","message":"budget","details":{"period":"MONTH","retryAfterSeconds":60,"reason":"BUDGET_UNAVAILABLE"}}}`, AiBudgetReasonUnavailable, "MONTH"},
		{"loop", `{"success":false,"error":{"code":"AI_BUDGET_EXCEEDED","message":"budget","details":{"retryAfterSeconds":60}}}`, nil, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, attempts := errorFixtureServer(t, aiVisionErrorFixture{Status: 429, Headers: map[string]string{"Retry-After": "60"}, Body: json.RawMessage(tc.body)})
			c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 2})
			_, err := c.Ai.Messages(AiMessagesInput{FeatureKey: "ops-assistant", Messages: []AiNativeMessage{{Role: "user", Content: "hi"}}})
			var apiErr *APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("want *APIError, got %v", err)
			}
			if apiErr.Code != ErrCodeAiBudgetExceeded || apiErr.RetryAfter != 60 {
				t.Errorf("Code = %q, RetryAfter = %v", apiErr.Code, apiErr.RetryAfter)
			}
			if got := apiErr.Details["reason"]; got != tc.wantReason {
				t.Errorf("Details[reason] = %v, want %v", got, tc.wantReason)
			}
			if got := apiErr.Details["period"]; got != tc.wantPeriod {
				t.Errorf("Details[period] = %v, want %v", got, tc.wantPeriod)
			}
			if *attempts != 1 {
				t.Errorf("attempts = %d, want 1", *attempts)
			}
		})
	}
}

func TestAiCompleteWithToolsetsMatchesFixture(t *testing.T) {
	f := loadAiGatewayFixtures(t)
	srv := fixtureServer(t, f.CompleteWithToolsetsResult, nil, func(r *http.Request, body []byte) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/ai/complete" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		assertSameJSON(t, "complete body", body, f.CompleteWithToolsetsRequest)
	})
	var input AiCompleteInput
	decodeInto(t, f.CompleteWithToolsetsRequest, &input)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Ai.Complete(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.ToolRuns) != 1 || result.ToolRuns[0].DenyReason == nil || *result.ToolRuns[0].DenyReason != "DATA_CLASS" {
		t.Errorf("toolRuns not decoded: %#v", result.ToolRuns)
	}
	if result.ToolRuns[0].OutputBytes != nil || result.ToolRuns[0].ToolCallID == nil {
		t.Errorf("null outputBytes must decode as nil, toolCallId as set: %#v", result.ToolRuns[0])
	}
}

func TestAiProfileDecodesGovernanceFields(t *testing.T) {
	f := loadAiGatewayFixtures(t)
	srv := fixtureServer(t, f.ProfileView, nil, nil)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	view, err := c.Ai.Profile("guest-concierge")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if view.MaxEffort != "medium" || view.MaxToolTurns != 8 || view.FallbackModel != nil || !view.AllowCaching {
		t.Errorf("governance not decoded: %#v", view)
	}
}

func TestAiToolsListMatchesFixture(t *testing.T) {
	f := loadAiGatewayFixtures(t)
	srv := fixtureServer(t, f.ToolList, nil, func(r *http.Request, body []byte) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/ai/tools" || r.URL.RawQuery != "" {
			t.Errorf("unexpected %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
		if len(body) != 0 {
			t.Errorf("GET must send no body, got %s", body)
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.AiTools.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var want AiToolList
	decodeInto(t, f.ToolList, &want)
	if !reflect.DeepEqual(*result, want) {
		t.Errorf("result mismatch: %#v", result)
	}
	if result.Total != 1 || result.Items[0].Toolset != "platform.read" {
		t.Errorf("list not decoded: %#v", result)
	}
}

func TestAiToolsListMissingScopeIsForbidden(t *testing.T) {
	srv, attempts := errorFixtureServer(t, aiVisionErrorFixture{
		Status: 403,
		Body:   json.RawMessage(`{"success":false,"error":{"code":"FORBIDDEN","message":"Missing scope"}}`),
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 2})
	_, err := c.AiTools.List()
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != "FORBIDDEN" || apiErr.StatusCode != 403 {
		t.Fatalf("want 403 FORBIDDEN, got %v", err)
	}
	if *attempts != 1 {
		t.Errorf("attempts = %d, want 1", *attempts)
	}
}
