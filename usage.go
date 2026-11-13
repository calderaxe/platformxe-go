// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

// UsageService handles usage metering and billing summaries.
type UsageService struct {
	client *Client
}

// Summary returns usage metrics for a billing month.
func (s *UsageService) Summary(month string) (*UsageSummary, error) {
	var result UsageSummary
	err := s.client.doRequestTyped("GET", "/api/v1/usage/summary", nil, map[string]string{"month": month}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
