// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-13, v1.12.0 pending):
// - Created. Identity.VerifyBusiness (POST /api/v1/identity/verify-business,
//   scope business:verify) sends the route body, defaults country to NG, and
//   sends x-idempotency-key when one is given.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

const verifyBusinessFixture = `{"status":"failed","reason":"tin_mismatch","matchScore":100,` +
	`"registeredName":"ACME TRADING LIMITED","registrationStatus":"ACTIVE","registeredAt":"2015-03-01",` +
	`"tinStatus":"NOT_MATCHED","provider":"prembly","verificationId":"iv_9"}`

func verifyBusinessClient(t *testing.T, gotBody *map[string]interface{}, gotKey *string) *Client {
	t.Helper()
	srv := fixtureServer(t, json.RawMessage(verifyBusinessFixture), nil, func(r *http.Request, body []byte) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/identity/verify-business" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		*gotKey = r.Header.Get("x-idempotency-key")
		if err := json.Unmarshal(body, gotBody); err != nil {
			t.Errorf("body is not JSON: %v", err)
		}
	})
	return NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
}

func TestVerifyBusinessSendsRouteBodyAndDecodes(t *testing.T) {
	var body map[string]interface{}
	var key string
	c := verifyBusinessClient(t, &body, &key)
	got, err := c.Identity.VerifyBusiness(VerifyBusinessRequest{
		SubjectID:          "prov_1",
		RegistrationType:   "RC",
		RegistrationNumber: "1234567",
		ExpectedName:       "Acme Trading Ltd",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]interface{}{
		"subjectId":          "prov_1",
		"country":            "NG",
		"registrationType":   "RC",
		"registrationNumber": "1234567",
		"expectedName":       "Acme Trading Ltd",
	}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("body = %#v, want %#v", body, want)
	}
	if key != "" {
		t.Errorf("no idempotency key expected, got %q", key)
	}
	if got.Status != "failed" || got.Reason == nil || *got.Reason != "tin_mismatch" {
		t.Errorf("status/reason not decoded: %#v", got)
	}
	if got.TINStatus == nil || *got.TINStatus != "NOT_MATCHED" || got.MatchScore == nil || *got.MatchScore != 100 {
		t.Errorf("tinStatus/matchScore not decoded: %#v", got)
	}
	if got.RegisteredAt == nil || *got.RegisteredAt != "2015-03-01" || got.Provider != "prembly" {
		t.Errorf("fields not decoded: %#v", got)
	}
}

func TestVerifyBusinessSendsTINThresholdAndIdempotencyKey(t *testing.T) {
	var body map[string]interface{}
	var key string
	c := verifyBusinessClient(t, &body, &key)
	_, err := c.Identity.VerifyBusinessWithIdempotencyKey(VerifyBusinessRequest{
		SubjectID:          "prov_1",
		Country:            "NG",
		RegistrationType:   "BN",
		RegistrationNumber: "3456789",
		TIN:                "12345678-0001",
		ExpectedName:       "Acme Trading Ltd",
		Threshold:          80,
	}, "kyb-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body["tin"] != "12345678-0001" || body["threshold"] != float64(80) || body["registrationType"] != "BN" {
		t.Errorf("body = %#v", body)
	}
	if key != "kyb-1" {
		t.Errorf("x-idempotency-key = %q, want kyb-1", key)
	}
}
