// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-10-08, v1.7.0):
// - Created (caldera-platformxe#22, NO-DRIFT for the vision route).
//   VisionService.AnalyzeRoom for POST /api/v1/vision/analyze-room — scope
//   vision:analyze, 2,000 requests/hr/key. AnalyzeRoomInput /
//   AnalyzeRoomResult (types.go) mirror @caldera/platformxe-types 3.5.0
//   src/vision.ts. AnalyzeRoomOptions.IdempotencyKey is sent as
//   x-idempotency-key (client.go doRequestTypedWithHeaders); a FALLBACK
//   answer is never cached under it server-side.
//
// CHANGELOG (2026-10-09, v1.7.0 pending — final review):
// - AnalyzeRoom waits up to 35 s (the route runs up to 30 s; the client
//   default is 10 s) and is retried only with an idempotency key — without
//   one a retry is a second billed analysis.
// =============================================================================

package platformxe

import "time"

// VisionService handles the image analysis API calls.
type VisionService struct {
	client *Client
}

// AnalyzeRoom classifies a property photo into one of input.RoomCategories
// and (when input.WantAltText) writes alt text. A provider timeout is not an
// error: it returns AiStatus "FALLBACK". opts may be nil.
func (s *VisionService) AnalyzeRoom(input AnalyzeRoomInput, opts *AnalyzeRoomOptions) (*AnalyzeRoomResult, error) {
	var headers map[string]string
	if opts != nil && opts.IdempotencyKey != "" {
		headers = map[string]string{"x-idempotency-key": opts.IdempotencyKey}
	}
	var result AnalyzeRoomResult
	// The route runs up to 30 s; without an idempotency key a retry is a
	// second billed analysis.
	call := callOptions{timeout: 35 * time.Second}
	if headers == nil {
		call.rateLimitRetriesOnly = true
	}
	if err := s.client.doRequestTypedOpts("POST", "/api/v1/vision/analyze-room", input, nil, headers, call, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
