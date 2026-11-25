// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-09-15): gofmt alignment of WhoamiResponse only — `gofmt -l`
// listed this file in CI.
// =============================================================================

package platformxe

// WhoamiService verifies the calling API key and surfaces caller identity.
//
// Designed for boot-time self-checks. The response includes grace-period
// and expiry warnings, plus platform-side env hints so consumers can
// detect environment drift.
type WhoamiService struct {
	client *Client
}

// WhoamiResponse is the response from GET /api/v1/whoami.
type WhoamiResponse struct {
	CallerService      string   `json:"callerService"`
	OrganizationID     string   `json:"organizationId"`
	APIKeyID           string   `json:"apiKeyId"`
	Label              *string  `json:"label"`
	Scopes             []string `json:"scopes"`
	RequestCount       int      `json:"requestCount"`
	LastUsedAt         *string  `json:"lastUsedAt"`
	ExpiresAt          *string  `json:"expiresAt"`
	GracePeriodWarning *string  `json:"gracePeriodWarning"`
	KeyExpiryWarning   *string  `json:"keyExpiryWarning"`
	PlatformxDbHost    *string  `json:"platformxDbHost"`
	PlatformxDbEnv     string   `json:"platformxDbEnv"` // "dev" | "prod" | "unknown"
	PlatformxNodeEnv   string   `json:"platformxNodeEnv"`
}

// Get returns the calling API key's identity, scopes, and metadata.
func (s *WhoamiService) Get() (*WhoamiResponse, error) {
	var result WhoamiResponse
	err := s.client.doRequestTyped("GET", "/api/v1/whoami", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
