// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-20, v1.13.0 pending, fidelity register): DryRunResult
// carries what POST /api/v1/workflows/{id} answers under dryRun (WouldMatch,
// ConditionsEvaluated, ActionsWouldExecute, ActionsPreview, Validation,
// Reason); TriggerID, TriggerName, Matched and Actions, which the route never
// answers, are deprecated. DryRunActionPreview and DryRunValidation added.
// Fix round (review M1, O1): each deprecated DryRunResult field has its own
// Deprecated line (one comment group only marked TriggerID), and the SDK
// fills TriggerID, Matched and Actions. WorkflowListResponse decodes the
// route's triggers and pagination (it read "workflows", which the route
// never answered, so List was always empty); Workflows is deprecated.
//
// CHANGELOG (2026-09-15):
// - ClientConfig.DisableRetries added: Retries: 0 is the zero value and has
//   always meant "use the default of 2", so a caller could not configure zero
//   retries. A new bool (not a *int, which would break every v1 caller's
//   struct literal) is the explicit switch; a negative Retries now also means
//   no retries instead of making zero request attempts.
// - Restored FolderListResponse and the exported fields removed earlier in
//   this release (DocumentFolder.Name/CreatedAt, DeletionOverride.RequestedBy/
//   Reason/ReviewedBy/ReviewNote/CreatedAt/UpdatedAt) as Deprecated: removing
//   exported identifiers breaks compilation for v1 module consumers. The API
//   never populated them, so they stay empty.
// - gofmt alignment of RenderResult, Thread, EntityEventResult,
//   ThreadReadState, DryRunResult and DeletionOverride.
// - ClientConfig.FailOpen's comment said "default true"; a Go bool's zero
//   value is false and NewClient copies it as-is, so the default is false.
//   Comment corrected to match the code (and apps/docs sdk/configuration).
// - APIError gained StatusCode (json:"-") and, in client.go, an Error()
//   method, so it is the error every non-2xx response returns (MR !199 round
//   9): the client used to ignore the HTTP status and hand back a 5xx whose
//   JSON was not an envelope as data with a nil error. Adding a field only
//   breaks unkeyed APIError{...} literals, which go vet already flags.
//
// CHANGELOG (2026-09-21):
// - Added RasterizePdfInput/RasterizePdfResult for the new PdfService.
//   Rasterize method (pdf.go) — POST /api/v1/pdf/rasterize. Unlike the
//   existing PdfResult (a generic, seemingly-unused placeholder left over
//   from an earlier pass — OfferLetter/PropertyFlyer both return a plain
//   map[string]interface{}), rasterize's request/response are both small
//   and fixed, so it gets typed structs following the IdentityDocVerifyResult
//   precedent (ocr.go's VerifyIdentity: typed input handled via a plain map
//   there, typed *IdentityDocVerifyResult output here) rather than the PDF
//   domain's freeform-template map convention — flyer/letter bodies are
//   arbitrary caller-supplied template data with no fixed shape; rasterize's
//   four request fields and five response fields are not. Field names/types
//   checked field-for-field against RasterizePdfRequest/RasterizePdfResponse
//   in src/app/api/v1/pdf/rasterize/route.ts.
//
// CHANGELOG (2026-10-20, pending release — webhooks v2, NO-DRIFT train):
// - Fixed the drifted webhook types to the real API shape: Webhook (callerService,
//   status, events, engineVersion, recipientRef, delivery counters; no
//   organizationId / isActive / secret), WebhookListResponse (items, total,
//   page, limit), WebhookTestResult (success, responseStatus, responseTimeMs,
//   error). The old fields stay as Deprecated: the API never populated them
//   and removing exported identifiers breaks v1 module consumers.
// - Added WebhookCreated (create / rotate: the secret is returned ONCE),
//   DispatchWebhookInput / DispatchWebhookResult, WebhookDelivery,
//   WebhookDeliveriesResponse and RedeliverWebhookResult.
// - Review fix wave 2 (still pending release; keeps v1 callers compiling):
//   WebhookCreated EMBEDS Webhook (tag json:"webhook", so the JSON is
//   unchanged) - code that read wh.ID / wh.Name off the old *Webhook result of
//   Create / RotateSecret still compiles, and wh.Secret is the one-time
//   secret; only code that stored the result in a *Webhook variable must
//   change. Webhook keeps OrganizationID, IsActive and Secret as Deprecated
//   (the API never populated them). DispatchWebhookInput / Result and
//   WebhookDelivery document the event-name and eventId rules, non-PAUSED
//   targeting and the null snippet on engine-1 rows.
//
// CHANGELOG (2026-10-08, v1.7.0 — caldera-platformxe#22):
// - APIError gained Details (json:"details,omitempty" — the envelope's
//   error.details) and RetryAfter (json:"-", the Retry-After header in
//   seconds). Adding fields only breaks unkeyed APIError{...} literals,
//   which go vet already flags.
// - Added the AI gateway and vision request / response types (AiMessage,
//   AiCompleteInput/Result, AiExtractInput/Result, AiImageInput, AiUsage,
//   AiProfileView, AiUsageQuery, AiUsageSummary/Row/Totals,
//   AnalyzeRoomInput/Result/Metrics, AnalyzeRoomOptions) and the closed
//   error-code constants (ErrCodeAi*, ErrCodeVision*). Field names, JSON tags
//   and optionality match @caldera/platformxe-types 3.5.0 (src/ai.ts,
//   src/vision.ts) field-for-field; nullable fields are pointers, and a
//   field whose zero value is meaningful (Temperature 0) is a pointer so
//   omitempty cannot drop it.
//
// CHANGELOG (2026-10-08, v1.7.0 pending — calderasuite/caldera-xadmin#29):
// - Tool calling: AiToolCall, AiToolDefinition, AiMessage.ToolCalls /
//   ToolCallID, AiCompleteInput.Tools / ToolChoice (a string "auto" | "none"
//   | "required", or AiToolChoiceName{Name}), AiCompleteResult.ToolCalls /
//   FinishReason, AiProfileView.AllowTools.
// - Knowledge retrieval: KnowledgeUpsertDocumentInput/Result,
//   KnowledgeDocumentSummary, KnowledgeDocumentList,
//   KnowledgeDeleteDocumentResult, KnowledgeSearchInput, KnowledgeSearchHit,
//   KnowledgeSearchResult, KnowledgeListQuery — @caldera/platformxe-types
//   3.6.0 field-for-field.
//
// CHANGELOG (2026-10-09, pending — calderasuite/caldera-xadmin#31):
// - AiCompleteInput.Provider documents the new "claude" value (ADR-0014;
//   ProductAiProvider gains 'claude' in the pending @caldera/platformxe-types
//   3.7.0). Comment only — the field is a string.
//
// CHANGELOG (2026-11-01, pending — caldera-platformxe#33):
// - ErrCodeAiModelUnavailable ("AI_MODEL_UNAVAILABLE", 503): the AI gateway
//   account cannot use the requested model (free tier — paid gateway credits
//   are required). Distinct from ErrCodeAiNotConfigured (key missing or
//   rejected); APIError.Message carries the gateway's reason and
//   Details["model"] the model. Mirrors the pending @caldera/platformxe-types
//   ProductAiErrorCode value.
// - 2026-11-01 review (caldera-platformxe#33): APIError.Message is now the
//   fixed text "The AI service account cannot use this model right now." —
//   the gateway's reason is logged server-side, never returned. Comment only.
//
// CHANGELOG (2026-11-01, v1.7.2 pending):
// - IdentityProviderInfo gains Healthy (json:"healthy"). GET
//   /api/v1/identity/providers returns { name, healthy } per provider
//   (resolver getProviderHealth; @caldera/platformxe-types
//   IdentityProviderStatus), so Healthy was never readable. Status,
//   LastChecked and ResponseTime were never returned: kept as Deprecated
//   (removing exported fields breaks v1 consumers), matching the types
//   package's 4.0.0 removal note.
// - EventSubscription gains ConsecutiveFailures / LastTriggeredAt (returned,
//   never decoded); UpdatedAt is never returned (Deprecated).
//   EventSubscriptionListResponse reads Items / Total — the API never sent
//   `subscriptions`, so the list always came back empty; Subscriptions stays
//   (Deprecated) and ListSubscriptions copies Items into it.
//   EventsService.CreateSubscription decodes the API's { subscription, secret }
//   (it decoded the envelope as a flat subscription, so ID was always empty)
//   and returns the subscription with Secret set — same signature.
// - FederationMember gains ID / GroupID and FederationMemberResult gains ID:
//   the federation routes return the member row with both (federation-crud
//   addMember / getGroup; @caldera/platformxe-types PermissionFederationMember).
// - PermissionsService.ListOverrides: GET /permissions/overrides/:adminId
//   answers a bare PermissionOverride[] — decoding it into the
//   { overrides } struct failed on every call. It now decodes the array and
//   wraps it in OverrideListResponse (same signature).
//
// CHANGELOG (2026-11-10, pending — portal completeness wave 2):
// - ErrCodeConflict ("CONFLICT", 409) joins the closed AI code set:
//   Knowledge.UpsertDocument answers it when the collection + sourceRef is
//   held by another organisation's document under the same caller service
//   (until the organisation-keyed contract migration). Final: 409 is never
//   retried and retrying does not fix it.
//
// CHANGELOG (2026-11-12, pending: portal completeness wave 2, Task 4):
// - WorkflowEvalDetail gains DryRun, DurationMs and ActionsExecuted
//   ([]WorkflowExecutedAction: type, config, status, error), all returned by
//   /api/v1/workflows/evaluate and never decoded. Step statuses as constants:
//   WorkflowExecutedActionStatus (SUCCESS, FAILED, SKIPPED, DEFERRED) and the
//   recorded-execution spelling WorkflowActionResultStatus (success, failed,
//   skipped, deferred). ExecutionID is empty when no execution was recorded.
// - Task 4 fix round 4 (N3-I1, m3-4): a recorded execution's steps are in the
//   spelling of the engine that ran it: upper case for a run recorded by
//   /api/v1/workflows/evaluate or the event relay, lower case for an event or
//   scheduled run. WorkflowActionResultStatus.Normalize upper-cases either.
//   WorkflowActionSkipped never means a timeout in an evaluate result (an
//   evaluation has no time limit; only event and scheduled runs do).
// - WorkflowEvalResult gains Message ("No active triggers for event ...") and
//   WorkflowEvalDetail ActionsWouldExecute (a dry run's stored actions), both
//   returned and never decoded. WorkflowEvaluateInput is the request of
//   WorkflowsService.EvaluateWithOptions.
//
// CHANGELOG (2026-11-16, pending: portal completeness wave 2, Task 5b):
// - SendMessageResult gains Deduplicated: true with Queued when your identical
//   email (same recipients, subject and content) was already queued, so
//   nothing new was stored.
// - The email send routes' error codes as constants: ErrCodeEmailRateLimited
//   (429), ErrCodeEmailNotAccepted (503) and ErrCodeEmailProviderUnavailable
//   (503). Each carries RetryAfter and is never retried by the client.
// - (fix round 1, ruling c1) EmailHealthResponse: a key with an organisation
//   reads each route as its chain position ("primary", "fallback_1", ...) in
//   ProviderOrder and as the Providers keys; a platform key reads the names.
//   The shape is unchanged.
// CHANGELOG (2026-11-18, v1.12.0 pending):
// - IdentityLookupResult has the lookup route's real shape (Found, the
//   identifier fields, Profile, Raw{Provider}, Attempted, Outcome, NotFound,
//   NotFoundBy). Type / Value / Data / Timestamp were never returned and are
//   Deprecated. IdentityOutcome* constants.
// - ErrCodeReplayInProgress, ErrCodeInternalError, ErrCodeCaseAtMlroStage.
// - RasterizePdfInput.PdfBase64 states the real limit (the 4.5 MB request
//   body, about 3.3 MB of PDF), not 10 MB.
// CHANGELOG (2026-11-19, v1.13.0 pending, X12):
// - ErrCodeRestrictedEventType (400, event subscriptions),
//   ErrCodeReservedSearchApp (400, Search.Index), ErrCodeFeedMissing and
//   ErrCodeFeedNotRealtime (409, Fraud.Rules.Transition).
// - EventSubscription.RestrictedEventTypes; EventReplayResult.Replayed (the
//   route answers { replayed }; the three old counts were never sent).
// CHANGELOG (2026-11-19, v1.13.0 pending):
// - ErrCodeVersionMismatch: 409 from Fraud.Rules.TransitionAtVersion and
//   UpdateAtVersion when the rule was edited since the version named.
// CHANGELOG (2026-11-21, pending, spec 03 PM-1):
// - ErrCodeCardNumberRefused: 400 from Fraud.Decide / Fraud.ShadowDecide.
// CHANGELOG (2026-11-23, v1.18.0 pending, AI gateway P6):
// - AiMessagesInput.Actors (map app id -> AiToolActorsEntry: Ref, Permissions,
//   Grant) for Caldera's operator console (a platform key): one person per
//   app, each entry forwarded only to its own app; a nil Permissions is sent
//   as [] — @caldera/platformxe-types 3.20.0.
// CHANGELOG (2026-11-23, v1.17.0 pending, AI gateway P5):
// - AiMessagesInput.Actor (AiToolActor: Ref, Permissions, Grant), the person
//   a first-party Caldera app's own tools act for; AiToolListItem.SourceApp
//   and RequiredPermission — @caldera/platformxe-types 3.19.0.
// - Review round 1 (M-13): a nil AiToolActor.Permissions is sent as [] and
//   never null (AiToolActor.MarshalJSON).
// CHANGELOG (2026-11-22, v1.16.0 pending, AI gateway T4):
// - AiMessagesInput / AiMessagesResult and their parts (AiSystemBlock,
//   AiNativeMessage, AiNativeTool, AiNativeReply, AiStopDetails,
//   AiMessagesUsage), AiToolRunSummary, AiToolListItem / AiToolList,
//   AiCompleteInput.Toolsets, AiCompleteResult.ToolRuns, the AiProfileView
//   governance fields, ErrCodeAiBudgetExceeded / Refused / ToolsetDenied /
//   ToolFailed / LoopLimit — @caldera/platformxe-types 3.18.0 field-for-field.
// - Amended before release (T5 review): AiBudgetReasonUnavailable
//   (Details["reason"] of AI_BUDGET_EXCEEDED when a hard budget could not be
//   checked; a tool-loop refusal has no Details["period"]); the OutputSchema,
//   ToolChoice-with-Toolsets, reply-after-fallback and AI_TOOL_FAILED docs.
//
// CHANGELOG (2026-11-20, v1.14.0 pending, fidelity W3):
// - GeneratePdfInput / GeneratePdfOptions / GeneratePdfMargins /
//   GeneratePdfResult (POST /api/v1/pdf/generate, both answer modes).
// - QrBatchInput / QrBatchItem / QrBatchItemOptions / QrUtmParams and the
//   answer types QrBatchItemResult / QrBatchItemError / QrBatchZip;
//   BatchQRResult gains Items, TotalFailed, Zip, GeneratedAt (Results and
//   Errors were never sent: Deprecated).
// - BillingInvoice, BillingInvoiceLineItem, BillingInvoiceDetail,
//   BillingInvoiceList, ListInvoicesParams (GET /api/v1/billing/invoices).
// - Error codes of those routes (ErrCodeValidationError,
//   ErrCodePayloadTooLarge, ErrCodeTemplateNotFound, ErrCodeRenderTimeout,
//   ErrCodeOutputTooLarge, ErrCodeStorageUnavailable,
//   ErrCodeRendererUnavailable, ErrCodeServiceDisabled, ErrCodeRateLimited,
//   ErrCodeStorageFailed, ErrCodeForbidden, ErrCodeNotFound).
// Mirrors @caldera/platformxe-types 3.16.0 src/pdf.ts and src/invoicing.ts.
// - W3 review: QrBatchItemResult.EntityType / EntityID and BatchQRResult.Zip
//   comments say what the server sends (any string as sent, cut to 128
//   characters on a failed item; Zip nil also when it could not be stored).
// CHANGELOG (2026-11-22, v1.15.0 pending, Enterprise License):
// - LicenseView, LicenseEffectiveEdition, LicenseCoverage and their constants (GET
//   /api/v1/license). Mirrors @caldera/platformxe-types 3.17.0 src/license.ts.
// - ErrCodeLicenseRequired (403 from Fraud.Activate after the licensing
//   cut-over) and ErrCodeContactLicensing (409 from Issues.Create for a
//   plan-change request naming PRO or ENTERPRISE).
// CHANGELOG (2026-11-24, v1.19.0 pending, provided-SSO code review B4):
// - ErrCodeDomainAlreadyRegistered (409) and ErrCodeDomainLimitReached (503)
//   from Domains.Add: POST /api/v1/domains answers them now (it answered 409
//   CONFLICT and a generic 502 PROVIDER_ERROR before).
// =============================================================================

