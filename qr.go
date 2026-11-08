// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

// QrService handles QR code generation.
type QrService struct {
	client *Client
}

// GetProcessor returns the QR processor configuration.
func (s *QrService) GetProcessor() (*ProcessorConfig, error) {
	var result ProcessorConfig
	err := s.client.doRequestTyped("GET", "/api/v1/qr/processor", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateProcessor updates the QR processor configuration.
func (s *QrService) UpdateProcessor(input map[string]interface{}) (*ProcessorConfig, error) {
	var result ProcessorConfig
	err := s.client.doRequestTyped("PUT", "/api/v1/qr/processor", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Generate generates a QR code.
func (s *QrService) Generate(input map[string]interface{}) (*QRCodeResult, error) {
	var result QRCodeResult
	err := s.client.doRequestTyped("POST", "/api/v1/qr", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GenerateBatch generates QR codes in batch.
func (s *QrService) GenerateBatch(input map[string]interface{}) (*BatchQRResult, error) {
	var result BatchQRResult
	err := s.client.doRequestTyped("POST", "/api/v1/qr/batch", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
