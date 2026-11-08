// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
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
// =============================================================================

package platformxe

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

// EmailHealthResponse is the health status of the email service.
type EmailHealthResponse struct {
	ProviderOrder []string                       `json:"providerOrder"`
	Providers     map[string]EmailProviderStatus `json:"providers"`
	Queue         QueueStats                     `json:"queue"`
}

// EmailProviderStatus is the status of an individual email provider.
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
}

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
	ReplayedCount int `json:"replayedCount"`
	SuccessCount  int `json:"successCount"`
	FailedCount   int `json:"failedCount"`
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

// EmitEventResult is the result of emitting a custom event.
type EmitEventResult struct {
	EventType         string              `json:"eventType"`
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

// WorkflowListResponse wraps a list of workflow triggers.
type WorkflowListResponse struct {
	Workflows []WorkflowTrigger `json:"workflows"`
}

// WorkflowEvalResult is the result of evaluating workflows against an event.
type WorkflowEvalResult struct {
	Matched int                  `json:"matched"`
	Total   int                  `json:"total"`
	Results []WorkflowEvalDetail `json:"results"`
}

// WorkflowEvalDetail is the evaluation result for a single workflow trigger.
type WorkflowEvalDetail struct {
	TriggerID   string `json:"triggerId"`
	TriggerCode string `json:"triggerCode"`
	TriggerName string `json:"triggerName"`
	Matched     bool   `json:"matched"`
	Reason      string `json:"reason,omitempty"`
	ExecutionID string `json:"executionId,omitempty"`
	Error       string `json:"error,omitempty"`
}

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
	Resolved          bool                   `json:"resolved"`
	LinkedIdentifiers []IdentityIdentifier   `json:"linkedIdentifiers"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
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

// IdentityLookupResult is the result of looking up an identity.
type IdentityLookupResult struct {
	Type      string                 `json:"type"`
	Value     string                 `json:"value"`
	Data      map[string]interface{} `json:"data"`
	Timestamp string                 `json:"timestamp"`
}

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

// BatchQRResult is the result of batch QR code generation.
type BatchQRResult struct {
	Results        []QRCodeResult `json:"results"`
	TotalGenerated int            `json:"totalGenerated"`
	Errors         []string       `json:"errors"`
}

// DryRunResult is the result of a workflow dry-run.
type DryRunResult struct {
	TriggerID   string                   `json:"triggerId"`
	TriggerName string                   `json:"triggerName"`
	Matched     bool                     `json:"matched"`
	Reason      string                   `json:"reason,omitempty"`
	Actions     []map[string]interface{} `json:"actions,omitempty"`
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
	// PdfBase64 is the base64-encoded PDF, <=10MB decoded.
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
	ErrCodeAiOutputBlocked     = "AI_OUTPUT_BLOCKED"     // 422
	ErrCodeAiOutputInvalid     = "AI_OUTPUT_INVALID"     // 422, Details["path"]
	ErrCodeAiQuotaExceeded     = "AI_QUOTA_EXCEEDED"     // 429, RetryAfter; never retried
	ErrCodeAiNotConfigured     = "AI_NOT_CONFIGURED"     // 503
	ErrCodeAiModelUnavailable  = "AI_MODEL_UNAVAILABLE"  // 503, Details["model"]; gateway account needs paid credits
	ErrCodeAiUpstreamError     = "AI_UPSTREAM_ERROR"     // 502
	ErrCodeVisionNotConfigured = "VISION_NOT_CONFIGURED" // 503
	ErrCodeServerError         = "SERVER_ERROR"          // 500
)

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
	Kind         string  `json:"kind"` // "ai" | "vision" | "embedding"
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
