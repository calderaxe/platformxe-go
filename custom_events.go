// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-09-15): gofmt alignment only (SubscribeCustomEventInput,
// SubscribeCustomEventResult, ForkMarketplaceResult,
// EventFederationPushSummary) — `gofmt -l` listed this file in CI.
// 2026-11-17: the marketplace and federation sub-services moved to
// custom_events_federation.go (custom events gap review, CE-B).
// 2026-11-17: core gap review (CE-A) — Update / Publish (PATCH), DryRunEmit,
// ListPage (limit / offset, pagination), the {id} path template URL-escaped,
// health's daily count and subscription block; doc comments name the 202
// durable delivery, the idempotency rules and the RE2 pattern dialect.
// =============================================================================

package platformxe

import "strconv"

// CustomEventsService is the typed client for the tenant custom events
// surface (POST /api/v1/events/custom/*). Accessed via Client.Events.Custom.
type CustomEventsService struct {
	client *Client
	// Marketplace is the typed sub-service for the Phase 9C marketplace
	// (publish + browse + fork). Mounted at Client.Events.Custom.Marketplace.
	Marketplace *MarketplaceService
	// Federation is the typed sub-service for the Phase 9D Custom Event
	// Federation Push surface (ENTERPRISE-only). Mounted at
	// Client.Events.Custom.Federation.
	//
	// Distinct from the v1.x.x Permission Federation surface available
	// at Client.Permissions.Federation.
	Federation *EventFederationService
}

// CustomEventStatus is the lifecycle state of a registered event.
type CustomEventStatus string

const (
	CustomEventStatusDraft     CustomEventStatus = "draft"
	CustomEventStatusPublished CustomEventStatus = "published"
	CustomEventStatusArchived  CustomEventStatus = "archived"
)

// RegisterCustomEventInput is the body of POST /api/v1/events/custom.
type RegisterCustomEventInput struct {
	Namespace      string                 `json:"namespace"`
	Name           string                 `json:"name"`
	Version        string                 `json:"version"`
	Status         *CustomEventStatus     `json:"status,omitempty"`
	Description    *string                `json:"description,omitempty"`
	PayloadSchema  map[string]interface{} `json:"payloadSchema"`
	PayloadExample map[string]interface{} `json:"payloadExample,omitempty"`
}

// RegisterCustomEventResult is the response shape of register/dry-run hits.
type RegisterCustomEventResult struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organizationId"`
	CanonicalName  string  `json:"canonicalName"`
	Namespace      string  `json:"namespace"`
	Name           string  `json:"name"`
	Version        string  `json:"version"`
	Status         string  `json:"status"`
	Description    *string `json:"description"`
	RegisteredBy   string  `json:"registeredBy"`
	SchemaHash     string  `json:"schemaHash"`
	SchemaFlavour  string  `json:"schemaFlavour"`
	SemverBump     string  `json:"semverBump"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

// CustomEventSummary is the per-row entry returned by List.
type CustomEventSummary struct {
	ID            string  `json:"id"`
	Namespace     string  `json:"namespace"`
	Name          string  `json:"name"`
	Version       string  `json:"version"`
	Status        string  `json:"status"`
	Description   *string `json:"description"`
	CanonicalName string  `json:"canonicalName"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

// CustomEventListResult is the body of GET /api/v1/events/custom.
type CustomEventListResult struct {
	Events     []CustomEventSummary `json:"events"`
	Pagination struct {
		Limit   int  `json:"limit"`
		Offset  int  `json:"offset"`
		Total   int  `json:"total"`
		HasMore bool `json:"hasMore"`
	} `json:"pagination"`
}

// customEventPath is the registration route template; the id is URL-escaped.
const customEventPath = "/api/v1/events/custom/{id}"

// ListCustomEventsParams narrows and pages GET /api/v1/events/custom.
type ListCustomEventsParams struct {
	Namespace string
	Status    string
	// Limit is 1..200 (0 = the default, 100).
	Limit  int
	Offset int
}

// UpdateCustomEventInput is the body of PATCH /api/v1/events/custom/{id}:
// publish a draft and/or change the description. ClearDescription sends
// "description": null.
type UpdateCustomEventInput struct {
	Status           *CustomEventStatus
	Description      *string
	ClearDescription bool
}

// UpdateCustomEventResult is the body PATCH answers.
type UpdateCustomEventResult struct {
	ID          string  `json:"id"`
	Status      string  `json:"status"`
	Description *string `json:"description"`
	UpdatedAt   string  `json:"updatedAt"`
	PublishedAt *string `json:"publishedAt,omitempty"`
}

// DryRunEmitResult is the body of an emit-mode dry-run that would be accepted.
type DryRunEmitResult struct {
	Valid         bool   `json:"valid"`
	CanonicalName string `json:"canonicalName"`
	Version       string `json:"version"`
}

