// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-24, pending release; provided-SSO code review B4):
// - Created: Domains.Add surfaces 409 DOMAIN_ALREADY_REGISTERED and 503
//   DOMAIN_LIMIT_REACHED as their closed codes, and neither is retried (the
//   provider's plan limit is not transient; a re-POST would register again).
// =============================================================================

package platformxe

import (
	"encoding/json"
	"testing"
)

func TestDomainRegistrationRefusalsAreFinal(t *testing.T) {
	if ErrCodeDomainAlreadyRegistered != "DOMAIN_ALREADY_REGISTERED" || ErrCodeDomainLimitReached != "DOMAIN_LIMIT_REACHED" {
		t.Fatalf("codes: %q %q", ErrCodeDomainAlreadyRegistered, ErrCodeDomainLimitReached)
	}
	cases := []struct {
		status  int
		code    string
		message string
	}{
		{409, ErrCodeDomainAlreadyRegistered, "This domain is already registered."},
		{503, ErrCodeDomainLimitReached, "Domains can't be added at the moment. Contact support."},
	}
	for _, tc := range cases {
		srv, attempts := errorFixtureServer(t, aiVisionErrorFixture{
			Status: tc.status,
			Body:   json.RawMessage(`{"success":false,"error":{"code":"` + tc.code + `","message":"` + tc.message + `"}}`),
		})
		c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 2})

		_, err := c.Domains.Add("mail.example.com")
		apiErr := asAPIError(t, err)
		if apiErr.Code != tc.code || apiErr.StatusCode != tc.status || apiErr.Message != tc.message {
			t.Errorf("%s: unexpected APIError: %#v", tc.code, apiErr)
		}
		if *attempts != 1 {
			t.Errorf("%s must not be retried, got %d attempts", tc.code, *attempts)
		}
	}
}
