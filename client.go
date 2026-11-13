// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-09-15):
// NewClient treated an explicit Retries: 0 as unset and defaulted it to 2,
// so retries could never be disabled (TestNewClientCustomConfig failed on
// development too). ClientConfig.DisableRetries is now the explicit switch
// and a negative Retries clamps to 0; omitting both still yields 2. Also
// gofmt alignment of the v1.1.0 service fields.
//
// Also 2026-09-15 (MR !199 round 9): doRequest never checked the HTTP status —
// it only failed when the body said "success": false — so a non-2xx whose
// JSON was not an envelope (for example a 5xx {"error":"..."}) was returned as
// data with a nil error. Any non-2xx is now an *APIError with the status and
// the parsed code/message (the status text for a non-JSON body). 5xx and 429
// are retried on the existing transport backoff and, once retries run out,
// return the last *APIError (FailOpen still covers transport failures only);
// other 4xx return immediately. 2xx handling is unchanged, and the
// success:false error keeps its "[code] message (HTTP n)" text.
//
// Also 2026-09-15 (MR !199 round 10, finding 3): retrying every 5xx repeated
// non-idempotent requests. The server answers a handler throw with 500
// precisely because side effects may have landed, and this client sends no
// idempotency key, so a retried POST could send an email or create a record
// twice. 429 and 502/503/504 (the request did not reach a handler, or was
// refused before it ran) are retried for every method; 500 only for
// GET/HEAD/PUT/DELETE/OPTIONS; a POST/PATCH 500 and any other 5xx return the
// *APIError at once. Released as module v1.6.0.
//
// CHANGELOG (2026-10-08, v1.7.0 — caldera-platformxe#22):
// - Client.Ai (AiService) and Client.Vision (VisionService) for the AI
//   gateway and vision routes.
// - doRequest takes optional per-request headers (doRequestWithHeaders), so
//   VisionService.AnalyzeRoom can send x-idempotency-key. Every existing
//   call goes through the unchanged doRequest signature.
// - A non-2xx *APIError now carries the envelope's error.details (Details —
//   e.g. "path" on 422 AI_OUTPUT_INVALID) and the Retry-After header in
//   seconds (RetryAfter).
// - A 429 AI_QUOTA_EXCEEDED is returned after one attempt: it is an AI
//   profile's DAILY cap (Retry-After runs to the next UTC midnight), so the
//   200/400 ms backoff can never succeed. Every other 429 is retried as before.
//
// CHANGELOG (2026-10-08, v1.7.0 pending — calderasuite/caldera-xadmin#29):
// - Client.Knowledge (KnowledgeService) for /api/v1/ai/knowledge/*.
//
// CHANGELOG (2026-10-09, v1.7.0 pending — final review):
// - Per-call options (callOptions: timeout, retries, no fail-open) through
//   doRequestOpts. Ai.Complete / Ai.Extract wait up to 155 s (the routes run
//   up to 150 s; the client default is 10 s), are not re-sent after a
//   transport error or a 5xx (every attempt is a billed model call counted
//   against the daily cap — those were retried; a 429 RATE_LIMITED, refused
//   before any handler ran, still is) and never return the FailOpen
//   placeholder (it unmarshalled into an empty result with a nil error).
//   Vision.AnalyzeRoom waits 35 s and re-sends only with an idempotency key;
//   Knowledge upsert / search wait 65 s / 35 s.
//
// CHANGELOG (2026-10-26, v1.7.1 — caldera-keys-service#2 G16):
// - The http.Client follows a redirect only to the SAME scheme + host + port
//   as the original request, up to 3 hops; any other redirect is not
//   followed and the caller gets the 3xx as an *APIError. Why: Go's default
//   policy strips only Authorization / Cookie on a cross-host redirect, so
//   x-api-key was forwarded to whatever host a redirect named.
// - Review of !243: the origin is compared as lowercased scheme + hostname
//   (an IP literal by value, so two spellings of one IPv6 address match) +
//   effective port (80 / 443 filled in). Why: the raw Host string made
//   https://h and https://h:443 differ, so a same-origin redirect was refused.
//   Userinfo never takes part (it is not in URL.Host).
//
// CHANGELOG (2026-11-01, v1.7.2 pending):
// - doRequestTypedArray for endpoints whose envelope data is an array:
//   doRequestOpts unwraps only object data, so doRequestTyped could never
//   decode a list. Used by PermissionsService.ListOverrides. Federation push
//   (and pull without a target) also answer arrays but decode into a struct
//   today — a separate fix, since correcting it changes their signatures.
//
// CHANGELOG (2026-11-01, v1.7.2 pending — caldera-platformxe#33 review):
// - finalErrorCodes gains AI_MODEL_UNAVAILABLE and AI_NOT_CONFIGURED: both
//   are 503s that describe the deployment / gateway account (a key missing or
//   rejected, a model the account cannot use), so the transport backoff that
//   retries a 503 on idempotent calls (Ai.Profile, Ai.Usage) could never
//   succeed and only repeated the request.
// =============================================================================

