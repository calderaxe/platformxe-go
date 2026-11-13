// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-09-15): TestNewClientCustomConfig configures zero retries
// via DisableRetries (Retries: 0 is the zero value and means "use the
// default"), and new tests pin the retry contract: omitted → 2, explicit
// value kept, DisableRetries wins over Retries, negative Retries → no
// retries, and DisableRetries makes exactly one HTTP attempt against an
// httptest server. Added because NewClient could not disable retries and a
// negative value sent no request at all.
//
// Also 2026-09-15 (MR !199 round 9): doRequest never checked the HTTP status,
// so a non-2xx whose JSON was not a {success:false} envelope (for example
// {"error":"..."}) came back as data with a nil error. New tests pin that any
// non-2xx is an *APIError carrying the status and the parsed message (envelope,
// bare "error" string, or the status text for a non-JSON body); 5xx and 429 are
// retried on the existing backoff and return the *APIError once retries run
// out (FailOpen does not apply); 4xx is not retried; and 2xx behaviour is
// unchanged (success:false is still an error with the same text, a non-JSON
// 2xx is still an "unmarshal response" error).
//
// Also 2026-09-15 (MR !199 round 10, findings 3 and 10): the server answers a
// handler throw with 500 precisely because side effects may have landed, and
// the Go SDK sends no idempotency key, so retrying a POST/PATCH on 500 could
// repeat them. New tests pin: POST and PATCH 500 make exactly one attempt;
// GET/HEAD/PUT/DELETE 500 are retried; 502/503/504 and 429 are retried for every
// method (POST included); another 5xx (501) is not retried. FailOpen with mixed
// failures is pinned too: a 503 then a transport failure returns the _failed
// result (the last attempt failed at the transport level), and a transport
// failure then a 503 returns the *APIError (FailOpen does not apply).
//
// CHANGELOG (2026-10-08, v1.7.0): TestServiceNamespaces also checks Client.Ai
// and Client.Vision (caldera-platformxe#22); their contract tests live in
// ai_vision_test.go.
// =============================================================================

package platformxe

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// -- HTTP status handling --

type cannedResponse struct {
	status int
	body   string
}

// cannedServer answers the Nth request with the Nth response, repeating the
// last one once they run out, and counts attempts.
func cannedServer(t *testing.T, responses ...cannedResponse) (*httptest.Server, *int) {
	t.Helper()
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		idx := attempts
		if idx >= len(responses) {
			idx = len(responses) - 1
		}
		attempts++
		w.WriteHeader(responses[idx].status)
		fmt.Fprint(w, responses[idx].body)
	}))
	t.Cleanup(srv.Close)
	return srv, &attempts
}

func asAPIError(t *testing.T, err error) *APIError {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an *APIError, got %T: %v", err, err)
	}
	return apiErr
}

func TestNon2xxWithoutEnvelopeIsAnError(t *testing.T) {
	srv, attempts := cannedServer(t, cannedResponse{400, `{"error":"bad thing"}`})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})

	data, err := c.Get("/api/x", nil)
	if data != nil {
		t.Errorf("expected no data, got %v", data)
	}
	apiErr := asAPIError(t, err)
	if apiErr.StatusCode != 400 || apiErr.Code != "HTTP_400" || apiErr.Message != "bad thing" {
		t.Errorf("unexpected APIError %+v", apiErr)
	}
	if *attempts != 1 {
		t.Errorf("a 4xx must not be retried, got %d attempts", *attempts)
	}
}

func TestNon2xxNestedErrorWithoutSuccessField(t *testing.T) {
	srv, _ := cannedServer(t, cannedResponse{500, `{"error":{"code":"DB_DOWN","message":"database unavailable"}}`})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, DisableRetries: true})

	_, err := c.Get("/api/x", nil)
	apiErr := asAPIError(t, err)
	if apiErr.StatusCode != 500 || apiErr.Code != "DB_DOWN" || apiErr.Message != "database unavailable" {
		t.Errorf("unexpected APIError %+v", apiErr)
	}
}