package platformxe

import (
	"encoding/json"
	"strings"
)

// APIResponse is the standard PlatformXe response envelope.
type APIResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   *APIError              `json:"error,omitempty"`
}

// APIError represents an error from the API. Every non-2xx response, and a
// 2xx {"success": false} envelope, is returned as an *APIError (it implements
// error), so callers can errors.As it for the status and code.
type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"` // HTTP status of the response
	// Details is the envelope's error.details object, when the server sent
	// one — e.g. {"path": "/total"} on 422 AI_OUTPUT_INVALID (v1.7.0).
	Details map[string]interface{} `json:"details,omitempty"`
	// RetryAfter is the Retry-After header in seconds, 0 when absent — set on
	// a 429 (RATE_LIMITED, AI_QUOTA_EXCEEDED) (v1.7.0).
	RetryAfter int `json:"-"`
}

// ClientConfig holds configuration for the PlatformXe client.
//
// Retries is the number of retries after the first attempt. Because 0 is
// Go's zero value it means "use the default" (2); a negative value means no
// retries. To configure zero retries explicitly, set DisableRetries, which
// forces a single attempt regardless of Retries.
type ClientConfig struct {
	APIKey         string
	BaseURL        string
	Timeout        int  // seconds, default 10
	Retries        int  // default 2 when 0; negative = no retries
	FailOpen       bool // default false
	DisableRetries bool // true = exactly one attempt (Retries ignored)
}

// -- Pagination --

// Pagination holds standard pagination metadata.
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// -- Delete Result --

// DeleteResult is the standard response for delete operations.
type DeleteResult struct {
	Deleted bool `json:"deleted"`
}

// -- Processor Configuration --

// ProcessorConfig represents the runtime processor configuration for a service.
type ProcessorConfig struct {
	ID             string                 `json:"id"`
	OrganizationID string                 `json:"organizationId"`
	Service        string                 `json:"service"`
	Enabled        bool                   `json:"enabled"`
	Config         map[string]interface{} `json:"config"`
	CreatedAt      string                 `json:"createdAt"`
	UpdatedAt      string                 `json:"updatedAt"`
}

// -- Health --

// HealthCheckResult is the response from the platform health endpoint.
type HealthCheckResult struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// -- Messaging --

// EmailHealthResponse is the health status of the email service. A key with
// an organisation reads each route as its chain position ("primary",
// "fallback_1", ...), never a provider name, and its own organisation's queue
// counts; a platform key reads the route names and the whole queue.
type EmailHealthResponse struct {
	ProviderOrder []string                       `json:"providerOrder"`
	Providers     map[string]EmailProviderStatus `json:"providers"`
	Queue         QueueStats                     `json:"queue"`
}

// EmailProviderStatus is the status of one email route (keyed by its label
// in ProviderOrder). State is its circuit breaker: CLOSED, OPEN or HALF_OPEN.
type EmailProviderStatus struct {
	Configured bool   `json:"configured"`
	State      string `json:"state"`
	Failures   int    `json:"failures"`
	Position   int    `json:"position"`
}

// SmsHealthResponse is the health status of SMS and WhatsApp services.
type SmsHealthResponse struct {
	Sms      ProviderHealthStatus `json:"sms"`
	Whatsapp ProviderHealthStatus `json:"whatsapp"`
}

// ProviderHealthStatus is the health status of a messaging provider.
type ProviderHealthStatus struct {
	Provider    string `json:"provider"`
	Status      string `json:"status"`
	LastChecked string `json:"lastChecked"`
}

// QueueStats holds email queue statistics.
type QueueStats struct {
	Pending    int `json:"pending"`
	Processing int `json:"processing"`
	Sent       int `json:"sent"`
	DeadLetter int `json:"deadLetter"`
}

// QueueProcessResult is the result of processing the email retry queue.
type QueueProcessResult struct {
	Processed    int `json:"processed"`
	Sent         int `json:"sent"`
	Failed       int `json:"failed"`
	DeadLettered int `json:"deadLettered"`
}

// SendMessageResult is the result of sending an email, SMS, or WhatsApp message.
type SendMessageResult struct {
	MessageID string `json:"messageId"`
	// Provider is the chain position that sent the email (primary,
	// fallback_1, ...; none when queued). Email sends only.
	Provider string `json:"provider,omitempty"`
	// Queued is true when every provider failed and the email was queued for
	// retry. Email sends only.
	Queued bool `json:"queued,omitempty"`
	// Deduplicated is true, with Queued, when your identical email (same
	// recipients, subject and content) was already queued: nothing new was
	// stored and the queued one will be sent. Email sends only.
	Deduplicated bool `json:"deduplicated,omitempty"`
}

// Error codes of the email send routes (POST /api/v1/messaging/email/send and
// POST /api/v1/templates/{id}/send) for an email neither sent nor queued, as
// APIError.Code. Each carries RetryAfter (seconds) and is returned at once:
// the client never retries them, so wait RetryAfter before sending again.
const (
	// ErrCodeEmailRateLimited: 429. The sending limit for the recipient was
	// reached; RetryAfter is the time until it frees a place.
	ErrCodeEmailRateLimited = "EMAIL_RATE_LIMITED"
	// ErrCodeEmailNotAccepted: 503. Every route failed and the retry queue
	// could not store the email.
	ErrCodeEmailNotAccepted = "EMAIL_NOT_ACCEPTED"
	// ErrCodeEmailProviderUnavailable: 503. No email route is available.
	ErrCodeEmailProviderUnavailable = "EMAIL_PROVIDER_UNAVAILABLE"
)

// Error codes of the identity DLQ replay and the fraud case transition, as
// APIError.Code (v1.12.0). Neither is retried by the client (409, and a 500
// on a POST, are final).
const (
	// ErrCodeReplayInProgress: 409 from Identity.Dlq.Replay. Another replay of
	// the row is running; nothing was called or billed. Read the row (Dlq.Get)
	// and replay again only if it is still unreplayed.
	ErrCodeReplayInProgress = "REPLAY_IN_PROGRESS"
	// ErrCodeInternalError: 500. From Identity.Dlq.Replay it means the replay
	// did not complete and the row stays unreplayed, so it can be replayed again.
	ErrCodeInternalError = "INTERNAL_ERROR"
	// ErrCodeCaseAtMlroStage: 409 from Fraud.Cases.Transition for a case at an
	// MLRO stage (mlro, sar_filed, le_referred).
	ErrCodeCaseAtMlroStage = "CASE_AT_MLRO_STAGE"
	// ErrCodeVersionMismatch: 409 from Fraud.Rules.TransitionAtVersion or
	// Fraud.Rules.UpdateAtVersion. The rule was edited since the version the
	// caller named, and nothing changed: read the rule again (Get) and review
	// the new version before retrying.
	ErrCodeVersionMismatch = "VERSION_MISMATCH"
	// ErrCodeCardNumberRefused: 400 from Fraud.Decide and Fraud.ShadowDecide.
	// The request carries a card number or a card field (pan, cvv ...) in
	// subject, action, resource, context or the idempotency key (only
	// context.amount.value skips the number test); nothing was
	// decided or recorded. Send a token, never card data. The message names the
	// path, never the value.
	ErrCodeCardNumberRefused = "CARD_NUMBER_REFUSED"
	// ErrCodeFeedMissing: 409 from Fraud.Rules.Transition (publish). The rule
	// reads a feed (an event type) the organisation has not sent recently;
	// the message names the event types.
	ErrCodeFeedMissing = "FEED_MISSING"
	// ErrCodeFeedNotRealtime: 409 from Fraud.Rules.Transition (publish). The
	// rule needs a real-time payment feed and the pack's paymentFeed setting
	// is not realtime (set in the portal's pack settings).
	ErrCodeFeedNotRealtime = "FEED_NOT_REALTIME"
	// ErrCodeRestrictedEventType: 400 from Events.CreateSubscription and
	// UpdateSubscription. eventTypes names a fraud or compliance event that is
	// never delivered (the message lists the entries to remove).
	ErrCodeRestrictedEventType = "RESTRICTED_EVENT_TYPE"
	// ErrCodeReservedSearchApp: 400 from Search.Index. The app "fraud" is
	// reserved for PlatformXe's fraud desk.
	ErrCodeReservedSearchApp = "RESERVED_SEARCH_APP"
)

// -- Webhooks --

// Webhook is a webhook endpoint as the API returns it. It never carries a
// secret (WebhookCreated does, once).
type Webhook struct {
	ID            string   `json:"id"`
	CallerService string   `json:"callerService"`
	Name          string   `json:"name"`
	URL           string   `json:"url"`
	Events        []string `json:"events"`
	// Status is ACTIVE, PAUSED or FAILING.
	Status string `json:"status"`
	// EngineVersion is 1 (the original sender) or 2 (recipient endpoints).
	EngineVersion int `json:"engineVersion"`
	// RecipientRef is set on a v2 recipient endpoint, nil on an untargeted one.
	RecipientRef         *string `json:"recipientRef"`
	TotalDeliveries      int     `json:"totalDeliveries"`
	SuccessfulDeliveries int     `json:"successfulDeliveries"`
	ConsecutiveFailures  int     `json:"consecutiveFailures"`
	LastTriggeredAt      *string `json:"lastTriggeredAt"`
	LastFailureAt        *string `json:"lastFailureAt"`
	LastFailureReason    *string `json:"lastFailureReason"`
	CreatedAt            string  `json:"createdAt"`
	UpdatedAt            string  `json:"updatedAt"`

	// Deprecated: the API never returned this field; endpoints belong to a
	// calling service (CallerService).
	OrganizationID string `json:"organizationId,omitempty"`
	// Deprecated: the API never returned this field; use Status.
	IsActive bool `json:"isActive,omitempty"`
	// Deprecated: no list or get response carries a secret; Create and
	// RotateSecret return it once on WebhookCreated.Secret.
	Secret string `json:"secret,omitempty"`
}

// WebhookCreated is the create / rotate response: the endpoint and its
// signing secret, which is returned ONCE and never retrievable again.
//
// Webhook is embedded (JSON key "webhook"), so the endpoint's fields are
// promoted: wh.ID, wh.Name ... read the endpoint and wh.Secret the one-time
// secret (it shadows the deprecated Webhook.Secret).
type WebhookCreated struct {
	Webhook `json:"webhook"`
	Secret  string `json:"secret"`
}

