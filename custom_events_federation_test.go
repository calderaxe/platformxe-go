// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// Federation / marketplace request paths and query parameters (custom
// events gap review, 2026-11-17): member remove / pause / resume, the relay
// log, includeRemoved, marketplace mine and the opt-in example; ids are
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

type recordedCall struct {
	Method string
	Path   string
	Query  string
	Body   map[string]interface{}
}

func recordingClient(t *testing.T) (*Client, *[]recordedCall) {
	t.Helper()
	calls := []recordedCall{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]interface{}
		_ = json.Unmarshal(raw, &body)
		calls = append(calls, recordedCall{Method: r.Method, Path: r.URL.EscapedPath(), Query: r.URL.RawQuery, Body: body})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{}}`))
	}))
	t.Cleanup(srv.Close)
	return NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, DisableRetries: true}), &calls
}

func TestEventFederationMemberLifecyclePaths(t *testing.T) {
	c, calls := recordingClient(t)
	fed := c.Events.Custom.Federation
	if _, err := fed.RemoveMember("g 1", "org_b"); err != nil {
		t.Fatal(err)
	}
	if _, err := fed.PauseMember("g1", "org_b"); err != nil {
		t.Fatal(err)
	}
	if _, err := fed.ResumeMember("g1", "org_b"); err != nil {
		t.Fatal(err)
	}
	failed := EventFederationRelayStatusFailed
	limit := 10
	if _, err := fed.ListRelays("p1", &ListEventFederationRelaysParams{Status: &failed, Limit: &limit}); err != nil {
		t.Fatal(err)
	}
	if _, err := fed.PauseExternalPeer("m 1"); err != nil {
		t.Fatal(err)
	}
	yes := true
	if _, err := fed.ListGroups(&ListEventFederationGroupsParams{IncludeRemoved: &yes}); err != nil {
		t.Fatal(err)
	}
	want := []recordedCall{
		{Method: "DELETE", Path: "/api/v1/events/custom/federation/groups/g%201/members/org_b"},
		{Method: "POST", Path: "/api/v1/events/custom/federation/groups/g1/members/org_b/pause"},
		{Method: "POST", Path: "/api/v1/events/custom/federation/groups/g1/members/org_b/resume"},
		{Method: "GET", Path: "/api/v1/events/custom/federation/pushes/p1/relays", Query: "limit=10&status=failed"},
		{Method: "POST", Path: "/api/v1/events/custom/federation/external-peers/m%201/pause"},
		{Method: "GET", Path: "/api/v1/events/custom/federation/groups", Query: "includeRemoved=true"},
	}
	if len(*calls) != len(want) {
		t.Fatalf("got %d calls, want %d", len(*calls), len(want))
	}
	for i, w := range want {
		got := (*calls)[i]
		if got.Method != w.Method || got.Path != w.Path || got.Query != w.Query {
			t.Errorf("call %d: got %s %s?%s, want %s %s?%s", i, got.Method, got.Path, got.Query, w.Method, w.Path, w.Query)
		}
	}
}

func TestMarketplaceMineAndExample(t *testing.T) {
	c, calls := recordingClient(t)
	mkt := c.Events.Custom.Marketplace
	mine := true
	status := "unpublished"
	if _, err := mkt.List(ListMarketplaceParams{Mine: &mine, Status: &status}); err != nil {
		t.Fatal(err)
	}
	if _, err := mkt.Publish(PublishMarketplaceInput{RegistrationID: "reg_1", Title: "Title", PayloadExample: map[string]interface{}{"id": "x"}}); err != nil {
		t.Fatal(err)
	}
	if got := (*calls)[0]; got.Path != "/api/v1/events/custom/marketplace" || got.Query != "mine=true&status=unpublished" {
		t.Errorf("list: got %s?%s", got.Path, got.Query)
	}
	if got := (*calls)[1]; got.Path != "/api/v1/events/custom/marketplace/publish" || got.Body["payloadExample"] == nil {
		t.Errorf("publish: got %s %v", got.Path, got.Body)
	}
}
