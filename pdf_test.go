// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-09-21): Initial. Pins the ACTUAL wire shape PdfService.
// Rasterize sends/receives against a real net/http/httptest server, following
// the pattern documents_test.go established (newTestServer/decodeJSONBody are
// defined there and reused here — same package). Written for the
// caldera-platformxe NO-DRIFT policy alongside the route (MR !211), the
// Python SDK method, and docs.
// =============================================================================

package platformxe

import (
	"net/http"
	"testing"
)

func TestRasterizeSendsPdfBase64AndFormat(t *testing.T) {
	srv := newTestServer(t, func(r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/pdf/rasterize" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body := decodeJSONBody(t, r)
		if body["pdfBase64"] != "JVBERi0=" {
			t.Errorf("unexpected pdfBase64: %v", body["pdfBase64"])
		}
		if body["format"] != "jpeg" {
			t.Errorf("unexpected format: %v", body["format"])
		}
		if _, hasPage := body["page"]; hasPage {
			t.Errorf("page should be omitted when zero, got %v", body["page"])
		}
		if _, hasDPI := body["dpi"]; hasDPI {
			t.Errorf("dpi should be omitted when zero, got %v", body["dpi"])
		}
	}, map[string]interface{}{
		"imageBase64": "aW1n", "contentType": "image/jpeg",
		"width": 1240, "height": 1754, "pageCount": 3,
	})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Pdf.Rasterize(RasterizePdfInput{
		PdfBase64: "JVBERi0=",
		Format:    "jpeg",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ImageBase64 != "aW1n" {
		t.Errorf("expected imageBase64 aW1n, got %q", result.ImageBase64)
	}
	if result.ContentType != "image/jpeg" {
		t.Errorf("expected contentType image/jpeg, got %q", result.ContentType)
	}
	if result.Width != 1240 || result.Height != 1754 {
		t.Errorf("unexpected dimensions: %dx%d", result.Width, result.Height)
	}
	if result.PageCount != 3 {
		t.Errorf("expected pageCount 3, got %d", result.PageCount)
	}
}

func TestRasterizeIncludesPageAndDPIWhenGiven(t *testing.T) {
	srv := newTestServer(t, func(r *http.Request) {
		body := decodeJSONBody(t, r)
		if body["page"] != float64(2) {
			t.Errorf("expected page=2, got %v", body["page"])
		}
		if body["dpi"] != float64(300) {
			t.Errorf("expected dpi=300, got %v", body["dpi"])
		}
	}, map[string]interface{}{
		"imageBase64": "x", "contentType": "image/png",
		"width": 1, "height": 1, "pageCount": 1,
	})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Pdf.Rasterize(RasterizePdfInput{
		PdfBase64: "JVBERi0=",
		Format:    "png",
		Page:      2,
		DPI:       300,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRasterizeErrorPropagatesAsAPIError(t *testing.T) {
	// 413 PAYLOAD_TOO_LARGE (oversized PDF / too many pages) must surface as
	// *APIError like every other endpoint's error path (client.go's
	// httpStatusError), not be swallowed or returned as a nil-error zero
	// value. Uses the cannedServer/asAPIError helpers from client_test.go
	// (same package).
	srv, _ := cannedServer(t, cannedResponse{413, `{"success":false,"error":{"code":"PAYLOAD_TOO_LARGE","message":"Decoded PDF is too large"}}`})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})

	result, err := c.Pdf.Rasterize(RasterizePdfInput{PdfBase64: "JVBERi0=", Format: "jpeg"})
	if result != nil {
		t.Errorf("expected nil result, got %#v", result)
	}
	apiErr := asAPIError(t, err)
	if apiErr.Code != "PAYLOAD_TOO_LARGE" || apiErr.StatusCode != 413 {
		t.Errorf("unexpected APIError: %#v", apiErr)
	}
}