package platformxe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client is the PlatformXe API client.
type Client struct {
	apiKey   string
	baseURL  string
	retries  int
	failOpen bool
	http     *http.Client

	Documents   *DocumentsService
	Domains     *DomainsService
	Events      *EventsService
	Exports     *ExportsService
	Fraud       *FraudService
	Identity    *IdentityService
	Messaging   *MessagingService
	Ocr         *OcrService
	Pdf         *PdfService
	Permissions *PermissionsService
	Qr          *QrService
	Storage     *StorageService
	Templates   *TemplatesService
	Threads     *ThreadsService
	Usage       *UsageService
	Webhooks    *WebhooksService
	Workflows   *WorkflowsService
	// v1.1.0
	Audit  *AuditService
	Issues *IssuesService
	Search *SearchService
	Whoami *WhoamiService
	// v1.7.0
	Ai        *AiService
	Vision    *VisionService
	Knowledge *KnowledgeService
}

// maxSameOriginRedirects caps how many same-origin redirects one request
// follows.
const maxSameOriginRedirects = 3

// sameOriginRedirectsOnly is the client's redirect policy. Go's default
// policy strips only Authorization and Cookie on a cross-host hop, so the
// x-api-key header would be forwarded to whatever host a redirect names. A
// redirect is followed only when its target has the same scheme, host and
// port as the original request (up to maxSameOriginRedirects hops); any other
// redirect is not followed and the caller gets the 3xx (an *APIError).
func sameOriginRedirectsOnly(req *http.Request, via []*http.Request) error {
	if !sameOrigin(req.URL, via[0].URL) {
		return http.ErrUseLastResponse
	}
	if len(via) > maxSameOriginRedirects {
		return http.ErrUseLastResponse
	}
	return nil
}

// sameOrigin reports whether a and b share scheme, host and effective port.
// The host comes from URL.Hostname (userinfo and IPv6 brackets excluded); an
// IP literal is compared by value; an absent port takes the scheme default.
func sameOrigin(a, b *url.URL) bool {
	schemeA, schemeB := strings.ToLower(a.Scheme), strings.ToLower(b.Scheme)
	if schemeA != schemeB {
		return false
	}
	if effectivePort(schemeA, a.Port()) != effectivePort(schemeB, b.Port()) {
		return false
	}
	hostA, hostB := a.Hostname(), b.Hostname()
	if ipA, ipB := net.ParseIP(hostA), net.ParseIP(hostB); ipA != nil || ipB != nil {
		return ipA != nil && ipB != nil && ipA.Equal(ipB)
	}
	return hostA != "" && strings.EqualFold(hostA, hostB)
}

func effectivePort(scheme, port string) string {
	if port != "" {
		return port
	}
	switch scheme {
	case "http":
		return "80"
	case "https":
		return "443"
	}
	return ""
}

