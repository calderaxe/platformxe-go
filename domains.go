// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

import "fmt"

// DomainsService handles sending domain management.
type DomainsService struct {
	client *Client
}

// List returns all registered sending domains.
func (s *DomainsService) List() (*DomainListResponse, error) {
	var result DomainListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/domains", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Add registers a new sending domain.
func (s *DomainsService) Add(domain string) (*SendingDomain, error) {
	var result SendingDomain
	err := s.client.doRequestTyped("POST", "/api/v1/domains", map[string]interface{}{"domain": domain}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a domain by ID.
func (s *DomainsService) Get(id string) (*SendingDomain, error) {
	var result SendingDomain
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/domains/%s", id), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a sending domain.
func (s *DomainsService) Delete(id string) (*DeleteResult, error) {
	var result DeleteResult
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/domains/%s", id), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Verify triggers DNS verification for a domain.
func (s *DomainsService) Verify(id string) (*DomainVerifyResult, error) {
	var result DomainVerifyResult
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/domains/%s/verify", id), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
