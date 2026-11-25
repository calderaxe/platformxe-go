# PlatformXe Go SDK

Official Go SDK for the [PlatformXe](https://platformxe.com) API — messaging, storage, OCR, PDF, product AI, vision, identity, fraud detection, webhooks, workflows, custom events, and authorization (RBAC + ABAC + ReBAC + Federation).

[![Go Reference](https://pkg.go.dev/badge/github.com/calderaxe/platformxe-go.svg)](https://pkg.go.dev/github.com/calderaxe/platformxe-go)
[![Go version](https://img.shields.io/badge/go-1.22%2B-blue)](https://go.dev/)

## Install

```bash
go get github.com/calderaxe/platformxe-go@v1.7.1
```

The SDK targets Go 1.22+ and has no third-party dependencies — only the standard library.

## Quick start

```go
package main

import (
    "fmt"

    px "github.com/calderaxe/platformxe-go"
)

func main() {
    client := px.NewClient(px.ClientConfig{
        APIKey: "pxk_live_…",
    })

    // Send a transactional email
    _, err := client.Messaging.SendEmail(map[string]interface{}{
        "from":    "Your App <hello@yourapp.com>",
        "to":      []string{"user@example.com"},
        "subject": "Welcome",
        "html":    "<h1>Hello</h1>",
    })
    if err != nil {
        panic(err)
    }

    // Resolve a Nigerian identity
    // The key must name identity:resolve; purpose is required.
    profile, err := client.Identity.ResolveRequest(platformxe.IdentityResolveRequest{
        Type:    "BVN",
        Value:   "22012345678",
        Purpose: "tenant onboarding",
        Consent: platformxe.IdentityResolveConsent{Granted: true, Reference: "consent_abc"},
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(profile)
}
```

## Configuration

```go
client := px.NewClient(px.ClientConfig{
    APIKey:         "pxk_live_…",             // required
    BaseURL:        "https://platformxe.com", // optional — defaults to production
    Timeout:        10,                       // optional — seconds
    Retries:        2,                        // optional — 0 means the default (2); negative = none
    DisableRetries: false,                    // optional — true = exactly one attempt
    FailOpen:       false,                    // optional — a transport failure returns a _failed result
})
```

`BaseURL` accepts the regional production endpoint or a self-hosted PlatformXe instance.

Retries back off from 200 ms, doubling each time. Transport failures, `429`, `502`, `503` and `504` are retried for every method; `500` only for `GET`, `HEAD`, `PUT`, `DELETE` and `OPTIONS`, because PlatformXe answers a handler failure with `500` after side effects may already have landed. A `POST`/`PATCH` `500` and every other non-2xx return an error after one attempt.

## Surface

| Namespace | Domain |
|-----------|--------|
| `client.Messaging` | Transactional email, SMS, WhatsApp click-to-chat |
| `client.Storage` / `client.Documents` | Media + fixed-storage upload, signed URLs; `ListCategories` / `CreateCategory` / `UpdateCategory` / `DeactivateCategory`, `GetSettings` / `UpdateSettings` (v1.12.0) |
| `client.Ocr` / `client.Pdf` / `client.Qr` | OCR, PDF generation (`Pdf.Generate`: HTML or a stored template to a stored PDF with a signed URL, or base64, v1.14.0), QR encoding (`Qr.GenerateBatchItems`: stored QR images and an optional zip, v1.14.0) |
| `client.Ai` | AI & Vision gateway — `Complete` (with `Toolsets`, v1.16.0), `Messages` (native Claude messages, v1.16.0; `Actor`, v1.17.0; `Actors`, v1.18.0), `Extract`, `Profile`, `Usage` (v1.7.0) |
| `client.AiTools` | The platform tools a key may use in the server-side tool loop: `List()` (scope `tools:platform.read`, v1.16.0) |
| `client.Vision` | Image analysis — `AnalyzeRoom` (v1.7.0) |
| `client.Identity` | Identity resolution + Nigerian KYC (BVN, NIN, liveness, face match) |
| `client.Audit` | Decision + mutation audit log query/export |
| `client.Webhooks` | Outbound webhook endpoints |
| `client.Templates` | Content templates |
| `client.Workflows` | Event-driven automations |
| `client.Domains` | Sending domain management |
| `client.Events` | Built-in event ingestion + log + subscriptions |
| `client.Events.Custom` | Tenant-defined custom events (Phase 9A) |
| `client.Events.Custom.Marketplace` | Cross-tenant marketplace (Phase 9C, PRO+) |
| `client.Events.Custom.Federation` | Cross-org event fan-out (Phase 9D + Pattern 3, ENTERPRISE) |
| `client.Threads` | Caldera Threads contextual messaging |
| `client.Exports` | Async data export jobs |
| `client.Usage` | Usage + billing reads |
| `client.Billing` | The key's organisation's invoices: `ListInvoices` / `GetInvoice` (scope `billing:read`, money as kobo strings, v1.14.0) |
| `client.License` | The key's organisation's Enterprise License: `Get()` (scope `license:read`; effective plan and its source, exemption, and the covering license's state, validity, edition and the plugins granted to this organisation; never money, v1.15.0) |
| `client.Search` | Federated search query |
| `client.Issues` | Tenant-reported issues / support cases |
| `client.Whoami` | API-key resolution + telemetry helpers |
| `px.Register` | Package-level standalone bootstrap helper |

Fraud Detection (`client.Fraud`) and Authorization (`client.Permissions`) are documented in the gated sections of [docs.platformxe.com](https://docs.platformxe.com/services/fraud-detection/overview), available on request (sales@platformxe.com).

## Product AI and vision (v1.7.0)

Every call runs under the AI profile for *(your API key's caller service, `FeatureKey`)*, which sets the allowed providers / models, size limits, policy and daily caps. Scopes: `ai:complete` (600 requests/hr/key) for `client.Ai`, `vision:analyze` (2,000/hr/key) for `client.Vision`.

```go
reply, err := client.Ai.Complete(px.AiCompleteInput{
    FeatureKey: "guest-concierge",
    Provider:   "gemini",
    Messages:   []px.AiMessage{{Role: "user", Content: "What time is check-out?"}},
})

invoice, err := client.Ai.Extract(px.AiExtractInput{
    FeatureKey:       "document.freeform",
    SystemPrompt:     "Extract the invoice fields.",
    UserPrompt:       text,
    OutputJSONSchema: map[string]interface{}{"type": "object"},
    ModelTier:        "FAST",
})

profile, err := client.Ai.Profile("guest-concierge")                        // limits + today's usage
usage, err := client.Ai.Usage(px.AiUsageQuery{FeatureKey: "guest-concierge"}) // default: last 30 days

room, err := client.Vision.AnalyzeRoom(
    px.AnalyzeRoomInput{ImageURL: imageURL, RoomCategories: []string{"KITCHEN", "BEDROOM"}, WantAltText: true},
    &px.AnalyzeRoomOptions{IdempotencyKey: "photo-" + photoID}, // sent as x-idempotency-key
)
if err == nil && room.AiStatus == "FALLBACK" {
    // provider timeout: a deterministic answer, never cached under the key
}
```

Errors are `*APIError` with the closed code set (`px.ErrCodeAi*`, `px.ErrCodeVisionNotConfigured`, `px.ErrCodeBadRequest`, `px.ErrCodeConflict`, `px.ErrCodeServerError`). `px.ErrCodeConflict` (`CONFLICT`, 409) comes from `Knowledge.UpsertDocument` when the collection + sourceRef is held by another organisation's document under the same caller service (until the organisation-keyed contract migration); it is never retried, and a different `SourceRef` is the fix. `apiErr.Details["path"]` names the failing value on `AI_OUTPUT_INVALID`; `apiErr.RetryAfter` is the Retry-After seconds (to the next UTC midnight) on `AI_QUOTA_EXCEEDED`, which is never retried. `px.ErrCodeAiModelUnavailable` (`AI_MODEL_UNAVAILABLE`, 503) means the AI gateway account cannot use the requested model — paid gateway credits are required; `apiErr.Message` is the fixed text "The AI service account cannot use this model right now." and `apiErr.Details["model"]` names the model. It is not `AI_NOT_CONFIGURED` (key missing or rejected) and retrying does not fix it.

### Native messages and server-side tools (v1.16.0)

`client.Ai.Messages(px.AiMessagesInput{...})` (`POST /api/v1/ai/messages`, scope `ai:complete`) takes the Anthropic Messages request shape on the Claude 5.5 models, governed by the same AI profile as `Complete`: `System`, `Messages` (content a string or content blocks), `Effort` (capped by the profile's `maxEffort`), `OutputSchema` (when the profile allows structured output), caller `Tools` with `ToolChoice` "auto" or "none" ("auto" or empty with `Toolsets`), `MaxTokens`. No streaming; sent once. After a server-side fallback the reply's `Content` starts at the last `fallback` block (that block included). `OutputSchema` needs an object root, every object node `additionalProperties: false`; `pattern`, `minItems` 0 or 1 and the supported `format`s pass, numeric and length bounds are refused (400). `Toolsets: []string{"platform.read"}` (on `Messages`, or on `Complete` with a Claude 5.5 model named by its gateway slug, e.g. `"anthropic/claude-sonnet-5.5"`; `Messages` takes the Anthropic id, e.g. `"claude-sonnet-5-5"`) runs PlatformXe's read-only tools in a loop over the key's own organisation; the key needs `tools:platform.read` and each toolset must be on the profile. The answer's `ToolRuns` lists each tool call. `client.AiTools.List()` (`GET /api/v1/ai/tools`) answers the tools the key may use.

```go
tools, err := client.AiTools.List() // Items: Name, Version, Description, InputSchema, DataClasses, Toolset
reply, err := client.Ai.Messages(px.AiMessagesInput{
    FeatureKey: "ops-assistant",
    Model:      "claude-sonnet-5-5",
    Messages:   []px.AiNativeMessage{{Role: "user", Content: "How many webhook deliveries failed today?"}},
    Toolsets:   []string{"platform.read"},
})
```

New closed codes: `ErrCodeAiBudgetExceeded` (429, a spend budget; `RetryAfter` runs to the period's end; also on vision; final; `Details["reason"]` is `AiBudgetReasonUnavailable` when a hard budget could not be checked, and a tool-loop refusal has no `Details["period"]`), `ErrCodeAiRefused` (422, the model declined; no content, the call is billed; `Details["category"]`), `ErrCodeAiToolsetDenied` (403), `ErrCodeAiToolFailed` (502, the tool round trip itself broke; final) and `ErrCodeAiLoopLimit` (422, `Details["reason"]` TURNS, CALLS or DEADLINE).

### Acting for a person (v1.17.0)

`AiMessagesInput.Actor` (`&px.AiToolActor{Ref, Permissions, Grant}`) names the person a first-party Caldera app's own tools act for (types 3.19.0 `actor`). `Ref` is the person's id in your app (1 to 64 of A-Z a-z 0-9 _ -); `Permissions` holds at most 32 canonical prefixed keys (e.g. `"CN:REQUESTS:VIEW"`; nil is sent as `[]`, never null); `Grant` is opaque (at most 4,096 base64url characters; `""` omits it), forwarded only to the tool's own app and never stored, logged or returned. A first-party app's tool is offered only when `Permissions` holds its `RequiredPermission`; `AiToolListItem` carries `SourceApp` and `RequiredPermission`. Only the app's own organisation can use its tools (scope `tools:{app}.read`, named on the key). When none of the requested toolsets has a usable tool, the call answers 403 `ErrCodeAiToolsetDenied` before any model call, and `Details["reason"]` says why: `"SOURCE_UNAVAILABLE"` (the app's tools are switched off or cannot be read: fall back to a path without tools), `"DATA_CLASS"` (refused by the profile's data classes) or `"NOT_ALLOWED"` (not on the profile, or the key or actor may use none of them). Fall back only on `"SOURCE_UNAVAILABLE"`. With several toolsets requested, the call runs (and is billed) when any one still has a usable tool. `ErrCodeAiToolFailed` (502) comes after model spend.

### Caldera's operator console (v1.18.0)

`AiMessagesInput.Actors` (`map[string]px.AiToolActorsEntry{"lettings": {Ref, Permissions, Grant}, ...}`) is for Caldera's own operator console, which calls with a platform key (types 3.20.0 `actors`). One entry per Caldera app (`"xadmin"`, `"lettings"`, `"concierge"`, `"chats"`): that app's id for the person, at most 64 permission keys with that app's prefix (`XA:`, `LT:`, `CN:`, `CH:`; nil is sent as `[]`) and that app's own `Grant` (required). Each entry is forwarded only to its own app's tools. `Actor` with `Actors` returns `ErrAiActorAndActors` before any request; an organisation key sending `actors`, or a platform key sending `actor`, is 400.

## Pattern 3 — external webhook peers (1.5.0)

Custom Event Federation supports peers addressed by URL + HMAC secret instead of a tenant org id — useful when the receiving system isn't on PlatformXe.

```go
result, err := client.Events.Custom.Federation.AddExternalPeer(
    groupID,
    px.AddEventFederationExternalPeerInput{
        Label:      "Booking.com",
        WebhookURL: "https://booking.example.com/inbound/platformxe",
        Headers:    map[string]string{"Authorization": "Bearer xyz"},
    },
)
if err != nil {
    panic(err)
}
fmt.Println(result.Secret)               // "whsec_…" — store immediately, shown ONCE
fmt.Println(result.Peer.PeerType)        // px.EventFederationPeerTypeExternalWebhook

// Remove later:
_, _ = client.Events.Custom.Federation.RemoveExternalPeer(result.Peer.ID)
```

The receiving endpoint verifies inbound POSTs with `X-Caldera-Signature-V2: sha256=<hex>` = HMAC-SHA256(secret, `{X-Caldera-Timestamp}.{raw body}`), a 5-minute timestamp window and `X-Caldera-Delivery-Id` as the replay key. The legacy body-only `X-Caldera-Signature` (HMAC of the raw body alone) is replayable and is sent for compatibility only until **2027-04-30**, when `X-Caldera-Signature` switches to the V2 value. See the [federation reference](https://docs.platformxe.com/sdk/federation) for the full wire format.

## Error handling

Every method returns an `error` on failure. Every non-2xx response — a JSON envelope, other JSON, or a non-JSON body such as a gateway's HTML error page — is an `*APIError` carrying `StatusCode`, plus `Code` and `Message` from the body when it has them (otherwise `HTTP_<status>` and the status text). Setting `FailOpen: true` makes a request whose last attempt failed at the transport level return a `{"_failed": true, "error": …}` result with a `nil` error — useful for non-blocking permission checks. It never covers an HTTP error response.

```go
decision, err := client.Permissions.Check(adminID, path, action)
if err != nil {
    var apiErr *px.APIError
    if errors.As(err, &apiErr) {
        // apiErr.Code, apiErr.Message, apiErr.StatusCode — plus apiErr.Details
        // (the envelope's error.details) and apiErr.RetryAfter (seconds, on a 429)
    }
    return err
}
```

## Versioning

| Package | Version |
|---------|---------|
| `github.com/calderaxe/platformxe-go` (this) | **v1.19.0** (domain registration codes), v1.18.0 (AI gateway P6), v1.17.0 (AI gateway P5), v1.16.0 (AI gateway), v1.15.0 (Enterprise License), v1.14.1, v1.14.0 and v1.13.0 *(pending, not tagged)*; v1.12.0 released (tag `sdk-go/v1.12.0`) |
| `@caldera/platformxe-sdk` (TypeScript) | 1.21.0 *(pending, AI gateway P5: `actor`, spoke tools)*; 1.20.0 *(pending, AI gateway, after 1.19.0)*; 1.19.0 *(pending, Enterprise License)*; 1.18.2 released |
| `@caldera/platformxe-types` (TypeScript shapes) | 3.19.0 *(pending, AI gateway P5: `actor`, spoke tools)*; 3.18.0 *(pending, AI gateway, after 3.17.0)*; 3.17.0 *(pending, Enterprise License)*; 3.16.2 released |
| `platformxe` (Python) | 1.18.0 *(pending, AI gateway P6: `actors`)*, 1.17.0 *(pending, AI gateway P5: `actor`)*, 1.16.0 *(pending, AI gateway)*, 1.15.0 *(pending, Enterprise License)*, 1.14.1, 1.14.0 and 1.13.0 *(pending, not tagged)*; 1.12.0 released |
| `calderaxe/platformxe` (Terraform) | 1.5.4 and 1.5.3 *(pending, not tagged)*; 1.5.2 released |

The PlatformXe ecosystem ships **5 published artefacts** that move in lockstep against every API change (the NO-DRIFT policy). The full coverage matrix lives at [docs.platformxe.com/sdk/alignment](https://docs.platformxe.com/sdk/alignment).

**v1.19.0 changes** *(pending, not tagged; 2026-11-24, sending-domain registration codes)*: `Domains.Add` (`POST /api/v1/domains`) now answers `ErrCodeDomainAlreadyRegistered` (409 `DOMAIN_ALREADY_REGISTERED`: the domain is already registered, by your organisation or another one; the message never says who holds it) and `ErrCodeDomainLimitReached` (503 `DOMAIN_LIMIT_REACHED`: no domain can be added at the moment; contact support). They replace 409 `CONFLICT` and a generic 502. *Behaviour change:* `DOMAIN_LIMIT_REACHED` joins the final codes (returned after one attempt; it was retried as any 503).

**v1.18.0 changes** *(pending, not tagged; 2026-11-23, AI gateway P6, NO-DRIFT with types 3.20.0)*: `AiMessagesInput.Actors` (`map[string]AiToolActorsEntry`: `Ref`, `Permissions`, `Grant`; a nil `Permissions` is sent as `[]`), `ErrAiActorAndActors`.

**v1.17.0 changes** *(pending, not tagged; 2026-11-23, AI gateway P5, NO-DRIFT with types 3.19.0)*: `AiMessagesInput.Actor` (`AiToolActor`: `Ref`, `Permissions`, `Grant`; a nil `Permissions` is sent as `[]`, never null), `AiToolListItem.SourceApp` and `RequiredPermission`.

**v1.16.0 changes** *(pending, not tagged; 2026-11-22, AI gateway, NO-DRIFT with types 3.18.0)*: `Ai.Messages(AiMessagesInput)` (`POST /api/v1/ai/messages`), `AiTools.List()` (`GET /api/v1/ai/tools`), `AiCompleteInput.Toolsets` / `AiCompleteResult.ToolRuns` (`AiToolRunSummary`), the `AiProfileView` governance fields (`MaxEffort`, `AllowCaching`, `AllowStructuredOutput`, `FallbackModel`, `Toolsets`, `MaxToolTurns`, `EndUserFacing`, `AllowedDataClasses`, `StrictIDRedaction`), and `ErrCodeAiBudgetExceeded`, `ErrCodeAiRefused`, `ErrCodeAiToolsetDenied`, `ErrCodeAiToolFailed`, `ErrCodeAiLoopLimit`. *Behaviour change:* `AI_BUDGET_EXCEEDED` and `AI_TOOL_FAILED` join the final codes (returned after one attempt).

**v1.15.0 changes** *(pending, not tagged; 2026-11-22, Enterprise License)*: new `client.License.Get()` (`GET /api/v1/license`, scope `license:read`; `LicenseView{Edition{Plan, Source}, Bypass, License}`, `License` nil when no license covers the organisation). `ErrCodeLicenseRequired` (403 from `Fraud.Activate` after the licensing cut-over, no live license grant naming a fraud pack; its `billing` may be `LICENSED`) and `ErrCodeContactLicensing` (409 from `Issues.Create` for a plan-change request naming PRO or ENTERPRISE: sent to the licensing team as an issue, never a plan change; `Details["email"]` is the licensing address, `Details["id"]` the request, `Details["existing"]` true when an open one was answered). `Exports.Create` of `dataType` `invoices` also needs `billing:read` (403), and `Exports.Get` of one is 404 without it.

**v1.14.1 changes** *(pending, not tagged; 2026-11-22)*: docs and tests only, no behaviour change. `Workflows.Evaluate` / `EvaluateWithOptions` and `Events.Ingest` document `entityData` as your own parameters for the event (what the triggers' conditions and templates read; for a business event type PlatformXe defines no fields) and `eventType` as one of the values on [Workflow Event Types](https://docs.platformxe.com/services/workflows/event-types); the 400 for any other value, or an unknown entity type, never lists the values. `Evaluate` does not send a nil or empty map (`omitempty`; the route reads none); `Ingest` sends an empty map as `{}` and omits a nil one.

**v1.14.0 changes** *(pending, not tagged; 2026-11-20)*: Spec 03 PM-1 (2026-11-21): **Fix:** `DecideRequest` marshals to the body the decide routes read, `{subject: {id, kind}, action, resource: {kind, id?}, context?}` (it sent flat `subjectId` / `subjectKind` / `resourceKind`: every `Fraud.Decide` / `ShadowDecide` was a 400); the struct is unchanged. `ErrCodeCardNumberRefused` (400 `CARD_NUMBER_REFUSED`: card data in a decide request). `Pdf.Generate(GeneratePdfInput, ...*GeneratePdfCallOptions)` (`POST /api/v1/pdf/generate`, scope `pdf:generate`, 100/hr per key): `HTML` or `TemplateID` + `Data`, `GeneratePdfOptions` (size, orientation, margins, header / footer, filename, remote images, base64, URL lifetime); stored answer `FileID`, `PdfURL`, `ExpiresAt`, base64 answer `Content`, `ContentType`, both with `Size`, `Pages`, `GeneratedAt`. **Fix:** `Qr.GenerateBatchItems(QrBatchInput, ...*QrBatchCallOptions)` for `POST /api/v1/qr/batch` (scope `storage:upload`, 120/hr per key; the route did not exist before, so `GenerateBatch` was a 404): `BatchQRResult` gains `Items` (`QrBatchItemResult`: generated with `FileID` / `SignedURL` / `ExpiresAt`, or failed with `Error`), `TotalFailed`, `Zip`, `GeneratedAt`; `Results` and `Errors` were never sent and are Deprecated; `GenerateBatch(map)` is kept for the same route. New `Billing.ListInvoices(ListInvoicesParams)` / `GetInvoice(id)` (`GET /api/v1/billing/invoices` and `/{id}`, id path-escaped, scope `billing:read`; money as kobo strings). `IdempotencyKey` sends `x-idempotency-key`; without one, a 5xx or a transport failure on these POSTs is not re-sent. New codes `ErrCodeValidationError`, `ErrCodeForbidden`, `ErrCodeNotFound`, `ErrCodeTemplateNotFound`, `ErrCodeRenderTimeout`, `ErrCodePayloadTooLarge`, `ErrCodeOutputTooLarge`, `ErrCodeServiceDisabled`, `ErrCodeRateLimited`, `ErrCodeStorageUnavailable`, `ErrCodeRendererUnavailable`, `ErrCodeStorageFailed`.

**v1.13.0 changes** *(pending, not tagged; 2026-11-19)*: **Fix:** `Events.Replay` sends `fromTimestamp` / `toTimestamp`, the fields the route reads (it sent `from` / `to`: every replay was a 400); new `Events.ReplayTypes` also sends `eventTypes`; `EventReplayResult.Replayed` is the count the route answers (`ReplayedCount`, `SuccessCount`, `FailedCount` were never sent and are Deprecated). `EventSubscription.RestrictedEventTypes`. Fraud rules: `Rules.TransitionAtVersion` / `UpdateAtVersion` send `expectedVersion` (409 `ErrCodeVersionMismatch`), `Rules.ShadowReportForVersion` and `FraudRuleVersion`. New codes `ErrCodeRestrictedEventType` (400), `ErrCodeReservedSearchApp` (400), `ErrCodeFeedMissing` and `ErrCodeFeedNotRealtime` (409 on a publish). Subscription and rule ids are path-escaped. `Workflows.DryRun` posts `{entityData}` to `/api/v1/workflows/{id}` (it posted to a `/dry-run` path that does not exist: every call was a 404) and returns the route's result (`DryRunResult.WouldMatch`, `ConditionsEvaluated`, `ActionsWouldExecute`, `ActionsPreview`, `Validation`; `TriggerID`, `TriggerName`, `Matched` and `Actions` were never answered and are Deprecated); new `Workflows.DryRunEntityData`. `Storage.ReorderFiles` sends `PATCH` (it sent `POST`, which the route does not take). **Fix:** `Workflows.List` decodes `{triggers, pagination}` into `WorkflowListResponse.Triggers` / `Pagination` (it read `workflows`, which the route never answered, so the list was always empty; `Workflows` is Deprecated and filled from `Triggers`); `Create`, `Get`, `Update` and `Delete` decode the trigger under `trigger` (they decoded the top level, so the trigger came back empty and `Delete` said `Deleted: false` on success; `Delete` answers `Deleted: true` when the trigger was switched off). The workflow trigger id is path-escaped and an empty id returns `ErrWorkflowIDRequired` before any request. `DryRun` reads the context as Python does (a context whose every key is a trigger-context key (`eventType`, `entityType`, `entityId`, `entityData`, `payload`, `metadata`, `sourceApp`, `dryRun`, `organizationId`, `callerService`) and that carries an `entityData` (else `payload`) object sends that object; any other map is the entity data, sent whole) and fills the deprecated `TriggerID`, `Matched` and `Actions`. `Events.Ingest` sends a deprecated `payload` key as `entityData` when the map has none (the route reads `entityData` only, so the event data never reached a trigger) and never sends `payload`; `EmitEventResult` gains `EntityType` and `EntityID`. `Workflows.List` takes an optional `*WorkflowListOptions` (`Page`, `Limit`, `IsActive`, `TriggerType`, `EventType`: the route's query); `List()` still lists the first page (it could not reach page 2), and more than one option is an error. **Compile-time note:** the method's type is now variadic, so code that stores `client.Workflows.List` as a `func() (*WorkflowListResponse, error)` value, or an interface with `List() (...)`, must change; calls `List()` compile as before. Every request refuses a path with a `.` or `..` segment before it is sent (`ErrDotSegmentPath`): an id of `.` or `..` would address another route. `Workflows.Update` can now change `code`, `triggerType`, `eventType` and `entityType` (the PATCH route took none of them; a code in use is 409 `DUPLICATE`), and a `nil` map value clears a nullable field. Pairs with the Terraform provider 1.5.3, whose `go.mod` requires this version.

**v1.12.0 changes** *(pending, not tagged; 2026-11-18)*: `Identity.ResolveRequest(IdentityResolveRequest)` sends the route's body with the required `purpose` (1 to 200 characters; the API answers 400 without it since 2026-11-18). The key must name the `identity:resolve` scope: a key of an organisation with no scopes or `*` is answered 403. `Identity.Resolve` is Deprecated: it has no purpose, so the API answers every call 400; it now sends the route's field names (`type`, `consent.granted`, `consent.reference`) where it sent `identifierType` / `consentRef`. `IdentityResolveResult` gains `Identifiers`, `Profile`, `Confidence`, `Provider`, `ResolvedAt` and `Outcome` (`found`, `not_found` billed, `provider_down` not billed); `LinkedIdentifiers` and `Metadata` were never returned and are Deprecated. Also `Identity.VerifyBusiness` (2026-11-13). Pairs with a `@caldera/platformxe-types` minor adding `IdentityResolveRequest.purpose` and the `outcome` field. Custom-events federation: `EventFederationGroupSummary.CreatedBy` and `EventFederationPushSummary.PushedBy` are `*string` (breaking at compile time for code that reads them as `string`): the API shows the actor labels (`createdBy`, `pushedBy`, `unpushedBy`) only to their own organisation (the group owner; the push's source) and null to every other member.

**Fix:** `Fraud.Rules.Publish` and `Archive` post to `/api/v1/fraud/rules/:id/transition` with `{"target": ...}` (they posted to `/publish` and `/archive`, which do not exist: every call was a 404). New `Fraud.Rules.Transition(id, target)`, `Draft`, `Shadow`, the `FraudRuleStatus*` constants and `FraudRuleTransitions` (a draft rule goes to shadow before it can be published), and `Fraud.Rules.ShadowReport(id, from, to)`. **Fix:** `IdentityLookupResult` has the lookup route's real shape (`Found`, the identifier fields, `PhoneNumbers`, `Profile`, `Raw.Provider`, `Attempted`, `Outcome`, `NotFound`, `NotFoundBy`); its `Type` / `Value` / `Data` / `Timestamp` were never returned and are Deprecated. New fixed-storage category registry and settings: `Documents.ListCategories`, `CreateCategory`, `UpdateCategory`, `DeactivateCategory`, `GetSettings`, `UpdateSettings`. Error codes `ErrCodeReplayInProgress` (409 from `Identity.Dlq.Replay`), `ErrCodeInternalError` (a replay that did not complete; the row stays unreplayed) and `ErrCodeCaseAtMlroStage` (409 from `Fraud.Cases.Transition`). `RasterizePdfInput.PdfBase64` documents the real limit: the 4.5 MB request body, about 3.3 MB of PDF.

**v1.11.0 changes** *(released — tag `sdk-go/v1.11.0`, 2026-11-08)*: `Permissions.CheckBatch` decodes the bare array the route answers into `Results` (it decoded `{"results": [...]}`, so `Results` was always empty), and still reads an object with `results`. `BatchCheckResult` gains `Source` and `RoleID`. No signature change.

**v1.9.0 changes** *(released)*: `Fraud.Activate(idempotencyKey)` turns the Detection Pack addon on for the key's own organisation (`POST /api/v1/fraud/activate`, scope `fraud:manage`); `Fraud.Terms.Status()` now answers `detectionPackEnabled` / `detectionPackEnabledAt`, also before the pack is on. Pairs with `@caldera/platformxe-types` 3.10.0.

**v1.8.0 changes** *(pending — tag `sdk-go/v1.8.0`)*: `Identity.VerifyRequest(IdentityVerifyRequest)` sends `matchAgainst.dateOfBirth` and `supporting.nin` (the holder's 11-digit NIN, `PASSPORT` only; any other type is a 400), which a Nigerian passport check needs. `Identity.Verify` now sends the route's real body `{ type, value, matchAgainst }`; it used to send `identifierType` and flat names, which the API answered 400. `IdentityVerifyResult` gains `Profile`, `Provider` and `AttemptedProviders`; `MatchDetails` was never returned and is Deprecated. Pairs with `@caldera/platformxe-types` 3.9.0.

**v1.7.2 changes** *(pending — tag `sdk-go/v1.7.2` cut on merge; first release under `github.com/calderaxe/platformxe-go`)*: response shapes corrected to what the API returns, with no signature change. `Identity.Providers()` entries carry `Healthy` (the API never sent `status` / `lastChecked` / `responseTime`; those stay as Deprecated). `Events.CreateSubscription` decodes `{ subscription, secret }`, so `ID` is set and `Secret` carries the one-time signing secret. `Events.ListSubscriptions` reads `Items` / `Total`, and the Deprecated `Subscriptions` is filled from them. `Permissions.ListOverrides` decodes the bare array the API returns; it used to fail on every call. Federation members carry `ID` / `GroupID`.

**v1.7.1 changes:** a redirect is followed only to the same scheme, host and port as the request (at most 3 hops); any other redirect is returned to the caller as the 3xx `*APIError`. Go's default policy forwarded the `x-api-key` header to a cross-host redirect target (it strips only `Authorization` / `Cookie`). No API shape change.

**v1.7.0 changes:** `client.Ai` / `client.Vision`; `*APIError` gains `Details` and `RetryAfter`; a 429 `AI_QUOTA_EXCEEDED` is returned after one attempt (every other 429 is still retried).

**v1.6.0 behaviour changes** (released with the `sdk-go/v1.6.0` tag, cut on merge):

- Any non-2xx response is an `*APIError`. Before, a non-2xx whose body was not a `{"success": false}` envelope (for example a 5xx `{"error": "..."}`) came back as data with a `nil` error, and a non-JSON body surfaced as an unmarshal error.
- A `POST`/`PATCH` `500` is no longer retried, and neither is a 5xx other than `500`/`502`/`503`/`504`. `429` and `502`/`503`/`504` are retried for every method.
- `FailOpen` covers transport failures only: a retried HTTP error status still failing after the last retry returns the `*APIError`.
- `DisableRetries: true` configures a single attempt (`Retries: 0` means the default of 2).

## Distribution

The Go SDK is hosted at `github.com/calderaxe/platformxe-go`. The source of truth lives in the PlatformXe monorepo at `packages/sdk-go/` and is mirrored to GitHub on every `sdk-go/v*` tag push so `go get …@v<semver>` resolves directly through the Go module proxy.

```bash
go get github.com/calderaxe/platformxe-go@v1.7.1
```

```go
import px "github.com/calderaxe/platformxe-go"
```

## Documentation

- **API reference + per-domain SDK guides:** [docs.platformxe.com](https://docs.platformxe.com)
- **Go reference:** [pkg.go.dev/github.com/calderaxe/platformxe-go](https://pkg.go.dev/github.com/calderaxe/platformxe-go)
- **OpenAPI 3.1 spec:** `https://api.platformxe.com/api/docs/openapi.json`
- **Status + incidents:** [status.platformxe.com](https://status.platformxe.com)

## License

Proprietary — © 2026 Caldera Technologies Ltd. All rights reserved.