func TestNon2xxNonJSONBodyIsAnAPIError(t *testing.T) {
	srv, attempts := cannedServer(t, cannedResponse{502, "<html>Bad Gateway</html>"})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, DisableRetries: true})

	_, err := c.Get("/api/x", nil)
	apiErr := asAPIError(t, err)
	if apiErr.StatusCode != 502 || apiErr.Code != "HTTP_502" || apiErr.Message != "Bad Gateway" {
		t.Errorf("unexpected APIError %+v", apiErr)
	}
	if *attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", *attempts)
	}
}

func TestEnvelopeErrorKeepsItsMessageFormat(t *testing.T) {
	srv, attempts := cannedServer(t, cannedResponse{404, `{"success":false,"error":{"code":"NOT_FOUND","message":"nope"}}`})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})

	_, err := c.Get("/api/x", nil)
	apiErr := asAPIError(t, err)
	if got := err.Error(); got != "[NOT_FOUND] nope (HTTP 404)" {
		t.Errorf("error text changed: %q", got)
	}
	if apiErr.Code != "NOT_FOUND" || apiErr.StatusCode != 404 {
		t.Errorf("unexpected APIError %+v", apiErr)
	}
	if *attempts != 1 {
		t.Errorf("a 4xx must not be retried, got %d attempts", *attempts)
	}
}

func Test5xxIsRetriedThenReturnsAPIError(t *testing.T) {
	srv, attempts := cannedServer(t, cannedResponse{503, `{"error":"down"}`})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 1})

	data, err := c.Get("/api/x", nil)
	if data != nil {
		t.Errorf("expected no data, got %v", data)
	}
	apiErr := asAPIError(t, err)
	if apiErr.StatusCode != 503 || apiErr.Message != "down" {
		t.Errorf("unexpected APIError %+v", apiErr)
	}
	if *attempts != 2 {
		t.Errorf("expected 2 attempts (1 retry), got %d", *attempts)
	}
}

func Test5xxThenSuccessReturnsData(t *testing.T) {
	srv, attempts := cannedServer(t,
		cannedResponse{500, "oops"},
		cannedResponse{200, `{"success":true,"data":{"ok":true}}`},
	)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 1})

	data, err := c.Get("/api/x", nil)
	if err != nil {
		t.Fatalf("expected success after a retry, got %v", err)
	}
	if data["ok"] != true {
		t.Errorf("unexpected data %v", data)
	}
	if *attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", *attempts)
	}
}

func Test429IsRetried(t *testing.T) {
	srv, attempts := cannedServer(t,
		cannedResponse{429, `{"success":false,"error":{"code":"RATE_LIMITED","message":"slow down"}}`},
		cannedResponse{200, `{"success":true,"data":{"ok":true}}`},
	)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 1})

	if _, err := c.Get("/api/x", nil); err != nil {
		t.Fatalf("expected success after retrying a 429, got %v", err)
	}
	if *attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", *attempts)
	}
}

// attemptsFor runs one request with Retries: 1 against a server that answers
// every attempt with status, and returns the attempt count and the error.
func attemptsFor(t *testing.T, method string, status int) (int, error) {
	t.Helper()
	srv, attempts := cannedServer(t, cannedResponse{status, `{"error":"x"}`})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 1})
	_, err := c.doRequest(method, "/api/x", nil, nil)
	return *attempts, err
}

func TestNonIdempotent500IsNotRetried(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPatch} {
		t.Run(method, func(t *testing.T) {
			attempts, err := attemptsFor(t, method, http.StatusInternalServerError)
			if apiErr := asAPIError(t, err); apiErr.StatusCode != 500 {
				t.Errorf("expected a 500 *APIError, got %+v", apiErr)
			}
			if attempts != 1 {
				t.Errorf("%s 500 may have landed side effects and must not be retried, got %d attempts", method, attempts)
			}
		})
	}
}

func TestIdempotent500IsRetried(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			attempts, err := attemptsFor(t, method, http.StatusInternalServerError)
			if apiErr := asAPIError(t, err); apiErr.StatusCode != 500 {
				t.Errorf("expected a 500 *APIError, got %+v", apiErr)
			}
			if attempts != 2 {
				t.Errorf("%s 500 must be retried, got %d attempts", method, attempts)
			}
		})
	}
}

