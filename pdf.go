// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-09-21):
// - Added Rasterize — POST /api/v1/pdf/rasterize, converts one page of a
//   base64-encoded PDF to a JPEG/PNG image. Added alongside the Python SDK
//   method and docs for the caldera-platformxe NO-DRIFT policy (route added
//   in MR !211). Requires scope pdf:rasterize (200 req/hr/key). Uses typed
//   RasterizePdfInput/RasterizePdfResult (types.go) rather than the
//   map[string]interface{} form OfferLetter/PropertyFlyer use — see
//   types.go's CHANGELOG for why.
// =============================================================================

package platformxe

// PdfService handles PDF generation API calls.
type PdfService struct {
	client *Client
}

// GetProcessor returns the PDF processor configuration.
func (s *PdfService) GetProcessor() (*ProcessorConfig, error) {
	var result ProcessorConfig
	err := s.client.doRequestTyped("GET", "/api/v1/pdf/processor", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateProcessor updates the PDF processor configuration.
func (s *PdfService) UpdateProcessor(input map[string]interface{}) (*ProcessorConfig, error) {
	var result ProcessorConfig
	err := s.client.doRequestTyped("PUT", "/api/v1/pdf/processor", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// OfferLetter generates an offer letter PDF.
func (s *PdfService) OfferLetter(input map[string]interface{}) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/pdf/offer-letter", input, nil)
}

// PropertyFlyer generates a property flyer PDF.
func (s *PdfService) PropertyFlyer(input map[string]interface{}) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/pdf/property-flyer", input, nil)
}

// Rasterize converts one page of a base64-encoded PDF into a JPEG/PNG image.
// Requires the pdf:rasterize scope (200 requests/hr/key).
func (s *PdfService) Rasterize(input RasterizePdfInput) (*RasterizePdfResult, error) {
	var result RasterizePdfResult
	err := s.client.doRequestTyped("POST", "/api/v1/pdf/rasterize", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
