// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-22, v1.16.0 pending, AI gateway T4):
// - Created (NO-DRIFT). AiToolsService.List for GET /api/v1/ai/tools (scope
//   tools:platform.read, the default 1,000 requests/hr/key tier). AiToolList
//   mirrors @caldera/platformxe-types 3.18.0.
// =============================================================================

package platformxe

// AiToolsService lists the platform tools an API key may use in the
// server-side tool loop (Toolsets on Ai.Messages and Ai.Complete).
type AiToolsService struct {
	client *Client
}

// List returns the platform tools this key may use: name, version,
// description, the full input schema, data classes and toolset. An
// organisation key holding tools:platform.read gets the platform.read tools,
// which read only its own organisation; a platform key gets none. No AI
// profile is read here, so a profile can narrow the list at call time. A key
// without the scope is an *APIError with Code FORBIDDEN (403).
func (s *AiToolsService) List() (*AiToolList, error) {
	var result AiToolList
	if err := s.client.doRequestTyped("GET", "/api/v1/ai/tools", nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