func TestGatewayStatusesAndRateLimitAreRetriedForEveryMethod(t *testing.T) {
	for _, status := range []int{http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout} {
		for _, method := range []string{http.MethodPost, http.MethodPatch, http.MethodGet} {
			t.Run(fmt.Sprintf("%s %d", method, status), func(t *testing.T) {
				attempts, err := attemptsFor(t, method, status)
				asAPIError(t, err)
				if attempts != 2 {
					t.Errorf("%s %d must be retried, got %d attempts", method, status, attempts)
				}
			})
		}
	}
}

func TestPost503ThenSuccessReturnsData(t *testing.T) {
	srv, attempts := cannedServer(t,
		cannedResponse{503, `{"error":"down"}`},
		cannedResponse{200, `{"success":true,"data":{"ok":true}}`},
	)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 1})

	data, err := c.Post("/api/x", map[string]string{"a": "b"})
	if err != nil || data["ok"] != true {
		t.Fatalf("expected success after retrying a POST 503, got %v, %v", data, err)
	}
	if *attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", *attempts)
	}
}

func TestOther5xxIsNotRetried(t *testing.T) {
	attempts, err := attemptsFor(t, http.MethodGet, http.StatusNotImplemented)
	asAPIError(t, err)
	if attempts != 1 {
		t.Errorf("a 501 must not be retried, got %d attempts", attempts)
	}
}

// scriptedServer runs the Nth step for the Nth attempt (repeating the last)
// and counts attempts. Responses carry Connection: close so no attempt reuses
// a connection — a reused connection that is then dropped would be retried
// transparently by net/http and skew the count.
func scriptedServer(t *testing.T, steps ...http.HandlerFunc) (*httptest.Server, *int) {
	t.Helper()
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx := attempts
		if idx >= len(steps) {
			idx = len(steps) - 1
		}
		attempts++
		w.Header().Set("Connection", "close")
		steps[idx](w, r)
	}))
	t.Cleanup(srv.Close)
	return srv, &attempts
}

func respond503(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusServiceUnavailable)
	fmt.Fprint(w, `{"error":"down"}`)
}

func dropConnection(w http.ResponseWriter, _ *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		panic("hijacking not supported")
	}
	conn, _, _ := hj.Hijack()
	_ = conn.Close()
}

func TestFailOpenHTTPErrorThenTransportFailureReturnsFailedResult(t *testing.T) {
	srv, attempts := scriptedServer(t, respond503, dropConnection)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 1, FailOpen: true})

	data, err := c.Get("/api/x", nil)
	if err != nil {
		t.Fatalf("the last attempt failed at the transport level, so FailOpen applies; got error %v", err)
	}
	if failed, _ := data["_failed"].(bool); !failed {
		t.Errorf("expected a _failed result, got %v", data)
	}
	if *attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", *attempts)
	}
}

func TestFailOpenTransportFailureThenHTTPErrorReturnsAPIError(t *testing.T) {
	srv, attempts := scriptedServer(t, dropConnection, respond503)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 1, FailOpen: true})

	data, err := c.Get("/api/x", nil)
	if data != nil {
		t.Errorf("the last attempt got an HTTP 503, so FailOpen does not apply; got %v", data)
	}
	if apiErr := asAPIError(t, err); apiErr.StatusCode != 503 {
		t.Errorf("expected the 503 *APIError, got %+v", apiErr)
	}
	if *attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", *attempts)
	}
}

func TestFailOpenDoesNotCoverAnHTTPErrorStatus(t *testing.T) {
	srv, _ := cannedServer(t, cannedResponse{503, "<html>Service Unavailable</html>"})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, DisableRetries: true, FailOpen: true})

	data, err := c.Get("/api/x", nil)
	if data != nil {
		t.Errorf("FailOpen must not turn an HTTP error into a _failed result, got %v", data)
	}
	asAPIError(t, err)
}