// NewClient creates a new PlatformXe API client.
func NewClient(config ClientConfig) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "https://platformxe.com"
	}
	if config.Timeout == 0 {
		config.Timeout = 10
	}
	switch {
	case config.DisableRetries || config.Retries < 0:
		// Explicit "no retries". A negative Retries used to make the request
		// loop run zero attempts and return without ever sending.
		config.Retries = 0
	case config.Retries == 0:
		// Zero value = option omitted → default. Use DisableRetries for zero.
		config.Retries = 2
	}

	c := &Client{
		apiKey:   config.APIKey,
		baseURL:  config.BaseURL,
		retries:  config.Retries,
		failOpen: config.FailOpen,
		http: &http.Client{
			Timeout:       time.Duration(config.Timeout) * time.Second,
			CheckRedirect: sameOriginRedirectsOnly,
		},
	}

	c.Documents = &DocumentsService{client: c}
	c.Domains = &DomainsService{client: c}
	c.Events = &EventsService{client: c}
	c.Events.Custom = &CustomEventsService{client: c}
	c.Events.Custom.Marketplace = &MarketplaceService{client: c}
	c.Events.Custom.Federation = &EventFederationService{client: c}
	c.Exports = &ExportsService{client: c}
	c.Fraud = newFraudService(c)
	c.Identity = newIdentityService(c)
	c.Messaging = &MessagingService{client: c}
	c.Ocr = &OcrService{client: c}
	c.Pdf = &PdfService{client: c}
	c.Permissions = &PermissionsService{client: c}
	c.Qr = &QrService{client: c}
	c.Storage = &StorageService{client: c}
	c.Templates = &TemplatesService{client: c}
	c.Threads = &ThreadsService{client: c}
	c.Usage = &UsageService{client: c}
	c.Webhooks = &WebhooksService{client: c}
	c.Workflows = &WorkflowsService{client: c}
	// v1.1.0
	c.Audit = &AuditService{client: c}
	c.Issues = &IssuesService{client: c}
	c.Search = &SearchService{client: c}
	c.Whoami = &WhoamiService{client: c}
	// v1.7.0
	c.Ai = &AiService{client: c}
	c.Vision = &VisionService{client: c}
	c.Knowledge = &KnowledgeService{client: c}

	return c
}

