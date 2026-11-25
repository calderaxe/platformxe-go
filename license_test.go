// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-22, v1.15.0 pending, Enterprise License): Initial. Pins
// what License.Get sends (GET /api/v1/license, no query, no body) and how the
// answer decodes (a covering license, a nil one with the bypass), and that an
// error answer is an *APIError with its code.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLicenseGetSendsGetAndDecodes(t *testing.T) {
	srv := newTestServer(t, func(r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/license" || r.URL.RawQuery != "" {
			t.Errorf("unexpected URL: %s", r.URL.String())
		}
		if body, _ := io.ReadAll(r.Body); len(body) != 0 {
			t.Errorf("no body expected, got %q", body)
		}
	}, map[string]interface{}{
		"edition": map[string]interface{}{"plan": "ENTERPRISE", "source": "license"},
		"bypass":  nil,
		"license": map[string]interface{}{
			"licenseId": "lic_1", "kind": "standard",
			"issuedAt": "2026-11-01T00:00:00.000Z", "expiresAt": "2027-11-01T00:00:00.000Z",
			"graceEndsAt": "2027-12-01T00:00:00.000Z", "state": "valid", "live": true,
			"daysLeft": 344, "renewalDue": false, "holderName": "Acme Holdings", "isHolder": true,
			"edition": "ENTERPRISE", "plugins": []interface{}{"banking", "banking.payments"},
		},
	})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	got, err := c.License.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Edition.Plan != "ENTERPRISE" || got.Edition.Source != LicenseEditionSourceLicense || got.Bypass != nil {
		t.Fatalf("unexpected edition / bypass: %#v", got)
	}
	l := got.License
	if l == nil || l.LicenseID != "lic_1" || l.State != LicenseStateValid || !l.Live || l.DaysLeft != 344 || !l.IsHolder {
		t.Fatalf("unexpected license: %#v", l)
	}
	if l.Edition == nil || *l.Edition != "ENTERPRISE" || len(l.Plugins) != 2 || l.Plugins[1] != "banking.payments" {
		t.Fatalf("unexpected grant: %#v", l)
	}
}

func TestLicenseGetFirstPartyHasNoLicense(t *testing.T) {
	srv := newTestServer(t, nil, map[string]interface{}{
		"edition": map[string]interface{}{"plan": "ENTERPRISE", "source": "first_party"},
		"bypass":  "first_party",
		"license": nil,
	})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	got, err := c.License.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.License != nil || got.Bypass == nil || *got.Bypass != LicenseBypassFirstParty || got.Edition.Source != LicenseEditionSourceFirstParty {
		t.Fatalf("unexpected view: %#v", got)
	}
}

func TestLicenseGetErrorIsAnAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   map[string]interface{}{"code": "FORBIDDEN", "message": "API key does not have scope: license:read"},
		})
	}))
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	_, err := c.License.Get()
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != ErrCodeForbidden {
		t.Fatalf("expected a FORBIDDEN *APIError, got %v", err)
	}
}

func TestLicensingErrorCodes(t *testing.T) {
	if ErrCodeLicenseRequired != "LICENSE_REQUIRED" || ErrCodeContactLicensing != "CONTACT_LICENSING" {
		t.Fatal("licensing error codes changed")
	}
}
