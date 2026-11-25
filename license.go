// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-22, v1.15.0 pending, Enterprise License):
// - Initial: LicenseService.Get for GET /api/v1/license (scope license:read).
// =============================================================================

package platformxe

const licensePath = "/api/v1/license"

// LicenseService reads the key's own organisation's Enterprise License
// (scope license:read; billing:read does not grant it). Read-only: a license
// is installed by its holder in the portal (Billing, License card), never
// through the API. The answer never carries money, the license text or
// another organisation's grants.
type LicenseService struct {
	client *Client
}

// Get returns the organisation's license position: its effective plan and
// where it comes from, whether it is exempt (first_party, demo), and the
// license that covers it (nil when none does) with the plugins granted to
// this organisation only. Errors: ErrCodeForbidden (403: no license:read, or
// a platform key with no organisation), ErrCodeNotFound (404),
// ErrCodeRateLimited (429), ErrCodeInternalError (500).
func (s *LicenseService) Get() (*LicenseView, error) {
	var result LicenseView
	if err := s.client.doRequestTyped("GET", licensePath, nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