// SendTelemetry sends client telemetry metrics as a full batch.
func (c *Client) SendTelemetry(batch TelemetryBatch) (*TelemetryResult, error) {
	var result TelemetryResult
	err := c.doRequestTyped("POST", "/api/v1/telemetry", batch, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// HealthCheck returns the platform health status.
func (c *Client) HealthCheck() (*HealthCheckResult, error) {
	var result HealthCheckResult
	err := c.doRequestTyped("GET", "/api/health", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Error implements error. The text matches the fmt.Errorf form the client
// returned before *APIError, so callers matching on it keep working.
func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s (HTTP %d)", e.Code, e.Message, e.StatusCode)
}

// isRetryableStatus reports whether a non-2xx status is retried on the
// transport backoff. 429 and 502/503/504 are retried for every method: the
// request was rate-limited or never reached a handler. 500 means a handler
// threw and its side effects may have landed, so it is retried only for an
// idempotent method. Every other status is final.
func isRetryableStatus(method string, status int) bool {
	switch status {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	case http.StatusInternalServerError:
		return isIdempotentMethod(method)
	default:
		return false
	}
}

// finalErrorCodes are error codes returned at once even when their status
// would be retried: a 429 AI_QUOTA_EXCEEDED is a daily cap, not a transient
// rate limit; a 503 AI_NOT_CONFIGURED / AI_MODEL_UNAVAILABLE is the
// deployment's or the gateway account's state, which a retry cannot change.
var finalErrorCodes = map[string]bool{
	ErrCodeAiQuotaExceeded:    true,
	ErrCodeAiNotConfigured:    true,
	ErrCodeAiModelUnavailable: true,
}

// isIdempotentMethod reports whether repeating a request with this method has
// the same effect as sending it once (RFC 9110 §9.2.2).
func isIdempotentMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
		return true
	default:
		return false
	}
}

// httpStatusError builds the *APIError for a non-2xx response. Code and
// message come from a JSON body when it has them — an {"error":{"code",
// "message"}} object (the envelope included), an {"error":"..."} string, or a
// top-level "message" — otherwise the code is HTTP_<status> and the message is
// the status text (a proxy's HTML error page, for example).
func httpStatusError(status int, body []byte) *APIError {
	apiErr := &APIError{Code: fmt.Sprintf("HTTP_%d", status), Message: http.StatusText(status), StatusCode: status}
	if apiErr.Message == "" {
		apiErr.Message = fmt.Sprintf("HTTP %d", status)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return apiErr
	}
	switch e := parsed["error"].(type) {
	case map[string]interface{}:
		if code, ok := e["code"].(string); ok && code != "" {
			apiErr.Code = code
		}
		if message, ok := e["message"].(string); ok && message != "" {
			apiErr.Message = message
		}
		if details, ok := e["details"].(map[string]interface{}); ok {
			apiErr.Details = details
		}
	case string:
		if e != "" {
			apiErr.Message = e
		}
	default:
		if message, ok := parsed["message"].(string); ok && message != "" {
			apiErr.Message = message
		}
	}
	return apiErr
}

// backoff sleeps before the next attempt (200 ms doubled per attempt), and not
// at all after the last one.
func (c *Client) backoff(attempt, retries int) {
	if attempt < retries {
		time.Sleep(time.Duration(200*(1<<attempt)) * time.Millisecond)
	}
}

func (c *Client) doRequest(method, path string, body interface{}, params map[string]string) (map[string]interface{}, error) {
	return c.doRequestWithHeaders(method, path, body, params, nil)
}

// retryAfterSeconds parses a Retry-After header given in seconds; 0 when it
// is absent or not a number.
func retryAfterSeconds(h http.Header) int {
	n, err := strconv.Atoi(strings.TrimSpace(h.Get("Retry-After")))
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// doRequestWithHeaders is doRequest with extra request headers (for example
// x-idempotency-key). The x-api-key, Accept and Content-Type headers are set
// after them and cannot be overridden.
func (c *Client) doRequestWithHeaders(method, path string, body interface{}, params map[string]string, headers map[string]string) (map[string]interface{}, error) {
	return c.doRequestOpts(method, path, body, params, headers, callOptions{})
}

// callOptions override the client's settings for one call. The zero value
// keeps them all.
type callOptions struct {
	// timeout replaces the client's http timeout for this call (0 = keep it).
	timeout time.Duration
	// rateLimitRetriesOnly: only a 429 (refused by the rate limiter before any
	// handler ran) is retried; a transport failure or a 5xx is NOT re-sent —
	// the handler may have run, and for a model call that was billed.
	rateLimitRetriesOnly bool
	// noFailOpen: a transport failure is an error even when FailOpen is on.
	noFailOpen bool
}

// modelCallOptions: Ai.Complete / Ai.Extract. The routes run up to 150 s, and
// a retried or fail-open call is a second billed model call (or a silent empty
// result).
func modelCallOptions() callOptions {
	return callOptions{timeout: 155 * time.Second, rateLimitRetriesOnly: true, noFailOpen: true}
}

// doRequestOpts is doRequestWithHeaders with per-call options.
func (c *Client) doRequestOpts(method, path string, body interface{}, params map[string]string, headers map[string]string, opts callOptions) (map[string]interface{}, error) {
	retries := c.retries
	httpClient := c.http
	if opts.timeout > 0 {
		perCall := *c.http
		perCall.Timeout = opts.timeout
		httpClient = &perCall
	}
	var lastErr error
	// lastErrIsHTTP: the last attempt got a retryable HTTP error status
	// (429/500/502/503/504) rather than a transport failure. FailOpen does not
	// cover it.
	lastErrIsHTTP := false

	attemptsMade := 0
	for attempt := 0; attempt <= retries; attempt++ {
		attemptsMade++
		var reqBody io.Reader
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("marshal body: %w", err)
			}
			reqBody = bytes.NewReader(b)
		}

		fullURL := c.baseURL + path
		if len(params) > 0 {
			q := url.Values{}
			for k, v := range params {
				q.Set(k, v)
			}
			fullURL += "?" + q.Encode()
		}

		req, err := http.NewRequest(method, fullURL, reqBody)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}
		req.Header.Set("x-api-key", c.apiKey)
		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr, lastErrIsHTTP = err, false
			if opts.rateLimitRetriesOnly {
				break
			}
			c.backoff(attempt, retries)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr, lastErrIsHTTP = err, false
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			apiErr := httpStatusError(resp.StatusCode, respBody)
			apiErr.RetryAfter = retryAfterSeconds(resp.Header)
			if !isRetryableStatus(method, resp.StatusCode) || finalErrorCodes[apiErr.Code] {
				return nil, apiErr
			}
			if opts.rateLimitRetriesOnly && resp.StatusCode != http.StatusTooManyRequests {
				return nil, apiErr
			}
			lastErr, lastErrIsHTTP = apiErr, true
			c.backoff(attempt, retries)
			continue
		}

		// 2xx — unchanged.
		var result map[string]interface{}
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("unmarshal response: %w", err)
		}

		if success, ok := result["success"].(bool); ok && !success {
			errMap, _ := result["error"].(map[string]interface{})
			code, _ := errMap["code"].(string)
			message, _ := errMap["message"].(string)
			return nil, &APIError{Code: code, Message: message, StatusCode: resp.StatusCode}
		}

		if data, ok := result["data"].(map[string]interface{}); ok {
			return data, nil
		}
		return result, nil
	}

	if lastErrIsHTTP {
		// A 5xx or 429 on every attempt is an API error, not an outage of the
		// transport: FailOpen does not apply.
		return nil, lastErr
	}
	if c.failOpen && !opts.noFailOpen {
		return map[string]interface{}{"_failed": true, "error": fmt.Sprintf("%v", lastErr)}, nil
	}
	return nil, fmt.Errorf("request failed after %d attempts: %w", attemptsMade, lastErr)
}

