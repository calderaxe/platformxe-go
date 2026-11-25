// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-20, v1.14.0 pending, fidelity W3): Initial. Pins what
// Pdf.Generate sends (POST /api/v1/pdf/generate: body, x-idempotency-key)
// and decodes (stored and base64 answers), the local html/templateId check,
// and the retry rule (a 5xx is re-sent only with an idempotency key).
// =============================================================================

package platformxe

import (
	"net/http"
	"testing"
)

func TestPdfGenerateSendsHTMLAndOptions(t *testing.T) {
	bg := true
	srv := newTestServer(t, func(r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/pdf/generate" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("x-idempotency-key"); got != "" {
			t.Errorf("no idempotency key expected, got %q", got)
		}
		body := decodeJSONBody(t, r)
		if body["html"] != "<h1>Hi</h1>" {
			t.Errorf("html = %v", body["html"])
		}
		for _, k := range []string{"templateId", "data"} {
			if _, ok := body[k]; ok {
				t.Errorf("%s must not be sent", k)
			}
		}
		opts, _ := body["options"].(map[string]interface{})
		if opts["size"] != "A5" || opts["printBackground"] != true || opts["remoteImages"] != false {
			t.Errorf("options = %#v", opts)
		}
		if opts["urlTtlSeconds"] != float64(3600) || opts["headerHtml"] != "<span class=\"title\"></span>" {
			t.Errorf("options = %#v", opts)
		}
		margins, _ := opts["margins"].(map[string]interface{})
		if margins["top"] != "10mm" {
			t.Errorf("margins = %#v", margins)
		}
		if _, ok := opts["returnBase64"]; ok {
			t.Error("returnBase64 false must not be sent")
		}
	}, map[string]interface{}{
		"fileId": "f_1", "pdfUrl": "https://acct.blob.core.windows.net/x.pdf?sp=r&sig=s",
		"expiresAt": "2026-11-21T14:30:00.000Z", "size": 45230, "pages": 2,
		"generatedAt": "2026-11-20T14:30:00.000Z",
	})
	defer srv.Close()

	noRemote := false
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Pdf.Generate(GeneratePdfInput{
		HTML: "<h1>Hi</h1>",
		Options: &GeneratePdfOptions{
			Size: "A5", PrintBackground: &bg, RemoteImages: &noRemote, URLTTLSeconds: 3600,
			HeaderHTML: "<span class=\"title\"></span>",
			Margins:    &GeneratePdfMargins{Top: "10mm"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FileID != "f_1" || result.Size != 45230 || result.Pages != 2 || result.ExpiresAt == "" || result.PdfURL == "" {
		t.Errorf("unexpected result: %#v", result)
	}
}

func TestPdfGenerateSendsTemplateDataAndIdempotencyKey(t *testing.T) {
	srv := newTestServer(t, func(r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/pdf/generate" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("x-idempotency-key"); got != "idem-1" {
			t.Errorf("x-idempotency-key = %q", got)
		}
		body := decodeJSONBody(t, r)
		if body["templateId"] != "tpl_abc" {
			t.Errorf("templateId = %v", body["templateId"])
		}
		if _, ok := body["html"]; ok {
			t.Error("html must not be sent")
		}
		data, _ := body["data"].(map[string]interface{})
		if data["customerName"] != "Ada" {
			t.Errorf("data = %#v", data)
		}
		opts, _ := body["options"].(map[string]interface{})
		if opts["returnBase64"] != true {
			t.Errorf("options = %#v", opts)
		}
	}, map[string]interface{}{
		"content": "JVBERi0xLjcK", "contentType": "application/pdf", "size": 9, "pages": 1,
		"generatedAt": "2026-11-20T14:30:00.000Z",
	})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Pdf.Generate(GeneratePdfInput{
		TemplateID: "tpl_abc",
		Data:       map[string]interface{}{"customerName": "Ada"},
		Options:    &GeneratePdfOptions{ReturnBase64: true},
	}, &GeneratePdfCallOptions{IdempotencyKey: "idem-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "JVBERi0xLjcK" || result.ContentType != "application/pdf" || result.FileID != "" {
		t.Errorf("unexpected result: %#v", result)
	}
}

func TestPdfGenerateRefusesHTMLTemplateMisuseWithoutSending(t *testing.T) {
	srv, attempts := cannedServer(t, cannedResponse{200, `{"success":true,"data":{}}`})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	cases := []GeneratePdfInput{
		{},
		{HTML: "<p/>", TemplateID: "tpl"},
		{HTML: "<p/>", Data: map[string]interface{}{"a": "b"}},
	}
	for i, in := range cases {
		if _, err := c.Pdf.Generate(in); err == nil {
			t.Errorf("case %d: expected an error", i)
		}
	}
	if *attempts != 0 {
		t.Errorf("nothing should be sent, got %d requests", *attempts)
	}
}

func TestPdfGenerateErrorCodeAndRetryRule(t *testing.T) {
	notFound := `{"success":false,"error":{"code":"TEMPLATE_NOT_FOUND","message":"Template not found"}}`
	srv, attempts := cannedServer(t, cannedResponse{404, notFound})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL, Retries: 2})
	_, err := c.Pdf.Generate(GeneratePdfInput{TemplateID: "tpl_x"})
	if apiErr := asAPIError(t, err); apiErr.Code != ErrCodeTemplateNotFound || apiErr.StatusCode != 404 {
		t.Errorf("unexpected APIError: %#v", apiErr)
	}
	if *attempts != 1 {
		t.Errorf("a 404 is final, got %d attempts", *attempts)
	}

	// A 503 without an idempotency key is not re-sent (a second render is a
	// second stored file); with one it is.
	unavailable := `{"success":false,"error":{"code":"RENDERER_UNAVAILABLE","message":"PDF renderer unavailable"}}`
	srv2, attempts2 := cannedServer(t, cannedResponse{503, unavailable})
	c2 := NewClient(ClientConfig{APIKey: "test", BaseURL: srv2.URL, Retries: 1})
	_, err = c2.Pdf.Generate(GeneratePdfInput{HTML: "<p/>"})
	if apiErr := asAPIError(t, err); apiErr.Code != ErrCodeRendererUnavailable {
		t.Errorf("unexpected APIError: %#v", apiErr)
	}
	if *attempts2 != 1 {
		t.Errorf("without a key a 503 must not be re-sent, got %d attempts", *attempts2)
	}
	_, _ = c2.Pdf.Generate(GeneratePdfInput{HTML: "<p/>"}, &GeneratePdfCallOptions{IdempotencyKey: "k"})
	if *attempts2 != 3 {
		t.Errorf("with a key a 503 is retried once, got %d total attempts", *attempts2)
	}
}
