// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-20, v1.14.0 pending, fidelity W3):
// - Added GenerateBatchItems(QrBatchInput, ...*QrBatchCallOptions) for POST
//   /api/v1/qr/batch (scope storage:upload): each item stored with a signed
//   URL, optional zip, optional x-idempotency-key. GenerateBatch(map) is kept
//   (same route, same answer) and its doc corrected.
// - W3 review: GenerateBatchItems' doc: a sandbox key counts one write per
//   item plus the zip; Zip is nil also when the zip could not be built or
//   stored; an all-failed answer is never stored for idempotency, a partial
//   one is (re-send its failed items under a new key); 409 means the first
//   request is still running.
// =============================================================================

package platformxe

const qrBatchPath = "/api/v1/qr/batch"

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

// GenerateBatchItems generates and stores QR codes in one call (POST
// /api/v1/qr/batch). It requires the storage:upload scope (120 requests/hr
// per key, plus the key budget). For a pxk_dev_ sandbox key it counts one of
// the 50 daily storage:upload writes per item, plus one for the zip; a batch
// that does not fit is refused whole (429 SANDBOX_QUOTA_EXCEEDED).
//
// The answer holds one QrBatchItemResult per item, in request order: a
// failing item (ErrCodeValidationError or ErrCodeStorageFailed) never fails
// the call. Zip is set when input.Zip was true, something was generated and
// the zip could be built and stored (nil otherwise; the images are stored
// either way). opts may be omitted.
//
// Idempotency: an answer in which nothing was generated is never stored, so
// a retry with the same key runs the batch again. A partial batch is stored:
// re-send its failed items under a NEW key.
//
// Errors (APIError.Code): ErrCodeValidationError (400), ErrCodeServiceDisabled
// (403), ErrCodeConflict (409: the first request with this idempotency key is
// still running; call again with the same key after about 30 s),
// ErrCodeRateLimited (429; SANDBOX_QUOTA_EXCEEDED for a sandbox key),
// ErrCodeStorageUnavailable (503), ErrCodeInternalError (500).
func (s *QrService) GenerateBatchItems(input QrBatchInput, opts ...*QrBatchCallOptions) (*BatchQRResult, error) {
	key := ""
	if len(opts) > 0 && opts[0] != nil {
		key = opts[0].IdempotencyKey
	}
	return s.postBatch(input, key)
}

// GenerateBatch posts input as the body of POST /api/v1/qr/batch, unchanged:
// {"items": [...], "format": "png" | "svg", "zip": bool, "urlTtlSeconds": n},
// with the item fields as QrBatchItem names them. The answer is the same as
// GenerateBatchItems'. Before v1.14.0 the route did not exist, so every call
// was a 404. Prefer the typed GenerateBatchItems; this form sends no
// idempotency key.
func (s *QrService) GenerateBatch(input map[string]interface{}) (*BatchQRResult, error) {
	return s.postBatch(input, "")
}

func (s *QrService) postBatch(input interface{}, idempotencyKey string) (*BatchQRResult, error) {
	headers, call := generationCall(idempotencyKey)
	var result BatchQRResult
	if err := s.client.doRequestTypedOpts("POST", qrBatchPath, input, nil, headers, call, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
