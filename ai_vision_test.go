// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-10-08, v1.7.0):
// - Created (caldera-platformxe#22). Contract tests for AiService and
//   VisionService against testdata/ai_vision.json — the shared fixture set
//   (byte-identical copy in packages/sdk-python/tests/fixtures/). Each test
//   runs a real net/http/httptest server and pins the method, path, query,
//   headers and the exact JSON body against the fixture request (decoded and
//   compared as generic JSON, so a renamed or dropped field fails), decodes
//   the fixture response into the typed result, and pins the error surface:
//   Code / StatusCode, Details["path"] on 422 AI_OUTPUT_INVALID, RetryAfter
//   on 429 AI_QUOTA_EXCEEDED (one attempt only — not retried) while a 429
//   RATE_LIMITED still is.
//
// CHANGELOG (2026-11-01, pending — caldera-platformxe#33):
// - TestAiClosedErrorCodes covers 503 AI_NOT_CONFIGURED and the new 503
//   AI_MODEL_UNAVAILABLE.
// - Review: TestAiConfigurationErrorsAreNotRetried — a 503 AI_NOT_CONFIGURED
//   or AI_MODEL_UNAVAILABLE on an idempotent call (Ai.Profile) is returned
//   after one attempt, while a plain 503 is still retried.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

type aiVisionErrorFixture struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    json.RawMessage   `json:"body"`
}

type aiVisionFixtures struct {
	CompleteRequest          json.RawMessage      `json:"completeRequest"`
	CompleteResult           json.RawMessage      `json:"completeResult"`
	ExtractRequest           json.RawMessage      `json:"extractRequest"`
	ExtractResult            json.RawMessage      `json:"extractResult"`
	ProfileView              json.RawMessage      `json:"profileView"`
	UsageSummary             json.RawMessage      `json:"usageSummary"`
	AnalyzeRoomRequest       json.RawMessage      `json:"analyzeRoomRequest"`
	AnalyzeRoomResult        json.RawMessage      `json:"analyzeRoomResult"`
	AnalyzeRoomFallback      json.RawMessage      `json:"analyzeRoomFallback"`
	OutputInvalidError       aiVisionErrorFixture `json:"outputInvalidError"`
	QuotaExceededError       aiVisionErrorFixture `json:"quotaExceededError"`
	VisionNotConfiguredError aiVisionErrorFixture `json:"visionNotConfiguredError"`
}

func loadAiVisionFixtures(t *testing.T) aiVisionFixtures {
	t.Helper()
	raw, err := os.ReadFile("testdata/ai_vision.json")
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	var f aiVisionFixtures
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("parse fixtures: %v", err)
	}
	return f
}

// decodeInto unmarshals a fixture into v, failing the test on error.
func decodeInto(t *testing.T, raw json.RawMessage, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
}

// assertSameJSON compares two JSON documents structurally.
func assertSameJSON(t *testing.T, label string, got []byte, want json.RawMessage) {
	t.Helper()
	var g, w interface{}
	if err := json.Unmarshal(got, &g); err != nil {
		t.Fatalf("%s: decode got: %v", label, err)
	}
	if err := json.Unmarshal(want, &w); err != nil {
		t.Fatalf("%s: decode want: %v", label, err)
	}
	if !reflect.DeepEqual(g, w) {
		t.Errorf("%s mismatch\n got: %s\nwant: %s", label, got, want)
	}
}

// fixtureServer answers every request with {success:true, data:<data>} after
// running check on it.
func fixtureServer(t *testing.T, data json.RawMessage, headers map[string]string, check func(r *http.Request, body []byte)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if check != nil {
			check(r, body)
		}
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success":true,"data":%s}`, data)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// errorFixtureServer answers every request with an error fixture and counts attempts.
func errorFixtureServer(t *testing.T, f aiVisionErrorFixture) (*httptest.Server, *int) {
	t.Helper()
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		for k, v := range f.Headers {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(f.Status)
		_, _ = w.Write(f.Body)
	}))
	t.Cleanup(srv.Close)
	return srv, &attempts
}

