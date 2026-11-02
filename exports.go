// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

import "fmt"

// ExportsService handles data export jobs.
type ExportsService struct {
	client *Client
}

// GetProcessor returns the exports processor configuration.
func (s *ExportsService) GetProcessor() (*ProcessorConfig, error) {
	var result ProcessorConfig
	err := s.client.doRequestTyped("GET", "/api/v1/exports/processor", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateProcessor updates the exports processor configuration.
func (s *ExportsService) UpdateProcessor(input map[string]interface{}) (*ProcessorConfig, error) {
	var result ProcessorConfig
	err := s.client.doRequestTyped("PUT", "/api/v1/exports/processor", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new data export job.
func (s *ExportsService) Create(input map[string]interface{}) (*ExportResult, error) {
	var result ExportResult
	err := s.client.doRequestTyped("POST", "/api/v1/exports", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns the status and download URL of an export job.
func (s *ExportsService) Get(exportId string) (*ExportResult, error) {
	var result ExportResult
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/exports/%s", exportId), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
