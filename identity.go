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

// IdentityMatchAgainst is the biographical data a verify call matches.
type IdentityMatchAgainst struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	// DateOfBirth is YYYY-MM-DD. A PASSPORT check needs it with Supporting.NIN.
	DateOfBirth string `json:"dateOfBirth,omitempty"`
}

// IdentityVerifySupporting carries extra identifiers a provider needs.
type IdentityVerifySupporting struct {
	// NIN is the holder's 11-digit NIN, accepted only for type PASSPORT
	// (any other type is a 400). With MatchAgainst.DateOfBirth it lets the
	// Nigerian passport check run on the primary provider.
	NIN string `json:"nin,omitempty"`
}

// IdentityVerifyRequest is the body for POST /api/v1/identity/verify
// (@caldera/platformxe-types IdentityVerifyRequest, 3.9.0).
type IdentityVerifyRequest struct {
	Type         string                    `json:"type"`
	Value        string                    `json:"value"`
	MatchAgainst IdentityMatchAgainst      `json:"matchAgainst"`
	Supporting   *IdentityVerifySupporting `json:"supporting,omitempty"`
}

// Verify verifies an identifier with name matching.
//
// v1.8.0: it sends the route's real body ({ type, value, matchAgainst });
// before, it sent identifierType / firstName / lastName at the top level,
// which the route answers 400. Use VerifyRequest for a date of birth or a
// PASSPORT check's supporting NIN.
func (s *IdentityService) Verify(identifierType, value, firstName, lastName string) (*IdentityVerifyResult, error) {
	return s.VerifyRequest(IdentityVerifyRequest{
		Type:         identifierType,
		Value:        value,
		MatchAgainst: IdentityMatchAgainst{FirstName: firstName, LastName: lastName},
	})
}

// VerifyRequest posts a full IdentityVerifyRequest (v1.8.0).
func (s *IdentityService) VerifyRequest(req IdentityVerifyRequest) (*IdentityVerifyResult, error) {
	var result IdentityVerifyResult
	err := s.client.doRequestTyped("POST", "/api/v1/identity/verify", req, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Lookup is a raw registry lookup of one identifier (v1.10.0). It needs the
// identity:lookup scope, the Detection Pack and its terms, and a purpose and
// consent reference (1 to 200 characters each), which PlatformXe logs; each
// call is metered. The answer never carries a photo.
func (s *IdentityService) Lookup(identifierType, value, purpose, consentReference string) (*IdentityLookupResult, error) {
	var result IdentityLookupResult
	err := s.client.doRequestTyped("GET", "/api/v1/identity/lookup", nil, map[string]string{
		"type": identifierType, "value": value,
		"purpose": purpose, "consentReference": consentReference,
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
// Match must carry firstName, lastName and dateOfBirth (YYYY-MM-DD, required
// since v1.10.0); the status is "verified" only when all three match, and the
// answer carries "checks" and "reason".
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

// VerifyBusinessRequest is the body for POST /api/v1/identity/verify-business
// (a company registration check, scope business:verify).
type VerifyBusinessRequest struct {
	SubjectID string `json:"subjectId"`
	// Country is "NG" (the only country for now); VerifyBusiness sets it when empty.
	Country string `json:"country"`
	// RegistrationType is "RC" (company), "BN" (business name) or "IT" (incorporated trustees).
	RegistrationType   string `json:"registrationType"`
	RegistrationNumber string `json:"registrationNumber"`
	// TIN is optional.
	TIN          string  `json:"tin,omitempty"`
	ExpectedName string  `json:"expectedName"`
	Threshold    float64 `json:"threshold,omitempty"`
}

// VerifyBusinessResult is the answer of VerifyBusiness. Status is "verified",
// "failed" or "provider_down"; Reason is "name_mismatch", "not_found",
// "inactive" or "tin_mismatch" (nil when verified or provider_down);
// RegistrationStatus is ACTIVE, INACTIVE or UNKNOWN; TINStatus is MATCHED,
// NOT_MATCHED or NOT_CHECKED.
type VerifyBusinessResult struct {
	Status             string   `json:"status"`
	Reason             *string  `json:"reason"`
	MatchScore         *float64 `json:"matchScore"`
	RegisteredName     *string  `json:"registeredName"`
	RegistrationStatus *string  `json:"registrationStatus"`
	RegisteredAt       *string  `json:"registeredAt"`
	TINStatus          *string  `json:"tinStatus"`
	Provider           string   `json:"provider"`
	VerificationID     *string  `json:"verificationId"`
}

// VerifyBusiness checks a business registration and optional TIN.
func (s *IdentityService) VerifyBusiness(req VerifyBusinessRequest) (*VerifyBusinessResult, error) {
	return s.VerifyBusinessWithIdempotencyKey(req, "")
}

// VerifyBusinessWithIdempotencyKey sends x-idempotency-key when key is not
// empty, so a retried check replays the stored answer instead of a second lookup.
func (s *IdentityService) VerifyBusinessWithIdempotencyKey(req VerifyBusinessRequest, key string) (*VerifyBusinessResult, error) {
	if req.Country == "" {
		req.Country = "NG"
	}
	var headers map[string]string
	if key != "" {
		headers = map[string]string{"x-idempotency-key": key}
	}
	var result VerifyBusinessResult
	if err := s.client.doRequestTypedWithHeaders("POST", "/api/v1/identity/verify-business", req, nil, headers, &result); err != nil {
		return nil, err
	}
	return &result, nil
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
// Reference is an image ({"url": ...} or {"base64": ..., "contentType": ...})
// or, since v1.10.0, {"verificationId": "iv_..."}: an earlier verified NIN
// check for the same subject, whose registry photo PlatformXe holds.
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