// EmitCustomEventInput is the body of POST /api/v1/events/custom/emit.
type EmitCustomEventInput struct {
	Name           string                 `json:"name"`
	Version        *string                `json:"version,omitempty"`
	Payload        map[string]interface{} `json:"payload"`
	IdempotencyKey *string                `json:"idempotencyKey,omitempty"`
}

// EmitCustomEventResult is the response shape of emit hits.
type EmitCustomEventResult struct {
	EmitID        string `json:"emitId"`
	CanonicalName string `json:"canonicalName"`
	Status        string `json:"status"`
}

// DryRunCustomEventResult is the body returned by /events/custom/dry-run.
type DryRunCustomEventResult struct {
	Valid               bool   `json:"valid"`
	SchemaHash          string `json:"schemaHash"`
	SchemaFlavour       string `json:"schemaFlavour"`
	SemverBump          string `json:"semverBump"`
	SchemaDepth         int    `json:"schemaDepth"`
	SchemaPropertyCount int    `json:"schemaPropertyCount"`
}

// ArchiveCustomEventResult is the body returned by DELETE /events/custom/[id].
type ArchiveCustomEventResult struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	ArchivedAt string `json:"archivedAt"`
}

// CustomEventDetail is the body returned by GET /events/custom/[id].
type CustomEventDetail struct {
	ID             string                 `json:"id"`
	OrganizationID string                 `json:"organizationId"`
	CanonicalName  string                 `json:"canonicalName"`
	Namespace      string                 `json:"namespace"`
	Name           string                 `json:"name"`
	Version        string                 `json:"version"`
	Status         string                 `json:"status"`
	Description    *string                `json:"description"`
	PayloadSchema  map[string]interface{} `json:"payloadSchema"`
	PayloadExample map[string]interface{} `json:"payloadExample"`
	RegisteredBy   string                 `json:"registeredBy"`
	CreatedAt      string                 `json:"createdAt"`
	UpdatedAt      string                 `json:"updatedAt"`
	ArchivedAt     *string                `json:"archivedAt"`
}

// CustomEventHealth is the body returned by GET /events/custom/health.
type CustomEventHealth struct {
	Plan          string `json:"plan"`
	Registrations struct {
		Used         int  `json:"used"`
		Limit        *int `json:"limit"`
		CreatedToday int  `json:"createdToday"`
		DailyLimit   int  `json:"dailyLimit"`
	} `json:"registrations"`
	Emits struct {
		ThisMonth    int  `json:"thisMonth"`
		MonthlyLimit *int `json:"monthlyLimit"`
	} `json:"emits"`
	Capabilities struct {
		CanRegister        bool `json:"canRegister"`
		SchemaVersioning   bool `json:"schemaVersioning"`
		MarketplacePublish bool `json:"marketplacePublish"`
		MarketplaceFork    bool `json:"marketplaceFork"`
		FederationPush     bool `json:"federationPush"`
	} `json:"capabilities"`
	RecentFailures []struct {
		ID            string `json:"id"`
		CanonicalName string `json:"canonicalName"`
		ErrorMessage  string `json:"errorMessage"`
		CreatedAt     string `json:"createdAt"`
	} `json:"recentFailures"`
	// Subscriptions is the health of the tenant's custom-event webhooks:
	// counts by status, how many the breaker isolated (FAILING), and the
	// latest deliveries that exhausted their retries.
	Subscriptions struct {
		ByStatus               map[string]int `json:"byStatus"`
		Isolated               int            `json:"isolated"`
		RecentDeliveryFailures []struct {
			SubscriptionID     string `json:"subscriptionId"`
			EventType          string `json:"eventType"`
			LastResponseStatus *int   `json:"lastResponseStatus"`
			Attempts           int    `json:"attempts"`
			CreatedAt          string `json:"createdAt"`
		} `json:"recentDeliveryFailures"`
	} `json:"subscriptions"`
}

