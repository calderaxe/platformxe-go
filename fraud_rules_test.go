// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-18, v1.12.0 pending):
// - Created. Every rule status change posts to /rules/:id/transition {target};
//   ShadowReport reads /rules/:id/shadow-report; a case at an MLRO stage is a
//   409 CASE_AT_MLRO_STAGE *APIError, sent once.
//
// CHANGELOG (2026-11-19, v1.13.0 pending):
// - TransitionAtVersion and UpdateAtVersion send expectedVersion; a 409
//   VERSION_MISMATCH is returned once as ErrCodeVersionMismatch;
//   ShadowReportForVersion sends ?version=; FraudRuleVersion.
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

func TestFraudRuleStatusHelpersPostToTransition(t *testing.T) {
	cases := []struct {
		name   string
		call   func(r *FraudRulesService) (map[string]interface{}, error)
		target string
	}{
		{"Draft", func(r *FraudRulesService) (map[string]interface{}, error) { return r.Draft("frl_1") }, "draft"},
		{"Shadow", func(r *FraudRulesService) (map[string]interface{}, error) { return r.Shadow("frl_1") }, "shadow"},
		{"Publish", func(r *FraudRulesService) (map[string]interface{}, error) { return r.Publish("frl_1") }, "published"},
		{"Archive", func(r *FraudRulesService) (map[string]interface{}, error) { return r.Archive("frl_1") }, "archived"},
		{"Transition", func(r *FraudRulesService) (map[string]interface{}, error) { return r.Transition("frl_1", "shadow") }, "shadow"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotMethod, gotPath string
			var gotBody map[string]interface{}
			srv := fixtureServer(t, json.RawMessage(`{"id":"frl_1","status":"`+tc.target+`"}`), nil,
				func(r *http.Request, body []byte) {
					gotMethod, gotPath = r.Method, r.URL.Path
					_ = json.Unmarshal(body, &gotBody)
				})
			c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
			res, err := tc.call(c.Fraud.Rules)
			if err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if gotMethod != "POST" || gotPath != "/api/v1/fraud/rules/frl_1/transition" {
				t.Errorf("sent %s %s", gotMethod, gotPath)
			}
			if !reflect.DeepEqual(gotBody, map[string]interface{}{"target": tc.target}) {
				t.Errorf("body = %v", gotBody)
			}
			if res["status"] != tc.target {
				t.Errorf("status = %v", res["status"])
			}
		})
	}
}

func TestFraudRuleTransitionsMatchTheAPI(t *testing.T) {
	want := map[string][]string{
		"draft":     {"shadow", "archived"},
		"shadow":    {"draft", "published", "archived"},
		"published": {"archived"},
		"archived":  {},
	}
	if !reflect.DeepEqual(FraudRuleTransitions, want) {
		t.Errorf("FraudRuleTransitions = %v", FraudRuleTransitions)
	}
}