// WebhookTestResult is the result of a webhook test delivery.
type WebhookTestResult struct {
	Success        bool   `json:"success"`
	ResponseStatus int    `json:"responseStatus,omitempty"`
	ResponseTimeMs int    `json:"responseTimeMs,omitempty"`
	Error          string `json:"error,omitempty"`

	// Deprecated: the API never returned these; use Success / ResponseStatus.
	Delivered bool `json:"delivered,omitempty"`
	// Deprecated: use ResponseStatus.
	StatusCode int `json:"statusCode,omitempty"`
	// Deprecated: use ResponseTimeMs.
	ResponseTime int `json:"responseTime,omitempty"`
	// Deprecated: use Error.
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// WebhookListResponse is a page of webhook endpoints.
type WebhookListResponse struct {
	Items []Webhook `json:"items"`
	Total int       `json:"total"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`

	// Deprecated: the API never returned this field; use Items.
	Webhooks []Webhook `json:"webhooks,omitempty"`
}

// DispatchWebhookInput is the body of Dispatch: one event for ONE recipient.
type DispatchWebhookInput struct {
	RecipientRef string `json:"recipientRef"`
	// EventType is a built-in event type, a custom.* name or any dotted name
	// such as booking.created (case-insensitive, max 200 characters); the
	// webhook. prefix is reserved.
	EventType string `json:"eventType"`
	// EventID is 1-200 visible ASCII characters, no spaces (it is sent as the
	// X-Webhook-Id header), and idempotent per endpoint: the same id never
	// creates a second delivery.
	EventID string `json:"eventId"`
	// Data is required by the API (null is refused); Dispatch sends a nil
	// map as {}.
	Data map[string]interface{} `json:"data"`
}

// DispatchWebhookResult is the outcome of Dispatch. The API answers once the
// delivery rows are recorded; the attempts run after the response.
type DispatchWebhookResult struct {
	// Targeted is the number of the recipient's v2 endpoints that subscribe to
	// the event and are not PAUSED (ACTIVE or FAILING).
	Targeted int `json:"targeted"`
	// Created is the number of delivery rows this call created (0 on a repeat).
	Created int `json:"created"`
}

// WebhookDelivery is one row of a delivery log. It never carries the payload
// or a secret; ResponseSnippet is redacted and capped at 2 KB, and always nil
// for an engine-1 endpoint's rows.
type WebhookDelivery struct {
	ID              string  `json:"id"`
	EventType       string  `json:"eventType"`
	EventID         string  `json:"eventId"`
	Status          string  `json:"status"` // PENDING, SUCCESS, RETRYING or FAILED
	AttemptCount    int     `json:"attemptCount"`
	ResponseStatus  *int    `json:"responseStatus"`
	ResponseTimeMs  *int    `json:"responseTimeMs"`
	ResponseSnippet *string `json:"responseSnippet"`
	NextRetryAt     *string `json:"nextRetryAt"`
	DeliveredAt     *string `json:"deliveredAt"`
	CreatedAt       string  `json:"createdAt"`
}

// WebhookDeliveriesResponse is a page of a webhook's delivery log.
type WebhookDeliveriesResponse struct {
	Items []WebhookDelivery `json:"items"`
	Total int               `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
}

// RedeliverWebhookResult is the outcome of one redelivery attempt.
type RedeliverWebhookResult struct {
	DeliveryID     string  `json:"deliveryId"`
	OK             bool    `json:"ok"`
	Status         *int    `json:"status"`
	ResponseTimeMs int     `json:"responseTimeMs"`
	Reason         *string `json:"reason"` // URL_REFUSED, TIMEOUT, NETWORK, HTTP or REDIRECT
}

// -- Templates --

// Template represents an email/message template.
type Template struct {
	ID             string                   `json:"id"`
	OrganizationID string                   `json:"organizationId"`
	Name           string                   `json:"name"`
	Subject        string                   `json:"subject"`
	BodyBlocks     []map[string]interface{} `json:"bodyBlocks"`
	Variables      []string                 `json:"variables"`
	CreatedAt      string                   `json:"createdAt"`
	UpdatedAt      string                   `json:"updatedAt"`
}

// TemplateListResponse wraps a list of templates.
type TemplateListResponse struct {
	Templates []Template `json:"templates"`
}

// RenderResult is the result of rendering a template.
type RenderResult struct {
	Subject string `json:"subject"`
	HTML    string `json:"html"`
	Text    string `json:"text"`
}

// TemplateSendResult is the result of rendering and sending a template.
type TemplateSendResult struct {
	Rendered TemplateSendRendered `json:"rendered"`
	Send     TemplateSendInfo     `json:"send"`
}

// TemplateSendRendered holds the rendered fields of a template send.
type TemplateSendRendered struct {
	Subject string `json:"subject"`
}

// TemplateSendInfo holds the send result of a template send.
type TemplateSendInfo struct {
	MessageID string `json:"messageId,omitempty"`
	Provider  string `json:"provider"`
	Queued    bool   `json:"queued"`
}

// -- Sending Domains --

// SendingDomain represents a registered sending domain.
type SendingDomain struct {
	ID             string      `json:"id"`
	OrganizationID string      `json:"organizationId"`
	Domain         string      `json:"domain"`
	Status         string      `json:"status"`
	DNSRecords     []DNSRecord `json:"dnsRecords"`
	VerifiedAt     *string     `json:"verifiedAt"`
	CreatedAt      string      `json:"createdAt"`
	UpdatedAt      string      `json:"updatedAt"`
}

// DNSRecord represents a DNS record for domain verification.
type DNSRecord struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Verified bool   `json:"verified"`
}

// DomainListResponse wraps a list of sending domains.
type DomainListResponse struct {
	Domains []SendingDomain `json:"domains"`
}

// DomainVerifyResult is the result of a domain verification check.
type DomainVerifyResult struct {
	Domain      SendingDomain `json:"domain"`
	AllVerified bool          `json:"allVerified"`
}

// -- Event Subscriptions --

// EventSubscription represents an event subscription.
type EventSubscription struct {
	ID             string   `json:"id"`
	OrganizationID string   `json:"organizationId"`
	Name           string   `json:"name"`
	WebhookURL     string   `json:"webhookUrl"`
	EventTypes     []string `json:"eventTypes"`
	// Status is ACTIVE, PAUSED or FAILING.
	Status              string  `json:"status"`
	ConsecutiveFailures int     `json:"consecutiveFailures"`
	LastTriggeredAt     *string `json:"lastTriggeredAt"`
	// RestrictedEventTypes are the stored EventTypes entries that name a
	// restricted (fraud or compliance) type: never delivered, and a PATCH
	// re-sending them is refused. Empty when none; not set on the create
	// answer.
	RestrictedEventTypes []string `json:"restrictedEventTypes,omitempty"`
	// Secret is the signing secret. CreateSubscription sets it (the API
	// returns it once, at creation); every other call leaves it empty.
	Secret    string `json:"secret,omitempty"`
	CreatedAt string `json:"createdAt"`
	// Deprecated: the API never returns this field.
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// EventSubscriptionListResponse is GET /api/v1/event-subscriptions.
type EventSubscriptionListResponse struct {
	Items []EventSubscription `json:"items"`
	Total int                 `json:"total"`
	// Deprecated: the API returns the list as Items. ListSubscriptions copies
	// Items here so v1 callers that read Subscriptions keep working.
	Subscriptions []EventSubscription `json:"-"`
}

// EventReplayResult is the result of replaying events.
type EventReplayResult struct {
	// Replayed is the number of events re-sent to the subscription.
	Replayed int `json:"replayed"`
	// Deprecated: never returned by the API; use Replayed.
	ReplayedCount int `json:"replayedCount"`
	// Deprecated: never returned by the API.
	SuccessCount int `json:"successCount"`
	// Deprecated: never returned by the API.
	FailedCount int `json:"failedCount"`
}

// EventLogEntry represents a single event log entry.
type EventLogEntry struct {
	ID        string                 `json:"id"`
	EventType string                 `json:"eventType"`
	SourceApp string                 `json:"sourceApp"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt string                 `json:"createdAt"`
}

// EventLogResponse is the paginated event log response.
type EventLogResponse struct {
	Items []EventLogEntry `json:"items"`
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}

// EmitEventResult is the POST /api/v1/events (Events.Ingest) answer: the
// workflow engine's outcome, one Results item per trigger evaluated.
type EmitEventResult struct {
	EventType         string              `json:"eventType"`
	EntityType        string              `json:"entityType,omitempty"`
	EntityID          string              `json:"entityId,omitempty"`
	SourceApp         string              `json:"sourceApp"`
	TriggersEvaluated int                 `json:"triggersEvaluated"`
	TriggersMatched   int                 `json:"triggersMatched"`
	TriggersExecuted  int                 `json:"triggersExecuted"`
	Results           []TriggerExecResult `json:"results"`
}

// TriggerExecResult is the result of a single trigger execution.
type TriggerExecResult struct {
	TriggerID     string `json:"triggerId"`
	TriggerName   string `json:"triggerName"`
	Matched       bool   `json:"matched"`
	Executed      bool   `json:"executed"`
	ExecutionID   string `json:"executionId,omitempty"`
	SkippedReason string `json:"skippedReason,omitempty"`
	Error         string `json:"error,omitempty"`
}

// -- Storage --

// StorageFile represents a stored media file.
type StorageFile struct {
	ID               string `json:"id"`
	CallerService    string `json:"callerService"`
	Module           string `json:"module"`
	EntityID         string `json:"entityId"`
	OriginalFilename string `json:"originalFilename"`
	MimeType         string `json:"mimeType"`
	FileSize         int64  `json:"fileSize"`
	StorageProvider  string `json:"storageProvider"`
	PublicURL        string `json:"publicUrl"`
	DisplayOrder     int    `json:"displayOrder"`
	ModerationStatus string `json:"moderationStatus"`
	MediaType        string `json:"mediaType"`
	Width            *int   `json:"width,omitempty"`
	Height           *int   `json:"height,omitempty"`
	CreatedAt        string `json:"createdAt"`
}

// StorageFileListResponse wraps a list of storage files.
type StorageFileListResponse struct {
	Files []StorageFile `json:"files"`
}

// SignUploadResult is the result of signing an upload URL.
type SignUploadResult struct {
	UploadURL string `json:"uploadUrl"`
	FileID    string `json:"fileId"`
}

// ReorderFilesResult is the result of reordering files.
type ReorderFilesResult struct {
	Reordered bool `json:"reordered"`
}

// Document represents a fixed-storage document.
type Document struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	Category     string `json:"category"`
	FileType     string `json:"fileType"`
	Space        string `json:"space"`
	FolderID     string `json:"folderId,omitempty"`
	OwnerType    string `json:"ownerType"`
	OwnerID      string `json:"ownerId"`
	AccessLevel  string `json:"accessLevel"`
	Visibility   string `json:"visibility"`
	FileURL      string `json:"fileUrl,omitempty"`
	FileName     string `json:"fileName,omitempty"`
	FileMimeType string `json:"fileMimeType,omitempty"`
	FileSize     int64  `json:"fileSize,omitempty"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// DocumentListResponse wraps a list of documents.
type DocumentListResponse struct {
	Documents []Document `json:"documents"`
}

// DocumentFolder represents a storage folder, scoped to one (OwnerType, OwnerID)
// pair — there is no folder hierarchy or nesting. GET .../folders?ownerType=&ownerId=
// and GET .../folders/:id both return this shape.
type DocumentFolder struct {
	ID                string   `json:"id"`
	OwnerType         string   `json:"ownerType"`
	OwnerID           string   `json:"ownerId"`
	CallerService     string   `json:"callerService"`
	StorageLimitBytes int64    `json:"storageLimitBytes"`
	StorageUsedBytes  int64    `json:"storageUsedBytes"`
	MaxFileCount      int      `json:"maxFileCount"`
	MaxFileSizeBytes  int64    `json:"maxFileSizeBytes"`
	AllowedCategories []string `json:"allowedCategories"`
	DocumentCount     int      `json:"documentCount"`
	UsagePercentage   int      `json:"usagePercentage"`
	RemainingBytes    int64    `json:"remainingBytes"`
	CanUpload         bool     `json:"canUpload"`

	// Deprecated: Name and CreatedAt were declared before v1.6 but the API
	// has never returned them for a folder — they are always empty. Kept only
	// so existing v1 code that references them still compiles.
	Name      string `json:"name,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

// FolderListResponse wraps a list of folders.
//
// Deprecated: returned only by the deprecated ListFolders, which always
// errors — the API has no list-folders route. Use GetFolderByOwner.
type FolderListResponse struct {
	Folders []DocumentFolder `json:"folders"`
}

// CreateFolderResult is the response from creating a folder: just the new
// (or, if one already existed for this owner, existing) folder's ID.
type CreateFolderResult struct {
	FolderID string `json:"folderId"`
}

// -- Workflows --

// WorkflowTrigger represents a workflow automation trigger.
type WorkflowTrigger struct {
	ID               string                   `json:"id"`
	Name             string                   `json:"name"`
	Description      string                   `json:"description,omitempty"`
	Code             string                   `json:"code"`
	TriggerType      string                   `json:"triggerType"`
	EventType        string                   `json:"eventType,omitempty"`
	CronExpression   string                   `json:"cronExpression,omitempty"`
	Conditions       map[string]interface{}   `json:"conditions,omitempty"`
	EntityType       string                   `json:"entityType,omitempty"`
	Actions          []map[string]interface{} `json:"actions"`
	IsActive         bool                     `json:"isActive"`
	Priority         int                      `json:"priority"`
	CooldownMinutes  *int                     `json:"cooldownMinutes,omitempty"`
	MaxExecutionsDay *int                     `json:"maxExecutionsDay,omitempty"`
	SourceApp        string                   `json:"sourceApp,omitempty"`
	CreatedAt        string                   `json:"createdAt"`
	UpdatedAt        string                   `json:"updatedAt"`
}

// WorkflowListResponse is GET /api/v1/workflows: the triggers and the page.
type WorkflowListResponse struct {
	// Triggers is the page of triggers the route answers.
	Triggers []WorkflowTrigger `json:"triggers"`
	// Pagination is the page asked for and the total.
	Pagination *Pagination `json:"pagination,omitempty"`
	// Deprecated: the route answers triggers; the SDK copies them here.
	Workflows []WorkflowTrigger `json:"workflows,omitempty"`
}

// WorkflowEvalResult is the result of evaluating workflows against an event.
type WorkflowEvalResult struct {
	Matched int                  `json:"matched"`
	Total   int                  `json:"total"`
	Results []WorkflowEvalDetail `json:"results"`
	// Message is set when no trigger is active for the event.
	Message string `json:"message,omitempty"`
}

// WorkflowEvaluateInput is the POST /api/v1/workflows/evaluate request. EventType is
// required; EntityData is what the triggers' conditions and {{field}} templates read.
type WorkflowEvaluateInput struct {
	EventType  string                 `json:"eventType"`
	EntityType string                 `json:"entityType,omitempty"`
	EntityID   string                 `json:"entityId,omitempty"`
	EntityData map[string]interface{} `json:"entityData,omitempty"`
	SourceApp  string                 `json:"sourceApp,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	// DryRun reports which triggers would run and their actions, and runs nothing.
	DryRun bool `json:"dryRun,omitempty"`
}

// WorkflowEvalDetail is the evaluation result for a single workflow trigger.
// ExecutionID is empty when no execution could be recorded (Error is then set).
type WorkflowEvalDetail struct {
	TriggerID   string `json:"triggerId"`
	TriggerCode string `json:"triggerCode"`
	TriggerName string `json:"triggerName"`
	Matched     bool   `json:"matched"`
	Reason      string `json:"reason,omitempty"`
	DryRun      bool   `json:"dryRun,omitempty"`
	// ActionsWouldExecute is a dry run's stored actions (set instead of an execution).
	ActionsWouldExecute []map[string]interface{} `json:"actionsWouldExecute,omitempty"`
	ExecutionID         string                   `json:"executionId,omitempty"`
	DurationMs          int                      `json:"durationMs,omitempty"`
	ActionsExecuted     []WorkflowExecutedAction `json:"actionsExecuted,omitempty"`
	Error               string                   `json:"error,omitempty"`
}

// WorkflowExecutedActionStatus is the outcome of one action in an evaluate result.
type WorkflowExecutedActionStatus string

const (
	// WorkflowActionSucceeded: the action ran (an email sent or queued for retry).
	WorkflowActionSucceeded WorkflowExecutedActionStatus = "SUCCESS"
	// WorkflowActionFailed: the action failed; a run whose every action failed counts toward auto-disable.
	WorkflowActionFailed WorkflowExecutedActionStatus = "FAILED"
	// WorkflowActionSkipped: not run (an unknown type, or no recipient resolved). In a recorded
	// event or scheduled run, also an action not reached before its 30 s limit; an evaluation has no limit.
	WorkflowActionSkipped WorkflowExecutedActionStatus = "SKIPPED"
	// WorkflowActionDeferred: the email was declined for capacity; not retried, never counted.
	WorkflowActionDeferred WorkflowExecutedActionStatus = "DEFERRED"
)

// WorkflowExecutedAction is a stored action and its outcome, as /api/v1/workflows/evaluate returns it.
type WorkflowExecutedAction struct {
	Type   string                       `json:"type"`
	Config map[string]interface{}       `json:"config,omitempty"`
	Status WorkflowExecutedActionStatus `json:"status"`
	Error  string                       `json:"error,omitempty"`
}

// WorkflowActionResultStatus is the outcome of one step in a recorded execution. Its spelling
// follows the engine that ran it: lower case (the constants below) for an event or scheduled
// run, upper case (as WorkflowExecutedActionStatus) for a run recorded by
// /api/v1/workflows/evaluate or the event relay. Compare through Normalize.
type WorkflowActionResultStatus string

// Normalize returns the status in the upper-case spelling, whichever engine recorded it.
func (s WorkflowActionResultStatus) Normalize() WorkflowExecutedActionStatus {
	return WorkflowExecutedActionStatus(strings.ToUpper(string(s)))
}

const (
	WorkflowStepSucceeded WorkflowActionResultStatus = "success"
	WorkflowStepFailed    WorkflowActionResultStatus = "failed"
	WorkflowStepSkipped   WorkflowActionResultStatus = "skipped"
	WorkflowStepDeferred  WorkflowActionResultStatus = "deferred"
)

// -- Permissions --

// PermissionRole represents a permission role.
type PermissionRole struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	Model          string `json:"model"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

// RoleListResponse wraps a list of permission roles.
type RoleListResponse struct {
	Roles []PermissionRole `json:"roles"`
}

// PermissionCheckResult is the result of a single permission check.
type PermissionCheckResult struct {
	Allowed     bool   `json:"allowed"`
	Source      string `json:"source"`
	RoleID      string `json:"roleId,omitempty"`
	EvaluatedAt string `json:"evaluatedAt"`
}

// BatchCheckResponse wraps a list of batch permission check results. The
// route answers the list itself; CheckBatch puts it in Results.
type BatchCheckResponse struct {
	Results []BatchCheckResult `json:"results"`
}

// BatchCheckResult is the result of a single check within a batch.
type BatchCheckResult struct {
	AdminID string `json:"adminId"`
	Path    string `json:"path"`
	Action  string `json:"action"`
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
	// Source is where the decision came from (v1.11.0), e.g. "role",
	// "override_grant", "default".
	Source string `json:"source,omitempty"`
	// RoleID is the deciding role, when one decided (v1.11.0).
	RoleID string `json:"roleId,omitempty"`
}

// ResolvedCapabilities is the full resolved permission set for an admin.
type ResolvedCapabilities struct {
	AdminID string `json:"adminId"`
	// RoleCapabilities are the role's own capabilities, before overrides.
	RoleCapabilities []string `json:"roleCapabilities,omitempty"`
	// Capabilities are the effective capabilities after overrides. They may
	// still contain role wildcards ("*:*", "path:*") that a deny overrides.
	Capabilities []string `json:"capabilities"`
	// DeniedCapabilities ("path:action") are live DENY overrides. A local
	// evaluator MUST refuse any request one covers — same action, same path or
	// a sub-path, case-insensitive, trailing '/' ignored — before looking at
	// Capabilities or Permissions.
	DeniedCapabilities []string             `json:"deniedCapabilities,omitempty"`
	Permissions        []ResolvedPermission `json:"permissions,omitempty"`
	TTL                int                  `json:"ttl"`
}

// ResolvedPermission represents a resolved module permission.
type ResolvedPermission struct {
	ModuleID string   `json:"moduleId"`
	Actions  []string `json:"actions"`
}

// RoleCapabilities holds capabilities for a role (Simple model).
type RoleCapabilities struct {
	RoleID       string   `json:"roleId"`
	Capabilities []string `json:"capabilities"`
}

// RoleModulePermissions holds module permissions for a role (Full model).
type RoleModulePermissions struct {
	RoleID  string             `json:"roleId"`
	Modules []ModulePermission `json:"modules"`
}

// ModulePermission represents a single module permission assignment.
type ModulePermission struct {
	ModuleID string   `json:"moduleId"`
	Actions  []string `json:"actions"`
}

// PermissionOverride represents an admin-level permission override.
type PermissionOverride struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organizationId"`
	AdminID        string  `json:"adminId"`
	Path           string  `json:"path"`
	Action         string  `json:"action"`
	Effect         string  `json:"effect"`
	Reason         string  `json:"reason"`
	ExpiresAt      *string `json:"expiresAt,omitempty"`
	CreatedAt      string  `json:"createdAt"`
	// UpdatedAt moves on every write; CreatedAt stays the first creation.
	UpdatedAt string `json:"updatedAt,omitempty"`
	CreatedBy string `json:"createdBy"`
}

// OverrideListResponse wraps a list of permission overrides.
type OverrideListResponse struct {
	Overrides []PermissionOverride `json:"overrides"`
}

// PermissionModule represents a registered permission module.
type PermissionModule struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Actions     []string `json:"actions"`
}

