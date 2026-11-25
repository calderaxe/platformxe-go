// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-20, v1.14.0 pending, fidelity W3): Initial. Pins what
// Billing.ListInvoices / GetInvoice send (GET /api/v1/billing/invoices with
// its query; /{id} path-escaped) and how the answers decode (kobo strings).
// W3 review: an empty id, "." and ".." are refused before anything is sent.
// =============================================================================

package platformxe

import (
	"errors"
	"net/http"
	"testing"
)

var billingInvoiceJSON = map[string]interface{}{
	"id": "inv_1", "status": "PAID", "currency": "NGN",
	"periodStart": "2026-10-01T00:00:00.000Z", "periodEnd": "2026-10-31T23:59:59.999Z",
	"subtotal": "9900000", "tax": "742500", "total": "10642500",
	"paidAt": "2026-11-02T08:15:00.000Z", "paidVia": "FLUTTERWAVE", "pdfUrl": nil,
	"createdAt": "2026-11-01T00:05:00.000Z", "updatedAt": "2026-11-02T08:15:00.000Z",
}

func TestBillingListInvoicesSendsQuery(t *testing.T) {
	srv := newTestServer(t, func(r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/billing/invoices" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("page") != "2" || q.Get("limit") != "50" || q.Get("status") != "PAID" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
	}, map[string]interface{}{
		"invoices":   []interface{}{billingInvoiceJSON},
		"pagination": map[string]interface{}{"page": 2, "limit": 50, "total": 51, "totalPages": 2},
	})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Billing.ListInvoices(ListInvoicesParams{Page: 2, Limit: 50, Status: "PAID"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Invoices) != 1 || result.Pagination.TotalPages != 2 || result.Pagination.Total != 51 {
		t.Fatalf("unexpected result: %#v", result)
	}
	inv := result.Invoices[0]
	if inv.Total != "10642500" || inv.Tax != "742500" || inv.PdfURL != nil || inv.PaidVia == nil || *inv.PaidVia != "FLUTTERWAVE" {
		t.Errorf("unexpected invoice: %#v", inv)
	}
}

func TestBillingListInvoicesZeroParamsSendsNoQuery(t *testing.T) {
	srv := newTestServer(t, func(r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("no query expected, got %q", r.URL.RawQuery)
		}
	}, map[string]interface{}{"invoices": []interface{}{}, "pagination": map[string]interface{}{"page": 1, "limit": 20, "total": 0, "totalPages": 0}})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Billing.ListInvoices(ListInvoicesParams{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBillingGetInvoiceEscapesID(t *testing.T) {
	var gotEscaped string
	srv := newTestServer(t, func(r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		gotEscaped = r.URL.EscapedPath()
	}, map[string]interface{}{
		"invoice": billingInvoiceJSON,
		"lineItems": []interface{}{map[string]interface{}{
			"id": "li_1", "service": "plan", "description": "Pro plan", "quantity": 1,
			"unitPrice": "9900000", "total": "9900000",
		}},
	})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Billing.GetInvoice("inv_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotEscaped != "/api/v1/billing/invoices/inv_1" {
		t.Errorf("path = %s", gotEscaped)
	}
	if result.Invoice.ID != "inv_1" || len(result.LineItems) != 1 || result.LineItems[0].UnitPrice != "9900000" || result.LineItems[0].Quantity != 1 {
		t.Errorf("unexpected result: %#v", result)
	}

	if _, err := c.Billing.GetInvoice("a/b?c"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotEscaped != "/api/v1/billing/invoices/a%2Fb%3Fc" {
		t.Errorf("id not path-escaped: %s", gotEscaped)
	}
}

func TestBillingGetInvoiceNotFound(t *testing.T) {
	srv, _ := cannedServer(t, cannedResponse{404, `{"success":false,"error":{"code":"NOT_FOUND","message":"Invoice not found"}}`})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	_, err := c.Billing.GetInvoice("inv_other_org")
	if apiErr := asAPIError(t, err); apiErr.Code != ErrCodeNotFound || apiErr.StatusCode != 404 {
		t.Errorf("unexpected APIError: %#v", apiErr)
	}
}

func TestBillingGetInvoiceRefusesEmptyAndDotIDsWithoutSending(t *testing.T) {
	sent := 0
	srv := newTestServer(t, func(r *http.Request) { sent++ }, map[string]interface{}{})
	defer srv.Close()
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	for _, id := range []string{"", "   ", ".", ".."} {
		result, err := c.Billing.GetInvoice(id)
		if err == nil || result != nil {
			t.Errorf("GetInvoice(%q) = %#v, %v; want a refusal", id, result, err)
		}
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			t.Errorf("GetInvoice(%q) refusal is an APIError: %#v", id, apiErr)
		}
	}
	if sent != 0 {
		t.Errorf("%d requests sent, want 0", sent)
	}
}

func TestBillingGetInvoiceKeepsDotsInsideAnID(t *testing.T) {
	var gotEscaped string
	srv := newTestServer(t, func(r *http.Request) { gotEscaped = r.URL.EscapedPath() }, map[string]interface{}{
		"invoice": billingInvoiceJSON, "lineItems": []interface{}{},
	})
	defer srv.Close()
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Billing.GetInvoice("inv.1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotEscaped != "/api/v1/billing/invoices/inv.1" {
		t.Errorf("path = %s", gotEscaped)
	}
}