func TestFraudRuleTransitionConflictIsReturnedOnce(t *testing.T) {
	srv, attempts := errorFixtureServer(t, aiVisionErrorFixture{
		Status: http.StatusConflict,
		Body:   json.RawMessage(`{"success":false,"error":{"code":"CONFLICT","message":"Transition not allowed from current status"}}`),
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	_, err := c.Fraud.Rules.Publish("frl_1")
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusConflict || apiErr.Code != "CONFLICT" {
		t.Fatalf("err = %v", err)
	}
	if *attempts != 1 {
		t.Errorf("attempts = %d, want 1", *attempts)
	}
}

func TestFraudRuleShadowReport(t *testing.T) {
	var gotPath, gotFrom, gotTo string
	srv := fixtureServer(t, json.RawMessage(`{"ruleId":"frl_1","evaluated":10,"triggered":2}`), nil,
		func(r *http.Request, _ []byte) {
			if r.Method != "GET" {
				t.Errorf("method = %s", r.Method)
			}
			gotPath = r.URL.Path
			gotFrom, gotTo = r.URL.Query().Get("from"), r.URL.Query().Get("to")
		})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	res, err := c.Fraud.Rules.ShadowReport("frl_1", "2026-11-01T00:00:00Z", "")
	if err != nil {
		t.Fatalf("ShadowReport: %v", err)
	}
	if gotPath != "/api/v1/fraud/rules/frl_1/shadow-report" || gotFrom != "2026-11-01T00:00:00Z" || gotTo != "" {
		t.Errorf("sent %s from=%q to=%q", gotPath, gotFrom, gotTo)
	}
	if res["triggered"] != float64(2) {
		t.Errorf("triggered = %v", res["triggered"])
	}
}

func TestFraudCaseTransitionAtMlroStage(t *testing.T) {
	srv, attempts := errorFixtureServer(t, aiVisionErrorFixture{
		Status: http.StatusConflict,
		Body:   json.RawMessage(`{"success":false,"error":{"code":"CASE_AT_MLRO_STAGE","message":"The case is at an MLRO stage"}}`),
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	_, err := c.Fraud.Cases.Transition("fc_1", map[string]interface{}{"target": "resolved"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != ErrCodeCaseAtMlroStage {
		t.Fatalf("err = %v", err)
	}
	if *attempts != 1 {
		t.Errorf("attempts = %d, want 1", *attempts)
	}
}

func TestFraudDecideSendsNestedBody(t *testing.T) {
	var got map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/fraud/decide" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"verdict":"allow"}}`))
	}))
	defer srv.Close()
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Fraud.Decide(DecideRequest{SubjectID: "u1", SubjectKind: "user", Action: "withdraw", ResourceKind: "transaction", ResourceID: "txn_1", Context: map[string]interface{}{"ip": "1.2.3.4"}}); err != nil {
		t.Fatalf("Decide: %v", err)
	}
	want := `{"action":"withdraw","context":{"ip":"1.2.3.4"},"resource":{"id":"txn_1","kind":"transaction"},"subject":{"id":"u1","kind":"user"}}`
	b, _ := json.Marshal(got)
	if string(b) != want {
		t.Errorf("body = %s, want %s", b, want)
	}
	b, _ = json.Marshal(DecideRequest{SubjectID: "u1", SubjectKind: "user", Action: "login", ResourceKind: "session"})
	if string(b) != `{"action":"login","resource":{"kind":"session"},"subject":{"id":"u1","kind":"user"}}` {
		t.Errorf("minimal body = %s", b)
	}
}

func TestFraudDecideCardNumberRefused(t *testing.T) {
	for _, call := range []func(c *Client) error{
		func(c *Client) error { _, err := c.Fraud.Decide(DecideRequest{}); return err },
		func(c *Client) error { _, err := c.Fraud.ShadowDecide(DecideRequest{}); return err },
	} {
		srv, attempts := errorFixtureServer(t, aiVisionErrorFixture{
			Status: http.StatusBadRequest,
			Body:   json.RawMessage(`{"success":false,"error":{"code":"CARD_NUMBER_REFUSED","message":"A value at resource.attributes.pan looks like a card number or card field. Send a token, never card data."}}`),
		})
		c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
		err := call(c)
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Code != ErrCodeCardNumberRefused || apiErr.StatusCode != http.StatusBadRequest {
			t.Fatalf("err = %v", err)
		}
		if *attempts != 1 {
			t.Errorf("attempts = %d, want 1 (a 400 is never retried)", *attempts)
		}
	}
}

func TestFraudRuleTransitionAtVersionSendsExpectedVersion(t *testing.T) {
	var gotBody map[string]interface{}
	var gotPath string
	srv := fixtureServer(t, json.RawMessage(`{"id":"frl_1","status":"published","version":4}`), nil,
		func(r *http.Request, body []byte) {
			gotPath = r.URL.Path
			_ = json.Unmarshal(body, &gotBody)
		})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	res, err := c.Fraud.Rules.TransitionAtVersion("frl_1", FraudRuleStatusPublished, 4)
	if err != nil {
		t.Fatalf("TransitionAtVersion: %v", err)
	}
	if gotPath != "/api/v1/fraud/rules/frl_1/transition" {
		t.Errorf("path = %s", gotPath)
	}
	if !reflect.DeepEqual(gotBody, map[string]interface{}{"target": "published", "expectedVersion": float64(4)}) {
		t.Errorf("body = %v", gotBody)
	}
	if v, ok := FraudRuleVersion(res); !ok || v != 4 {
		t.Errorf("FraudRuleVersion = %d, %v", v, ok)
	}
}

func TestFraudRuleUpdateAtVersionSendsExpectedVersionAndKeepsThePayload(t *testing.T) {
	var gotMethod string
	var gotBody map[string]interface{}
	srv := fixtureServer(t, json.RawMessage(`{"id":"frl_1","version":5}`), nil,
		func(r *http.Request, body []byte) {
			gotMethod = r.Method
			_ = json.Unmarshal(body, &gotBody)
		})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	payload := map[string]interface{}{"weight": 20}
	if _, err := c.Fraud.Rules.UpdateAtVersion("frl_1", payload, 4); err != nil {
		t.Fatalf("UpdateAtVersion: %v", err)
	}
	if gotMethod != "PATCH" || !reflect.DeepEqual(gotBody, map[string]interface{}{"weight": float64(20), "expectedVersion": float64(4)}) {
		t.Errorf("sent %s %v", gotMethod, gotBody)
	}
	if _, touched := payload["expectedVersion"]; touched {
		t.Error("UpdateAtVersion modified the caller's payload")
	}
}

func TestFraudRuleVersionMismatchIsReturnedOnce(t *testing.T) {
	srv, attempts := errorFixtureServer(t, aiVisionErrorFixture{
		Status: http.StatusConflict,
		Body:   json.RawMessage(`{"success":false,"error":{"code":"VERSION_MISMATCH","message":"The rule was edited since the version you reviewed"}}`),
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	_, err := c.Fraud.Rules.TransitionAtVersion("frl_1", FraudRuleStatusPublished, 2)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != ErrCodeVersionMismatch || apiErr.StatusCode != http.StatusConflict {
		t.Fatalf("err = %v", err)
	}
	if *attempts != 1 {
		t.Errorf("attempts = %d, want 1", *attempts)
	}
}

func TestFraudRuleShadowReportForVersion(t *testing.T) {
	var gotVersion string
	var hadVersion bool
	srv := fixtureServer(t, json.RawMessage(`{"ruleId":"frl_1","ruleVersion":3,"currentVersion":4}`), nil,
		func(r *http.Request, _ []byte) {
			gotVersion = r.URL.Query().Get("version")
			_, hadVersion = r.URL.Query()["version"]
		})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Fraud.Rules.ShadowReportForVersion("frl_1", "", "", 3); err != nil {
		t.Fatalf("ShadowReportForVersion: %v", err)
	}
	if gotVersion != "3" {
		t.Errorf("version = %q, want 3", gotVersion)
	}
	if _, err := c.Fraud.Rules.ShadowReport("frl_1", "", ""); err != nil {
		t.Fatalf("ShadowReport: %v", err)
	}
	if hadVersion {
		t.Error("ShadowReport sent ?version= (the current revision is the default)")
	}
}

func TestFraudRuleVersion(t *testing.T) {
	for _, tc := range []struct {
		rule map[string]interface{}
		want int64
		ok   bool
	}{
		{map[string]interface{}{"version": float64(7)}, 7, true},
		{map[string]interface{}{"version": int64(2)}, 2, true},
		{map[string]interface{}{"version": "7"}, 0, false},
		{map[string]interface{}{}, 0, false},
	} {
		if got, ok := FraudRuleVersion(tc.rule); got != tc.want || ok != tc.ok {
			t.Errorf("FraudRuleVersion(%v) = %d, %v", tc.rule, got, ok)
		}
	}
}
