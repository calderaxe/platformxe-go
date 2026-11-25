// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-18, v1.12.0 pending):
// - Created. ResolveRequest sends the route's body with the required
//   purpose; the deprecated Resolve sends the right field names (type,
//   consent) but no purpose, so the API answers it 400.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func resolveClient(t *testing.T, gotBody *map[string]interface{}) *Client {
	t.Helper()
	srv := fixtureServer(t,
		json.RawMessage(`{"resolved":false,"identifiers":{"nin":null,"bvn":null,"tin":null,"passportNumber":null,"voterCardNumber":null,"driversLicense":null,"phoneNumbers":[]},"profile":null,"confidence":0,"provider":"none","resolvedAt":"2026-11-18T00:00:00Z","outcome":"not_found"}`),
		nil,
		func(r *http.Request, body []byte) {
			if r.Method != "POST" || r.URL.Path != "/api/v1/identity/resolve" {
				t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			}
			if err := json.Unmarshal(body, gotBody); err != nil {
				t.Errorf("body is not JSON: %v", err)
			}
		})
	return NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
}

func TestIdentityResolveRequestSendsPurpose(t *testing.T) {
	var body map[string]interface{}
	c := resolveClient(t, &body)
	got, err := c.Identity.ResolveRequest(IdentityResolveRequest{
		Type:          "NIN",
		Value:         "12345678901",
		Purpose:       "tenant onboarding",
		ResolveLinked: true,
		Consent:       IdentityResolveConsent{Granted: true, Reference: "consent_abc"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]interface{}{
		"type":          "NIN",
		"value":         "12345678901",
		"purpose":       "tenant onboarding",
		"resolveLinked": true,
		"consent":       map[string]interface{}{"granted": true, "reference": "consent_abc"},
	}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("body = %#v, want %#v", body, want)
	}
	if got.Resolved || got.Outcome != "not_found" || got.Provider != "none" {
		t.Errorf("result not decoded: %#v", got)
	}
}

func TestIdentityResolveRequestOmitsResolveLinkedWhenFalse(t *testing.T) {
	var body map[string]interface{}
	c := resolveClient(t, &body)
	if _, err := c.Identity.ResolveRequest(IdentityResolveRequest{
		Type: "BVN", Value: "22012345678", Purpose: "kyc",
		Consent: IdentityResolveConsent{Granted: true, Reference: "c"},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := body["resolveLinked"]; ok {
		t.Errorf("resolveLinked sent when false: %#v", body)
	}
}

func TestIdentityResolveDeprecatedSendsRouteFieldNames(t *testing.T) {
	var body map[string]interface{}
	c := resolveClient(t, &body)
	if _, err := c.Identity.Resolve("BVN", "22012345678", false, "consent_abc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]interface{}{
		"type":    "BVN",
		"value":   "22012345678",
		"consent": map[string]interface{}{"granted": true, "reference": "consent_abc"},
	}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("body = %#v, want %#v", body, want)
	}
}
