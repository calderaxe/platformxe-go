// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-01, v1.7.2 pending):
// - Created. Pins the response shapes the API really returns for the calls
//   the Terraform provider depends on, each found decoding wrongly:
//   identity providers { name, healthy }, event subscriptions (create
//   { subscription, secret }, list { items, total }), permission overrides
//   (a bare array) and federation members (id / groupId). Payloads follow
//   @caldera/platformxe-types and the route handlers.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"net/http"
	"testing"
)

func newShapeClient(t *testing.T, method, path string, data string) *Client {
	t.Helper()
	srv := fixtureServer(t, json.RawMessage(data), nil, func(r *http.Request, _ []byte) {
		if r.Method != method || r.URL.Path != path {
			t.Errorf("unexpected %s %s (want %s %s)", r.Method, r.URL.Path, method, path)
		}
	})
	return NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
}

func TestIdentityProvidersDecodesHealthy(t *testing.T) {
	c := newShapeClient(t, "GET", "/api/v1/identity/providers",
		`{"providers":[{"name":"smile-id","healthy":true},{"name":"nimc","healthy":false}]}`)
	got, err := c.Identity.Providers()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Providers) != 2 || got.Providers[0].Name != "smile-id" || !got.Providers[0].Healthy || got.Providers[1].Healthy {
		t.Errorf("providers not decoded: %#v", got.Providers)
	}
}

func TestCreateSubscriptionDecodesNestedSubscriptionAndSecret(t *testing.T) {
	c := newShapeClient(t, "POST", "/api/v1/event-subscriptions",
		`{"subscription":{"id":"sub_1","organizationId":"org_1","name":"orders","eventTypes":["order.created"],"webhookUrl":"https://example.com/h","status":"ACTIVE","consecutiveFailures":0,"lastTriggeredAt":null,"createdAt":"2026-11-01T00:00:00Z"},"secret":"whsec_abc"}`)
	got, err := c.Events.CreateSubscription(map[string]interface{}{"name": "orders"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "sub_1" || got.Name != "orders" || got.Status != "ACTIVE" {
		t.Errorf("subscription not decoded: %#v", got)
	}
	if got.Secret != "whsec_abc" {
		t.Errorf("secret = %q, want whsec_abc", got.Secret)
	}
}

func TestListSubscriptionsReadsItems(t *testing.T) {
	c := newShapeClient(t, "GET", "/api/v1/event-subscriptions",
		`{"items":[{"id":"sub_1","name":"orders","status":"PAUSED"}],"total":1}`)
	got, err := c.Events.ListSubscriptions()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Total != 1 || len(got.Items) != 1 || got.Items[0].ID != "sub_1" {
		t.Errorf("items not decoded: %#v", got)
	}
	if len(got.Subscriptions) != 1 || got.Subscriptions[0].ID != "sub_1" {
		t.Errorf("deprecated Subscriptions not filled: %#v", got.Subscriptions)
	}
}

func TestListOverridesDecodesBareArray(t *testing.T) {
	c := newShapeClient(t, "GET", "/api/v1/permissions/overrides/adm_1",
		`[{"id":"ovr_1","adminId":"adm_1","path":"chat/session","action":"READ","effect":"DENY","reason":null,"expiresAt":null,"createdAt":"2026-11-01T00:00:00Z"}]`)
	got, err := c.Permissions.ListOverrides("adm_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Overrides) != 1 || got.Overrides[0].ID != "ovr_1" || got.Overrides[0].Effect != "DENY" || got.Overrides[0].ExpiresAt != nil {
		t.Errorf("overrides not decoded: %#v", got.Overrides)
	}
}

func TestListOverridesEmptyArray(t *testing.T) {
	c := newShapeClient(t, "GET", "/api/v1/permissions/overrides/adm_1", `[]`)
	got, err := c.Permissions.ListOverrides("adm_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Overrides == nil || len(got.Overrides) != 0 {
		t.Errorf("want an empty, non-nil list: %#v", got.Overrides)
	}
}

func TestFederationMembersCarryID(t *testing.T) {
	c := newShapeClient(t, "GET", "/api/v1/permissions/federation/groups/grp_1",
		`{"id":"grp_1","name":"Caldera","ownerOrgId":"org_1","createdAt":"2026-11-01T00:00:00Z","members":[{"id":"mem_1","groupId":"grp_1","organizationId":"org_2","prefix":"LT","role":"member","joinedAt":"2026-11-01T00:00:00Z"}]}`)
	got, err := c.Permissions.GetFederationGroup("grp_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Members) != 1 || got.Members[0].ID != "mem_1" || got.Members[0].GroupID != "grp_1" || got.Members[0].OrganizationID != "org_2" {
		t.Errorf("member not decoded: %#v", got.Members)
	}
}
