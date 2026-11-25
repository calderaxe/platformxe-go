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
//
// CHANGELOG (2026-11-20, v1.14.0 pending, fidelity W3):
// - Added Generate — POST /api/v1/pdf/generate (scope pdf:generate, 100
//   requests/hr per key): HTML or a stored template to a stored PDF with a
//   signed URL, or base64. Optional GeneratePdfCallOptions.IdempotencyKey is
//   sent as x-idempotency-key; generationCall is shared with
//   Qr.GenerateBatchItems.
// - W3 review: Generate's doc says what 408 and 409 mean for a retry, that a
//   base64 answer re-renders and meters again, and that deleting the stored
//   file needs storage:upload.
// =============================================================================

package platformxe

import (
	"errors"
	"time"
)

const pdfGeneratePath = "/api/v1/pdf/generate"

// PdfService handles PDF generation API calls.
type PdfService struct {
	client *Client
}

// generationCall is the header and call options of the rendering POSTs
// (Pdf.Generate, Qr.GenerateBatchItems). The routes run up to 60 s, longer
// than the client default. Without an idempotency key only a 429 (refused
// before any handler ran) is re-sent, since a second render is a second
// stored file; with one, a retry answers the stored result. A transport
// failure is always an error, never the FailOpen placeholder (it would read
// as an empty result).
func generationCall(idempotencyKey string) (map[string]string, callOptions) {
	call := callOptions{timeout: 65 * time.Second, noFailOpen: true}
	if idempotencyKey == "" {
		call.rateLimitRetriesOnly = true
		return nil, call
	}
	return map[string]string{"x-idempotency-key": idempotencyKey}, call
}

// Generate renders input.HTML, or the stored template input.TemplateID with
// input.Data, to a PDF (POST /api/v1/pdf/generate). It requires the
// pdf:generate scope (100 requests/hr per key, plus the key budget).
//
// Stored mode (the default) answers FileID, PdfURL (a read-only signed URL)
// and ExpiresAt; Options.ReturnBase64 answers Content (at most 3 MiB of PDF).
// Both carry Size, Pages and GeneratedAt. opts may be omitted. The stored file
// stays until deleted with DELETE /api/v1/storage/media/files/{fileId}, which
// needs storage:upload (a key holding only pdf:generate cannot delete it).
//
// Idempotency: with a key, a retry answers the stored result. A base64 answer
// and a 408 are never stored, so their retry renders (and meters) again.
//
// Errors (APIError.Code): ErrCodeValidationError (400), ErrCodeTemplateNotFound
// (404), ErrCodeRenderTimeout (408, never stored for idempotency: retry with
// the same key or a new one), ErrCodeConflict (409: the first request with
// this idempotency key is still running; call again with the same key after
// about 30 s), ErrCodePayloadTooLarge / ErrCodeOutputTooLarge (413),
// ErrCodeStorageUnavailable / ErrCodeRendererUnavailable (503),
// ErrCodeInternalError (500). Neither or both of HTML and TemplateID, or Data
// without TemplateID, is refused before sending.
func (s *PdfService) Generate(input GeneratePdfInput, opts ...*GeneratePdfCallOptions) (*GeneratePdfResult, error) {
	if (input.HTML == "") == (input.TemplateID == "") {
		return nil, errors.New("platformxe: Pdf.Generate needs exactly one of HTML or TemplateID")
	}
	if input.Data != nil && input.TemplateID == "" {
		return nil, errors.New("platformxe: Pdf.Generate sends Data only with TemplateID")
	}
	key := ""
	if len(opts) > 0 && opts[0] != nil {
		key = opts[0].IdempotencyKey
	}
	headers, call := generationCall(key)
	var result GeneratePdfResult
	if err := s.client.doRequestTypedOpts("POST", pdfGeneratePath, input, nil, headers, call, &result); err != nil {
		return nil, err
	}
	return &result, nil
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
