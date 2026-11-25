// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-10-09, v1.7.0 pending — final review):
// - Created. The per-call options of the billed AI / vision calls: a model
//   call is not re-sent after a 5xx or a transport failure and never returns
//   the FailOpen placeholder; a 429 RATE_LIMITED is still retried; analyze-
//   room re-sends only with an idempotency key; other calls keep the
//   client's retries.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"testing"
)

func upstreamError() aiVisionErrorFixture {
	return aiVisionErrorFixture{
		Status: 502,
		Body:   json.RawMessage(`{"success":false,"error":{"code":"AI_UPSTREAM_ERROR","message":"AI Gateway error"}}`),
	}
}

func TestModelCallIsNotResentAfterA5xx(t *testing.T) {
	srv, attempts := errorFixtureServer(t, upstreamError())
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 2})

	if _, err := c.Ai.Complete(AiCompleteInput{FeatureKey: "guest-concierge", Provider: "gemini"}); err == nil {
		t.Fatal("expected an error")
	}
	if *attempts != 1 {
		t.Errorf("a billed model call must be sent once, got %d attempts", *attempts)
	}
	if _, err := c.Ai.Extract(AiExtractInput{FeatureKey: "document.freeform"}); err == nil {
		t.Fatal("expected an error")
	}
	if *attempts != 2 {
		t.Errorf("extract must be sent once too, got %d attempts in total", *attempts)
	}
}

func TestModelCallTransportFailureIsAnErrorEvenWithFailOpen(t *testing.T) {
	// Nothing listens on this address: every attempt is a transport failure.
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: "http://127.0.0.1:1", Retries: 2, FailOpen: true})
	result, err := c.Ai.Complete(AiCompleteInput{FeatureKey: "guest-concierge", Provider: "gemini"})
	if err == nil || result != nil {
		t.Fatalf("a model call must not fail open into an empty result (result=%v, err=%v)", result, err)
	}
}

func TestAnalyzeRoomResendsOnlyWithAnIdempotencyKey(t *testing.T) {
	srv, attempts := errorFixtureServer(t, aiVisionErrorFixture{
		Status: 503,
		Body:   json.RawMessage(`{"success":false,"error":{"code":"SERVICE_UNAVAILABLE","message":"busy"}}`),
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 1})
	input := AnalyzeRoomInput{ImageURL: "https://x.test/a.jpg", RoomCategories: []string{"KITCHEN"}}

	_, _ = c.Vision.AnalyzeRoom(input, nil)
	if *attempts != 1 {
		t.Errorf("without an idempotency key analyze-room is sent once, got %d", *attempts)
	}
	_, _ = c.Vision.AnalyzeRoom(input, &AnalyzeRoomOptions{IdempotencyKey: "photo-1"})
	if *attempts != 3 {
		t.Errorf("with an idempotency key it is retried, got %d attempts in total", *attempts)
	}
}

func TestOtherCallsKeepTheClientRetries(t *testing.T) {
	srv, attempts := errorFixtureServer(t, upstreamError())
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 1})
	_, _ = c.Ai.Profile("ai-canvas")
	if *attempts != 2 {
		t.Errorf("a GET keeps the client's retries, got %d attempts", *attempts)
	}
}
