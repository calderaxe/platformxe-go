// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-07, v1.9.0 pending):
// - Created. Fraud.Activate posts to /api/v1/fraud/activate.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestFraudActivate(t *testing.T) {
	var gotKey string
	srv := fixtureServer(t,
		json.RawMessage(`{"detectionPackEnabled":true,"alreadyActive":false,"billing":"FIRST_PARTY_METERED","invoiceId":null,"firstParty":true}`),
		nil,
		func(r *http.Request, _ []byte) {
			if r.Method != "POST" || r.URL.Path != "/api/v1/fraud/activate" {
				t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			}
			gotKey = r.Header.Get("x-idempotency-key")
		})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	res, err := c.Fraud.Activate("errandone-activate-1")
	if err != nil {
		t.Fatalf("Activate: %v", err)
	}
	if gotKey != "errandone-activate-1" {
		t.Errorf("idempotency key = %q", gotKey)
	}
	data, _ := res["data"].(map[string]interface{})
	if data == nil {
		data = res
	}
	if data["billing"] != "FIRST_PARTY_METERED" {
		t.Errorf("billing = %v", data["billing"])
	}
}
