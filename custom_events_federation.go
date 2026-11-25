// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// Tenant custom events: the marketplace (Phase 9C) and federation push
// (Phase 9D) sub-services. Moved out of custom_events.go on 2026-11-17
// (custom events gap review, CE-B) with: member remove / pause / resume and
// the relay log (F4, F8); FederatedTypeName on pushes (F2); marketplace
// `mine`, opt-in PayloadExample and PublisherName (F11, F18, F21); signature
// header docs that match what the relay sends (F15); path templates with
// URL-escaped ids. Error codes are SCREAMING_SNAKE since 2026-11-17.
// 2026-11-18 (v1.12.0 pending): EventFederationGroupSummary.CreatedBy and
// EventFederationPushSummary.PushedBy are *string. The API returns the actor
// labels (createdBy, pushedBy, unpushedBy) only to their own organisation (the
// group owner; the push's source) and null to every other member.
// =============================================================================

package platformxe

import (
	"fmt"
	"net/url"
	"strings"
)

// Path templates of the marketplace and federation routes.
const (
	marketplacePath          = "/api/v1/events/custom/marketplace"
	marketplacePublishPath   = "/api/v1/events/custom/marketplace/publish"
	marketplaceListingPath   = "/api/v1/events/custom/marketplace/{id}"
	marketplaceRepublishPath = "/api/v1/events/custom/marketplace/{id}/republish"
	marketplaceForkPath      = "/api/v1/events/custom/marketplace/{id}/fork"

	federationGroupsPath             = "/api/v1/events/custom/federation/groups"
	federationGroupPath              = "/api/v1/events/custom/federation/groups/{id}"
	federationInvitePath             = "/api/v1/events/custom/federation/groups/{id}/invite"
	federationAcceptPath             = "/api/v1/events/custom/federation/groups/{id}/accept"
	federationLeavePath              = "/api/v1/events/custom/federation/groups/{id}/leave"
	federationPushesPath             = "/api/v1/events/custom/federation/groups/{id}/pushes"
	federationExternalPeersPath      = "/api/v1/events/custom/federation/groups/{id}/external-peers"
	federationMemberPath             = "/api/v1/events/custom/federation/groups/{id}/members/{orgId}"
	federationMemberPausePath        = "/api/v1/events/custom/federation/groups/{id}/members/{orgId}/pause"
	federationMemberResumePath       = "/api/v1/events/custom/federation/groups/{id}/members/{orgId}/resume"
	federationPushPath               = "/api/v1/events/custom/federation/pushes/{id}"
	federationPushRelaysPath         = "/api/v1/events/custom/federation/pushes/{id}/relays"
	federationExternalPeerPath       = "/api/v1/events/custom/federation/external-peers/{id}"
	federationExternalPeerPausePath  = "/api/v1/events/custom/federation/external-peers/{id}/pause"
	federationExternalPeerResumePath = "/api/v1/events/custom/federation/external-peers/{id}/resume"
)

// fillPath replaces {id} (and {orgId} when given) with URL-escaped values.
// url.PathEscape leaves "." and ".." as they are; the client refuses a path
// with such a segment before any request (ErrDotSegmentPath), as the Python
// SDK's fill_path refuses those ids.
func fillPath(template, id string, orgID ...string) string {
	out := strings.Replace(template, "{id}", url.PathEscape(id), 1)
	if len(orgID) > 0 {
		out = strings.Replace(out, "{orgId}", url.PathEscape(orgID[0]), 1)
	}
	return out
}

// =============================================================================
// MARKETPLACE (Phase 9C)
// =============================================================================
// PRO+ tenants publish a registered event so other PRO+ tenants can fork the
// schema into their own org. Browsing is open to all plans. Another tenant's
// listing names its publisher (PublisherName), never its organisation id or
// caller label; an unpublished listing is hidden from every other tenant.

// MarketplaceListingStatus is the lifecycle state of a marketplace listing.
type MarketplaceListingStatus string

const (
	MarketplaceListingStatusPublished   MarketplaceListingStatus = "published"
	MarketplaceListingStatusUnpublished MarketplaceListingStatus = "unpublished"
	MarketplaceListingStatusArchived    MarketplaceListingStatus = "archived"
)

