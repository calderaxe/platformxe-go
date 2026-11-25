// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-18, v1.12.0 pending):
// - Created. Lookup decodes the route's real answer; Dlq.Replay returns 409
//   REPLAY_IN_PROGRESS and a 500 at once (never re-sent).
// =============================================================================

package platformxe

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestIdentityLookupDecodesTheRealAnswer(t *testing.T) {
	srv := fixtureServer(t, json.RawMessage(`{
		"found": true,
		"nin": "12345678901",
		"bvn": null,
		"phoneNumbers": ["+2348000000000"],
		"outcome": "found",
		"profile": {"firstName": "ADA", "lastName": "OBI", "photo": null},
		"raw": {"provider": "prembly"}
	}`), nil, func(r *http.Request, _ []byte) {
		if r.Method != "GET" || r.URL.Path != "/api/v1/identity/lookup" {
			t.Errorf("sent %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("purpose") != "kyc" || r.URL.Query().Get("consentReference") != "c1" {
			t.Errorf("query = %v", r.URL.RawQuery)
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	res, err := c.Identity.Lookup("NIN", "12345678901", "kyc", "c1")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if !res.Found || res.Outcome != IdentityOutcomeFound || res.Raw.Provider != "prembly" {
		t.Errorf("result = %+v", res)
	}
	if res.NIN == nil || *res.NIN != "12345678901" || res.BVN != nil {
		t.Errorf("identifiers = %v %v", res.NIN, res.BVN)
	}
	if len(res.PhoneNumbers) != 1 || res.Profile["firstName"] != "ADA" {
		t.Errorf("phones = %v profile = %v", res.PhoneNumbers, res.Profile)
	}
}

func TestIdentityLookupNotFound(t *testing.T) {
	srv := fixtureServer(t, json.RawMessage(`{
		"found": false, "notFound": true, "notFoundBy": "prembly",
		"outcome": "not_found", "profile": null, "raw": {"provider": "prembly"}
	}`), nil, nil)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	res, err := c.Identity.Lookup("NIN", "12345678901", "kyc", "c1")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if res.Found || !res.NotFound || res.NotFoundBy != "prembly" || res.Outcome != IdentityOutcomeNotFound || res.Profile != nil {
		t.Errorf("result = %+v", res)
	}
}

func TestIdentityDlqReplayErrorsAreReturnedOnce(t *testing.T) {
	cases := []struct {
		status int
		code   string
	}{
		{http.StatusConflict, ErrCodeReplayInProgress},
		{http.StatusInternalServerError, ErrCodeInternalError},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			srv, attempts := errorFixtureServer(t, aiVisionErrorFixture{
				Status: tc.status,
				Body:   json.RawMessage(`{"success":false,"error":{"code":"` + tc.code + `","message":"x"}}`),
			})
			c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
			_, err := c.Identity.Dlq.Replay("idlq_1")
			var apiErr *APIError
			if !errors.As(err, &apiErr) || apiErr.Code != tc.code || apiErr.StatusCode != tc.status {
				t.Fatalf("err = %v", err)
			}
			if *attempts != 1 {
				t.Errorf("attempts = %d, want 1", *attempts)
			}
		})
	}
}
