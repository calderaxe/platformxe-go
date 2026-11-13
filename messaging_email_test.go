// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-11, pending release):
// - Created (Batch A closing fix, fix5 code review M4, NO-DRIFT): SendEmail
//   sends "from" and a "replyTo" list as given and reads { messageId,
//   provider, queued }.
// =============================================================================

package platformxe

import (
	"net/http"
	"testing"
)

func TestMessagingSendEmailFromAndReplyToList(t *testing.T) {
	srv := newTestServer(t, func(r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/messaging/email/send" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		body := decodeJSONBody(t, r)
		if body["from"] != "App <hello@app.ng>" {
			t.Errorf("from not sent: %#v", body["from"])
		}
		reply, ok := body["replyTo"].([]interface{})
		if !ok || len(reply) != 2 || reply[0] != "desk@app.ng" || reply[1] != "Ops <ops@app.ng>" {
			t.Errorf("replyTo list not sent as a list: %#v", body["replyTo"])
		}
	}, map[string]interface{}{"messageId": "m1", "provider": "primary", "queued": false})
	t.Cleanup(srv.Close)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})

	got, err := c.Messaging.SendEmail(map[string]interface{}{
		"from":    "App <hello@app.ng>",
		"to":      []string{"jane@acme.ng"},
		"subject": "Hi",
		"html":    "<p>Hi</p>",
		"replyTo": []string{"desk@app.ng", "Ops <ops@app.ng>"},
	})
	if err != nil {
		t.Fatalf("SendEmail: %v", err)
	}
	if got.MessageID != "m1" || got.Provider != "primary" || got.Queued {
		t.Errorf("unexpected result: %#v", got)
	}
}
