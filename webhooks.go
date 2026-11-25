// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-10-20, pending release — webhooks v2, NO-DRIFT train):
// - Added Dispatch (POST /api/v1/webhooks/dispatch, scope webhooks:dispatch,
//   idempotent per endpoint by EventID, optional x-idempotency-key),
//   Deliveries (GET /api/v1/webhooks/{id}/deliveries) and Redeliver (POST
//   /api/v1/webhooks/deliveries/{deliveryId}/redeliver). List takes
//   ListWebhooksOptions (RecipientRef filter, Limit, Page); Create and
//   RotateSecret return *WebhookCreated (the signing secret is returned ONCE).
// - Fixed drift: RotateSecret called /rotate-secret (the route is /rotate) and
//   the Webhook / WebhookListResponse / WebhookTestResult types did not match
//   the API (see types.go).
// - The path templates keep their {id} / {deliveryId} placeholders (the
//   route's own names, matched by openapi-route-coverage.test.ts) and are
//   filled by webhookPath / redeliverPath.
// - Review fix wave 2: List is variadic, List(opts ...*ListWebhooksOptions),
//   so the v1 call List() still compiles (WebhookCreated embeds Webhook for
//   the same reason, see types.go); Dispatch sends a nil Data as {} (the API
//   refuses null and an absent key alike); doc comments carry the event-name
//   rule, the FORBIDDEN / NOT_FOUND split and the redelivery 409.
// =============================================================================

package platformxe

import (
	"net/url"
	"strconv"
	"strings"
)

const (
	webhooksPath          = "/api/v1/webhooks"
	webhookByIDPath       = "/api/v1/webhooks/{id}"
	webhookRotatePath     = "/api/v1/webhooks/{id}/rotate"
	webhookTestPath       = "/api/v1/webhooks/{id}/test"
	webhookDeliveriesPath = "/api/v1/webhooks/{id}/deliveries"
	webhookDispatchPath   = "/api/v1/webhooks/dispatch"
	webhookRedeliverPath  = "/api/v1/webhooks/deliveries/{deliveryId}/redeliver"
)

func webhookPath(template, id string) string {
	return strings.Replace(template, "{id}", url.PathEscape(id), 1)
}

func redeliverPath(deliveryID string) string {
	return strings.Replace(webhookRedeliverPath, "{deliveryId}", url.PathEscape(deliveryID), 1)
}

// WebhooksService handles webhook endpoint management (scope webhooks:manage)
// and recipient dispatch (scope webhooks:dispatch).
type WebhooksService struct {
	client *Client
}

// ListWebhooksOptions filters and pages List. A nil value lists the first
// page of the untargeted endpoints.
type ListWebhooksOptions struct {
	// RecipientRef lists that recipient's v2 endpoints instead of the
	// untargeted ones.
	RecipientRef string
	Limit        int
	Page         int
}

func pageParams(recipientRef string, limit, page int) map[string]string {
	params := map[string]string{}
	if recipientRef != "" {
		params["recipientRef"] = recipientRef
	}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	if page > 0 {
		params["page"] = strconv.Itoa(page)
	}
	if len(params) == 0 {
		return nil
	}
	return params
}

