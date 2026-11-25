// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
//
// CHANGELOG (2026-11-19, v1.13.0 pending, X12):
// - Created. Replay sends fromTimestamp / toTimestamp (and eventTypes from
//   ReplayTypes) and reads { replayed }; a subscription read carries
//   RestrictedEventTypes; a restricted entry is a 400 RESTRICTED_EVENT_TYPE
//   *APIError, sent once; ids are path-escaped; the publish 409s FEED_MISSING
//   and FEED_NOT_REALTIME and the index 400 RESERVED_SEARCH_APP are returned
//   with their codes.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"
)

func TestEventsReplaySendsTheRouteFields(t *testing.T) {
	var gotPath string
	var gotBody map[string]interface{}
	srv := fixtureServer(t, json.RawMessage(`{"replayed":3}`), nil, func(r *http.Request, body []byte) {
		gotPath = r.URL.Path
		_ = json.Unmarshal(body, &gotBody)
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	res, err := c.Events.Replay("esub_1", "2026-11-01T00:00:00Z", "2026-11-02T00:00:00Z")
	if err != nil {
		t.Fatalf("Replay: %v", err)
	}
	if gotPath != "/api/v1/event-subscriptions/esub_1/replay" {
		t.Errorf("path = %s", gotPath)
	}
	want := map[string]interface{}{"fromTimestamp": "2026-11-01T00:00:00Z", "toTimestamp": "2026-11-02T00:00:00Z"}
	if !reflect.DeepEqual(gotBody, want) {
		t.Errorf("body = %v, want %v", gotBody, want)
	}
	if res.Replayed != 3 {
		t.Errorf("Replayed = %d, want 3", res.Replayed)
	}
}

func TestEventsReplayTypesSendsEventTypesAndAnOpenWindow(t *testing.T) {
	var gotBody map[string]interface{}
	srv := fixtureServer(t, json.RawMessage(`{"replayed":0}`), nil, func(_ *http.Request, body []byte) {
		_ = json.Unmarshal(body, &gotBody)
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Events.ReplayTypes("esub_1", "2026-11-01T00:00:00Z", "", []string{"messaging.email.sent"}); err != nil {
		t.Fatalf("ReplayTypes: %v", err)
	}
	want := map[string]interface{}{"fromTimestamp": "2026-11-01T00:00:00Z", "eventTypes": []interface{}{"messaging.email.sent"}}
	if !reflect.DeepEqual(gotBody, want) {
		t.Errorf("body = %v, want %v", gotBody, want)
	}
}

func TestEventsGetSubscriptionReadsRestrictedEventTypesAndEscapesTheID(t *testing.T) {
	var gotPath string
	srv := fixtureServer(t, json.RawMessage(`{"id":"a/b","eventTypes":["FRAUD_*","A"],"restrictedEventTypes":["FRAUD_*"]}`), nil,
		func(r *http.Request, _ []byte) { gotPath = r.URL.EscapedPath() })
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	sub, err := c.Events.GetSubscription("a/b")
	if err != nil {
		t.Fatalf("GetSubscription: %v", err)
	}
	if gotPath != "/api/v1/event-subscriptions/a%2Fb" {
		t.Errorf("path = %s", gotPath)
	}
	if !reflect.DeepEqual(sub.RestrictedEventTypes, []string{"FRAUD_*"}) {
		t.Errorf("RestrictedEventTypes = %v", sub.RestrictedEventTypes)
	}
}

func TestEventsRestrictedEventTypeIsReturnedOnce(t *testing.T) {
	srv, attempts := errorFixtureServer(t, aiVisionErrorFixture{
		Status: http.StatusBadRequest,
		Body:   json.RawMessage(`{"success":false,"error":{"code":"RESTRICTED_EVENT_TYPE","message":"eventTypes names a restricted event type: FRAUD_*"}}`),
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	_, err := c.Events.UpdateSubscription("esub_1", map[string]interface{}{"eventTypes": []string{"FRAUD_*"}})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != ErrCodeRestrictedEventType || apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("err = %v", err)
	}
	if *attempts != 1 {
		t.Errorf("attempts = %d, want 1", *attempts)
	}
}

func TestFraudRulePublishFeedRefusalsCarryTheirCodes(t *testing.T) {
	for _, code := range []string{ErrCodeFeedMissing, ErrCodeFeedNotRealtime} {
		srv, attempts := errorFixtureServer(t, aiVisionErrorFixture{
			Status: http.StatusConflict,
			Body:   json.RawMessage(`{"success":false,"error":{"code":"` + code + `","message":"refused"}}`),
		})
		c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
		_, err := c.Fraud.Rules.Transition("frl_1", FraudRuleStatusPublished)
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Code != code || apiErr.StatusCode != http.StatusConflict {
			t.Fatalf("%s: err = %v", code, err)
		}
		if *attempts != 1 {
			t.Errorf("%s: attempts = %d, want 1", code, *attempts)
		}
	}
}

func TestSearchReservedAppCode(t *testing.T) {
	if ErrCodeReservedSearchApp != "RESERVED_SEARCH_APP" {
		t.Errorf("ErrCodeReservedSearchApp = %q", ErrCodeReservedSearchApp)
	}
}
