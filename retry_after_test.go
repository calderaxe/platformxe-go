// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-16, pending release):
// - Created (W2 Task 5b, Batch A final code review F2, NO-DRIFT): the client
//   never retries sooner than a Retry-After header says, returns at once when
//   the wait is longer than it will sleep, and never retries the email send
//   routes' EMAIL_RATE_LIMITED / EMAIL_NOT_ACCEPTED /
//   EMAIL_PROVIDER_UNAVAILABLE; SendEmail reads "deduplicated".
// =============================================================================

package platformxe

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var sendEmailInput = map[string]interface{}{
	"from":    "App <hello@app.ng>",
	"to":      []string{"jane@acme.ng"},
	"subject": "Hi",
	"html":    "<p>Hi</p>",
}

func TestEmailNotAcceptedCodesAreNeverRetried(t *testing.T) {
	cases := []struct {
		status     int
		code       string
		retryAfter string
		want       int
	}{
		{429, ErrCodeEmailRateLimited, "2710", 2710},
		{503, ErrCodeEmailNotAccepted, "60", 60},
		{503, ErrCodeEmailProviderUnavailable, "300", 300},
		// A proxy that drops Retry-After changes nothing: still one attempt.
		{503, ErrCodeEmailNotAccepted, "", 0},
		{429, ErrCodeEmailRateLimited, "", 0},
	}
	for _, tc := range cases {
		headers := map[string]string{}
		if tc.retryAfter != "" {
			headers["Retry-After"] = tc.retryAfter
		}
		srv, attempts := errorFixtureServer(t, aiVisionErrorFixture{
			Status:  tc.status,
			Headers: headers,
			Body:    json.RawMessage(`{"success":false,"error":{"code":"` + tc.code + `","message":"not accepted"}}`),
		})
		c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 2})

		started := time.Now()
		_, err := c.Messaging.SendEmail(sendEmailInput)
		apiErr := asAPIError(t, err)
		if apiErr.Code != tc.code || apiErr.StatusCode != tc.status || apiErr.RetryAfter != tc.want {
			t.Errorf("%s: unexpected APIError: %#v", tc.code, apiErr)
		}
		if *attempts != 1 {
			t.Errorf("%s (Retry-After %q) must not be retried, got %d attempts", tc.code, tc.retryAfter, *attempts)
		}
		if time.Since(started) > time.Second {
			t.Errorf("%s: the client slept before returning", tc.code)
		}
	}
}

func TestRetryIsNeverSoonerThanRetryAfter(t *testing.T) {
	var times []time.Time
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		times = append(times, time.Now())
		w.Header().Set("Content-Type", "application/json")
		if len(times) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"success":false,"error":{"code":"RATE_LIMITED","message":"slow down"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"success":true,"data":{"ok":true}}`))
	}))
	t.Cleanup(srv.Close)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 1})

	if _, err := c.Get("/api/x", nil); err != nil {
		t.Fatalf("expected success after waiting out Retry-After, got %v", err)
	}
	if len(times) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(times))
	}
	// The backoff alone is 200 ms: the retry must wait the whole second.
	if gap := times[1].Sub(times[0]); gap < 950*time.Millisecond {
		t.Errorf("retried %v after a Retry-After of 1 s", gap)
	}
}

func TestRetryAfterLongerThanTheClientWaitsIsReturnedAtOnce(t *testing.T) {
	for _, status := range []int{http.StatusTooManyRequests, http.StatusServiceUnavailable} {
		srv, attempts := errorFixtureServer(t, aiVisionErrorFixture{
			Status:  status,
			Headers: map[string]string{"Retry-After": "3600"},
			Body:    json.RawMessage(`{"success":false,"error":{"code":"RATE_LIMITED","message":"slow down"}}`),
		})
		c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 2})

		started := time.Now()
		_, err := c.Get("/api/x", nil)
		apiErr := asAPIError(t, err)
		if apiErr.RetryAfter != 3600 || apiErr.StatusCode != status {
			t.Errorf("unexpected APIError: %#v", apiErr)
		}
		// Retrying inside the hour can only be refused again.
		if *attempts != 1 {
			t.Errorf("status %d with Retry-After 3600: expected 1 attempt, got %d", status, *attempts)
		}
		if time.Since(started) > time.Second {
			t.Errorf("status %d: the client slept before returning", status)
		}
	}
}

func TestRetryableStatusWithoutRetryAfterKeepsTheBackoff(t *testing.T) {
	srv, attempts := cannedServer(t,
		cannedResponse{503, `{"success":false,"error":{"code":"UNAVAILABLE","message":"x"}}`},
		cannedResponse{200, `{"success":true,"data":{"ok":true}}`},
	)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 1})

	started := time.Now()
	if _, err := c.Get("/api/x", nil); err != nil {
		t.Fatalf("expected success after one retry, got %v", err)
	}
	if *attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", *attempts)
	}
	if time.Since(started) > 900*time.Millisecond {
		t.Errorf("a retry with no Retry-After waited longer than the 200 ms backoff")
	}
}

func TestMessagingSendEmailReadsDeduplicated(t *testing.T) {
	srv := newTestServer(t, func(_ *http.Request) {}, map[string]interface{}{"provider": "none", "queued": true, "deduplicated": true})
	t.Cleanup(srv.Close)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})

	got, err := c.Messaging.SendEmail(sendEmailInput)
	if err != nil {
		t.Fatalf("SendEmail: %v", err)
	}
	if !got.Queued || !got.Deduplicated || got.Provider != "none" {
		t.Errorf("unexpected result: %#v", got)
	}
}

// W2-5b fix round 1 (N2): a Retry-After too large for time.Duration once
// multiplied by a second must not wrap into a short (or negative) wait and be
// retried at the backoff; it is longer than the client waits, so it returns.
func TestAbsurdRetryAfterIsReturnedAtOnceNotWrapped(t *testing.T) {
	for _, retryAfter := range []string{"9300000000", "18446744073", "9223372036"} {
		srv, attempts := errorFixtureServer(t, aiVisionErrorFixture{
			Status:  http.StatusServiceUnavailable,
			Headers: map[string]string{"Retry-After": retryAfter},
			Body:    json.RawMessage(`{"success":false,"error":{"code":"SERVICE_UNAVAILABLE","message":"later"}}`),
		})
		c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 2})

		_, err := c.Get("/api/x", nil)
		asAPIError(t, err)
		if *attempts != 1 {
			t.Errorf("Retry-After %s: expected 1 attempt, got %d", retryAfter, *attempts)
		}
	}
}