// doRequestTyped performs an HTTP request and unmarshals the response data into the typed result.
func (c *Client) doRequestTyped(method, path string, body interface{}, params map[string]string, result interface{}) error {
	return c.doRequestTypedWithHeaders(method, path, body, params, nil, result)
}

// doRequestTypedWithHeaders is doRequestTyped with extra request headers.
func (c *Client) doRequestTypedWithHeaders(method, path string, body interface{}, params map[string]string, headers map[string]string, result interface{}) error {
	return c.doRequestTypedOpts(method, path, body, params, headers, callOptions{}, result)
}

// doRequestTypedOpts is doRequestTypedWithHeaders with per-call options.
func (c *Client) doRequestTypedOpts(method, path string, body interface{}, params map[string]string, headers map[string]string, opts callOptions, result interface{}) error {
	raw, err := c.doRequestOpts(method, path, body, params, headers, opts)
	if err != nil {
		return err
	}
	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	return json.Unmarshal(jsonBytes, result)
}

// doRequestTypedArray decodes a response whose envelope `data` is a JSON
// array. doRequestOpts unwraps `data` only when it is an object and otherwise
// hands back the whole envelope, so doRequestTyped cannot decode a list.
func (c *Client) doRequestTypedArray(method, path string, body interface{}, params map[string]string, result interface{}) error {
	raw, err := c.doRequestOpts(method, path, body, params, nil, callOptions{})
	if err != nil {
		return err
	}
	data, ok := raw["data"].([]interface{})
	if !ok {
		return fmt.Errorf("%s %s: expected the response data to be an array", method, path)
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	return json.Unmarshal(jsonBytes, result)
}

// Get performs an HTTP GET request.
func (c *Client) Get(path string, params map[string]string) (map[string]interface{}, error) {
	return c.doRequest("GET", path, nil, params)
}

// Post performs an HTTP POST request.
func (c *Client) Post(path string, body interface{}) (map[string]interface{}, error) {
	return c.doRequest("POST", path, body, nil)
}

// Put performs an HTTP PUT request.
func (c *Client) Put(path string, body interface{}) (map[string]interface{}, error) {
	return c.doRequest("PUT", path, body, nil)
}

// Patch performs an HTTP PATCH request.
func (c *Client) Patch(path string, body interface{}) (map[string]interface{}, error) {
	return c.doRequest("PATCH", path, body, nil)
}

// Delete performs an HTTP DELETE request.
func (c *Client) Delete(path string) (map[string]interface{}, error) {
	return c.doRequest("DELETE", path, nil, nil)
}