func TestAiCompleteMatchesFixture(t *testing.T) {
	f := loadAiVisionFixtures(t)
	srv := fixtureServer(t, f.CompleteResult, nil, func(r *http.Request, body []byte) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/ai/complete" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		assertSameJSON(t, "complete body", body, f.CompleteRequest)
	})
	var input AiCompleteInput
	decodeInto(t, f.CompleteRequest, &input)

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Ai.Complete(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var want AiCompleteResult
	decodeInto(t, f.CompleteResult, &want)
	if !reflect.DeepEqual(*result, want) {
		t.Errorf("result mismatch: %#v", result)
	}
	if result.Usage == nil || result.Usage.TotalTokens != 51 {
		t.Errorf("usage not decoded: %#v", result.Usage)
	}
}

func TestAiExtractMatchesFixture(t *testing.T) {
	f := loadAiVisionFixtures(t)
	srv := fixtureServer(t, f.ExtractResult, nil, func(r *http.Request, body []byte) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/ai/extract" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		assertSameJSON(t, "extract body", body, f.ExtractRequest)
	})
	var input AiExtractInput
	decodeInto(t, f.ExtractRequest, &input)

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Ai.Extract(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Data["invoiceNumber"] != "INV-001" || result.Data["total"] != float64(45000) {
		t.Errorf("data not decoded: %#v", result.Data)
	}
}

