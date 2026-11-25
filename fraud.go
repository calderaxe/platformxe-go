// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-09-15): gofmt only — the FraudService sub-namespace list
// is now a tab-indented doc-comment code block, as gofmt (Go 1.19+) requires.
// CHANGELOG (2026-11-07, v1.9.0 pending): Activate turns the Detection Pack
// on for the key's own organisation (POST /api/v1/fraud/activate).
// CHANGELOG (2026-11-18, v1.12.0 pending): Rules.Publish and Rules.Archive
// post to /api/v1/fraud/rules/:id/transition {target}; they posted to
// /rules/:id/publish and /rules/:id/archive, which do not exist (every call was
// a 404). Rules.Transition, Rules.Draft, Rules.Shadow, Rules.ShadowReport, the
// FraudRuleStatus* constants and FraudRuleTransitions added. Cases.Transition
// documents 409 CASE_AT_MLRO_STAGE (ErrCodeCaseAtMlroStage).
// CHANGELOG (2026-11-18, v1.12.0 pending, review round 1): FraudRuleTransitions
// carries a DUPLICATION marker naming the server table it mirrors;
// src/lib/api/sdk-fraud-rule-transitions.test.ts fails CI when they differ.
// CHANGELOG (2026-11-19, v1.13.0 pending, X12): the rule paths are constants
// (fraudRuleByIDPath, ...; the id is path-escaped); Rules.Transition and
// TransitionAtVersion document 409 FEED_MISSING (ErrCodeFeedMissing) and
// FEED_NOT_REALTIME (ErrCodeFeedNotRealtime); Cases.Transition documents that
// a resolve records the outcome at once.
// CHANGELOG (2026-11-21, pending, spec 03 PM-1): Decide and ShadowDecide
// document 400 CARD_NUMBER_REFUSED (ErrCodeCardNumberRefused). Fix:
// DecideRequest marshals to the nested body the routes read (subject and
// resource objects); the flat body it sent was refused (every call a 400).
// CHANGELOG (2026-11-19, v1.13.0 pending): a rule's `version` is its revision
// (it goes up on every PATCH that changes something). FraudRuleVersion reads
// it from a rule. Rules.TransitionAtVersion and Rules.UpdateAtVersion send
// expectedVersion: a rule edited since is a 409 VERSION_MISMATCH
// (ErrCodeVersionMismatch) and nothing changes. Rules.ShadowReportForVersion
// counts one revision's decisions (the report counts the current one by
// default, and returns ruleVersion and currentVersion).
// =============================================================================

package platformxe

import (
	"encoding/json"
	"fmt"
)

const (
	fraudRulesPath            = "/api/v1/fraud/rules"
	fraudRuleByIDPath         = "/api/v1/fraud/rules/{id}"
	fraudRuleTransitionPath   = "/api/v1/fraud/rules/{id}/transition"
	fraudRuleShadowReportPath = "/api/v1/fraud/rules/{id}/shadow-report"
)

// FraudService handles all Fraud Detection Engine endpoints (Phase 6A–6H).
//
// Sub-namespaces:
//
//	client.Fraud.Cases       — cases workflow (Phase 6E)
//	client.Fraud.Rules       — rule CRUD (Phase 6B)
//	client.Fraud.Lists       — tenant blocklists / allowlists (Phase 6C)
//	client.Fraud.Devices     — device fingerprint registry (Phase 6D)
//	client.Fraud.Terms       — T&Cs status / accept (Phase 6H)
//	client.Fraud.Federation  — federation push / preview (Phase 6G)
//
// Note: Idempotency-Key headers are not surfaced in v1; the server-side
// withIdempotency middleware degrades cleanly when no key is provided.
type FraudService struct {
	client *Client

	Cases      *FraudCasesService
	Rules      *FraudRulesService
	Lists      *FraudListsService
	Devices    *FraudDevicesService
	Terms      *FraudTermsService
	Federation *FraudFederationService
}

func newFraudService(c *Client) *FraudService {
	s := &FraudService{client: c}
	s.Cases = &FraudCasesService{client: c}
	s.Rules = &FraudRulesService{client: c}
	s.Lists = &FraudListsService{client: c}
	s.Devices = &FraudDevicesService{client: c}
	s.Terms = &FraudTermsService{client: c}
	s.Federation = &FraudFederationService{client: c}
	return s
}

