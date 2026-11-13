// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-07, v1.8.0 pending):
// - Created. Verify sends the route's { type, value, matchAgainst } body
//   (it used to send identifierType / firstName / lastName, a 400), and
//   VerifyRequest carries supporting.nin for a PASSPORT check.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func verifyClient(t *testing.T, gotBody *map[string]interface{}) *Client {
	t.Helper()
	srv := fixtureServer(t,
		json.RawMessage(`{"verified":true,"matchScore":1,"profile":{"firstName":"ADA"},"provider":"prembly"}`),
		nil,
		func(r *http.Request, body []byte) {
			if r.Method != "POST" || r.URL.Path != "/api/v1/identity/verify" {
				t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			}
			if err := json.Unmarshal(body, gotBody); err != nil {
				t.Errorf("body is not JSON: %v", err)
			}
		})
	return NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
}

func TestIdentityVerifySendsRouteShape(t *testing.T) {
	var body map[string]interface{}
	c := verifyClient(t, &body)
	got, err := c.Identity.Verify("NIN", "12345678901", "Ada", "Obi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]interface{}{
		"type":         "NIN",
		"value":        "12345678901",
		"matchAgainst": map[string]interface{}{"firstName": "Ada", "lastName": "Obi"},
	}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("body = %#v, want %#v", body, want)
	}
	if !got.Verified || got.Provider != "prembly" || got.Profile["firstName"] != "ADA" {
		t.Errorf("result not decoded: %#v", got)
	}
}

func TestIdentityVerifyRequestSendsSupportingNIN(t *testing.T) {
	var body map[string]interface{}
	c := verifyClient(t, &body)
	_, err := c.Identity.VerifyRequest(IdentityVerifyRequest{
		Type:  "PASSPORT",
		Value: "A12345678",
		MatchAgainst: IdentityMatchAgainst{
			FirstName: "Ada", LastName: "Obi", DateOfBirth: "1990-03-15",
		},
		Supporting: &IdentityVerifySupporting{NIN: "12345678901"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]interface{}{
		"type":  "PASSPORT",
		"value": "A12345678",
		"matchAgainst": map[string]interface{}{
			"firstName": "Ada", "lastName": "Obi", "dateOfBirth": "1990-03-15",
		},
		"supporting": map[string]interface{}{"nin": "12345678901"},
	}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("body = %#v, want %#v", body, want)
	}
}

// 2026-11-07 v1.10.0: Lookup sends the purpose and consent reference the
// route requires (types 3.11.0).
func TestIdentityLookupSendsConsent(t *testing.T) {
	srv := fixtureServer(t, json.RawMessage(`{"found":false}`), nil, func(r *http.Request, _ []byte) {
		q := r.URL.Query()
		if r.URL.Path != "/api/v1/identity/lookup" || q.Get("type") != "NIN" || q.Get("value") != "12345678901" ||
			q.Get("purpose") != "onboarding" || q.Get("consentReference") != "c_1" {
			t.Errorf("unexpected request %s?%s", r.URL.Path, r.URL.RawQuery)
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Identity.Lookup("NIN", "12345678901", "onboarding", "c_1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
