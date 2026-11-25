// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

// OcrService handles OCR identity document verification.
type OcrService struct {
	client *Client
}

// GetProcessor returns the OCR processor configuration.
func (s *OcrService) GetProcessor() (*ProcessorConfig, error) {
	var result ProcessorConfig
	err := s.client.doRequestTyped("GET", "/api/v1/ocr/processor", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateProcessor updates the OCR processor configuration.
func (s *OcrService) UpdateProcessor(input map[string]interface{}) (*ProcessorConfig, error) {
	var result ProcessorConfig
	err := s.client.doRequestTyped("PUT", "/api/v1/ocr/processor", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// VerifyIdentity submits an identity document for OCR verification.
func (s *OcrService) VerifyIdentity(input map[string]interface{}) (*IdentityDocVerifyResult, error) {
	var result IdentityDocVerifyResult
	err := s.client.doRequestTyped("POST", "/api/v1/ocr/verify-identity", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
