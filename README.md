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
    _, err := client.Messaging.SendEmail(px.SendEmailInput{
        To:      []string{"user@example.com"},
        Subject: "Welcome",
        HTML:    "<h1>Hello</h1>",
    })
    if err != nil {
        panic(err)
    }

    // Check a permission
    decision, err := client.Permissions.Check("usr_123", "chat/session", "READ")
    if err != nil {
        panic(err)
    }
    fmt.Println(decision.Allowed)

    // Resolve a Nigerian identity
    profile, err := client.Identity.Resolve("BVN", "22012345678", true, "consent_abc")
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
| `client.Storage` / `client.Documents` | Media + fixed-storage upload, signed URLs |
| `client.Ocr` / `client.Pdf` / `client.Qr` | OCR, PDF generation, QR encoding |
| `client.Ai` | Product AI gateway — `Complete`, `Extract`, `Profile`, `Usage` (v1.7.0) |
| `client.Vision` | Image analysis — `AnalyzeRoom` (v1.7.0) |
| `client.Identity` | Identity resolution + Nigerian KYC (BVN, NIN, liveness, face match) |
| `client.Fraud` | Fraud Detection — rules, screens, devices, federation, terms |
| `client.Permissions` | RBAC + ABAC + ReBAC + Federation. `Check`, `CheckBatch`, `Resolve`, roles, modules, overrides, policies, relationships |
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
| `client.Search` | Federated search query |
| `client.Issues` | Tenant-reported issues / support cases |
| `client.Whoami` | API-key resolution + telemetry helpers |
| `px.Register` | Package-level standalone bootstrap helper |

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

Errors are `*APIError` with the closed code set (`px.ErrCodeAi*`, `px.ErrCodeVisionNotConfigured`, `px.ErrCodeBadRequest`, `px.ErrCodeServerError`). `apiErr.Details["path"]` names the failing value on `AI_OUTPUT_INVALID`; `apiErr.RetryAfter` is the Retry-After seconds (to the next UTC midnight) on `AI_QUOTA_EXCEEDED`, which is never retried. `px.ErrCodeAiModelUnavailable` (`AI_MODEL_UNAVAILABLE`, 503) means the AI gateway account cannot use the requested model — paid gateway credits are required; `apiErr.Message` is the fixed text "The AI service account cannot use this model right now." and `apiErr.Details["model"]` names the model. It is not `AI_NOT_CONFIGURED` (key missing or rejected) and retrying does not fix it.

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
| `github.com/calderaxe/platformxe-go` (this) | **v1.11.0** *(pending release — tag `sdk-go/v1.11.0`)*; v1.10.0 released (tag `sdk-go/v1.10.0`) |
| `@caldera/platformxe-sdk` (TypeScript) | 1.14.0 *(pending release)*; 1.13.0 published |
| `@caldera/platformxe-types` (TypeScript shapes) | 3.11.0 |
| `platformxe` (Python) | 1.11.0 *(pending release — tag `sdk-python/v1.11.0`)*; 1.10.0 released |
| `calderaxe/platformxe` (Terraform) | 1.5.1 |

The PlatformXe ecosystem ships **5 published artefacts** that move in lockstep against every API change (the NO-DRIFT policy). The full coverage matrix lives at [docs.platformxe.com/sdk/alignment](https://docs.platformxe.com/sdk/alignment).

**v1.11.0 changes** *(pending — tag `sdk-go/v1.11.0`, 2026-11-08)*: `Permissions.CheckBatch` decodes the bare array the route answers into `Results` (it decoded `{"results": [...]}`, so `Results` was always empty), and still reads an object with `results`. `BatchCheckResult` gains `Source` and `RoleID`. No signature change.

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