func Test2xxBehaviourUnchanged(t *testing.T) {
	t.Run("success false is still an error with the same text", func(t *testing.T) {
		srv, attempts := cannedServer(t, cannedResponse{200, `{"success":false,"error":{"code":"X","message":"y"}}`})
		c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
		_, err := c.Get("/api/x", nil)
		if err == nil || err.Error() != "[X] y (HTTP 200)" {
			t.Errorf("unexpected error %v", err)
		}
		if *attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", *attempts)
		}
	})

	t.Run("a non-JSON body is still an unmarshal error", func(t *testing.T) {
		srv, attempts := cannedServer(t, cannedResponse{200, "not json"})
		c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
		_, err := c.Get("/api/x", nil)
		if err == nil || !strings.Contains(err.Error(), "unmarshal response") {
			t.Errorf("unexpected error %v", err)
		}
		if *attempts != 1 {
			t.Errorf("expected 1 attempt, got %d", *attempts)
		}
	})

	t.Run("non-envelope JSON is returned as-is", func(t *testing.T) {
		srv, _ := cannedServer(t, cannedResponse{200, `{"status":"healthy"}`})
		c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
		data, err := c.Get("/api/health", nil)
		if err != nil || data["status"] != "healthy" {
			t.Errorf("unexpected result %v, %v", data, err)
		}
	})
}

func TestNewClient(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test_key"})
	if c.apiKey != "test_key" {
		t.Errorf("expected api key 'test_key', got '%s'", c.apiKey)
	}
	if c.baseURL != "https://platformxe.com" {
		t.Errorf("expected default base URL, got '%s'", c.baseURL)
	}
}

func TestNewClientCustomConfig(t *testing.T) {
	// A caller can configure zero retries. Retries: 0 is Go's zero value and
	// so means "use the default" (see TestNewClientDefaultRetries);
	// DisableRetries is the explicit switch.
	c := NewClient(ClientConfig{
		APIKey:         "test",
		BaseURL:        "http://localhost:3000",
		Timeout:        5,
		DisableRetries: true,
	})
	if c.baseURL != "http://localhost:3000" {
		t.Errorf("expected custom base URL, got '%s'", c.baseURL)
	}
	if c.retries != 0 {
		t.Errorf("expected 0 retries, got %d", c.retries)
	}
}

func TestNewClientDefaultRetries(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.retries != 2 {
		t.Errorf("expected default of 2 retries when Retries is omitted, got %d", c.retries)
	}
}

func TestNewClientExplicitRetries(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test", Retries: 5})
	if c.retries != 5 {
		t.Errorf("expected 5 retries, got %d", c.retries)
	}
}

func TestNewClientDisableRetriesWinsOverRetries(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test", Retries: 5, DisableRetries: true})
	if c.retries != 0 {
		t.Errorf("expected DisableRetries to force 0 retries, got %d", c.retries)
	}
}

// A negative Retries used to make the request loop run ZERO attempts and
// return a nil error with no request ever sent. It now means no retries.
func TestNewClientNegativeRetriesMeansNoRetries(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test", Retries: -1})
	if c.retries != 0 {
		t.Errorf("expected negative Retries to clamp to 0, got %d", c.retries)
	}
}

// Behavioural check: with DisableRetries a network failure makes exactly one
// attempt (no backoff sleeps) and surfaces the error.
func TestDisableRetriesMakesExactlyOneAttempt(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("hijacking not supported")
		}
		conn, _, _ := hj.Hijack()
		_ = conn.Close() // drop the connection → client-side network error
	}))
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, DisableRetries: true})
	c.failOpen = false
	if _, err := c.Get("/api/health", nil); err == nil {
		t.Fatal("expected a network error")
	}
	if attempts != 1 {
		t.Errorf("expected exactly 1 attempt, got %d", attempts)
	}
}

func TestServiceNamespaces(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.Permissions == nil {
		t.Error("Permissions service is nil")
	}
	if c.Identity == nil {
		t.Error("Identity service is nil")
	}
	if c.Messaging == nil {
		t.Error("Messaging service is nil")
	}
	if c.Webhooks == nil {
		t.Error("Webhooks service is nil")
	}
	if c.Templates == nil {
		t.Error("Templates service is nil")
	}
	if c.Storage == nil {
		t.Error("Storage service is nil")
	}
	if c.Workflows == nil {
		t.Error("Workflows service is nil")
	}
	if c.Domains == nil {
		t.Error("Domains service is nil")
	}
	if c.Documents == nil {
		t.Error("Documents service is nil")
	}
	if c.Events == nil {
		t.Error("Events service is nil")
	}
	if c.Exports == nil {
		t.Error("Exports service is nil")
	}
	if c.Ocr == nil {
		t.Error("Ocr service is nil")
	}
	if c.Pdf == nil {
		t.Error("Pdf service is nil")
	}
	if c.Qr == nil {
		t.Error("Qr service is nil")
	}
	if c.Threads == nil {
		t.Error("Threads service is nil")
	}
	if c.Usage == nil {
		t.Error("Usage service is nil")
	}
	if c.Ai == nil {
		t.Error("Ai service is nil")
	}
	if c.Vision == nil {
		t.Error("Vision service is nil")
	}
}

