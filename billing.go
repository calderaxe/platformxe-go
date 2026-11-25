// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-20, v1.14.0 pending, fidelity W3):
// - Initial: BillingService.ListInvoices / GetInvoice for GET
//   /api/v1/billing/invoices and /{id} (scope billing:read). The detail path
//   keeps its {id} placeholder (matched by openapi-route-coverage.test.ts)
//   and is filled, path-escaped, at call time.
// - W3 review: GetInvoice refuses an empty id, "." and ".." before sending
//   (path escaping leaves dots as they are, and a dot segment resolves to the
//   list route or another one).
// =============================================================================

package platformxe

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

const (
	billingInvoicesPath    = "/api/v1/billing/invoices"
	billingInvoiceByIDPath = "/api/v1/billing/invoices/{id}"
)

// BillingService reads the invoices of the key's own organisation (scope
// billing:read). Money fields are integer strings in kobo. A platform key
// with no organisation is answered 403 FORBIDDEN.
type BillingService struct {
	client *Client
}

// billingInvoicePath fills the detail path with id, path-escaped (a "/" in
// id cannot reach another route).
func billingInvoicePath(id string) string {
	return strings.Replace(billingInvoiceByIDPath, "{id}", url.PathEscape(id), 1)
}

// ListInvoices lists invoices, newest first (createdAt then id descending).
// DRAFT invoices are listed too. Errors: ErrCodeValidationError (400, unknown
// Status), ErrCodeForbidden (403), ErrCodeInternalError (500).
func (s *BillingService) ListInvoices(params ListInvoicesParams) (*BillingInvoiceList, error) {
	query := map[string]string{}
	if params.Page > 0 {
		query["page"] = strconv.Itoa(params.Page)
	}
	if params.Limit > 0 {
		query["limit"] = strconv.Itoa(params.Limit)
	}
	if params.Status != "" {
		query["status"] = params.Status
	}
	if len(query) == 0 {
		query = nil
	}
	var result BillingInvoiceList
	if err := s.client.doRequestTyped("GET", billingInvoicesPath, nil, query, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// errInvoiceID is returned, before anything is sent, for an empty id, "."
// or "..".
var errInvoiceID = errors.New("platformxe: Billing.GetInvoice needs an invoice id (not empty, \".\" or \"..\")")

// GetInvoice returns one invoice and its line items. Errors:
// ErrCodeForbidden (403), ErrCodeNotFound (404: unknown id, or another
// organisation's invoice), ErrCodeInternalError (500). An empty id, "." or
// ".." is refused before sending (a plain error, not an *APIError).
func (s *BillingService) GetInvoice(id string) (*BillingInvoiceDetail, error) {
	if strings.TrimSpace(id) == "" || id == "." || id == ".." {
		return nil, errInvoiceID
	}
	var result BillingInvoiceDetail
	if err := s.client.doRequestTyped("GET", billingInvoicePath(id), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
