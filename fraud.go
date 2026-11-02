// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-09-15): gofmt only — the FraudService sub-namespace list
// is now a tab-indented doc-comment code block, as gofmt (Go 1.19+) requires.
// =============================================================================

package platformxe

import "fmt"

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

// Decide renders a synchronous fraud verdict.
func (s *FraudService) Decide(req DecideRequest) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/fraud/decide", req, nil)
}

// ShadowDecide renders a non-enforcing verdict for rule validation.
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

func (s *FraudCasesService) Transition(id string, payload map[string]interface{}) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/fraud/cases/"+id+"/transition", payload, nil)
}

// ----------------------------------------------------------------------------
// FraudRulesService — Phase 6B
// ----------------------------------------------------------------------------

type FraudRulesService struct{ client *Client }

func (s *FraudRulesService) List(status string) (map[string]interface{}, error) {
	q := map[string]string{}
	if status != "" {
		q["status"] = status
	}
	return s.client.doRequest("GET", "/api/v1/fraud/rules", nil, q)
}

func (s *FraudRulesService) Get(id string) (map[string]interface{}, error) {
	return s.client.doRequest("GET", "/api/v1/fraud/rules/"+id, nil, nil)
}

func (s *FraudRulesService) Create(payload map[string]interface{}) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/fraud/rules", payload, nil)
}

func (s *FraudRulesService) Update(id string, payload map[string]interface{}) (map[string]interface{}, error) {
	return s.client.doRequest("PATCH", "/api/v1/fraud/rules/"+id, payload, nil)
}

func (s *FraudRulesService) Publish(id string) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/fraud/rules/"+id+"/publish", map[string]interface{}{}, nil)
}

func (s *FraudRulesService) Archive(id string) (map[string]interface{}, error) {
	return s.client.doRequest("POST", "/api/v1/fraud/rules/"+id+"/archive", map[string]interface{}{}, nil)
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