// ModuleListResponse wraps a list of permission modules.
type ModuleListResponse struct {
	Modules []PermissionModule `json:"modules"`
}

// ResourcePolicy represents an ABAC resource policy.
type ResourcePolicy struct {
	ID             string                 `json:"id"`
	OrganizationID string                 `json:"organizationId"`
	Path           string                 `json:"path"`
	Action         string                 `json:"action"`
	Condition      map[string]interface{} `json:"condition"`
	Effect         string                 `json:"effect"`
	Priority       int                    `json:"priority,omitempty"`
	Description    string                 `json:"description,omitempty"`
	IsActive       bool                   `json:"isActive"`
	CreatedAt      string                 `json:"createdAt"`
	UpdatedAt      string                 `json:"updatedAt"`
}

// PolicyListResponse wraps a list of resource policies.
type PolicyListResponse struct {
	Policies []ResourcePolicy `json:"policies"`
}

// RelationshipTuple represents a ReBAC relationship tuple.
type RelationshipTuple struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organizationId"`
	ObjectType     string `json:"objectType"`
	ObjectID       string `json:"objectId"`
	Relation       string `json:"relation"`
	SubjectType    string `json:"subjectType"`
	SubjectID      string `json:"subjectId"`
	CreatedAt      string `json:"createdAt"`
}

// RelationshipListResponse wraps a list of relationship tuples.
type RelationshipListResponse struct {
	Relationships []RelationshipTuple `json:"relationships"`
}

// RelationshipUpdateResult is the result of a relationship batch operation.
type RelationshipUpdateResult struct {
	Created int `json:"created"`
	Deleted int `json:"deleted"`
	Errors  int `json:"errors"`
}

// PermissionAuditLog represents a permission decision audit log entry.
type PermissionAuditLog struct {
	ID        string                 `json:"id"`
	AdminID   string                 `json:"adminId"`
	Path      string                 `json:"path"`
	Action    string                 `json:"action"`
	Allowed   bool                   `json:"allowed"`
	Source    string                 `json:"source"`
	Context   map[string]interface{} `json:"context"`
	LatencyMs int                    `json:"latencyMs"`
	CreatedAt string                 `json:"createdAt"`
}

// AuditLogListResponse wraps a paginated list of audit logs.
type AuditLogListResponse struct {
	Items []PermissionAuditLog `json:"items"`
	Total int                  `json:"total"`
	Page  int                  `json:"page"`
	Limit int                  `json:"limit"`
}

// PermissionChangeLog represents a permission mutation change log entry.
type PermissionChangeLog struct {
	ID             string                 `json:"id"`
	OrganizationID string                 `json:"organizationId"`
	EntityType     string                 `json:"entityType"`
	EntityID       string                 `json:"entityId"`
	Changes        map[string]interface{} `json:"changes"`
	ChangedBy      string                 `json:"changedBy"`
	Timestamp      string                 `json:"timestamp"`
}

