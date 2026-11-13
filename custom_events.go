// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-09-15): gofmt alignment only (SubscribeCustomEventInput,
// SubscribeCustomEventResult, ForkMarketplaceResult,
// EventFederationPushSummary) — `gofmt -l` listed this file in CI.
// =============================================================================

package platformxe

import "fmt"

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
	Events []CustomEventSummary `json:"events"`
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
		Used  int  `json:"used"`
		Limit *int `json:"limit"`
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
}

// Register registers a new custom event (or new version).
func (s *CustomEventsService) Register(input RegisterCustomEventInput) (*RegisterCustomEventResult, error) {
	var result RegisterCustomEventResult
	err := s.client.doRequestTyped("POST", "/api/v1/events/custom", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// List returns every registration the API key's organisation has.
func (s *CustomEventsService) List(namespace, status string) (*CustomEventListResult, error) {
	params := map[string]string{}
	if namespace != "" {
		params["namespace"] = namespace
	}
	if status != "" {
		params["status"] = status
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
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/events/custom/%s", registrationID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Archive soft-deletes a registration.
func (s *CustomEventsService) Archive(registrationID string) (*ArchiveCustomEventResult, error) {
	var result ArchiveCustomEventResult
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/events/custom/%s", registrationID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Emit fires a payload against a registered event.
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

// =============================================================================
// MARKETPLACE (Phase 9C — Go SDK 1.3.0)
// =============================================================================
// PRO+ tenants publish a registered event so other PRO+ tenants can fork
// the schema into their own org. Browsing is open to all plans.
// Endpoints map to /api/v1/events/custom/marketplace/*.

// MarketplaceListingStatus is the lifecycle state of a marketplace listing.
type MarketplaceListingStatus string

const (
	MarketplaceListingStatusPublished   MarketplaceListingStatus = "published"
	MarketplaceListingStatusUnpublished MarketplaceListingStatus = "unpublished"
	MarketplaceListingStatusArchived    MarketplaceListingStatus = "archived"
)

// PublishMarketplaceInput is the body of POST /marketplace/publish.
type PublishMarketplaceInput struct {
	RegistrationID string   `json:"registrationId"`
	Title          string   `json:"title"`
	Description    *string  `json:"description,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}

// MarketplaceListingSummary is the per-row entry for list/detail.
type MarketplaceListingSummary struct {
	ID                      string   `json:"id"`
	PublisherOrganizationID string   `json:"publisherOrganizationId"`
	Namespace               string   `json:"namespace"`
	Name                    string   `json:"name"`
	Version                 string   `json:"version"`
	SourceCanonicalName     string   `json:"sourceCanonicalName"`
	Title                   string   `json:"title"`
	Description             *string  `json:"description"`
	Tags                    []string `json:"tags"`
	Status                  string   `json:"status"`
	ForkCount               int      `json:"forkCount"`
	PublishedBy             string   `json:"publishedBy"`
	PublishedAt             string   `json:"publishedAt"`
	UnpublishedAt           *string  `json:"unpublishedAt"`
	CreatedAt               string   `json:"createdAt"`
	UpdatedAt               string   `json:"updatedAt"`
}

// MarketplaceListingDetail extends summary with the schema body.
type MarketplaceListingDetail struct {
	MarketplaceListingSummary
	PayloadSchema  map[string]interface{} `json:"payloadSchema"`
	PayloadExample map[string]interface{} `json:"payloadExample,omitempty"`
}

// ListMarketplaceParams is the query params shape.
type ListMarketplaceParams struct {
	Namespace *string
	Status    *string
	Search    *string
	Limit     *int
	Offset    *int
}

// ListMarketplaceResult is the response of GET /marketplace.
type ListMarketplaceResult struct {
	Listings []MarketplaceListingSummary `json:"listings"`
	Total    int                         `json:"total"`
	Limit    int                         `json:"limit"`
	Offset   int                         `json:"offset"`
}

// ForkMarketplaceInput is the body of POST /marketplace/:id/fork.
type ForkMarketplaceInput struct {
	Namespace   string  `json:"namespace"`
	Name        string  `json:"name"`
	Version     *string `json:"version,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}

// ForkMarketplaceResult is the response of POST /marketplace/:id/fork.
type ForkMarketplaceResult struct {
	ForkID              string `json:"forkId"`
	ForkRegistrationID  string `json:"forkRegistrationId"`
	ForkCanonicalName   string `json:"forkCanonicalName"`
	SourceCanonicalName string `json:"sourceCanonicalName"`
}

// UnpublishMarketplaceResult is the response of DELETE /marketplace/:id.
type UnpublishMarketplaceResult struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	UnpublishedAt string `json:"unpublishedAt"`
}

// MarketplaceService is the typed client for the marketplace surface,
// nested under CustomEventsService at .Marketplace.
type MarketplaceService struct {
	client *Client
}

// Publish a registered event to the marketplace. PRO+ only.
func (s *MarketplaceService) Publish(input PublishMarketplaceInput) (*MarketplaceListingSummary, error) {
	var result MarketplaceListingSummary
	err := s.client.doRequestTyped("POST", "/api/v1/events/custom/marketplace/publish", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// List browses the marketplace. Open to all plans.
func (s *MarketplaceService) List(params ListMarketplaceParams) (*ListMarketplaceResult, error) {
	q := map[string]string{}
	if params.Namespace != nil {
		q["namespace"] = *params.Namespace
	}
	if params.Status != nil {
		q["status"] = *params.Status
	}
	if params.Search != nil {
		q["search"] = *params.Search
	}
	if params.Limit != nil {
		q["limit"] = fmt.Sprintf("%d", *params.Limit)
	}
	if params.Offset != nil {
		q["offset"] = fmt.Sprintf("%d", *params.Offset)
	}
	var result ListMarketplaceResult
	err := s.client.doRequestTyped("GET", "/api/v1/events/custom/marketplace", nil, q, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns full detail (incl. payloadSchema) for a listing.
func (s *MarketplaceService) Get(listingID string) (*MarketplaceListingDetail, error) {
	var result MarketplaceListingDetail
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/events/custom/marketplace/%s", listingID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Unpublish hides a listing (owner only). Existing forks unaffected.
func (s *MarketplaceService) Unpublish(listingID string) (*UnpublishMarketplaceResult, error) {
	var result UnpublishMarketplaceResult
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/events/custom/marketplace/%s", listingID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Republish re-activates a previously-unpublished listing (owner only).
func (s *MarketplaceService) Republish(listingID string) (*MarketplaceListingSummary, error) {
	var result MarketplaceListingSummary
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/events/custom/marketplace/%s/republish", listingID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Fork creates an independent copy of the listing's schema in the
// caller's org under the chosen destination shape. PRO+ only.
func (s *MarketplaceService) Fork(listingID string, input ForkMarketplaceInput) (*ForkMarketplaceResult, error) {
	var result ForkMarketplaceResult
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/events/custom/marketplace/%s/fork", listingID), input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// =============================================================================
// FEDERATION PUSH (Phase 9D, ENTERPRISE-only)
// =============================================================================
// Owner organisations create federation groups, invite peer Enterprise orgs,
// and declare per-version event-pushes. When the owner emits a pushed event,
// every accepted member receives the relay on their own bus channel.
//
// Endpoints map to /api/v1/events/custom/federation/*.

// EventFederationMemberStatus is the lifecycle state of a member's relationship
// with a federation group.
type EventFederationMemberStatus string

const (
	EventFederationMemberStatusPending  EventFederationMemberStatus = "pending"
	EventFederationMemberStatusAccepted EventFederationMemberStatus = "accepted"
	EventFederationMemberStatusPaused   EventFederationMemberStatus = "paused"
	EventFederationMemberStatusRemoved  EventFederationMemberStatus = "removed"
)

// EventFederationRelayStatus is the outcome of a single relay attempt.
type EventFederationRelayStatus string

const (
	EventFederationRelayStatusQueued  EventFederationRelayStatus = "queued"
	EventFederationRelayStatusRelayed EventFederationRelayStatus = "relayed"
	EventFederationRelayStatusFailed  EventFederationRelayStatus = "failed"
)

// CreateEventFederationGroupInput is the body of POST /federation/groups.
type CreateEventFederationGroupInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// EventFederationGroupSummary is the per-row entry for list/detail responses.
type EventFederationGroupSummary struct {
	ID                  string  `json:"id"`
	OwnerOrganizationID string  `json:"ownerOrganizationId"`
	Name                string  `json:"name"`
	Description         *string `json:"description"`
	CreatedBy           string  `json:"createdBy"`
	CreatedAt           string  `json:"createdAt"`
	UpdatedAt           string  `json:"updatedAt"`
	ArchivedAt          *string `json:"archivedAt"`
	// CallerRole is "owner" if the caller's org owns the group, else the
	// EventFederationMemberStatus value.
	CallerRole  string `json:"callerRole"`
	MemberCount int    `json:"memberCount"`
	PushCount   int    `json:"pushCount"`
}

// EventFederationPeerType discriminates the two member shapes:
//   - "tenant_org"       — peer is another PlatformXe tenant organization
//   - "external_webhook" — peer is an arbitrary HTTPS endpoint (Pattern 3,
//     added in 1.5.0). MemberOrganizationID is nil; ExternalWebhookURL,
//     ExternalWebhookLabel, and ExternalWebhookHeaderNames are populated.
type EventFederationPeerType string

const (
	EventFederationPeerTypeTenantOrg       EventFederationPeerType = "tenant_org"
	EventFederationPeerTypeExternalWebhook EventFederationPeerType = "external_webhook"
)

// EventFederationMemberSummary is one row of a group's member list. It is a
// polymorphic shape: PeerType determines which fields are populated.
type EventFederationMemberSummary struct {
	ID                         string                      `json:"id"`
	GroupID                    string                      `json:"groupId"`
	PeerType                   EventFederationPeerType     `json:"peerType"`
	MemberOrganizationID       *string                     `json:"memberOrganizationId"`
	ExternalWebhookURL         *string                     `json:"externalWebhookUrl"`
	ExternalWebhookLabel       *string                     `json:"externalWebhookLabel"`
	ExternalWebhookHeaderNames []string                    `json:"externalWebhookHeaderNames"`
	Status                     EventFederationMemberStatus `json:"status"`
	InvitedBy                  string                      `json:"invitedBy"`
	InvitedAt                  string                      `json:"invitedAt"`
	AcceptedBy                 *string                     `json:"acceptedBy"`
	AcceptedAt                 *string                     `json:"acceptedAt"`
	PausedBy                   *string                     `json:"pausedBy"`
	PausedAt                   *string                     `json:"pausedAt"`
	RemovedBy                  *string                     `json:"removedBy"`
	RemovedAt                  *string                     `json:"removedAt"`
}

// EventFederationPushSummary is one row of a group's push list.
type EventFederationPushSummary struct {
	ID                   string  `json:"id"`
	GroupID              string  `json:"groupId"`
	SourceRegistrationID *string `json:"sourceRegistrationId"`
	SourceOrganizationID string  `json:"sourceOrganizationId"`
	Namespace            string  `json:"namespace"`
	Name                 string  `json:"name"`
	Version              string  `json:"version"`
	SourceCanonicalName  string  `json:"sourceCanonicalName"`
	PushedBy             string  `json:"pushedBy"`
	PushedAt             string  `json:"pushedAt"`
	UnpushedBy           *string `json:"unpushedBy"`
	UnpushedAt           *string `json:"unpushedAt"`
	IsActive             bool    `json:"isActive"`
}

// EventFederationGroupDetail extends EventFederationGroupSummary with the full members
// + active pushes lists. Returned by GET /federation/groups/:id.
type EventFederationGroupDetail struct {
	EventFederationGroupSummary
	Members []EventFederationMemberSummary `json:"members"`
	Pushes  []EventFederationPushSummary   `json:"pushes"`
}

// ListEventFederationGroupsParams is the query-params shape for list_groups.
type ListEventFederationGroupsParams struct {
	IncludeArchived *bool
}

// ListEventFederationGroupsResult is the response of GET /federation/groups.
type ListEventFederationGroupsResult struct {
	Groups []EventFederationGroupSummary `json:"groups"`
}

// ListEventFederationPushesParams is the query-params shape for list_pushes.
type ListEventFederationPushesParams struct {
	IncludeInactive *bool
}

// ListEventFederationPushesResult is the response of GET /federation/groups/:id/pushes.
type ListEventFederationPushesResult struct {
	Pushes []EventFederationPushSummary `json:"pushes"`
}

// InviteEventFederationMemberInput is the body of POST /federation/groups/:id/invite.
type InviteEventFederationMemberInput struct {
	MemberOrganizationID string `json:"memberOrganizationId"`
}

// DeclareEventFederationPushInput is the body of POST /federation/groups/:id/pushes.
type DeclareEventFederationPushInput struct {
	RegistrationID string `json:"registrationId"`
}

// AddEventFederationExternalPeerInput is the body of
// POST /federation/groups/:id/external-peers (Pattern 3, added in 1.5.0).
//
// Headers, when supplied, are stored as opaque static values and replayed on
// every relay. Only the header NAMES are echoed back in
// EventFederationMemberSummary.ExternalWebhookHeaderNames; values are
// considered secret.
type AddEventFederationExternalPeerInput struct {
	Label      string            `json:"label"`
	WebhookURL string            `json:"webhookUrl"`
	Headers    map[string]string `json:"headers,omitempty"`
}

// AddEventFederationExternalPeerResult is the response of
// POST /federation/groups/:id/external-peers. The Secret is a one-time
// whsec_-prefixed HMAC key — it is shown ONLY here and cannot be retrieved
// again. Store it immediately.
type AddEventFederationExternalPeerResult struct {
	Peer   EventFederationMemberSummary `json:"peer"`
	Secret string                       `json:"secret"`
}

// EventFederationService is the typed client for the federation push surface.
// Accessed via Client.Events.Custom.Federation.
type EventFederationService struct {
	client *Client
}

// CreateGroup creates a new federation group owned by the calling org.
func (s *EventFederationService) CreateGroup(input CreateEventFederationGroupInput) (*EventFederationGroupSummary, error) {
	var result EventFederationGroupSummary
	err := s.client.doRequestTyped("POST", "/api/v1/events/custom/federation/groups", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListGroups lists groups visible to the caller (owned + member-of).
func (s *EventFederationService) ListGroups(params *ListEventFederationGroupsParams) (*ListEventFederationGroupsResult, error) {
	q := map[string]string{}
	if params != nil && params.IncludeArchived != nil {
		if *params.IncludeArchived {
			q["includeArchived"] = "true"
		} else {
			q["includeArchived"] = "false"
		}
	}
	var result ListEventFederationGroupsResult
	err := s.client.doRequestTyped("GET", "/api/v1/events/custom/federation/groups", nil, q, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetGroup returns the full detail (members + pushes) for one group.
func (s *EventFederationService) GetGroup(groupID string) (*EventFederationGroupDetail, error) {
	var result EventFederationGroupDetail
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/events/custom/federation/groups/%s", groupID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ArchiveGroup archives a group the caller owns. Stops fan-out;
// preserves history. Owner-only.
func (s *EventFederationService) ArchiveGroup(groupID string) (*EventFederationGroupSummary, error) {
	var result EventFederationGroupSummary
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/events/custom/federation/groups/%s", groupID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Invite invites a peer ENTERPRISE org into a group. Owner-only.
func (s *EventFederationService) Invite(groupID string, input InviteEventFederationMemberInput) (*EventFederationMemberSummary, error) {
	var result EventFederationMemberSummary
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/events/custom/federation/groups/%s/invite", groupID), input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Accept accepts a pending invitation. Caller must be the invited org.
func (s *EventFederationService) Accept(groupID string) (*EventFederationMemberSummary, error) {
	var result EventFederationMemberSummary
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/events/custom/federation/groups/%s/accept", groupID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Leave voluntarily removes the calling org from a group it's a member of.
func (s *EventFederationService) Leave(groupID string) (*EventFederationMemberSummary, error) {
	var result EventFederationMemberSummary
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/events/custom/federation/groups/%s/leave", groupID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeclarePush declares a per-version push of one of the owner's
// registrations. Owner-only.
func (s *EventFederationService) DeclarePush(groupID string, input DeclareEventFederationPushInput) (*EventFederationPushSummary, error) {
	var result EventFederationPushSummary
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/events/custom/federation/groups/%s/pushes", groupID), input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListPushes lists pushes declared in a group.
func (s *EventFederationService) ListPushes(groupID string, params *ListEventFederationPushesParams) (*ListEventFederationPushesResult, error) {
	q := map[string]string{}
	if params != nil && params.IncludeInactive != nil {
		if *params.IncludeInactive {
			q["includeInactive"] = "true"
		} else {
			q["includeInactive"] = "false"
		}
	}
	var result ListEventFederationPushesResult
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/events/custom/federation/groups/%s/pushes", groupID), nil, q, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UndeclarePush stops pushing a previously-declared per-version push. Owner-only.
func (s *EventFederationService) UndeclarePush(pushID string) (*EventFederationPushSummary, error) {
	var result EventFederationPushSummary
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/events/custom/federation/pushes/%s", pushID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// AddExternalPeer adds an external (non-tenant) peer to a federation group,
// addressed by webhook URL (Pattern 3, added in 1.5.0). Owner-only.
//
// The returned Secret is a one-time whsec_-prefixed HMAC key — store it
// immediately; it cannot be retrieved again. Use it to verify the
// X-Platformxe-Signature header on inbound webhooks.
func (s *EventFederationService) AddExternalPeer(groupID string, input AddEventFederationExternalPeerInput) (*AddEventFederationExternalPeerResult, error) {
	var result AddEventFederationExternalPeerResult
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/events/custom/federation/groups/%s/external-peers", groupID), input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RemoveExternalPeer archives an external webhook peer (sets status="removed").
// The encrypted secret is wiped from storage as part of the soft-delete.
// Owner-only.
func (s *EventFederationService) RemoveExternalPeer(memberID string) (*EventFederationMemberSummary, error) {
	var result EventFederationMemberSummary
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/events/custom/federation/external-peers/%s", memberID), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