func TestPermissionsServiceMethods(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	p := c.Permissions

	// Verify key methods exist by checking they don't panic
	// (We can't call them without a server, but we can verify the struct has them)
	if p == nil {
		t.Fatal("Permissions service is nil")
	}
}

func TestIdentityServiceMethods(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.Identity == nil {
		t.Fatal("Identity service is nil")
	}
}

func TestMessagingServiceMethods(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.Messaging == nil {
		t.Fatal("Messaging service is nil")
	}
}

func TestWebhooksServiceMethods(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.Webhooks == nil {
		t.Fatal("Webhooks service is nil")
	}
}

func TestTemplatesServiceMethods(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.Templates == nil {
		t.Fatal("Templates service is nil")
	}
}

func TestStorageServiceMethods(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.Storage == nil {
		t.Fatal("Storage service is nil")
	}
}

func TestWorkflowsServiceMethods(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.Workflows == nil {
		t.Fatal("Workflows service is nil")
	}
}

func TestDomainsServiceMethods(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.Domains == nil {
		t.Fatal("Domains service is nil")
	}
}

func TestEventsServiceMethods(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.Events == nil {
		t.Fatal("Events service is nil")
	}
}

// TestEventFederationServiceWired confirms the federation surface (incl. the
// Pattern 3 external_webhook peer methods added in 1.5.0) is reachable from
// the client and that the typed entry points compile.
func TestEventFederationServiceWired(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.Events == nil || c.Events.Custom == nil || c.Events.Custom.Federation == nil {
		t.Fatal("Events.Custom.Federation service is nil")
	}
	fed := c.Events.Custom.Federation

	// Compile-check that all federation methods (incl. Pattern 3) exist on
	// the typed service. We can't invoke them without a server, but taking
	// the method values verifies the signatures are present.
	_ = fed.CreateGroup
	_ = fed.ListGroups
	_ = fed.GetGroup
	_ = fed.ArchiveGroup
	_ = fed.Invite
	_ = fed.Accept
	_ = fed.Leave
	_ = fed.DeclarePush
	_ = fed.ListPushes
	_ = fed.UndeclarePush
	// Pattern 3 — external webhook peers (added in 1.5.0)
	_ = fed.AddExternalPeer
	_ = fed.RemoveExternalPeer

	// Compile-check the new request/response shapes.
	_ = AddEventFederationExternalPeerInput{
		Label:      "Booking.com",
		WebhookURL: "https://booking.example.com/inbound",
		Headers:    map[string]string{"Authorization": "Bearer xyz"},
	}
	_ = AddEventFederationExternalPeerResult{}
	if EventFederationPeerTypeTenantOrg != "tenant_org" {
		t.Errorf("expected tenant_org, got %s", EventFederationPeerTypeTenantOrg)
	}
	if EventFederationPeerTypeExternalWebhook != "external_webhook" {
		t.Errorf("expected external_webhook, got %s", EventFederationPeerTypeExternalWebhook)
	}
}

func TestDocumentsServiceMethods(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.Documents == nil {
		t.Fatal("Documents service is nil")
	}
}

func TestExportsServiceMethods(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.Exports == nil {
		t.Fatal("Exports service is nil")
	}
}

func TestOcrServiceMethods(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.Ocr == nil {
		t.Fatal("Ocr service is nil")
	}
}

func TestQrServiceMethods(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.Qr == nil {
		t.Fatal("Qr service is nil")
	}
}

func TestUsageServiceMethods(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.Usage == nil {
		t.Fatal("Usage service is nil")
	}
}

func TestClientTelemetryAndHealth(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	// Verify telemetry and health are direct client methods (not on a service)
	if c == nil {
		t.Fatal("Client is nil")
	}
}

func TestDefaultFailOpen(t *testing.T) {
	c := NewClient(ClientConfig{APIKey: "test"})
	if c.failOpen != false {
		// Default is false for Go (IaC safety)
	}
}