// PublishMarketplaceInput is the body of POST /marketplace/publish. Title is
// 3..120 characters, Description at most 2,000, at most 10 Tags of at most
// 32 characters. The registration's own example is never copied:
// PayloadExample is opt-in, must match the schema and is at most 16 KB.
type PublishMarketplaceInput struct {
	RegistrationID string                 `json:"registrationId"`
	Title          string                 `json:"title"`
	Description    *string                `json:"description,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	PayloadExample map[string]interface{} `json:"payloadExample,omitempty"`
}

// MarketplaceListingSummary is the per-row entry for list/detail.
// PublisherOrganizationID and PublishedBy are set on the caller's own
// listings only; SourceCanonicalName is `<namespace>.<name>@<version>` on
// other publishers' listings.
type MarketplaceListingSummary struct {
	ID                      string   `json:"id"`
	PublisherName           string   `json:"publisherName"`
	PublisherOrganizationID *string  `json:"publisherOrganizationId"`
	Namespace               string   `json:"namespace"`
	Name                    string   `json:"name"`
	Version                 string   `json:"version"`
	SourceCanonicalName     string   `json:"sourceCanonicalName"`
	Title                   string   `json:"title"`
	Description             *string  `json:"description"`
	Tags                    []string `json:"tags"`
	Status                  string   `json:"status"`
	ForkCount               int      `json:"forkCount"`
	PublishedBy             *string  `json:"publishedBy"`
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

// ListMarketplaceParams is the query params shape. Mine lists the caller's
// own listings in any status; a Status other than "published" needs Mine.
type ListMarketplaceParams struct {
	Namespace *string
	Mine      *bool
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

// ForkMarketplaceInput is the body of POST /marketplace/:id/fork. Version is
// MAJOR.MINOR.PATCH (default 1.0.0); Status is "draft" (default) or
// "published".
type ForkMarketplaceInput struct {
	Namespace   string  `json:"namespace"`
	Name        string  `json:"name"`
	Version     *string `json:"version,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}

// ForkMarketplaceResult is the response of POST /marketplace/:id/fork.
// SourceCanonicalName is the listing's `<namespace>.<name>@<version>`.
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
	if err := s.client.doRequestTyped("POST", marketplacePublishPath, input, nil, &result); err != nil {
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
	if params.Mine != nil {
		q["mine"] = fmt.Sprintf("%t", *params.Mine)
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
	if err := s.client.doRequestTyped("GET", marketplacePath, nil, q, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns full detail (incl. payloadSchema) for a listing. Another
// publisher's withdrawn listing is 404 LISTING_NOT_FOUND.
func (s *MarketplaceService) Get(listingID string) (*MarketplaceListingDetail, error) {
	var result MarketplaceListingDetail
	if err := s.client.doRequestTyped("GET", fillPath(marketplaceListingPath, listingID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Unpublish hides a listing (owner only). Existing forks unaffected.
func (s *MarketplaceService) Unpublish(listingID string) (*UnpublishMarketplaceResult, error) {
	var result UnpublishMarketplaceResult
	if err := s.client.doRequestTyped("DELETE", fillPath(marketplaceListingPath, listingID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Republish re-activates a previously-unpublished listing (owner only, PRO+).
func (s *MarketplaceService) Republish(listingID string) (*MarketplaceListingSummary, error) {
	var result MarketplaceListingSummary
	if err := s.client.doRequestTyped("POST", fillPath(marketplaceRepublishPath, listingID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Fork creates an independent copy of the listing's schema in the caller's
// org. PRO+ only. The fork is a registration: naming, the namespace
// allowlist, the schema rules, semver and backward compatibility against the
// caller's priors, and the plan's registration cap apply (400
// VALIDATION_FAILED, 402 REGISTRATION_QUOTA_EXCEEDED).
func (s *MarketplaceService) Fork(listingID string, input ForkMarketplaceInput) (*ForkMarketplaceResult, error) {
	var result ForkMarketplaceResult
	if err := s.client.doRequestTyped("POST", fillPath(marketplaceForkPath, listingID), input, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// =============================================================================
// FEDERATION PUSH (Phase 9D, ENTERPRISE-only)
// =============================================================================
// Owner organisations create federation groups, invite peer Enterprise orgs,
// and declare per-version event-pushes. When the owner emits a pushed event,
// every accepted member receives it as the federated type
// `fed-<ownerOrg>.<namespace>.<name>` (FederatedTypeName on each push) —
// never on the member's own `<namespace>.<name>` channel. Changes need the
// events:federate scope.

// EventFederationMemberStatus is the lifecycle state of a member's relationship
// with a federation group.
type EventFederationMemberStatus string

const (
	EventFederationMemberStatusPending  EventFederationMemberStatus = "pending"
	EventFederationMemberStatusAccepted EventFederationMemberStatus = "accepted"
	EventFederationMemberStatusPaused   EventFederationMemberStatus = "paused"
	EventFederationMemberStatusRemoved  EventFederationMemberStatus = "removed"
)

// EventFederationRelayStatus is the outcome of a relay (its latest attempt).
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
	// CreatedBy is set for the owner organisation only; nil for members (v1.12.0).
	CreatedBy  *string `json:"createdBy"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
	ArchivedAt *string `json:"archivedAt"`
	// CallerRole is "owner" if the caller's org owns the group, else the
	// EventFederationMemberStatus value.
	CallerRole  string `json:"callerRole"`
	MemberCount int    `json:"memberCount"`
	PushCount   int    `json:"pushCount"`
}

// EventFederationPeerType discriminates the two member shapes:
//   - "tenant_org"       — peer is another PlatformXe tenant organization
//   - "external_webhook" — peer is an arbitrary HTTPS endpoint (Pattern 3).
//     MemberOrganizationID is nil; ExternalWebhookURL and
//     ExternalWebhookHeaderNames are set for the owner only.
type EventFederationPeerType string

const (
	EventFederationPeerTypeTenantOrg       EventFederationPeerType = "tenant_org"
	EventFederationPeerTypeExternalWebhook EventFederationPeerType = "external_webhook"
)

// EventFederationMemberSummary is one row of a group's member list. It is a
// polymorphic shape: PeerType determines which fields are populated.
// PausedBy is `owner:<caller>`, `member:<caller>` or `system:plan_downgrade`.
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
// FederatedTypeName is what members receive and subscribe to.
type EventFederationPushSummary struct {
	ID                   string  `json:"id"`
	GroupID              string  `json:"groupId"`
	SourceRegistrationID *string `json:"sourceRegistrationId"`
	SourceOrganizationID string  `json:"sourceOrganizationId"`
	Namespace            string  `json:"namespace"`
	Name                 string  `json:"name"`
	Version              string  `json:"version"`
	SourceCanonicalName  string  `json:"sourceCanonicalName"`
	FederatedTypeName    string  `json:"federatedTypeName"`
	// PushedBy and UnpushedBy are set for the push's source organisation
	// only; nil for every other member (v1.12.0: PushedBy was a string).
	PushedBy   *string `json:"pushedBy"`
	PushedAt   string  `json:"pushedAt"`
	UnpushedBy *string `json:"unpushedBy"`
	UnpushedAt *string `json:"unpushedAt"`
	IsActive   bool    `json:"isActive"`
}

// EventFederationGroupDetail extends EventFederationGroupSummary with the
// members and active pushes the caller's role may see (the owner: every
// member; a member: live members, external peers by label only; a pending
// invitee: no members).
type EventFederationGroupDetail struct {
	EventFederationGroupSummary
	Members []EventFederationMemberSummary `json:"members"`
	Pushes  []EventFederationPushSummary   `json:"pushes"`
}

// ListEventFederationGroupsParams is the query-params shape for ListGroups.
type ListEventFederationGroupsParams struct {
	IncludeArchived *bool
	// IncludeRemoved also lists groups the caller left or was removed from.
	IncludeRemoved *bool
}

// ListEventFederationGroupsResult is the response of GET /federation/groups.
type ListEventFederationGroupsResult struct {
	Groups []EventFederationGroupSummary `json:"groups"`
}

// ListEventFederationPushesParams is the query-params shape for ListPushes.
type ListEventFederationPushesParams struct {
	IncludeInactive *bool
}

// ListEventFederationPushesResult is the response of GET /federation/groups/:id/pushes.
type ListEventFederationPushesResult struct {
	Pushes []EventFederationPushSummary `json:"pushes"`
}

// EventFederationRelay is one relay of one emit to one peer (latest attempt).
type EventFederationRelay struct {
	ID           string                     `json:"id"`
	PushID       string                     `json:"pushId"`
	EmitID       string                     `json:"emitId"`
	Peer         string                     `json:"peer"`
	PeerType     EventFederationPeerType    `json:"peerType"`
	Status       EventFederationRelayStatus `json:"status"`
	ErrorMessage *string                    `json:"errorMessage"`
	RelayedAt    *string                    `json:"relayedAt"`
	CreatedAt    string                     `json:"createdAt"`
}

// ListEventFederationRelaysParams is the query-params shape for ListRelays.
type ListEventFederationRelaysParams struct {
	Status *EventFederationRelayStatus
	Limit  *int
	Offset *int
}

// ListEventFederationRelaysResult is the response of GET /federation/pushes/:id/relays.
type ListEventFederationRelaysResult struct {
	Relays []EventFederationRelay `json:"relays"`
	Limit  int                    `json:"limit"`
	Offset int                    `json:"offset"`
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
// POST /federation/groups/:id/external-peers (Pattern 3).
//
// Headers (e.g. a partner bearer token) are stored sealed under the
// key-management envelope, sent on every relay and wiped when the peer is
// removed; only their NAMES are echoed back, to the owner. At most 10, no
// hop-by-hop or X-Caldera-* names, no CR/LF, at most 1 KiB each.
type AddEventFederationExternalPeerInput struct {
	Label      string            `json:"label"`
	WebhookURL string            `json:"webhookUrl"`
	Headers    map[string]string `json:"headers,omitempty"`
}

// AddEventFederationExternalPeerResult is the response of
// POST /federation/groups/:id/external-peers. The Secret is a one-time
// whsec_-prefixed key — shown ONLY here, never stored for an idempotent
// replay. Each relay POST carries X-Caldera-Signature-V2: sha256=<hex> =
// HMAC-SHA256(secret, "<X-Caldera-Timestamp>.<raw body>") with
// X-Caldera-Timestamp (reject beyond 5 minutes) and X-Caldera-Delivery-Id
// (de-duplicate on it). Until 2027-04-30 it also carries the legacy
// body-only X-Caldera-Signature: sha256=<hex> = HMAC-SHA256(secret, raw body).
type AddEventFederationExternalPeerResult struct {
	Peer   EventFederationMemberSummary `json:"peer"`
	Secret string                       `json:"secret"`
}

// EventFederationService is the typed client for the federation push surface.
// Accessed via Client.Events.Custom.Federation.
type EventFederationService struct {
	client *Client
}

func boolParam(q map[string]string, key string, v *bool) {
	if v != nil {
		q[key] = fmt.Sprintf("%t", *v)
	}
}

// CreateGroup creates a new federation group owned by the calling org.
func (s *EventFederationService) CreateGroup(input CreateEventFederationGroupInput) (*EventFederationGroupSummary, error) {
	var result EventFederationGroupSummary
	if err := s.client.doRequestTyped("POST", federationGroupsPath, input, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListGroups lists groups visible to the caller (owned + live member-of).
func (s *EventFederationService) ListGroups(params *ListEventFederationGroupsParams) (*ListEventFederationGroupsResult, error) {
	q := map[string]string{}
	if params != nil {
		boolParam(q, "includeArchived", params.IncludeArchived)
		boolParam(q, "includeRemoved", params.IncludeRemoved)
	}
	var result ListEventFederationGroupsResult
	if err := s.client.doRequestTyped("GET", federationGroupsPath, nil, q, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetGroup returns one group as the caller's role may see it. A caller with
// no live relationship (removed included) gets 404 GROUP_NOT_FOUND.
func (s *EventFederationService) GetGroup(groupID string) (*EventFederationGroupDetail, error) {
	var result EventFederationGroupDetail
	if err := s.client.doRequestTyped("GET", fillPath(federationGroupPath, groupID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ArchiveGroup archives a group the caller owns. Stops fan-out;
// preserves history. Owner-only.
func (s *EventFederationService) ArchiveGroup(groupID string) (*EventFederationGroupSummary, error) {
	var result EventFederationGroupSummary
	if err := s.client.doRequestTyped("DELETE", fillPath(federationGroupPath, groupID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Invite invites a peer ENTERPRISE org into the group. Owner-only. The
// invitee receives CUSTOM_EVENT_FEDERATION_MEMBER_INVITED.
func (s *EventFederationService) Invite(groupID string, input InviteEventFederationMemberInput) (*EventFederationMemberSummary, error) {
	var result EventFederationMemberSummary
	if err := s.client.doRequestTyped("POST", fillPath(federationInvitePath, groupID), input, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Accept accepts a pending invitation (the invited org). Every active push
// is registered on the caller as its federated type.
func (s *EventFederationService) Accept(groupID string) (*EventFederationMemberSummary, error) {
	var result EventFederationMemberSummary
	if err := s.client.doRequestTyped("POST", fillPath(federationAcceptPath, groupID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Leave voluntarily leaves a group the caller is a member of.
func (s *EventFederationService) Leave(groupID string) (*EventFederationMemberSummary, error) {
	var result EventFederationMemberSummary
	if err := s.client.doRequestTyped("POST", fillPath(federationLeavePath, groupID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RemoveMember removes a member or revokes a pending invitation. Owner-only.
func (s *EventFederationService) RemoveMember(groupID, memberOrganizationID string) (*EventFederationMemberSummary, error) {
	var result EventFederationMemberSummary
	if err := s.client.doRequestTyped("DELETE", fillPath(federationMemberPath, groupID, memberOrganizationID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PauseMember pauses an accepted membership (the owner, or the member
// itself); relays stop until the same side resumes it.
func (s *EventFederationService) PauseMember(groupID, memberOrganizationID string) (*EventFederationMemberSummary, error) {
	var result EventFederationMemberSummary
	if err := s.client.doRequestTyped("POST", fillPath(federationMemberPausePath, groupID, memberOrganizationID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ResumeMember resumes a paused membership. The side that paused resumes; a
// plan pause resumes on its own when both sides are Enterprise again.
func (s *EventFederationService) ResumeMember(groupID, memberOrganizationID string) (*EventFederationMemberSummary, error) {
	var result EventFederationMemberSummary
	if err := s.client.doRequestTyped("POST", fillPath(federationMemberResumePath, groupID, memberOrganizationID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeclarePush declares a per-version push of one of the owner's
// registrations. Owner-only. The generic event and federated types cannot be
// pushed (400 EVENT_NOT_PUSHABLE).
func (s *EventFederationService) DeclarePush(groupID string, input DeclareEventFederationPushInput) (*EventFederationPushSummary, error) {
	var result EventFederationPushSummary
	if err := s.client.doRequestTyped("POST", fillPath(federationPushesPath, groupID), input, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListPushes lists pushes declared in a group.
func (s *EventFederationService) ListPushes(groupID string, params *ListEventFederationPushesParams) (*ListEventFederationPushesResult, error) {
	q := map[string]string{}
	if params != nil {
		boolParam(q, "includeInactive", params.IncludeInactive)
	}
	var result ListEventFederationPushesResult
	if err := s.client.doRequestTyped("GET", fillPath(federationPushesPath, groupID), nil, q, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UndeclarePush stops pushing a previously-declared per-version push. Owner-only.
func (s *EventFederationService) UndeclarePush(pushID string) (*EventFederationPushSummary, error) {
	var result EventFederationPushSummary
	if err := s.client.doRequestTyped("DELETE", fillPath(federationPushPath, pushID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListRelays reads the relay log of a push, newest first. Owner-only.
func (s *EventFederationService) ListRelays(pushID string, params *ListEventFederationRelaysParams) (*ListEventFederationRelaysResult, error) {
	q := map[string]string{}
	if params != nil {
		if params.Status != nil {
			q["status"] = string(*params.Status)
		}
		if params.Limit != nil {
			q["limit"] = fmt.Sprintf("%d", *params.Limit)
		}
		if params.Offset != nil {
			q["offset"] = fmt.Sprintf("%d", *params.Offset)
		}
	}
	var result ListEventFederationRelaysResult
	if err := s.client.doRequestTyped("GET", fillPath(federationPushRelaysPath, pushID), nil, q, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AddExternalPeer adds an external (non-tenant) peer to a federation group,
// addressed by webhook URL (Pattern 3). Owner-only. See
// AddEventFederationExternalPeerResult for the one-time secret and the
// signature headers to verify.
func (s *EventFederationService) AddExternalPeer(groupID string, input AddEventFederationExternalPeerInput) (*AddEventFederationExternalPeerResult, error) {
	var result AddEventFederationExternalPeerResult
	if err := s.client.doRequestTyped("POST", fillPath(federationExternalPeersPath, groupID), input, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RemoveExternalPeer removes an external webhook peer (status "removed").
// Relays and in-flight retries stop; the sealed secret and headers are wiped
// (only the secret's hash is kept, for diagnostics). Owner-only.
func (s *EventFederationService) RemoveExternalPeer(memberID string) (*EventFederationMemberSummary, error) {
	var result EventFederationMemberSummary
	if err := s.client.doRequestTyped("DELETE", fillPath(federationExternalPeerPath, memberID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PauseExternalPeer pauses an external webhook peer (owner only). Relays and
// in-flight retries stop; the sealed secret and headers are kept. Works after a
// plan drop (an exit action).
func (s *EventFederationService) PauseExternalPeer(memberID string) (*EventFederationMemberSummary, error) {
	var result EventFederationMemberSummary
	if err := s.client.doRequestTyped("POST", fillPath(federationExternalPeerPausePath, memberID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ResumeExternalPeer resumes a paused external webhook peer (owner only); a
// plan pause resumes only when the owner is Enterprise again.
func (s *EventFederationService) ResumeExternalPeer(memberID string) (*EventFederationMemberSummary, error) {
	var result EventFederationMemberSummary
	if err := s.client.doRequestTyped("POST", fillPath(federationExternalPeerResumePath, memberID), nil, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
