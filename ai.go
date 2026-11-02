// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-10-08, v1.7.0):
// - Created (caldera-platformxe#22, NO-DRIFT for the AI gateway routes).
//   AiService.Complete / Extract / Profile / Usage for POST
//   /api/v1/ai/complete, POST /api/v1/ai/extract, GET /api/v1/ai/profile and
//   GET /api/v1/ai/usage — scope ai:complete, 600 requests/hr/key. Typed
//   request / response structs in types.go mirror
//   @caldera/platformxe-types 3.5.0 src/ai.ts. Errors are *APIError with the
//   closed ErrCodeAi* set; Details["path"] on AI_OUTPUT_INVALID, RetryAfter
//   on AI_QUOTA_EXCEEDED (which client.go never retries).
//
// CHANGELOG (2026-10-08, v1.7.0 pending — calderasuite/caldera-xadmin#29):
// - Complete carries tool calling: AiCompleteInput.Tools / ToolChoice,
//   AiMessage.ToolCalls / ToolCallID, AiCompleteResult.ToolCalls /
//   FinishReason (types.go). PlatformXe never executes a tool.
//
// CHANGELOG (2026-10-09, v1.7.0 pending — final review):
// - Complete and Extract run with modelCallOptions (client.go): a 155 s
//   timeout, sent once, and never the FailOpen placeholder — every attempt
//   is a billed model call counted against the daily cap.
// =============================================================================

package platformxe

// AiService handles the product AI gateway API calls.
type AiService struct {
	client *Client
}

// Complete runs one chat completion under the AI profile for (the API key's
// caller service, input.FeatureKey). With input.Tools (profile allowTools)
// the result may carry ToolCalls (FinishReason "tool_calls"): run each call
// yourself and send the result back as a "tool" message with ToolCallID.
func (s *AiService) Complete(input AiCompleteInput) (*AiCompleteResult, error) {
	var result AiCompleteResult
	if err := s.client.doRequestTypedOpts("POST", "/api/v1/ai/complete", input, nil, nil, modelCallOptions(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Extract runs a structured extraction. The reply is validated against
// input.OutputJSONSchema; an off-schema reply is an *APIError with Code
// AI_OUTPUT_INVALID (422) and Details["path"].
func (s *AiService) Extract(input AiExtractInput) (*AiExtractResult, error) {
	var result AiExtractResult
	if err := s.client.doRequestTypedOpts("POST", "/api/v1/ai/extract", input, nil, nil, modelCallOptions(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Profile returns the calling service's own AI profile for one feature:
// status, allow-lists, limits and today's (UTC) usage.
func (s *AiService) Profile(featureKey string) (*AiProfileView, error) {
	var result AiProfileView
	params := map[string]string{"featureKey": featureKey}
	if err := s.client.doRequestTyped("GET", "/api/v1/ai/profile", nil, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Usage returns the calling service's own AI and vision usage, grouped by
// feature, kind, provider and model. Empty query fields are omitted (server
// default: the last 30 days).
func (s *AiService) Usage(query AiUsageQuery) (*AiUsageSummary, error) {
	params := map[string]string{}
	if query.From != "" {
		params["from"] = query.From
	}
	if query.To != "" {
		params["to"] = query.To
	}
	if query.FeatureKey != "" {
		params["featureKey"] = query.FeatureKey
	}
	var result AiUsageSummary
	if err := s.client.doRequestTyped("GET", "/api/v1/ai/usage", nil, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