// Register registers a new custom event (or new version). An identical
// re-register of an existing version answers 200 with the stored row;
// another schema at that version is 409 version_already_exists. Patterns run
// on RE2 (no lookaround or backreferences). On Basic a published version
// archives the event's other live versions.
func (s *CustomEventsService) Register(input RegisterCustomEventInput) (*RegisterCustomEventResult, error) {
	var result RegisterCustomEventResult
	err := s.client.doRequestTyped("POST", "/api/v1/events/custom", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// List returns the first page (up to 100) of the organisation's
// registrations; see ListPage for paging.
func (s *CustomEventsService) List(namespace, status string) (*CustomEventListResult, error) {
	return s.ListPage(ListCustomEventsParams{Namespace: namespace, Status: status})
}

// ListPage returns one page of registrations (Pagination.HasMore says if
// another follows).
func (s *CustomEventsService) ListPage(p ListCustomEventsParams) (*CustomEventListResult, error) {
	params := map[string]string{}
	if p.Namespace != "" {
		params["namespace"] = p.Namespace
	}
	if p.Status != "" {
		params["status"] = p.Status
	}
	if p.Limit > 0 {
		params["limit"] = strconv.Itoa(p.Limit)
	}
	if p.Offset > 0 {
		params["offset"] = strconv.Itoa(p.Offset)
	}
	var result CustomEventListResult
	err := s.client.doRequestTyped("GET", "/api/v1/events/custom", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns the detail for a single registration.
func (s *CustomEventsService) Get(registrationID string) (*CustomEventDetail, error) {
	var result CustomEventDetail
	err := s.client.doRequestTyped("GET", webhookPath(customEventPath, registrationID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Archive soft-deletes a registration.
func (s *CustomEventsService) Archive(registrationID string) (*ArchiveCustomEventResult, error) {
	var result ArchiveCustomEventResult
	err := s.client.doRequestTyped("DELETE", webhookPath(customEventPath, registrationID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update publishes a draft and/or changes the description
// (PATCH /api/v1/events/custom/{id}).
func (s *CustomEventsService) Update(registrationID string, input UpdateCustomEventInput) (*UpdateCustomEventResult, error) {
	body := map[string]interface{}{}
	if input.Status != nil {
		body["status"] = *input.Status
	}
	if input.ClearDescription {
		body["description"] = nil
	} else if input.Description != nil {
		body["description"] = *input.Description
	}
	var result UpdateCustomEventResult
	err := s.client.doRequestTyped("PATCH", webhookPath(customEventPath, registrationID), body, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Publish publishes a draft registration.
func (s *CustomEventsService) Publish(registrationID string) (*UpdateCustomEventResult, error) {
	status := CustomEventStatusPublished
	return s.Update(registrationID, UpdateCustomEventInput{Status: &status})
}

// Emit fires a payload against a registered event. It answers 202 queued
// once the emit is stored and handed to durable delivery; webhooks and
// federation peers are called after. A retry with the same IdempotencyKey
// answers the original emitId (and re-hands an emit that never reached
// delivery, e.g. after a 502 bus_relay_failed); another payload under the
// same key is 422 idempotency_key_reused.
func (s *CustomEventsService) Emit(input EmitCustomEventInput) (*EmitCustomEventResult, error) {
	var result EmitCustomEventResult
	err := s.client.doRequestTyped("POST", "/api/v1/events/custom/emit", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DryRun validates a registration template without persisting.
func (s *CustomEventsService) DryRun(input RegisterCustomEventInput) (*DryRunCustomEventResult, error) {
	var result DryRunCustomEventResult
	err := s.client.doRequestTyped("POST", "/api/v1/events/custom/dry-run", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DryRunEmit checks an emit without sending it: the emit's own checks,
// nothing written, counted or relayed. A payload the emit would refuse
// answers the emit's error (400 payload_invalid, 404, 409, 402, 413).
func (s *CustomEventsService) DryRunEmit(input EmitCustomEventInput) (*DryRunEmitResult, error) {
	body := map[string]interface{}{"mode": "emit", "name": input.Name, "payload": input.Payload}
	if input.Version != nil {
		body["version"] = *input.Version
	}
	var result DryRunEmitResult
	err := s.client.doRequestTyped("POST", "/api/v1/events/custom/dry-run", body, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Health returns the tenant's usage diagnostics.
func (s *CustomEventsService) Health() (*CustomEventHealth, error) {
	var result CustomEventHealth
	err := s.client.doRequestTyped("GET", "/api/v1/events/custom/health", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// SubscribeCustomEventInput is the body of POST /api/v1/events/custom/subscribe.
type SubscribeCustomEventInput struct {
	Name        string  `json:"name"`              // "<namespace>.<name>"
	Version     *string `json:"version,omitempty"` // exact semver, "*", or omitted = all versions
	WebhookURL  string  `json:"webhookUrl"`
	DisplayName *string `json:"displayName,omitempty"`
}

// SubscribeCustomEventResult is the response shape for subscribe.
type SubscribeCustomEventResult struct {
	SubscriptionID string `json:"subscriptionId"`
	CanonicalName  string `json:"canonicalName"`
	Secret         string `json:"secret"` // returned ONCE — verify X-Event-Signature with this
	WebhookURL     string `json:"webhookUrl"`
}

// Subscribe registers a webhook against a custom event. The platform
// resolves the canonical bus name from the API key's organisation — no
// need to construct TENANT_CUSTOM:org:ns.name@v strings by hand.
//
// Pass version="*" or omit Version to subscribe to all versions of the
// named event.
func (s *CustomEventsService) Subscribe(input SubscribeCustomEventInput) (*SubscribeCustomEventResult, error) {
	var result SubscribeCustomEventResult
	err := s.client.doRequestTyped("POST", "/api/v1/events/custom/subscribe", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