// List returns the calling service's webhook endpoints. Without
// RecipientRef only the untargeted endpoints are listed. opts is optional
// (List() lists the first page of the untargeted endpoints); only the first
// value is read and it may be nil.
func (s *WebhooksService) List(opts ...*ListWebhooksOptions) (*WebhookListResponse, error) {
	var params map[string]string
	if len(opts) > 0 && opts[0] != nil {
		params = pageParams(opts[0].RecipientRef, opts[0].Limit, opts[0].Page)
	}
	var result WebhookListResponse
	if err := s.client.doRequestTyped("GET", webhooksPath, nil, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a webhook endpoint. input carries name, url, events and, for
// a v2 recipient endpoint, engineVersion (2) and recipientRef. The result
// carries the signing secret ONCE.
func (s *WebhooksService) Create(input map[string]interface{}) (*WebhookCreated, error) {
	var result WebhookCreated
	if err := s.client.doRequestTyped("POST", webhooksPath, input, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a webhook endpoint by ID. It never carries a secret. Another
// service's webhook is FORBIDDEN; a missing one NOT_FOUND.
func (s *WebhooksService) Get(id string) (*Webhook, error) {
	var result Webhook
	if err := s.client.doRequestTyped("GET", webhookPath(webhookByIDPath, id), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates a webhook endpoint (name, url, events, status).
// engineVersion / recipientRef are accepted only when unchanged; a changed
// value is VALIDATION_ERROR.
func (s *WebhooksService) Update(id string, input map[string]interface{}) (*Webhook, error) {
	var result Webhook
	if err := s.client.doRequestTyped("PATCH", webhookPath(webhookByIDPath, id), input, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a webhook endpoint.
func (s *WebhooksService) Delete(id string) (*DeleteResult, error) {
	var result DeleteResult
	if err := s.client.doRequestTyped("DELETE", webhookPath(webhookByIDPath, id), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RotateSecret rotates the signing secret. The new secret is returned ONCE.
func (s *WebhooksService) RotateSecret(id string) (*WebhookCreated, error) {
	var result WebhookCreated
	if err := s.client.doRequestTyped("POST", webhookPath(webhookRotatePath, id), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Test sends a test event to a webhook endpoint.
func (s *WebhooksService) Test(id string) (*WebhookTestResult, error) {
	var result WebhookTestResult
	if err := s.client.doRequestTyped("POST", webhookPath(webhookTestPath, id), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DispatchWebhookOptions are the per-call options of Dispatch.
type DispatchWebhookOptions struct {
	// IdempotencyKey is sent as x-idempotency-key: a repeat of the whole
	// request answers the stored response.
	IdempotencyKey string
}

// Dispatch sends one event to ONE recipient's v2 endpoints that subscribe to
// it and are not PAUSED. It is idempotent per endpoint by input.EventID. A nil
// input.Data is sent as {} (input is a copy; the caller's value is not
// changed). Besides the closed webhook codes the API can answer CONFLICT
// (409), RATE_LIMITED (429) and IDEMPOTENCY_UNAVAILABLE (503), all retryable.
// opts may be nil.
func (s *WebhooksService) Dispatch(input DispatchWebhookInput, opts *DispatchWebhookOptions) (*DispatchWebhookResult, error) {
	if input.Data == nil {
		input.Data = map[string]interface{}{}
	}
	var headers map[string]string
	if opts != nil && opts.IdempotencyKey != "" {
		headers = map[string]string{"x-idempotency-key": opts.IdempotencyKey}
	}
	var result DispatchWebhookResult
	if err := s.client.doRequestTypedWithHeaders("POST", webhookDispatchPath, input, nil, headers, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Deliveries returns the delivery log of one endpoint, newest first. Rows
// never carry the payload; an engine-1 endpoint's rows carry no snippet.
// limit and page of 0 use the server defaults.
func (s *WebhooksService) Deliveries(id string, limit, page int) (*WebhookDeliveriesResponse, error) {
	var result WebhookDeliveriesResponse
	if err := s.client.doRequestTyped("GET", webhookPath(webhookDeliveriesPath, id), nil, pageParams("", limit, page), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Redeliver re-sends one v2 delivery once, with the same delivery id. A
// delivery being processed or about to be retried is DELIVERY_IN_PROGRESS
// (409); a PAUSED endpoint may be redelivered to and stays PAUSED.
func (s *WebhooksService) Redeliver(deliveryID string) (*RedeliverWebhookResult, error) {
	var result RedeliverWebhookResult
	if err := s.client.doRequestTyped("POST", redeliverPath(deliveryID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