func TestAiExtractSendsZeroTemperatureWhenSet(t *testing.T) {
	zero := 0.0
	srv := fixtureServer(t, json.RawMessage(`{"data":{},"provider":"gemini","model":"m"}`), nil, func(_ *http.Request, body []byte) {
		var m map[string]interface{}
		_ = json.Unmarshal(body, &m)
		if v, ok := m["temperature"]; !ok || v != float64(0) {
			t.Errorf("temperature 0 should be sent, got %v (present=%v)", v, ok)
		}
		if _, ok := m["images"]; ok {
			t.Errorf("images should be omitted when empty")
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Ai.Extract(AiExtractInput{
		FeatureKey: "document.freeform", SystemPrompt: "s", UserPrompt: "u",
		OutputJSONSchema: map[string]interface{}{"type": "object"}, ModelTier: "FAST",
		Temperature: &zero,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAiProfileMatchesFixture(t *testing.T) {
	f := loadAiVisionFixtures(t)
	srv := fixtureServer(t, f.ProfileView, nil, func(r *http.Request, _ []byte) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/ai/profile" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if got := r.URL.Query().Get("featureKey"); got != "guest-concierge" {
			t.Errorf("featureKey = %q", got)
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Ai.Profile("guest-concierge")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var want AiProfileView
	decodeInto(t, f.ProfileView, &want)
	if !reflect.DeepEqual(*result, want) {
		t.Errorf("result mismatch: %#v", result)
	}
	if result.DailyTokenCap == nil || *result.DailyTokenCap != 2000000 {
		t.Errorf("dailyTokenCap not decoded: %v", result.DailyTokenCap)
	}
}

func TestAiUsageMatchesFixture(t *testing.T) {
	f := loadAiVisionFixtures(t)
	srv := fixtureServer(t, f.UsageSummary, nil, func(r *http.Request, _ []byte) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/ai/usage" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("from") != "2026-09-08T00:00:00.000Z" || q.Get("to") != "2026-10-08T00:00:00.000Z" || q.Get("featureKey") != "guest-concierge" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Ai.Usage(AiUsageQuery{
		From: "2026-09-08T00:00:00.000Z", To: "2026-10-08T00:00:00.000Z", FeatureKey: "guest-concierge",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Rows) != 2 || result.Totals.Requests != 57 || result.Totals.CostCredits != 101 {
		t.Errorf("summary not decoded: %#v", result)
	}
	if result.Rows[1].Kind != "vision" || result.Rows[1].Provider == nil || *result.Rows[1].Provider != "azure" {
		t.Errorf("row not decoded: %#v", result.Rows[1])
	}
}

func TestAiUsageOmitsEmptyQuery(t *testing.T) {
	f := loadAiVisionFixtures(t)
	srv := fixtureServer(t, f.UsageSummary, nil, func(r *http.Request, _ []byte) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query, got %q", r.URL.RawQuery)
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Ai.Usage(AiUsageQuery{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAiOutputInvalidCarriesDetailsPath(t *testing.T) {
	f := loadAiVisionFixtures(t)
	srv, attempts := errorFixtureServer(t, f.OutputInvalidError)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})

	result, err := c.Ai.Extract(AiExtractInput{FeatureKey: "document.freeform", ModelTier: "FAST"})
	if result != nil {
		t.Errorf("expected nil result, got %#v", result)
	}
	apiErr := asAPIError(t, err)
	if apiErr.Code != ErrCodeAiOutputInvalid || apiErr.StatusCode != 422 {
		t.Errorf("unexpected APIError: %#v", apiErr)
	}
	if apiErr.Details["path"] != "/total" {
		t.Errorf("details.path = %v", apiErr.Details["path"])
	}
	if *attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", *attempts)
	}
}

func TestAiQuotaExceededIsNotRetriedAndCarriesRetryAfter(t *testing.T) {
	f := loadAiVisionFixtures(t)
	srv, attempts := errorFixtureServer(t, f.QuotaExceededError)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 2})

	_, err := c.Ai.Complete(AiCompleteInput{FeatureKey: "guest-concierge", Provider: "gemini"})
	apiErr := asAPIError(t, err)
	if apiErr.Code != ErrCodeAiQuotaExceeded || apiErr.StatusCode != 429 {
		t.Errorf("unexpected APIError: %#v", apiErr)
	}
	if apiErr.RetryAfter != 3600 {
		t.Errorf("RetryAfter = %d", apiErr.RetryAfter)
	}
	if apiErr.Details["metric"] != "tokens" {
		t.Errorf("details.metric = %v", apiErr.Details["metric"])
	}
	if *attempts != 1 {
		t.Errorf("AI_QUOTA_EXCEEDED must not be retried, got %d attempts", *attempts)
	}
}

func TestRateLimited429IsStillRetried(t *testing.T) {
	srv, attempts := errorFixtureServer(t, aiVisionErrorFixture{
		Status:  429,
		Headers: map[string]string{"Retry-After": "1"},
		Body:    json.RawMessage(`{"success":false,"error":{"code":"RATE_LIMITED","message":"Too many requests"}}`),
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 1})

	_, err := c.Ai.Complete(AiCompleteInput{FeatureKey: "guest-concierge", Provider: "gemini"})
	apiErr := asAPIError(t, err)
	if apiErr.Code != "RATE_LIMITED" || apiErr.RetryAfter != 1 {
		t.Errorf("unexpected APIError: %#v", apiErr)
	}
	if *attempts != 2 {
		t.Errorf("RATE_LIMITED should be retried, got %d attempts", *attempts)
	}
}

func TestAiClosedErrorCodes(t *testing.T) {
	cases := []struct {
		status int
		code   string
	}{
		{403, ErrCodeAiProfileNotFound},
		{403, ErrCodeAiProfileDisabled},
		{403, ErrCodeAiPolicyDenied},
		{422, ErrCodeAiOutputBlocked},
		{400, ErrCodeBadRequest},
		{503, ErrCodeAiNotConfigured},
		{503, ErrCodeAiModelUnavailable},
	}
	for _, tc := range cases {
		srv, _ := cannedServer(t, cannedResponse{tc.status, fmt.Sprintf(`{"success":false,"error":{"code":%q,"message":"refused"}}`, tc.code)})
		c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
		_, err := c.Ai.Complete(AiCompleteInput{FeatureKey: "guest-concierge", Provider: "gemini"})
		apiErr := asAPIError(t, err)
		if apiErr.Code != tc.code || apiErr.StatusCode != tc.status {
			t.Errorf("%s: unexpected APIError: %#v", tc.code, apiErr)
		}
	}
}

func TestAiConfigurationErrorsAreNotRetried(t *testing.T) {
	for _, code := range []string{ErrCodeAiNotConfigured, ErrCodeAiModelUnavailable} {
		srv, attempts := cannedServer(t, cannedResponse{503, fmt.Sprintf(`{"success":false,"error":{"code":%q,"message":"refused"}}`, code)})
		c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
		_, err := c.Ai.Profile("guest-concierge")
		apiErr := asAPIError(t, err)
		if apiErr.Code != code || apiErr.StatusCode != 503 {
			t.Errorf("%s: unexpected APIError: %#v", code, apiErr)
		}
		if *attempts != 1 {
			t.Errorf("%s must not be retried, got %d attempts", code, *attempts)
		}
	}
	// A plain 503 on the same idempotent call is still retried.
	srv, attempts := cannedServer(t, cannedResponse{503, `{"success":false,"error":{"code":"SERVICE_UNAVAILABLE","message":"down"}}`})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	_, _ = c.Ai.Profile("guest-concierge")
	if *attempts < 2 {
		t.Errorf("a plain 503 should be retried, got %d attempts", *attempts)
	}
}

func TestVisionAnalyzeRoomMatchesFixture(t *testing.T) {
	f := loadAiVisionFixtures(t)
	srv := fixtureServer(t, f.AnalyzeRoomResult, nil, func(r *http.Request, body []byte) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/vision/analyze-room" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("x-idempotency-key") != "" {
			t.Errorf("no idempotency key expected, got %q", r.Header.Get("x-idempotency-key"))
		}
		assertSameJSON(t, "analyze-room body", body, f.AnalyzeRoomRequest)
	})
	var input AnalyzeRoomInput
	decodeInto(t, f.AnalyzeRoomRequest, &input)

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Vision.AnalyzeRoom(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var want AnalyzeRoomResult
	decodeInto(t, f.AnalyzeRoomResult, &want)
	if !reflect.DeepEqual(*result, want) {
		t.Errorf("result mismatch: %#v", result)
	}
}

func TestVisionAnalyzeRoomSendsIdempotencyKeyAndReturnsFallback(t *testing.T) {
	f := loadAiVisionFixtures(t)
	srv := fixtureServer(t, f.AnalyzeRoomFallback, map[string]string{"X-Idempotency-Cache": "skip"}, func(r *http.Request, _ []byte) {
		if got := r.Header.Get("x-idempotency-key"); got != "photo-123" {
			t.Errorf("x-idempotency-key = %q", got)
		}
		if got := r.Header.Get("x-api-key"); got != "test" {
			t.Errorf("x-api-key = %q", got)
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Vision.AnalyzeRoom(
		AnalyzeRoomInput{ImageURL: "https://cdn.example.com/a.jpg", RoomCategories: []string{"KITCHEN"}},
		&AnalyzeRoomOptions{IdempotencyKey: "photo-123"},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AiStatus != "FALLBACK" || result.RoomCategory != nil {
		t.Errorf("fallback not decoded: %#v", result)
	}
}

func TestVisionNotConfiguredSurfaces(t *testing.T) {
	f := loadAiVisionFixtures(t)
	srv, _ := errorFixtureServer(t, f.VisionNotConfiguredError)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, DisableRetries: true})
	_, err := c.Vision.AnalyzeRoom(AnalyzeRoomInput{ImageURL: "https://x.test/a.jpg", RoomCategories: []string{"KITCHEN"}}, nil)
	apiErr := asAPIError(t, err)
	if apiErr.Code != ErrCodeVisionNotConfigured || apiErr.StatusCode != 503 {
		t.Errorf("unexpected APIError: %#v", apiErr)
	}
}
