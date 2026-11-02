// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

import "fmt"

// IdentityService handles identity resolution and verification.
type IdentityService struct {
	client *Client

	// Dlq covers the identity verification dead-letter queue (Phase 6F.5b).
	Dlq *IdentityDlqService
}

func newIdentityService(c *Client) *IdentityService {
	s := &IdentityService{client: c}
	s.Dlq = &IdentityDlqService{client: c}
	return s
}

// GetProcessor returns the identity processor configuration.
func (s *IdentityService) GetProcessor() (*ProcessorConfig, error) {
	var result ProcessorConfig
	err := s.client.doRequestTyped("GET", "/api/v1/identity/processor", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateProcessor updates the identity processor configuration.
func (s *IdentityService) UpdateProcessor(input map[string]interface{}) (*ProcessorConfig, error) {
	var result ProcessorConfig
	err := s.client.doRequestTyped("PUT", "/api/v1/identity/processor", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Resolve resolves an identity by identifier type and value.
func (s *IdentityService) Resolve(identifierType, value string, resolveLinked bool, consentRef string) (*IdentityResolveResult, error) {
	var result IdentityResolveResult
	err := s.client.doRequestTyped("POST", "/api/v1/identity/resolve", map[string]interface{}{
		"identifierType": identifierType,
		"value":          value,
		"resolveLinked":  resolveLinked,
		"consentRef":     consentRef,
	}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Verify verifies an identity document.
func (s *IdentityService) Verify(identifierType, value, firstName, lastName string) (*IdentityVerifyResult, error) {
	var result IdentityVerifyResult
	err := s.client.doRequestTyped("POST", "/api/v1/identity/verify", map[string]interface{}{
		"identifierType": identifierType,
		"value":          value,
		"firstName":      firstName,
		"lastName":       lastName,
	}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Lookup looks up an identity by identifier type and value.
func (s *IdentityService) Lookup(identifierType, value string) (*IdentityLookupResult, error) {
	var result IdentityLookupResult
	err := s.client.doRequestTyped("GET", "/api/v1/identity/lookup", nil, map[string]string{
		"type": identifierType, "value": value,
	}, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Providers returns available identity verification providers.
func (s *IdentityService) Providers() (*IdentityProvidersResponse, error) {
	var result IdentityProvidersResponse
	err := s.client.doRequestTyped("GET", "/api/v1/identity/providers", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ProvidersHealth returns the per-(country, provider) breaker + latency
// rollup (Phase 6F.5).
func (s *IdentityService) ProvidersHealth() (map[string]interface{}, error) {
	return s.client.doRequest("GET", "/api/v1/identity/providers/health", nil, nil)
}

// ----------------------------------------------------------------------------
// Per-kind verifications (Phase 6F)
// ----------------------------------------------------------------------------

// VerifyBVNRequest is the body for POST /api/v1/identity/verify-bvn.
type VerifyBVNRequest struct {
	SubjectID string                 `json:"subjectId"`
	BVN       string                 `json:"bvn"`
	Match     map[string]interface{} `json:"matchAgainst"`
}

func (s *IdentityService) VerifyBVN(req VerifyBVNRequest) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/identity/verify-bvn", req, nil)
}

// VerifyNINRequest is the body for POST /api/v1/identity/verify-nin.
type VerifyNINRequest struct {
	SubjectID string                 `json:"subjectId"`
	NIN       string                 `json:"nin"`
	Match     map[string]interface{} `json:"matchAgainst"`
}

func (s *IdentityService) VerifyNIN(req VerifyNINRequest) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/identity/verify-nin", req, nil)
}

// VerifyAccountRequest is the body for POST /api/v1/identity/verify-account.
type VerifyAccountRequest struct {
	SubjectID     string                 `json:"subjectId"`
	AccountNumber string                 `json:"accountNumber"`
	BankCode      string                 `json:"bankCode"`
	ExpectedName  map[string]interface{} `json:"expectedName"`
	Threshold     float64                `json:"threshold,omitempty"`
}

func (s *IdentityService) VerifyAccount(req VerifyAccountRequest) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/identity/verify-account", req, nil)
}

// LivenessRequest is the body for POST /api/v1/identity/liveness.
type LivenessRequest struct {
	SubjectID string                 `json:"subjectId"`
	Image     map[string]interface{} `json:"image"`
}

func (s *IdentityService) Liveness(req LivenessRequest) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/identity/liveness", req, nil)
}

// FaceMatchRequest is the body for POST /api/v1/identity/face-match.
type FaceMatchRequest struct {
	SubjectID string                 `json:"subjectId"`
	Selfie    map[string]interface{} `json:"selfie"`
	Reference map[string]interface{} `json:"reference"`
	Threshold float64                `json:"threshold,omitempty"`
}

func (s *IdentityService) FaceMatch(req FaceMatchRequest) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/identity/face-match", req, nil)
}

// ----------------------------------------------------------------------------
// IdentityDlqService — Phase 6F.5b
// ----------------------------------------------------------------------------

// IdentityDlqService surfaces the dead-letter queue for terminal verification
// failures, plus the replay endpoint that re-runs them through the live chain.
type IdentityDlqService struct {
	client *Client
}

// ListDlqParams filters the DLQ list endpoint.
type ListDlqParams struct {
	UnreplayedOnly bool
	SubjectID      string
	Kind           string
	CountryCode    string
	Limit          int
	Offset         int
}

func (s *IdentityDlqService) List(p ListDlqParams) (map[string]interface{}, error) {
	q := map[string]string{}
	if p.UnreplayedOnly {
		q["unreplayedOnly"] = "true"
	}
	if p.SubjectID != "" {
		q["subjectId"] = p.SubjectID
	}
	if p.Kind != "" {
		q["kind"] = p.Kind
	}
	if p.CountryCode != "" {
		q["countryCode"] = p.CountryCode
	}
	if p.Limit > 0 {
		q["limit"] = fmt.Sprintf("%d", p.Limit)
	}
	if p.Offset > 0 {
		q["offset"] = fmt.Sprintf("%d", p.Offset)
	}
	return s.client.doRequest("GET", "/api/v1/identity/dlq", nil, q)
}

func (s *IdentityDlqService) Stats() (map[string]interface{}, error) {
	return s.client.doRequest("GET", "/api/v1/identity/dlq/stats", nil, nil)
}

func (s *IdentityDlqService) Get(id string) (map[string]interface{}, error) {
	return s.client.doRequest("GET", "/api/v1/identity/dlq/"+id, nil, nil)
}

// Replay re-runs a DLQ row through the live provider chain. Idempotent —
// re-replaying an already-replayed row returns the existing verification id
// without re-billing.
func (s *IdentityDlqService) Replay(id string) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/identity/dlq/"+id+"/replay", map[string]interface{}{}, nil)
}