// ----------------------------------------------------------------------------
// /decide and /shadow-decide (Phase 6A)
// ----------------------------------------------------------------------------

// DecideRequest is the body shape for POST /api/v1/fraud/decide.
type DecideRequest struct {
	SubjectID    string                 `json:"subjectId"`
	SubjectKind  string                 `json:"subjectKind"`
	Action       string                 `json:"action"`
	ResourceKind string                 `json:"resourceKind"`
	ResourceID   string                 `json:"resourceId,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

// MarshalJSON sends the body the decide routes read:
// {subject: {id, kind}, action, resource: {kind, id?}, context?}. The flat
// subjectId / subjectKind / resourceKind fields the struct tags name were never
// read by the API (every Decide was a 400 BAD_REQUEST before 2026-11-21).
func (r DecideRequest) MarshalJSON() ([]byte, error) {
	resource := map[string]interface{}{"kind": r.ResourceKind}
	if r.ResourceID != "" {
		resource["id"] = r.ResourceID
	}
	body := map[string]interface{}{
		"subject":  map[string]interface{}{"id": r.SubjectID, "kind": r.SubjectKind},
		"action":   r.Action,
		"resource": resource,
	}
	if len(r.Context) > 0 {
		body["context"] = r.Context
	}
	return json.Marshal(body)
}

// Activate turns the Detection Pack addon on for the key's own organisation
// (POST /api/v1/fraud/activate, scope fraud:manage). Idempotent: a pack
// already on answers alreadyActive true and bills nothing again. An empty
// idempotencyKey sends no header. Accept the terms next with Terms.Accept.
// "billing" is FIRST_PARTY_METERED, PLAN_INCLUDED, INVOICED, DEMO_SEED,
// LICENSED (under an Enterprise License: nothing invoiced here) or null.
// After the licensing cut-over, with no live license grant naming a fraud
// pack, the API answers 403 ErrCodeLicenseRequired (v1.15.0).
func (s *FraudService) Activate(idempotencyKey string) (map[string]interface{}, error) {
	var headers map[string]string
	if idempotencyKey != "" {
		headers = map[string]string{"x-idempotency-key": idempotencyKey}
	}
	return s.client.doRequestWithHeaders("POST", "/api/v1/fraud/activate", map[string]interface{}{}, nil, headers)
}

// Decide renders a production verdict. A request carrying a card number or a
// card field (pan, cvv ...) anywhere in Subject, Action, Resource or Context
// (only Context's amount value skips the number test) is refused with a 400 *APIError coded
// ErrCodeCardNumberRefused before anything is decided or recorded: send a
// token, never card data.
func (s *FraudService) Decide(req DecideRequest) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/fraud/decide", req, nil)
}

// ShadowDecide renders a non-enforcing verdict for rule validation. It refuses
// card data as Decide does (ErrCodeCardNumberRefused).
func (s *FraudService) ShadowDecide(req DecideRequest) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/fraud/shadow-decide", req, nil)
}

// ----------------------------------------------------------------------------
// /screen (Phase 6C)
// ----------------------------------------------------------------------------

type ScreenRequest struct {
	Name        string            `json:"name"`
	DOB         string            `json:"dob,omitempty"`
	Country     string            `json:"country,omitempty"`
	Identifiers map[string]string `json:"identifiers,omitempty"`
	Kinds       []string          `json:"kinds,omitempty"`
	Threshold   float64           `json:"threshold,omitempty"`
	DecisionID  string            `json:"decisionId,omitempty"`
}

// Screen runs a sanctions / PEP / blocklist screen.
func (s *FraudService) Screen(req ScreenRequest) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/fraud/screen", req, nil)
}

// ----------------------------------------------------------------------------
// /decisions (Phase 6A)
// ----------------------------------------------------------------------------

type ListDecisionsParams struct {
	SubjectID string
	Verdict   string
	Limit     int
	Offset    int
}

// ListDecisions returns the audit trail.
func (s *FraudService) ListDecisions(p ListDecisionsParams) (map[string]interface{}, error) {
	q := map[string]string{}
	if p.SubjectID != "" {
		q["subjectId"] = p.SubjectID
	}
	if p.Verdict != "" {
		q["verdict"] = p.Verdict
	}
	if p.Limit > 0 {
		q["limit"] = fmt.Sprintf("%d", p.Limit)
	}
	if p.Offset > 0 {
		q["offset"] = fmt.Sprintf("%d", p.Offset)
	}
	return s.client.doRequest("GET", "/api/v1/fraud/decisions", nil, q)
}

// GetDecision fetches a single decision by id.
func (s *FraudService) GetDecision(id string) (map[string]interface{}, error) {
	return s.client.doRequest("GET", "/api/v1/fraud/decisions/"+id, nil, nil)
}

// ----------------------------------------------------------------------------
// FraudCasesService — Phase 6E
// ----------------------------------------------------------------------------

type FraudCasesService struct{ client *Client }

type ListCasesParams struct {
	Status      string
	DecisionID  string
	OverdueOnly bool
	Limit       int
	Offset      int
}

func (s *FraudCasesService) List(p ListCasesParams) (map[string]interface{}, error) {
	q := map[string]string{}
	if p.Status != "" {
		q["status"] = p.Status
	}
	if p.DecisionID != "" {
		q["decisionId"] = p.DecisionID
	}
	if p.OverdueOnly {
		q["overdueOnly"] = "true"
	}
	if p.Limit > 0 {
		q["limit"] = fmt.Sprintf("%d", p.Limit)
	}
	if p.Offset > 0 {
		q["offset"] = fmt.Sprintf("%d", p.Offset)
	}
	return s.client.doRequest("GET", "/api/v1/fraud/cases", nil, q)
}

func (s *FraudCasesService) Get(id string) (map[string]interface{}, error) {
	return s.client.doRequest("GET", "/api/v1/fraud/cases/"+id, nil, nil)
}

func (s *FraudCasesService) Open(payload map[string]interface{}) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/fraud/cases", payload, nil)
}

func (s *FraudCasesService) Update(id string, payload map[string]interface{}) (map[string]interface{}, error) {
	return s.client.doRequest("PATCH", "/api/v1/fraud/cases/"+id, payload, nil)
}

// Transition moves a case (POST /api/v1/fraud/cases/:id/transition; payload
// target, optional resolution and note). A case at an MLRO stage (mlro,
// sar_filed, le_referred) is refused for every transition with a 409
// *APIError coded ErrCodeCaseAtMlroStage: that decision is the MLRO desk's in
// the portal, and an API key holds no desk role. A move not allowed from the
// current status, or one that lost a race, is 409 CONFLICT. A move to
// resolved records the case's outcome at once (as a portal close does), named
// api-key:<key id>; a false_positive close also withdraws the case's open
// recovery actions. The response is unchanged.
func (s *FraudCasesService) Transition(id string, payload map[string]interface{}) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/fraud/cases/"+id+"/transition", payload, nil)
}

// ----------------------------------------------------------------------------
// FraudRulesService — Phase 6B
// ----------------------------------------------------------------------------

// FraudRulesService manages tenant fraud rules: CRUD, status transitions and
// the shadow report.
type FraudRulesService struct{ client *Client }

// The fraud rule statuses.
const (
	FraudRuleStatusDraft     = "draft"
	FraudRuleStatusShadow    = "shadow"
	FraudRuleStatusPublished = "published"
	FraudRuleStatusArchived  = "archived"
)

// FraudRuleTransitions are the moves POST /rules/:id/transition allows from
// each status: draft to shadow or archived; shadow to draft, published or
// archived; published to archived; archived is terminal.
//
// DUPLICATION: mirrors TRANSITIONS in src/lib/services/fraud/rules/api.ts
// (caldera-platformxe), so callers and the Terraform provider can check a move
// before sending it. src/lib/api/sdk-fraud-rule-transitions.test.ts parses
// both and fails CI when they differ.
var FraudRuleTransitions = map[string][]string{
	FraudRuleStatusDraft:     {FraudRuleStatusShadow, FraudRuleStatusArchived},
	FraudRuleStatusShadow:    {FraudRuleStatusDraft, FraudRuleStatusPublished, FraudRuleStatusArchived},
	FraudRuleStatusPublished: {FraudRuleStatusArchived},
	FraudRuleStatusArchived:  {},
}

func (s *FraudRulesService) List(status string) (map[string]interface{}, error) {
	q := map[string]string{}
	if status != "" {
		q["status"] = status
	}
	return s.client.doRequest("GET", fraudRulesPath, nil, q)
}

func (s *FraudRulesService) Get(id string) (map[string]interface{}, error) {
	return s.client.doRequest("GET", fillPath(fraudRuleByIDPath, id), nil, nil)
}

func (s *FraudRulesService) Create(payload map[string]interface{}) (map[string]interface{}, error) {
	return s.client.doRequest("POST", fraudRulesPath, payload, nil)
}

func (s *FraudRulesService) Update(id string, payload map[string]interface{}) (map[string]interface{}, error) {
	return s.client.doRequest("PATCH", fillPath(fraudRuleByIDPath, id), payload, nil)
}

// UpdateAtVersion is Update guarded by the version the caller last read
// (FraudRuleVersion of the rule it read): when the rule was edited since,
// the API answers 409 VERSION_MISMATCH (ErrCodeVersionMismatch) and writes
// nothing, so a stale edit never overwrites a newer one. payload is not
// modified.
func (s *FraudRulesService) UpdateAtVersion(id string, payload map[string]interface{}, expectedVersion int64) (map[string]interface{}, error) {
	body := make(map[string]interface{}, len(payload)+1)
	for k, v := range payload {
		body[k] = v
	}
	body["expectedVersion"] = expectedVersion
	return s.client.doRequest("PATCH", fillPath(fraudRuleByIDPath, id), body, nil)
}

// FraudRuleVersion is a rule's version (its revision) from a rule the API
// returned, and whether the rule carried one.
func FraudRuleVersion(rule map[string]interface{}) (int64, bool) {
	switch v := rule["version"].(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	}
	return 0, false
}

// Transition moves a rule to target (POST /api/v1/fraud/rules/:id/transition
// with {"target": target}, scope fraud:manage) and returns the rule. target is
// one of the FraudRuleStatus* constants; FraudRuleTransitions lists the moves
// allowed from each status. A move not allowed from the rule's status is a
// 409 CONFLICT *APIError, an unknown target 400, a rule not found 404.
func (s *FraudRulesService) Transition(id, target string) (map[string]interface{}, error) {
	return s.client.doRequest("POST", fillPath(fraudRuleTransitionPath, id), map[string]interface{}{"target": target}, nil)
}

// Publishing a rule that reads a feed the organisation has not sent recently
// is 409 FEED_MISSING (ErrCodeFeedMissing); one that needs a real-time
// payment feed where the pack's paymentFeed setting is not realtime is 409
// FEED_NOT_REALTIME (ErrCodeFeedNotRealtime). Both also apply to
// TransitionAtVersion.
//
// TransitionAtVersion is Transition guarded by the revision the caller
// reviewed (expectedVersion, the rule's `version`): when the rule was edited
// since, the API answers 409 VERSION_MISMATCH (ErrCodeVersionMismatch) and
// nothing changes, so a publish never carries an edit nobody reviewed.
func (s *FraudRulesService) TransitionAtVersion(id, target string, expectedVersion int64) (map[string]interface{}, error) {
	return s.client.doRequest("POST", fillPath(fraudRuleTransitionPath, id),
		map[string]interface{}{"target": target, "expectedVersion": expectedVersion}, nil)
}

// Draft moves a shadow rule back to draft (Transition(id, "draft")).
func (s *FraudRulesService) Draft(id string) (map[string]interface{}, error) {
	return s.Transition(id, FraudRuleStatusDraft)
}

// Shadow moves a draft rule to shadow: evaluated, never enforced.
func (s *FraudRulesService) Shadow(id string) (map[string]interface{}, error) {
	return s.Transition(id, FraudRuleStatusShadow)
}

// Publish publishes a shadow rule (Transition(id, "published")). A draft rule
// must go to shadow first. Before v1.12.0 it posted to /rules/:id/publish,
// which does not exist (every call was a 404).
func (s *FraudRulesService) Publish(id string) (map[string]interface{}, error) {
	return s.Transition(id, FraudRuleStatusPublished)
}

// Archive archives a rule from any status but archived (terminal).
// Before v1.12.0 it posted to /rules/:id/archive, which does not exist.
func (s *FraudRulesService) Archive(id string) (map[string]interface{}, error) {
	return s.Transition(id, FraudRuleStatusArchived)
}

// ShadowReport returns the rule's trigger-rate metrics over a window
// (GET /api/v1/fraud/rules/:id/shadow-report, scope fraud:manage), to check a
// shadow rule before publishing it. from and to are ISO-8601 timestamps; an
// empty one is not sent (the window defaults to the last 7 days). A timestamp
// that does not parse is 400.
func (s *FraudRulesService) ShadowReport(id, from, to string) (map[string]interface{}, error) {
	return s.ShadowReportForVersion(id, from, to, 0)
}

// ShadowReportForVersion is ShadowReport for one revision's decisions: version
// (sent as ?version=) when it is above 0, else the rule's current one. The
// report's ruleVersion is the revision counted (the one to publish with
// TransitionAtVersion); currentVersion is the rule's version now.
func (s *FraudRulesService) ShadowReportForVersion(id, from, to string, version int64) (map[string]interface{}, error) {
	q := map[string]string{}
	if from != "" {
		q["from"] = from
	}
	if to != "" {
		q["to"] = to
	}
	if version > 0 {
		q["version"] = fmt.Sprintf("%d", version)
	}
	return s.client.doRequest("GET", fillPath(fraudRuleShadowReportPath, id), nil, q)
}

// ----------------------------------------------------------------------------
// FraudListsService — Phase 6C
// ----------------------------------------------------------------------------

type FraudListsService struct{ client *Client }

func (s *FraudListsService) List() (map[string]interface{}, error) {
	return s.client.doRequest("GET", "/api/v1/fraud/lists", nil, nil)
}

func (s *FraudListsService) Get(id string) (map[string]interface{}, error) {
	return s.client.doRequest("GET", "/api/v1/fraud/lists/"+id, nil, nil)
}

func (s *FraudListsService) Create(source, name, kind string) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/fraud/lists", map[string]interface{}{
		"source": source,
		"name":   name,
		"kind":   kind,
	}, nil)
}

func (s *FraudListsService) Update(id string, payload map[string]interface{}) (map[string]interface{}, error) {
	return s.client.doRequest("PATCH", "/api/v1/fraud/lists/"+id, payload, nil)
}

func (s *FraudListsService) Delete(id string) (map[string]interface{}, error) {
	return s.client.doRequest("DELETE", "/api/v1/fraud/lists/"+id, nil, nil)
}

func (s *FraudListsService) ListEntries(id string) (map[string]interface{}, error) {
	return s.client.doRequest("GET", "/api/v1/fraud/lists/"+id+"/entries", nil, nil)
}

func (s *FraudListsService) AddEntry(id string, payload map[string]interface{}) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/fraud/lists/"+id+"/entries", payload, nil)
}

func (s *FraudListsService) RemoveEntry(listID, entryID string) (map[string]interface{}, error) {
	return s.client.doRequest("DELETE", "/api/v1/fraud/lists/"+listID+"/entries/"+entryID, nil, nil)
}

// ----------------------------------------------------------------------------
// FraudDevicesService — Phase 6D
// ----------------------------------------------------------------------------

type FraudDevicesService struct{ client *Client }

func (s *FraudDevicesService) Seen(payload map[string]interface{}) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/fraud/devices/seen", payload, nil)
}

// ----------------------------------------------------------------------------
// FraudTermsService — Phase 6H
// ----------------------------------------------------------------------------

type FraudTermsService struct{ client *Client }

func (s *FraudTermsService) Status() (map[string]interface{}, error) {
	return s.client.doRequest("GET", "/api/v1/fraud/terms/status", nil, nil)
}

func (s *FraudTermsService) Accept(acceptedBy, version string) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/fraud/terms/accept", map[string]interface{}{
		"acceptedBy": acceptedBy,
		"version":    version,
	}, nil)
}

// ----------------------------------------------------------------------------
// FraudFederationService — Phase 6G (Enterprise)
// ----------------------------------------------------------------------------

type FraudFederationService struct{ client *Client }

func (s *FraudFederationService) Preview(groupID string) (map[string]interface{}, error) {
	return s.client.doRequest("GET", "/api/v1/permissions/federation/groups/"+groupID+"/fraud/preview", nil, nil)
}

func (s *FraudFederationService) Push(groupID string) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/permissions/federation/groups/"+groupID+"/fraud/push", map[string]interface{}{}, nil)
}