// ChangeLogListResponse wraps a paginated list of change logs.
type ChangeLogListResponse struct {
	Items []PermissionChangeLog `json:"items"`
	Total int                   `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}

// AuditExportResult is the result of exporting audit logs.
type AuditExportResult struct {
	Decisions   []PermissionAuditLog  `json:"decisions"`
	Changes     []PermissionChangeLog `json:"changes"`
	PeriodStart string                `json:"periodStart,omitempty"`
	PeriodEnd   string                `json:"periodEnd,omitempty"`
	ExportedAt  string                `json:"exportedAt"`
}

// ShadowCheckResult is the result of a shadow permission check.
type ShadowCheckResult struct {
	AdminID        string `json:"adminId"`
	Path           string `json:"path"`
	Action         string `json:"action"`
	LocalDecision  bool   `json:"localDecision"`
	RemoteDecision bool   `json:"remoteDecision"`
	Match          bool   `json:"match"`
	Discrepancy    bool   `json:"discrepancy"`
}

// -- Federation --

// FederationGroup represents a federation group.
type FederationGroup struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	OwnerOrgID string             `json:"ownerOrgId"`
	Members    []FederationMember `json:"members"`
	CreatedAt  string             `json:"createdAt"`
}

// FederationGroupListResponse wraps a list of federation groups.
type FederationGroupListResponse struct {
	Groups []FederationGroup `json:"groups"`
}

// FederationMemberStatus is the membership lifecycle state. A member is
// INVITED until it accepts; only ACTIVE members are pulled from or pushed to.
type FederationMemberStatus = string

const (
	// FederationMemberInvited — invited by the owner, not yet accepted.
	FederationMemberInvited FederationMemberStatus = "INVITED"
	// FederationMemberActive — accepted by the member organization.
	FederationMemberActive FederationMemberStatus = "ACTIVE"
)

// FederationMember represents a member of a federation group.
type FederationMember struct {
	ID             string                 `json:"id"`
	GroupID        string                 `json:"groupId"`
	OrganizationID string                 `json:"organizationId"`
	Prefix         string                 `json:"prefix"`
	Role           string                 `json:"role"`
	Status         FederationMemberStatus `json:"status"`
	JoinedAt       string                 `json:"joinedAt"`
}

// FederationMemberResult is the result of inviting or accepting a membership.
type FederationMemberResult struct {
	ID             string                 `json:"id"`
	OrganizationID string                 `json:"organizationId"`
	GroupID        string                 `json:"groupId"`
	Prefix         string                 `json:"prefix"`
	Role           string                 `json:"role"`
	Status         FederationMemberStatus `json:"status"`
	JoinedAt       string                 `json:"joinedAt"`
}

// FederationRemoveResult is the result of removing a member from a federation group.
type FederationRemoveResult struct {
	OrganizationID string `json:"organizationId"`
	GroupID        string `json:"groupId"`
	RemovedAt      string `json:"removedAt"`
}

// FederationPullResult is the result of pulling modules from a federation group.
type FederationPullResult struct {
	OrganizationID string                 `json:"organizationId"`
	Modules        []FederationModuleInfo `json:"modules"`
}

// FederationModuleInfo describes a module pulled from a federation group.
type FederationModuleInfo struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Paths []string `json:"paths"`
}

// FederationPushResult is the result of pushing permissions to federation members.
type FederationPushResult struct {
	OrganizationID string                `json:"organizationId"`
	AdminID        string                `json:"adminId"`
	Permissions    []FederationPermEntry `json:"permissions"`
	Status         string                `json:"status"`
	Error          string                `json:"error,omitempty"`
}

// FederationPermEntry describes a single permission entry in a federation push.
type FederationPermEntry struct {
	Path   string `json:"path"`
	Action string `json:"action"`
}

// FederationStatusResult is the result of querying federation group status.
type FederationStatusResult struct {
	Group       FederationGroupInfo `json:"group"`
	Members     []FederationMember  `json:"members"`
	RecentSyncs []FederationSync    `json:"recentSyncs"`
}

// FederationGroupInfo is a summary of a federation group.
type FederationGroupInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	OwnerOrgID string `json:"ownerOrgId"`
	CreatedAt  string `json:"createdAt"`
}

// FederationSync represents a federation sync log entry.
type FederationSync struct {
	ID        string                 `json:"id"`
	GroupID   string                 `json:"groupId"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// -- Identity --

// IdentityResolveResult is the result of resolving an identity.
type IdentityResolveResult struct {
	Resolved bool `json:"resolved"`
	// Identifiers are the linked identifiers (v1.12.0).
	Identifiers *IdentityResolvedIdentifiers `json:"identifiers,omitempty"`
	// Profile is the merged registry profile; its photo is always null (v1.12.0).
	Profile map[string]interface{} `json:"profile,omitempty"`
	// Confidence is 0..1 (v1.12.0).
	Confidence float64 `json:"confidence"`
	// Provider is the provider label; "none" when no provider answered (v1.12.0).
	Provider string `json:"provider"`
	// ResolvedAt is when the answer was made (v1.12.0).
	ResolvedAt string `json:"resolvedAt"`
	// Outcome is "found", "not_found" (a provider answered: billed) or
	// "provider_down" (no provider could answer: not billed) (v1.12.0).
	Outcome string `json:"outcome"`
	// Deprecated: the API never returns linkedIdentifiers; read Identifiers.
	LinkedIdentifiers []IdentityIdentifier `json:"linkedIdentifiers"`
	// Deprecated: the API never returns metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// IdentityResolvedIdentifiers are the identifiers a resolve links (null when unknown).
type IdentityResolvedIdentifiers struct {
	NIN             *string  `json:"nin"`
	BVN             *string  `json:"bvn"`
	TIN             *string  `json:"tin"`
	PassportNumber  *string  `json:"passportNumber"`
	VoterCardNumber *string  `json:"voterCardNumber"`
	DriversLicense  *string  `json:"driversLicense"`
	PhoneNumbers    []string `json:"phoneNumbers"`
}

// IdentityIdentifier represents a linked identity identifier.
type IdentityIdentifier struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// IdentityVerifyResult is the result of verifying an identity.
type IdentityVerifyResult struct {
	Verified   bool    `json:"verified"`
	MatchScore float64 `json:"matchScore"`
	// Profile is the provider's profile, nil when nothing matched (v1.8.0).
	Profile map[string]interface{} `json:"profile,omitempty"`
	// Provider is the internal provider label; "none" when the chain was
	// exhausted (v1.8.0).
	Provider string `json:"provider,omitempty"`
	// AttemptedProviders lists the providers tried when Provider is "none".
	AttemptedProviders []string `json:"attemptedProviders,omitempty"`
	// Deprecated: never returned by the API.
	MatchDetails *IdentityMatchDetails `json:"matchDetails,omitempty"`
}

// IdentityMatchDetails holds name match details for identity verification.
type IdentityMatchDetails struct {
	FirstName IdentityFieldMatch `json:"firstName"`
	LastName  IdentityFieldMatch `json:"lastName"`
}

// IdentityFieldMatch represents the match result for a single field.
type IdentityFieldMatch struct {
	Match      bool    `json:"match"`
	Confidence float64 `json:"confidence"`
}

// IdentityLookupResult is GET /api/v1/identity/lookup's answer (v1.12.0: the
// real shape). The identifier fields are set when the provider returned them.
type IdentityLookupResult struct {
	// Found is true when a provider found the identifier.
	Found           bool     `json:"found"`
	NIN             *string  `json:"nin,omitempty"`
	BVN             *string  `json:"bvn,omitempty"`
	TIN             *string  `json:"tin,omitempty"`
	PassportNumber  *string  `json:"passportNumber,omitempty"`
	VoterCardNumber *string  `json:"voterCardNumber,omitempty"`
	DriversLicense  *string  `json:"driversLicense,omitempty"`
	PhoneNumbers    []string `json:"phoneNumbers,omitempty"`
	// Profile is the registry profile, nil when there is none. It never
	// carries a photo (its photo field is always null).
	Profile map[string]interface{} `json:"profile"`
	// Raw is the provider's name only ({"provider": "..."}; "none" when no
	// provider answered), never the provider's payload.
	Raw IdentityLookupRaw `json:"raw"`
	// Attempted lists the providers tried when Found is false.
	Attempted []string `json:"attempted,omitempty"`
	// Outcome is "found", "not_found" (a provider answered that the identifier
	// is not on the register: billed) or "provider_down" (no provider could
	// answer: not billed).
	Outcome string `json:"outcome"`
	// NotFound is true with Outcome "not_found"; NotFoundBy names the provider.
	NotFound   bool   `json:"notFound,omitempty"`
	NotFoundBy string `json:"notFoundBy,omitempty"`

	// Deprecated: never returned by the API (always empty); read the fields above.
	Type string `json:"type,omitempty"`
	// Deprecated: never returned by the API (always empty).
	Value string `json:"value,omitempty"`
	// Deprecated: never returned by the API (always nil).
	Data map[string]interface{} `json:"data,omitempty"`
	// Deprecated: never returned by the API (always empty).
	Timestamp string `json:"timestamp,omitempty"`
}

// IdentityLookupRaw is the lookup answer's raw field: the provider's name only.
type IdentityLookupRaw struct {
	Provider string `json:"provider"`
}

// The identity lookup and resolve outcomes (IdentityLookupResult.Outcome,
// IdentityResolveResult.Outcome).
const (
	IdentityOutcomeFound        = "found"
	IdentityOutcomeNotFound     = "not_found"
	IdentityOutcomeProviderDown = "provider_down"
)

// IdentityProviderInfo is one entry of GET /api/v1/identity/providers.
type IdentityProviderInfo struct {
	// Name is the internal provider label ("unknown" when the health probe threw).
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`

	// Deprecated: never returned by the API; use Healthy.
	Status string `json:"status,omitempty"`
	// Deprecated: never returned by the API.
	LastChecked string `json:"lastChecked,omitempty"`
	// Deprecated: never returned by the API.
	ResponseTime *int `json:"responseTime,omitempty"`
}

// IdentityProvidersResponse wraps a list of identity providers.
type IdentityProvidersResponse struct {
	Providers []IdentityProviderInfo `json:"providers"`
}

// -- OCR --

// IdentityDocVerifyResult is the result of OCR identity document verification.
type IdentityDocVerifyResult struct {
	Verified        bool              `json:"verified"`
	Confidence      float64           `json:"confidence"`
	Names           DocVerifyNames    `json:"names"`
	DocumentDetails map[string]string `json:"documentDetails,omitempty"`
}

// DocVerifyNames holds name matching details for document verification.
type DocVerifyNames struct {
	DocumentName string  `json:"documentName"`
	ProfileName  string  `json:"profileName"`
	MatchScore   float64 `json:"matchScore"`
}

// -- QR --

// QRCodeResult is the result of generating a QR code.
type QRCodeResult struct {
	QR     string `json:"qr"`
	Format string `json:"format"`
}

// -- Telemetry --

// TelemetryBatch is the full telemetry payload required by the API.
type TelemetryBatch struct {
	SourceApp      string            `json:"sourceApp"`
	PackageName    string            `json:"packageName"`
	PackageVersion string            `json:"packageVersion,omitempty"`
	Metrics        []TelemetryMetric `json:"metrics"`
	PeriodStart    string            `json:"periodStart"`
	PeriodEnd      string            `json:"periodEnd"`
}

// TelemetryMetric is a single metric in a telemetry batch.
type TelemetryMetric struct {
	MetricType string `json:"metricType"`
	MetricKey  string `json:"metricKey"`
	Count      int    `json:"count"`
}

// TelemetryResult is the result of sending a telemetry batch.
type TelemetryResult struct {
	Ingested    int    `json:"ingested"`
	SourceApp   string `json:"sourceApp"`
	PackageName string `json:"packageName"`
}

// -- Exports --

// ExportResult is the result of creating or querying a data export.
type ExportResult struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	DataType    string `json:"dataType"`
	Format      string `json:"format"`
	DownloadURL string `json:"downloadUrl,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
}

// -- Usage --

// UsageSummary is the usage metrics response for a billing month.
type UsageSummary struct {
	Month               string         `json:"month"`
	Services            []ServiceUsage `json:"services"`
	TotalBillableEvents int            `json:"totalBillableEvents"`
}

// ServiceUsage represents usage for a single service.
type ServiceUsage struct {
	Service     string  `json:"service"`
	Used        int     `json:"used"`
	Limit       int     `json:"limit"`
	PercentUsed float64 `json:"percentUsed"`
}

// -- Threads --

// Thread represents a messaging thread.
type Thread struct {
	ID             string                  `json:"id"`
	OrganizationID string                  `json:"organizationId"`
	ChannelID      string                  `json:"channelId"`
	ChannelSlug    string                  `json:"channelSlug"`
	EntityID       string                  `json:"entityId"`
	Subject        string                  `json:"subject"`
	Status         string                  `json:"status"`
	Priority       string                  `json:"priority"`
	Metadata       map[string]interface{}  `json:"metadata,omitempty"`
	Participants   []ThreadParticipantInfo `json:"participants"`
	CreatedAt      string                  `json:"createdAt"`
	UpdatedAt      string                  `json:"updatedAt"`
	ClosedAt       *string                 `json:"closedAt,omitempty"`
}

// ThreadParticipantInfo represents a participant in a thread response.
type ThreadParticipantInfo struct {
	ID          string `json:"id"`
	Role        string `json:"role"`
	ExternalID  string `json:"externalId"`
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
	JoinedAt    string `json:"joinedAt"`
}

// ThreadListResponse wraps a paginated list of threads.
type ThreadListResponse struct {
	Threads []Thread `json:"threads"`
	Total   int      `json:"total"`
	Page    int      `json:"page"`
	Limit   int      `json:"limit"`
}

// ThreadMessage represents a message in a thread.
type ThreadMessage struct {
	ID               string   `json:"id"`
	ThreadID         string   `json:"threadId"`
	SenderExternalID string   `json:"senderExternalId"`
	SenderRole       string   `json:"senderRole"`
	Content          string   `json:"content"`
	Visibility       []string `json:"visibility"`
	IsSystem         bool     `json:"isSystem"`
	CreatedAt        string   `json:"createdAt"`
}

// ThreadCloseResult is the result of closing a thread.
type ThreadCloseResult struct {
	ThreadID string `json:"threadId"`
	Status   string `json:"status"`
	ClosedAt string `json:"closedAt"`
}

// ThreadReopenResult is the result of reopening a thread.
type ThreadReopenResult struct {
	ThreadID   string `json:"threadId"`
	Status     string `json:"status"`
	ReopenedAt string `json:"reopenedAt"`
}

// EntityEventResult is the result of forwarding an entity event.
type EntityEventResult struct {
	Evaluated  int    `json:"evaluated"`
	Matched    int    `json:"matched"`
	ActionsRun int    `json:"actionsRun"`
	ThreadID   string `json:"threadId,omitempty"`
}

// MarkReadResult is the result of marking a thread as read.
type MarkReadResult struct {
	ThreadID string `json:"threadId"`
	ReadAt   string `json:"readAt"`
}

// InboxResponse wraps a paginated inbox response.
type InboxResponse struct {
	Threads []InboxThread `json:"threads"`
	Total   int           `json:"total"`
	Page    int           `json:"page"`
	Limit   int           `json:"limit"`
}

// InboxThread represents a thread in the inbox with unread count.
type InboxThread struct {
	Thread      Thread `json:"thread"`
	UnreadCount int    `json:"unreadCount"`
	LastMessage string `json:"lastMessage"`
}

// UnreadCountResult is the result of querying total unread count.
type UnreadCountResult struct {
	Count int `json:"count"`
}

// ThreadFlag represents a flag on a message.
type ThreadFlag struct {
	ID                  string `json:"id"`
	MessageID           string `json:"messageId"`
	Reason              string `json:"reason"`
	Note                string `json:"note,omitempty"`
	FlaggedByExternalID string `json:"flaggedByExternalId"`
	FlaggedByRole       string `json:"flaggedByRole"`
	CreatedAt           string `json:"createdAt"`
}

// FlagListResponse wraps a list of thread flags.
type FlagListResponse struct {
	Flags []ThreadFlag `json:"flags"`
}

// FlagMessageResult is the result of flagging a message.
type FlagMessageResult struct {
	FlagID    string `json:"flagId"`
	MessageID string `json:"messageId"`
	ThreadID  string `json:"threadId"`
}

// EscalateThreadResult is the result of escalating a thread.
type EscalateThreadResult struct {
	ThreadID    string `json:"threadId"`
	Priority    string `json:"priority"`
	EscalatedAt string `json:"escalatedAt"`
}

// EscalationConfig represents escalation configuration for a channel.
type EscalationConfig struct {
	Enabled    bool                   `json:"enabled"`
	Rules      map[string]interface{} `json:"rules,omitempty"`
	WebhookURL string                 `json:"webhookUrl,omitempty"`
}

// ThreadChannel represents a channel configuration.
type ThreadChannel struct {
	ID                string                 `json:"id"`
	OrganizationID    string                 `json:"organizationId"`
	Slug              string                 `json:"slug"`
	DisplayName       string                 `json:"displayName"`
	EntityType        string                 `json:"entityType"`
	ParticipantRoles  []string               `json:"participantRoles"`
	DefaultVisibility []string               `json:"defaultVisibility"`
	LifecycleRules    map[string]interface{} `json:"lifecycleRules"`
	EscalationConfig  map[string]interface{} `json:"escalationConfig"`
	WebhookURL        string                 `json:"webhookUrl"`
	IsActive          bool                   `json:"isActive"`
	CreatedAt         string                 `json:"createdAt"`
	UpdatedAt         string                 `json:"updatedAt"`
}

// ChannelListResponse wraps a list of thread channels.
type ChannelListResponse struct {
	Channels []ThreadChannel `json:"channels"`
}

// UpdateThreadInput is the input for updating a thread.
type UpdateThreadInput struct {
	Subject  *string                `json:"subject,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// MessageListResponse wraps a paginated list of thread messages.
type MessageListResponse struct {
	Messages []ThreadMessage `json:"messages"`
	Total    int             `json:"total"`
}

// AddParticipantInput is the input for adding a participant to a thread.
type AddParticipantInput struct {
	Role        string `json:"role"`
	ExternalID  string `json:"externalId"`
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
}

// UpdateParticipantInput is the input for updating a thread participant.
type UpdateParticipantInput struct {
	DisplayName *string `json:"displayName,omitempty"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
	IsMuted     *bool   `json:"isMuted,omitempty"`
}

// ThreadReadState represents the read state for a participant in a thread.
type ThreadReadState struct {
	ThreadID   string `json:"threadId"`
	ExternalID string `json:"externalId"`
	Role       string `json:"role"`
	LastReadAt string `json:"lastReadAt"`
}

// ReadStatesResponse wraps a list of thread read states.
type ReadStatesResponse struct {
	ReadStates []ThreadReadState `json:"readStates"`
}

// ReviewFlagInput is the input for reviewing a flagged message.
type ReviewFlagInput struct {
	Status     string `json:"status"`
	ReviewNote string `json:"reviewNote,omitempty"`
	ReviewedBy string `json:"reviewedBy,omitempty"`
}

// UploadFile describes a file for batch upload in Go.
type UploadFile struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Data        string `json:"data"` // base64-encoded
}

// BatchUploadResult is the result of a batch file upload.
type BatchUploadResult struct {
	Files     []StorageFile `json:"files"`
	Succeeded int           `json:"succeeded"`
	Failed    int           `json:"failed"`
}

// UploadResult is the result of a single file upload.
type UploadResult struct {
	File StorageFile `json:"file"`
}

// BatchQRResult is the answer of POST /api/v1/qr/batch (v1.14.0): one entry
// per requested item, in request order, and the optional zip.
type BatchQRResult struct {
	// Items holds one result per requested item (Status "generated" or
	// "failed"); a failing item never fails the call.
	Items          []QrBatchItemResult `json:"items,omitempty"`
	TotalGenerated int                 `json:"totalGenerated"`
	TotalFailed    int                 `json:"totalFailed"`
	// Zip is nil when no zip was requested, nothing was generated, or the zip
	// could not be built or stored (the images are stored either way).
	Zip         *QrBatchZip `json:"zip,omitempty"`
	GeneratedAt string      `json:"generatedAt,omitempty"`
	// Deprecated: never sent (the route answers Items). Removal: v2.
	Results []QRCodeResult `json:"results"`
	// Deprecated: never sent (a failing item carries its Error). Removal: v2.
	Errors []string `json:"errors"`
}

// DryRunResult is the result of a workflow dry-run (the route's data.dryRun).
type DryRunResult struct {
	// WouldMatch is true when the trigger's conditions match the entity data.
	WouldMatch bool `json:"wouldMatch"`
	// ConditionsEvaluated maps each condition to whether it matched.
	ConditionsEvaluated map[string]bool `json:"conditionsEvaluated,omitempty"`
	// ActionsWouldExecute is the trigger's actions that would run.
	ActionsWouldExecute []map[string]interface{} `json:"actionsWouldExecute,omitempty"`
	// ActionsPreview is each action with its config interpolated.
	ActionsPreview []DryRunActionPreview `json:"actionsPreview,omitempty"`
	// Validation is the trigger's config verdict.
	Validation *DryRunValidation `json:"validation,omitempty"`
	// Reason is set when nothing would run, such as "Trigger not found".
	Reason string `json:"reason,omitempty"`

	// TriggerID is the trigger the dry run was asked for (set by the SDK).
	//
	// Deprecated: the route does not answer it; it is the id passed in.
	TriggerID string `json:"triggerId,omitempty"`
	// Deprecated: the route does not answer the trigger's name; it stays
	// empty. Use Workflows.Get.
	TriggerName string `json:"triggerName,omitempty"`
	// Deprecated: use WouldMatch (the SDK copies it here).
	Matched bool `json:"matched,omitempty"`
	// Deprecated: use ActionsWouldExecute (the SDK copies it here).
	Actions []map[string]interface{} `json:"actions,omitempty"`
}

// DryRunActionPreview is one action of a dry run with its config interpolated.
type DryRunActionPreview struct {
	Type               string                 `json:"type"`
	Config             map[string]interface{} `json:"config"`
	InterpolatedConfig map[string]interface{} `json:"interpolatedConfig"`
}

// DryRunValidation is the dry run's config verdict for the trigger.
type DryRunValidation struct {
	OK     bool     `json:"ok"`
	Errors []string `json:"errors"`
}

// DeletionOverride represents a deletion override request/result, as
// returned by RequestOverride and by ProcessOverride for the approve,
// reject, and witness actions. The execute action returns a DIFFERENT
// shape — see ExecuteOverrideResult and the ExecuteOverride method.
type DeletionOverride struct {
	ID                 string  `json:"id"`
	DocumentID         string  `json:"documentId"`
	DocumentTitle      string  `json:"documentTitle"`
	OverrideReason     string  `json:"overrideReason"`
	AuthorityReference *string `json:"authorityReference"`
	AuthorityType      *string `json:"authorityType"`
	RequestedByID      string  `json:"requestedById"`
	RequestedAt        string  `json:"requestedAt"`
	ApprovedByID       *string `json:"approvedById"`
	ApprovedAt         *string `json:"approvedAt"`
	WitnessedByID      *string `json:"witnessedById"`
	WitnessedAt        *string `json:"witnessedAt"`
	// Status is one of PENDING, APPROVED, REJECTED, EXECUTED.
	Status          string  `json:"status"`
	RejectionReason *string `json:"rejectionReason"`
	// ExecutedAt is not currently populated by the API (reserved) — check
	// Status == "EXECUTED" instead.
	ExecutedAt *string `json:"executedAt"`

	// Deprecated: these pre-v1.6 fields were never part of the API response
	// and are always empty. Kept only so existing v1 code that references
	// them still compiles. Use RequestedByID, OverrideReason, ApprovedByID /
	// WitnessedByID and RejectionReason instead.
	RequestedBy string  `json:"requestedBy,omitempty"`
	Reason      string  `json:"reason,omitempty"`
	ReviewedBy  *string `json:"reviewedBy,omitempty"`
	ReviewNote  *string `json:"reviewNote,omitempty"`
	CreatedAt   string  `json:"createdAt,omitempty"`
	UpdatedAt   string  `json:"updatedAt,omitempty"`
}

// ExecuteOverrideResult is the response from ProcessOverride/ExecuteOverride
// when action is "execute" — it performs the document's soft-delete
// directly, so its shape is unrelated to DeletionOverride's fields.
type ExecuteOverrideResult struct {
	Success    bool   `json:"success"`
	DocumentID string `json:"documentId"`
}

// PdfResult is a generic PDF generation result.
type PdfResult struct {
	URL      string `json:"url,omitempty"`
	FileID   string `json:"fileId,omitempty"`
	Filename string `json:"filename,omitempty"`
}

// RasterizePdfInput is the request body for PdfService.Rasterize — matches
// RasterizePdfRequest in src/app/api/v1/pdf/rasterize/route.ts field-for-field.
type RasterizePdfInput struct {
	// PdfBase64 is the base64-encoded PDF. The whole request body is limited
	// to 4.5 MB by the hosting platform (its own 413, before the route runs),
	// so the PDF can be at most about 3.3 MB (base64 adds a third).
	PdfBase64 string `json:"pdfBase64"`
	// Format is "jpeg" or "png".
	Format string `json:"format"`
	// Page is the 1-indexed page to render. Omit for page 1.
	Page int `json:"page,omitempty"`
	// DPI is the render resolution; the server clamps it to [1, 300] and
	// defaults to 150 when omitted.
	DPI int `json:"dpi,omitempty"`
}

// RasterizePdfResult is the response from PdfService.Rasterize — matches
// RasterizePdfResponse in src/app/api/v1/pdf/rasterize/route.ts field-for-field.
type RasterizePdfResult struct {
	ImageBase64 string `json:"imageBase64"`
	// ContentType is "image/jpeg" or "image/png", matching the requested
	// Format — mirrors the RasterizePdfResponse.contentType union exactly
	// ('image/jpeg' | 'image/png') as a plain string (Go has no union
	// types).
	ContentType string `json:"contentType"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	PageCount   int    `json:"pageCount"`
}

// ─── Product AI gateway (v1.7.0) ────────────────────────────────────────────
// Mirrors @caldera/platformxe-types 3.5.0 src/ai.ts.

// Closed error codes of /api/v1/ai/* (ProductAiErrorCode) and
// /api/v1/vision/* (VisionErrorCode), as APIError.Code.
const (
	ErrCodeBadRequest          = "BAD_REQUEST"           // 400
	ErrCodeAiProfileNotFound   = "AI_PROFILE_NOT_FOUND"  // 403
	ErrCodeAiProfileDisabled   = "AI_PROFILE_DISABLED"   // 403
	ErrCodeAiPolicyDenied      = "AI_POLICY_DENIED"      // 403
	ErrCodeConflict            = "CONFLICT"              // 409, Knowledge.UpsertDocument: collection + sourceRef held by another organisation; never retried
	ErrCodeAiOutputBlocked     = "AI_OUTPUT_BLOCKED"     // 422
	ErrCodeAiOutputInvalid     = "AI_OUTPUT_INVALID"     // 422, Details["path"]
	ErrCodeAiQuotaExceeded     = "AI_QUOTA_EXCEEDED"     // 429, RetryAfter; never retried
	ErrCodeAiNotConfigured     = "AI_NOT_CONFIGURED"     // 503
	ErrCodeAiModelUnavailable  = "AI_MODEL_UNAVAILABLE"  // 503, Details["model"]; gateway account needs paid credits
	ErrCodeAiUpstreamError     = "AI_UPSTREAM_ERROR"     // 502
	ErrCodeVisionNotConfigured = "VISION_NOT_CONFIGURED" // 503
	ErrCodeServerError         = "SERVER_ERROR"          // 500
	// v1.16.0 (types 3.18.0)
	ErrCodeAiBudgetExceeded = "AI_BUDGET_EXCEEDED" // 429, RetryAfter to the period's end, Details["period"] (absent on a tool-loop refusal), Details["reason"] AiBudgetReasonUnavailable when unchecked; never retried (also on vision)
	ErrCodeAiRefused        = "AI_REFUSED"         // 422, Details["category"]; no content, the call is billed
	ErrCodeAiToolsetDenied  = "AI_TOOLSET_DENIED"  // 403
	ErrCodeAiToolFailed     = "AI_TOOL_FAILED"     // 502: the tool round trip itself broke (a tool's own failure goes back to the model); never retried
	ErrCodeAiLoopLimit      = "AI_LOOP_LIMIT"      // 422, Details["reason"]: "TURNS" | "CALLS" | "DEADLINE"
)

// AiBudgetReasonUnavailable is Details["reason"] of an AI_BUDGET_EXCEEDED
// when a hard budget could not be checked (its spend could not be read;
// RetryAfter 60). A spent budget carries no reason. A tool loop stopped by its
// budget between model calls answers Details {"retryAfterSeconds": 60} with
// no "period". Since v1.16.0 (types 3.18.0 AiBudgetRefusalReason).
const AiBudgetReasonUnavailable = "BUDGET_UNAVAILABLE"

// AiMessage is one chat message (role: "system" | "user" | "assistant" |
// "tool"). Content may be "" on an assistant message carrying ToolCalls, or
// on a tool message.
type AiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	// ToolCalls: assistant messages only — echo the calls the model returned.
	ToolCalls []AiToolCall `json:"toolCalls,omitempty"`
	// ToolCallID: tool messages only — the call this is the result of.
	ToolCallID string `json:"toolCallId,omitempty"`
}

// AiToolCall is one tool call the model asked for. PlatformXe never
// executes it; Arguments is the JSON-encoded argument object.
type AiToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// AiToolDefinition is one function tool the model may call.
type AiToolDefinition struct {
	// Name: 1-64 letters, digits, "_", "." or "-"; unique in the request.
	Name string `json:"name"`
	// Description: at most 1,024 characters.
	Description string `json:"description,omitempty"`
	// Parameters is the JSON Schema of the arguments (type "object").
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// AiToolChoiceName forces one named tool (AiCompleteInput.ToolChoice).
type AiToolChoiceName struct {
	Name string `json:"name"`
}

// AiCompleteInput is the body of POST /api/v1/ai/complete
// (ProductAiCompleteRequest).
type AiCompleteInput struct {
	// FeatureKey names the AI profile for the calling service.
	FeatureKey string `json:"featureKey"`
	// Provider is "gemini", "grok" or "claude".
	Provider string      `json:"provider"`
	Messages []AiMessage `json:"messages"`
	// Model optionally overrides the profile / provider default.
	Model string `json:"model,omitempty"`
	// Tools: at most 16, only when the profile allows tools.
	Tools []AiToolDefinition `json:"tools,omitempty"`
	// ToolChoice: "auto" | "none" | "required" (a string) or
	// AiToolChoiceName{Name}; nil omits it. Requires Tools.
	ToolChoice interface{} `json:"toolChoice,omitempty"`
	// Toolsets: server-side toolsets (e.g. "platform.read") run by
	// PlatformXe's tool loop on a Claude 5.5 model (Model the gateway slug,
	// e.g. "anthropic/claude-sonnet-5.5"); each must be on the profile and the
	// key needs tools:platform.read. Not together with Tools; ToolChoice must
	// then be "auto" or nil (400 otherwise).
	Toolsets []string `json:"toolsets,omitempty"`
}

// AiUsage is the token and cost report of one call (ProductAiUsage).
type AiUsage struct {
	InputTokens  int     `json:"inputTokens,omitempty"`
	OutputTokens int     `json:"outputTokens,omitempty"`
	TotalTokens  int     `json:"totalTokens,omitempty"`
	CostCredits  float64 `json:"costCredits,omitempty"`
}

// AiCompleteResult is the data of POST /api/v1/ai/complete.
type AiCompleteResult struct {
	// Content is "" when the model answered only with tool calls.
	Content  string   `json:"content"`
	Provider string   `json:"provider"`
	Model    string   `json:"model"`
	Usage    *AiUsage `json:"usage,omitempty"`
	// ToolCalls is set when the model asked for tools.
	ToolCalls []AiToolCall `json:"toolCalls,omitempty"`
	// FinishReason: "stop" | "length" | "tool_calls" | "content_filter" | "other".
	FinishReason string `json:"finishReason,omitempty"`
	// ToolRuns: one per server-side tool call, when Toolsets ran.
	ToolRuns []AiToolRunSummary `json:"toolRuns,omitempty"`
}

// AiToolRunSummary is one server-side tool call as recorded: never its input
// or output (AiToolRunSummary in @caldera/platformxe-types 3.18.0).
type AiToolRunSummary struct {
	// ToolCallID is the ledger row; nil when it could not be written.
	ToolCallID  *string `json:"toolCallId"`
	ToolUseID   *string `json:"toolUseId"`
	ToolName    string  `json:"toolName"`
	ToolVersion int     `json:"toolVersion"`
	// Status: "OK" | "DENIED" | "FAILED" | "TIMEOUT".
	Status string `json:"status"`
	// DenyReason: "TENANT" | "SCOPE" | "TOOLSET" | "DATA_CLASS" |
	// "PERMISSION" | "BUDGET" | "SPOKE"; nil unless denied.
	DenyReason  *string `json:"denyReason"`
	IsError     bool    `json:"isError"`
	LatencyMs   int     `json:"latencyMs"`
	OutputBytes *int    `json:"outputBytes"`
}

// ─── Native Claude messages and platform tools (v1.16.0) ────────────────────
// Mirrors @caldera/platformxe-types 3.18.0 src/ai.ts.

// AiSystemBlock is one system prompt block (Type "text"). Cache is accepted
// for shape parity; caching is decided by the profile.
type AiSystemBlock struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Cache bool   `json:"cache,omitempty"`
}

// AiNativeMessage is one message in the Anthropic Messages shape (Role "user"
// | "assistant"). Content is a non-empty string or a slice of content blocks
// (map[string]interface{} with a "type": text, image, tool_use, tool_result,
// thinking, redacted_thinking, fallback); pass thinking blocks back unchanged.
type AiNativeMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// AiNativeTool is a caller-executed tool: its tool_use blocks are returned
// for the caller to run. Name may not start with a tool-source prefix
// (platformxe_, xadmin_, lettings_, concierge_, chats_).
type AiNativeTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// AiMessagesInput is the body of POST /api/v1/ai/messages
// (ProductAiMessagesRequest).
type AiMessagesInput struct {
	FeatureKey string `json:"featureKey"`
	// Model: "claude-haiku-5-5" | "claude-sonnet-5-5" | "claude-opus-5-5", as
	// the profile allows; "" = the profile's first Claude 5.5 model.
	Model string `json:"model,omitempty"`
	// System: a string or []AiSystemBlock; nil omits it.
	System   interface{}       `json:"system,omitempty"`
	Messages []AiNativeMessage `json:"messages"`
	// Effort: "low" | "medium" | "high" | "xhigh" | "max", capped by the
	// profile's maxEffort; "" = maxEffort.
	Effort string `json:"effort,omitempty"`
	// OutputSchema: structured output (JSON Schema, an object root), when the
	// profile allows it. Every object node needs additionalProperties: false
	// (required optional). pattern, minProperties / maxProperties, minItems 0
	// or 1 and the formats date-time, time, date, duration, email, hostname,
	// uri, ipv4, ipv6, uuid pass; numeric and length bounds, maxItems,
	// uniqueItems, minContains / maxContains, minItems above 1 and other
	// formats are refused (400).
	OutputSchema map[string]interface{} `json:"outputSchema,omitempty"`
	Toolsets     []string               `json:"toolsets,omitempty"`
	Tools        []AiNativeTool         `json:"tools,omitempty"`
	// ToolChoice: "auto" | "none"; forced tool use is refused (400). With
	// Toolsets only "auto" or "" (400 otherwise).
	ToolChoice string `json:"toolChoice,omitempty"`
	// MaxTokens: at most the profile's maxTokens; 0 omits it.
	MaxTokens int `json:"maxTokens,omitempty"`
	// Actor (v1.17.0): the person a first-party Caldera app's own tools act
	// for; nil omits it (no such tool is then offered).
	Actor *AiToolActor `json:"actor,omitempty"`
	// Actors (v1.18.0): for Caldera's operator console (a platform key), one
	// person per Caldera app keyed "xadmin", "lettings", "concierge" or
	// "chats"; each entry reaches only its own app's tools. Never with Actor
	// (Ai.Messages returns ErrAiActorAndActors; the server answers 400). Nil
	// or empty omits it.
	Actors map[string]AiToolActorsEntry `json:"actors,omitempty"`
}

// AiToolActorsEntry is one app's person in AiMessagesInput.Actors
// (AiToolActorsInput): that app's id for the person, the permissions that
// app's own engine gives them, and that app's own tool grant.
type AiToolActorsEntry struct {
	// Ref: the person's id in that app, 1 to 64 of A-Z a-z 0-9 _ -.
	Ref string `json:"ref"`
	// Permissions: at most 64 canonical keys with THAT app's prefix (XA:, LT:,
	// CN:, CH:); nil is sent as [].
	Permissions []string `json:"permissions"`
	// Grant: that app's own tool grant (required), at most 4,096 base64url
	// characters; forwarded only to that app, never stored, logged or returned.
	Grant string `json:"grant"`
}

// MarshalJSON sends a nil Permissions as [] and never null.
func (a AiToolActorsEntry) MarshalJSON() ([]byte, error) {
	type wire AiToolActorsEntry
	w := wire(a)
	if w.Permissions == nil {
		w.Permissions = []string{}
	}
	return json.Marshal(w)
}

// AiToolActor is the actor of POST /api/v1/ai/messages (AiToolActorInput).
// A first-party app's tool is offered only when Permissions holds its
// RequiredPermission exactly; the app checks the person again on every call.
type AiToolActor struct {
	// Ref: the person's id in your app, 1 to 64 of A-Z a-z 0-9 _ -.
	Ref string `json:"ref"`
	// Permissions: at most 32 canonical prefixed keys, e.g. "CN:REQUESTS:VIEW".
	// Nil is sent as [] (the server refuses null): no permission, so no
	// first-party tool is offered.
	Permissions []string `json:"permissions"`
	// Grant: opaque, at most 4,096 base64url characters; "" omits it. Forwarded
	// only to the tool's own app, never stored, logged or returned.
	Grant string `json:"grant,omitempty"`
}

// MarshalJSON sends a nil Permissions as [] and never null (the server answers
// a null with 400 VALIDATION_ERROR).
func (a AiToolActor) MarshalJSON() ([]byte, error) {
	type wire AiToolActor
	w := wire(a)
	if w.Permissions == nil {
		w.Permissions = []string{}
	}
	return json.Marshal(w)
}

// AiStopDetails is the reply's stop_details.
type AiStopDetails struct {
	Type     string  `json:"type,omitempty"`
	Category *string `json:"category"`
}

// AiNativeReply is the model's reply (restored and screened). After a
// server-side fallback, Content starts at the last "fallback" block (that
// block included); send the turn back as returned. A fallback block's
// "trigger" sent back is at most 512 characters as JSON; only "type" and
// "category" (at most 64 characters, or null) are kept.
type AiNativeReply struct {
	ID          string                   `json:"id"`
	Role        string                   `json:"role"`
	Content     []map[string]interface{} `json:"content"`
	StopReason  *string                  `json:"stop_reason"`
	StopDetails *AiStopDetails           `json:"stop_details"`
}

// AiMessagesUsage is AiUsage plus the prompt-cache token counts.
type AiMessagesUsage struct {
	AiUsage
	CacheReadTokens  int `json:"cacheReadTokens,omitempty"`
	CacheWriteTokens int `json:"cacheWriteTokens,omitempty"`
}

// AiMessagesResult is the data of POST /api/v1/ai/messages. There is no
// refused field: a refusal is always 422 AI_REFUSED.
type AiMessagesResult struct {
	Message AiNativeReply `json:"message"`
	// Model is the model that answered (a server-side fallback may differ).
	Model    string             `json:"model"`
	Usage    AiMessagesUsage    `json:"usage"`
	ToolRuns []AiToolRunSummary `json:"toolRuns,omitempty"`
}

// AiToolListItem is one platform tool a key may use (never its implementation).
type AiToolListItem struct {
	Name        string                 `json:"name"`
	Version     int                    `json:"version"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	// DataClasses: "PII_NAME" | "PII_CONTACT" | "PII_ID" | "MONEY" | "LOCATION".
	DataClasses []string `json:"dataClasses"`
	Toolset     string   `json:"toolset"`
	// SourceApp (v1.17.0): "platformx", or the Caldera app that serves the tool.
	SourceApp string `json:"sourceApp,omitempty"`
	// RequiredPermission (v1.17.0): the key the acting person must hold, or nil.
	RequiredPermission *string `json:"requiredPermission,omitempty"`
}

// AiToolList is the data of GET /api/v1/ai/tools.
type AiToolList struct {
	Items []AiToolListItem `json:"items"`
	Total int              `json:"total"`
}

// AiImageInput is one image for a multimodal extract: exactly one of URL
// (public https) / Base64 (at most 8 MiB decoded, with MediaType).
type AiImageInput struct {
	URL    string `json:"url,omitempty"`
	Base64 string `json:"base64,omitempty"`
	// MediaType is "image/jpeg", "image/png" or "image/webp"; required with Base64.
	MediaType string `json:"mediaType,omitempty"`
}

// AiExtractInput is the body of POST /api/v1/ai/extract
// (ProductAiExtractRequest).
type AiExtractInput struct {
	FeatureKey   string `json:"featureKey"`
	SystemPrompt string `json:"systemPrompt"`
	UserPrompt   string `json:"userPrompt"`
	// Images: at most 4, accepted only when the profile allows images.
	Images []AiImageInput `json:"images,omitempty"`
	// OutputJSONSchema is a draft-07 JSON Schema the reply is validated against.
	OutputJSONSchema map[string]interface{} `json:"outputJsonSchema"`
	// ModelTier is "FAST" or "PREMIUM".
	ModelTier string `json:"modelTier"`
	// Temperature is 0..2; nil omits it (server default 0.2). A pointer so 0 is sendable.
	Temperature *float64 `json:"temperature,omitempty"`
}

// AiExtractResult is the data of POST /api/v1/ai/extract.
type AiExtractResult struct {
	Data       map[string]interface{} `json:"data"`
	Provider   string                 `json:"provider"`
	Model      string                 `json:"model"`
	Usage      *AiUsage               `json:"usage,omitempty"`
	RawContent string                 `json:"rawContent,omitempty"`
}

// AiProfileView is the data of GET /api/v1/ai/profile: the calling
// service's own profile for one feature (policy text withheld).
type AiProfileView struct {
	CallerService     string   `json:"callerService"`
	FeatureKey        string   `json:"featureKey"`
	Status            string   `json:"status"` // "ACTIVE" | "DISABLED"
	AllowedProviders  []string `json:"allowedProviders"`
	AllowedModels     []string `json:"allowedModels"`
	MaxTokens         int      `json:"maxTokens"`
	MaxMessages       int      `json:"maxMessages"`
	MaxTotalChars     int      `json:"maxTotalChars"`
	AllowImages       bool     `json:"allowImages"`
	AllowTools        bool     `json:"allowTools"`
	DailyTokenCap     *int     `json:"dailyTokenCap"`   // nil = no cap
	DailyRequestCap   *int     `json:"dailyRequestCap"` // nil = no cap
	UsedTokensToday   int      `json:"usedTokensToday"`
	UsedRequestsToday int      `json:"usedRequestsToday"`
	// ResetsAt is the next UTC midnight (ISO-8601).
	ResetsAt string `json:"resetsAt"`
	// Governance (v1.16.0, types 3.18.0).
	// MaxEffort: the highest effort a Claude 5.5 call may use.
	MaxEffort             string `json:"maxEffort"`
	AllowCaching          bool   `json:"allowCaching"`
	AllowStructuredOutput bool   `json:"allowStructuredOutput"`
	// FallbackModel: "claude-opus-5" | "claude-opus-4-8"; nil = the provider's default routing.
	FallbackModel *string  `json:"fallbackModel"`
	Toolsets      []string `json:"toolsets"`
	// MaxToolTurns: 1..25.
	MaxToolTurns int `json:"maxToolTurns"`
	// EndUserFacing: output screening on, a non-empty allowedModels with no
	// Opus model, no fallbackModel, no platform tools.
	EndUserFacing      bool     `json:"endUserFacing"`
	AllowedDataClasses []string `json:"allowedDataClasses"`
	StrictIDRedaction  bool     `json:"strictIdRedaction"`
}

// AiUsageQuery is the query of GET /api/v1/ai/usage. From / To are ISO-8601
// (default: the last 30 days; at most 92 days); empty fields are omitted.
type AiUsageQuery struct {
	From       string
	To         string
	FeatureKey string
}

// AiUsageSummaryRow is one (feature, kind, provider, model) group.
type AiUsageSummaryRow struct {
	FeatureKey   string  `json:"featureKey"`
	Kind         string  `json:"kind"` // "ai" | "vision" | "embedding" | "messages" | "complete_with_tools"
	Provider     *string `json:"provider"`
	Model        *string `json:"model"`
	Requests     int     `json:"requests"`
	Refused      int     `json:"refused"`
	Failed       int     `json:"failed"`
	InputTokens  int     `json:"inputTokens"`
	OutputTokens int     `json:"outputTokens"`
	CostCredits  float64 `json:"costCredits"`
}

// AiUsageTotals sums an AiUsageSummary.
type AiUsageTotals struct {
	Requests     int     `json:"requests"`
	InputTokens  int     `json:"inputTokens"`
	OutputTokens int     `json:"outputTokens"`
	CostCredits  float64 `json:"costCredits"`
}

// AiUsageSummary is the data of GET /api/v1/ai/usage.
type AiUsageSummary struct {
	CallerService string              `json:"callerService"`
	From          string              `json:"from"`
	To            string              `json:"to"`
	Rows          []AiUsageSummaryRow `json:"rows"`
	Totals        AiUsageTotals       `json:"totals"`
}

// ─── Knowledge retrieval (v1.7.0, xadmin#29) ────────────────────────────────
// Mirrors @caldera/platformxe-types 3.6.0 src/ai.ts (Knowledge*).

// KnowledgeUpsertDocumentInput is the body of PUT
// /api/v1/ai/knowledge/documents. Text is stored as sent — no PII redaction
// at index time; never send Tier-1 / Tier-2 personal data.
type KnowledgeUpsertDocumentInput struct {
	FeatureKey string `json:"featureKey"`
	// Collection: 1-64 lowercase letters, digits, ".", "_" or "-".
	Collection string `json:"collection"`
	// SourceRef is the caller's own id, unique within the collection.
	SourceRef string `json:"sourceRef"`
	Title     string `json:"title"`
	URL       string `json:"url,omitempty"`
	// ACL: opaque tags; a searcher holding ANY of them sees the document. nil / empty = every searcher.
	ACL  []string `json:"acl,omitempty"`
	Text string   `json:"text"`
}

// KnowledgeUpsertDocumentResult is the data of PUT /api/v1/ai/knowledge/documents.
type KnowledgeUpsertDocumentResult struct {
	Collection        string `json:"collection"`
	SourceRef         string `json:"sourceRef"`
	Changed           bool   `json:"changed"`
	ChunkCount        int    `json:"chunkCount"`
	PendingEmbeddings int    `json:"pendingEmbeddings"`
}

// KnowledgeDocumentSummary is one document of a list (no text).
type KnowledgeDocumentSummary struct {
	Collection  string   `json:"collection"`
	SourceRef   string   `json:"sourceRef"`
	Title       string   `json:"title"`
	URL         *string  `json:"url"`
	ACL         []string `json:"acl"`
	FeatureKey  string   `json:"featureKey"`
	ContentHash string   `json:"contentHash"`
	ChunkCount  int      `json:"chunkCount"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

// KnowledgeDocumentList is the data of GET /api/v1/ai/knowledge/documents.
type KnowledgeDocumentList struct {
	Items []KnowledgeDocumentSummary `json:"items"`
	// NextCursor is nil on the last page.
	NextCursor *string `json:"nextCursor"`
}

// KnowledgeListQuery is the query of GET /api/v1/ai/knowledge/documents;
// empty Cursor / zero Limit are omitted.
type KnowledgeListQuery struct {
	Collection string
	Cursor     string
	Limit      int
}

// KnowledgeDeleteDocumentResult is the data of DELETE /api/v1/ai/knowledge/documents.
type KnowledgeDeleteDocumentResult struct {
	Collection string `json:"collection"`
	SourceRef  string `json:"sourceRef"`
	Deleted    bool   `json:"deleted"`
}

// KnowledgeSearchInput is the body of POST /api/v1/ai/knowledge/search.
type KnowledgeSearchInput struct {
	FeatureKey  string   `json:"featureKey"`
	Collections []string `json:"collections"`
	Query       string   `json:"query"`
	// K: 1-20; 0 omits it (server default 8).
	K int `json:"k,omitempty"`
	// ACLTags is always sent; an empty list sees only documents with no ACL.
	ACLTags []string `json:"aclTags"`
}

// KnowledgeSearchHit is one ranked chunk; SourceRef + Ordinal cite it.
type KnowledgeSearchHit struct {
	SourceRef   string  `json:"sourceRef"`
	Collection  string  `json:"collection"`
	Title       string  `json:"title"`
	URL         *string `json:"url"`
	Heading     *string `json:"heading"`
	Ordinal     int     `json:"ordinal"`
	Text        string  `json:"text"`
	Score       float64 `json:"score"`
	LexicalRank int     `json:"lexicalRank"`
	VectorRank  *int    `json:"vectorRank"`
}

// KnowledgeSearchResult is the data of POST /api/v1/ai/knowledge/search.
type KnowledgeSearchResult struct {
	// Mode is "hybrid" or "lexical".
	Mode string               `json:"mode"`
	Hits []KnowledgeSearchHit `json:"hits"`
}

// ─── Vision (v1.7.0) ────────────────────────────────────────────────────────
// Mirrors @caldera/platformxe-types 3.5.0 src/vision.ts.

// AnalyzeRoomInput is the body of POST /api/v1/vision/analyze-room
// (AnalyzeRoomRequest). Exactly one of ImageURL / ImageBase64.
type AnalyzeRoomInput struct {
	// ImageURL is a public https URL.
	ImageURL string `json:"imageUrl,omitempty"`
	// ImageBase64 is at most 8 MiB decoded (a data: prefix is accepted).
	ImageBase64    string   `json:"imageBase64,omitempty"`
	RoomCategories []string `json:"roomCategories"`
	WantAltText    bool     `json:"wantAltText"`
	// Locale is "en-NG", "en-GB" or "en-US".
	Locale string `json:"locale,omitempty"`
	// FeatureKey defaults server-side to "room-vision".
	FeatureKey string `json:"featureKey,omitempty"`
}

// AnalyzeRoomOptions are per-call options for VisionService.AnalyzeRoom.
type AnalyzeRoomOptions struct {
	// IdempotencyKey is sent as x-idempotency-key. A FALLBACK answer is never
	// cached under it (X-Idempotency-Cache: skip), so a retry reaches the
	// provider again.
	IdempotencyKey string
}

// AnalyzeRoomMetrics reports the provider call.
type AnalyzeRoomMetrics struct {
	LatencyMs           int     `json:"latencyMs"`
	TokensOrCostCredits float64 `json:"tokensOrCostCredits"`
}

// AnalyzeRoomResult is the data of POST /api/v1/vision/analyze-room.
type AnalyzeRoomResult struct {
	RoomCategory     *string  `json:"roomCategory"` // nil when no category fits
	RoomConfidence   float64  `json:"roomConfidence"`
	AltText          string   `json:"altText"`
	DetectedFeatures []string `json:"detectedFeatures"`
	ModelVersion     string   `json:"modelVersion"`
	Cached           bool     `json:"cached"`
	// AiStatus is "OK" or "FALLBACK" (a provider timeout answered with a
	// deterministic result, still a success).
	AiStatus string              `json:"aiStatus,omitempty"`
	Provider string              `json:"provider,omitempty"` // "gateway" | "azure"
	Metrics  *AnalyzeRoomMetrics `json:"metrics,omitempty"`
}

// ─── PDF generation (v1.14.0) ───────────────────────────────────────────────
// Mirrors @caldera/platformxe-types 3.16.0 src/pdf.ts (GeneratePdf*).

// GeneratePdfMargins are CSS lengths, e.g. "10mm" (mm, cm, in or px; at most
// 100 mm each).
type GeneratePdfMargins struct {
	Top    string `json:"top,omitempty"`
	Right  string `json:"right,omitempty"`
	Bottom string `json:"bottom,omitempty"`
	Left   string `json:"left,omitempty"`
}

// GeneratePdfOptions is the request's options object. Zero values are not
// sent (the server default applies).
type GeneratePdfOptions struct {
	// Size is A3, A4, A5, Letter or Legal (default A4).
	Size string `json:"size,omitempty"`
	// Orientation is portrait or landscape (default portrait).
	Orientation     string              `json:"orientation,omitempty"`
	Margins         *GeneratePdfMargins `json:"margins,omitempty"`
	PrintBackground *bool               `json:"printBackground,omitempty"`
	// HeaderHTML / FooterHTML: at most 16 KiB each, allow-listed elements.
	HeaderHTML string `json:"headerHtml,omitempty"`
	FooterHTML string `json:"footerHtml,omitempty"`
	// Filename is 1 to 100 word characters, dots or hyphens; .pdf is added
	// when missing.
	Filename string `json:"filename,omitempty"`
	// RemoteImages defaults to true on the server (https images fetched by
	// the server); set it to a false pointer to turn them off.
	RemoteImages *bool `json:"remoteImages,omitempty"`
	// ReturnBase64 answers the PDF inline (at most 3 MiB) instead of storing
	// it.
	ReturnBase64 bool `json:"returnBase64,omitempty"`
	// URLTTLSeconds is the signed URL lifetime, 60..604800 (default 86400).
	URLTTLSeconds int `json:"urlTtlSeconds,omitempty"`
}

// GeneratePdfInput is the body of POST /api/v1/pdf/generate. Exactly one of
// HTML (at most 2 MiB) and TemplateID; Data only with TemplateID (string,
// number or boolean values, at most 200 keys, 64 KiB serialised).
type GeneratePdfInput struct {
	HTML       string                 `json:"html,omitempty"`
	TemplateID string                 `json:"templateId,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Options    *GeneratePdfOptions    `json:"options,omitempty"`
}

// GeneratePdfCallOptions are the per-call options of Pdf.Generate.
type GeneratePdfCallOptions struct {
	// IdempotencyKey is sent as x-idempotency-key: a repeat answers the
	// stored result (base64 answers are not cached). Only with a key does the
	// client re-send after a transport error or a 5xx.
	IdempotencyKey string
}

// GeneratePdfResult is the answer of Pdf.Generate. Stored mode fills FileID,
// PdfURL (a read-only signed URL) and ExpiresAt; base64 mode fills Content
// and ContentType ("application/pdf").
type GeneratePdfResult struct {
	FileID      string `json:"fileId,omitempty"`
	PdfURL      string `json:"pdfUrl,omitempty"`
	ExpiresAt   string `json:"expiresAt,omitempty"`
	Content     string `json:"content,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Size        int64  `json:"size"`
	Pages       int    `json:"pages"`
	GeneratedAt string `json:"generatedAt"`
}

// ─── QR batch (v1.14.0) ─────────────────────────────────────────────────────

// QrUtmParams are appended to the encoded URL (at most 100 characters each).
type QrUtmParams struct {
	Source   string `json:"source,omitempty"`
	Medium   string `json:"medium,omitempty"`
	Campaign string `json:"campaign,omitempty"`
}

// QrBatchItemOptions are one item's QR options.
type QrBatchItemOptions struct {
	// Size is a named size ("display", "print", "flyer") or an int 64..1024.
	Size interface{} `json:"size,omitempty"`
	// ErrorCorrection is L, M, Q or H.
	ErrorCorrection string `json:"errorCorrection,omitempty"`
	// Margin is 0..10 (nil: the default).
	Margin *int `json:"margin,omitempty"`
	// DarkColor / LightColor are #RRGGBB or #RRGGBBAA.
	DarkColor  string       `json:"darkColor,omitempty"`
	LightColor string       `json:"lightColor,omitempty"`
	UtmParams  *QrUtmParams `json:"utmParams,omitempty"`
}

// QrBatchItem is one requested QR code. EntityType is AGENT, PARTNER,
// PRINCIPAL, PROPERTY, BOOKING or PAYMENT; EntityID (at most 128 characters)
// is required for the content types; BaseURL is http(s), at most 512.
type QrBatchItem struct {
	EntityType string              `json:"entityType"`
	EntityID   string              `json:"entityId,omitempty"`
	AccessCode string              `json:"accessCode,omitempty"`
	BaseURL    string              `json:"baseUrl,omitempty"`
	Options    *QrBatchItemOptions `json:"options,omitempty"`
}

// QrBatchInput is the body of POST /api/v1/qr/batch.
type QrBatchInput struct {
	// Items: 1 to the processor's maxBatchSize (at most 100).
	Items []QrBatchItem `json:"items"`
	// Format is png or svg (default: the processor's).
	Format string `json:"format,omitempty"`
	// Zip also stores one zip of every generated image.
	Zip bool `json:"zip,omitempty"`
	// URLTTLSeconds is the signed URL lifetime, 60..604800 (default 86400).
	URLTTLSeconds int `json:"urlTtlSeconds,omitempty"`
}

// QrBatchCallOptions are the per-call options of Qr.GenerateBatchItems.
type QrBatchCallOptions struct {
	// IdempotencyKey is sent as x-idempotency-key (see GeneratePdfCallOptions).
	IdempotencyKey string
}

// QrBatchItemError is a failed item's error: Code is ErrCodeValidationError
// or ErrCodeStorageFailed, with a fixed message.
type QrBatchItemError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// QrBatchItemResult is one item of the answer. Status "generated" fills
// TargetURL, Format, FileID, SignedURL, ExpiresAt and Size; "failed" fills
// Error. EntityType and EntityID echo whatever string the item sent (a
// failed item's cut to 128 characters), known type or not; each is nil when
// it was missing or not a string.
type QrBatchItemResult struct {
	Index      int               `json:"index"`
	Status     string            `json:"status"`
	EntityType *string           `json:"entityType"`
	EntityID   *string           `json:"entityId"`
	TargetURL  string            `json:"targetUrl,omitempty"`
	Format     string            `json:"format,omitempty"`
	FileID     string            `json:"fileId,omitempty"`
	SignedURL  string            `json:"signedUrl,omitempty"`
	ExpiresAt  string            `json:"expiresAt,omitempty"`
	Size       int64             `json:"size,omitempty"`
	Error      *QrBatchItemError `json:"error,omitempty"`
}

// QrBatchZip is the stored zip of a batch (entries stored uncompressed).
type QrBatchZip struct {
	FileID    string `json:"fileId"`
	SignedURL string `json:"signedUrl"`
	ExpiresAt string `json:"expiresAt"`
	Size      int64  `json:"size"`
	Entries   int    `json:"entries"`
}

// ─── Billing invoices (v1.14.0) ─────────────────────────────────────────────
// Mirrors @caldera/platformxe-types 3.16.0 src/invoicing.ts (BillingInvoice*).

// BillingInvoice is one invoice of the key's organisation. Money fields are
// integer strings in kobo (minor units); Tax is the 7.5% VAT. Status is
// DRAFT, PENDING, PAID, OVERDUE or CANCELLED. PdfURL is always nil today.
type BillingInvoice struct {
	ID          string  `json:"id"`
	Status      string  `json:"status"`
	Currency    string  `json:"currency"`
	PeriodStart string  `json:"periodStart"`
	PeriodEnd   string  `json:"periodEnd"`
	Subtotal    string  `json:"subtotal"`
	Tax         string  `json:"tax"`
	Total       string  `json:"total"`
	PaidAt      *string `json:"paidAt"`
	PaidVia     *string `json:"paidVia"`
	PdfURL      *string `json:"pdfUrl"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

// BillingInvoiceLineItem is one line of an invoice (money in kobo strings).
type BillingInvoiceLineItem struct {
	ID          string  `json:"id"`
	Service     string  `json:"service"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   string  `json:"unitPrice"`
	Total       string  `json:"total"`
}

// BillingInvoiceDetail is the answer of Billing.GetInvoice.
type BillingInvoiceDetail struct {
	Invoice   BillingInvoice           `json:"invoice"`
	LineItems []BillingInvoiceLineItem `json:"lineItems"`
}

// BillingPagination is the list's paging block.
type BillingPagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// BillingInvoiceList is the answer of Billing.ListInvoices, newest first.
type BillingInvoiceList struct {
	Invoices   []BillingInvoice  `json:"invoices"`
	Pagination BillingPagination `json:"pagination"`
}

// ListInvoicesParams is the list query. Page / Limit of 0 use the server
// defaults (1 / 20; Limit is clamped to 1..100); Status "" lists every
// status, and an unknown one is 400 VALIDATION_ERROR.
type ListInvoicesParams struct {
	Page   int
	Limit  int
	Status string
}

// Error codes of POST /api/v1/pdf/generate, POST /api/v1/qr/batch and the
// billing invoice routes, as APIError.Code (v1.14.0). ErrCodeConflict (409,
// idempotency key in flight) and ErrCodeInternalError (500) are declared
// above.
const (
	// ErrCodeValidationError: 400. A bad request; also a failed QR item's
	// QrBatchItemError.Code.
	ErrCodeValidationError = "VALIDATION_ERROR"
	// ErrCodeForbidden: 403. Missing scope, inactive organisation, or (billing)
	// a platform key with no organisation.
	ErrCodeForbidden = "FORBIDDEN"
	// ErrCodeNotFound: 404. Billing.GetInvoice: unknown id, or another
	// organisation's invoice.
	ErrCodeNotFound = "NOT_FOUND"
	// ErrCodeTemplateNotFound: 404 from Pdf.Generate (unknown template, or
	// another organisation's).
	ErrCodeTemplateNotFound = "TEMPLATE_NOT_FOUND"
	// ErrCodeRenderTimeout: 408 from Pdf.Generate (render past 25 s).
	ErrCodeRenderTimeout = "RENDER_TIMEOUT"
	// ErrCodePayloadTooLarge: 413 from Pdf.Generate (HTML over 2 MiB, data
	// over 64 KiB, header or footer over 16 KiB).
	ErrCodePayloadTooLarge = "PAYLOAD_TOO_LARGE"
	// ErrCodeOutputTooLarge: 413 from Pdf.Generate (over 100 pages; PDF over
	// 20 MiB stored or 3 MiB base64).
	ErrCodeOutputTooLarge = "OUTPUT_TOO_LARGE"
	// ErrCodeServiceDisabled: 403 from Qr.GenerateBatchItems (the
	// organisation's QR processor is disabled).
	ErrCodeServiceDisabled = "SERVICE_DISABLED"
	// ErrCodeRateLimited: 429 (key budget or route tier); RetryAfter is set.
	ErrCodeRateLimited = "RATE_LIMITED"
	// ErrCodeStorageUnavailable: 503 from Pdf.Generate (stored mode) and
	// Qr.GenerateBatchItems: storage not configured or the upload failed.
	ErrCodeStorageUnavailable = "STORAGE_UNAVAILABLE"
	// ErrCodeRendererUnavailable: 503 from Pdf.Generate (the renderer did not
	// start).
	ErrCodeRendererUnavailable = "RENDERER_UNAVAILABLE"
	// ErrCodeStorageFailed: a failed QR item's QrBatchItemError.Code (its
	// upload failed); never a whole-call code.
	ErrCodeStorageFailed = "STORAGE_FAILED"
)

// ─── Enterprise License (v1.15.0) ──────────────────────────────────────────
// Mirrors @caldera/platformxe-types 3.17.0 src/license.ts.

// Where the organisation's effective plan comes from (LicenseEffectiveEdition.Source).
const (
	LicenseEditionSourceLicense    = "license"
	LicenseEditionSourcePlan       = "plan"
	LicenseEditionSourceFirstParty = "first_party"
)

// The licensing exemption an organisation has (LicenseView.Bypass).
const (
	LicenseBypassFirstParty = "first_party"
	LicenseBypassDemo       = "demo"
)

// A license's state (LicenseCoverage.State): valid, then grace (everything
// still runs), then lapsed (its plugins pause); or revoked by Caldera.
const (
	LicenseStateValid   = "valid"
	LicenseStateGrace   = "grace"
	LicenseStateLapsed  = "lapsed"
	LicenseStateRevoked = "revoked"
)

// LicenseEffectiveEdition is the organisation's effective plan (FREE, BASIC, PRO or
// ENTERPRISE: the plan every edition-level gate reads) and its source.
type LicenseEffectiveEdition struct {
	Plan   string `json:"plan"`
	Source string `json:"source"`
}

// LicenseCoverage is the license that covers the key's organisation. It never
// carries money, the license text or another organisation's grants.
type LicenseCoverage struct {
	LicenseID string `json:"licenseId"`
	// Kind is "trial" or "standard".
	Kind        string `json:"kind"`
	IssuedAt    string `json:"issuedAt"`
	ExpiresAt   string `json:"expiresAt"`
	GraceEndsAt string `json:"graceEndsAt"`
	State       string `json:"state"`
	// Live: installed, valid or in grace, and the business registration
	// still matches.
	Live bool `json:"live"`
	// DaysLeft counts Lagos days to the expiry day (negative after it).
	DaysLeft   int    `json:"daysLeft"`
	RenewalDue bool   `json:"renewalDue"`
	HolderName string `json:"holderName"`
	IsHolder   bool   `json:"isHolder"`
	// Edition is this organisation's edition under the grant (PRO or
	// ENTERPRISE; nil on a coverage written before editions).
	Edition *string `json:"edition"`
	// Plugins granted to this organisation only, packs before add-ons.
	Plugins []string `json:"plugins"`
}

// LicenseView is the answer of License.Get.
type LicenseView struct {
	Edition LicenseEffectiveEdition `json:"edition"`
	// Bypass is "first_party" or "demo" when the organisation needs no
	// license, else nil.
	Bypass *string `json:"bypass"`
	// License is nil when no license covers the organisation.
	License *LicenseCoverage `json:"license"`
}

// Licensing error codes, as APIError.Code (v1.15.0).
const (
	// ErrCodeLicenseRequired: 403 from Fraud.Activate after the licensing
	// cut-over, when no live Enterprise License grant names a fraud pack.
	ErrCodeLicenseRequired = "LICENSE_REQUIRED"
	// ErrCodeContactLicensing: 409 from Issues.Create for a plan-change
	// request naming PRO or ENTERPRISE (those editions come only with a
	// license). It goes to the licensing team as an issue, never a plan
	// change: Details["email"] is the licensing address, Details["id"] the
	// request, Details["existing"] true when an open one was answered and
	// nothing was filed.
	ErrCodeContactLicensing = "CONTACT_LICENSING"
)

// Sending-domain registration error codes, as APIError.Code (v1.19.0).
const (
	// ErrCodeDomainAlreadyRegistered: 409 from Domains.Add when the domain is
	// already registered, by this organisation or another one; the message is
	// the same in every case and never says who holds it. Nothing was
	// registered.
	ErrCodeDomainAlreadyRegistered = "DOMAIN_ALREADY_REGISTERED"
	// ErrCodeDomainLimitReached: 503 from Domains.Add when no domain can be
	// added at the moment (the email provider account's plan limit). A retry
	// does not help, so it is never retried; contact support.
	ErrCodeDomainLimitReached = "DOMAIN_LIMIT_REACHED"
)
