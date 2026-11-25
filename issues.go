// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-22, v1.15.0 pending, Enterprise License): Create
// documents 409 CONTACT_LICENSING (no request change); review round 1:
// Details["existing"].
// =============================================================================

package platformxe

// IssuesService — federated bug / support-ticket workflow.
type IssuesService struct {
	client *Client
}

// IssueRecord mirrors the database row returned by GET /api/v1/issues.
type IssueRecord struct {
	ID             string                 `json:"id"`
	OrganizationID string                 `json:"organizationId"`
	App            string                 `json:"app"`
	ReporterID     string                 `json:"reporterId"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Priority       string                 `json:"priority"`
	Status         string                 `json:"status"`
	Tags           []string               `json:"tags"`
	Metadata       map[string]interface{} `json:"metadata"`
	CreatedAt      string                 `json:"createdAt"`
	UpdatedAt      string                 `json:"updatedAt"`
}

type listIssuesResponse struct {
	Issues []IssueRecord `json:"issues"`
}

// ListIssuesParams filters the issues list.
type ListIssuesParams struct {
	App    string `json:"app,omitempty"`
	Status string `json:"status,omitempty"`
}

// CreateIssueRequest is the body for POST /api/v1/issues.
//
// An issue with App "PX" whose Tags include "plan-change" (both in any case)
// is a plan-change request: Metadata["requestedPlan"] must be "BASIC", "PRO"
// or "ENTERPRISE" (otherwise HTTP 400 VALIDATION_ERROR), and while one of the
// organisation's plan-change requests is still open another is refused with
// HTTP 409 CONFLICT and nothing is filed.
type CreateIssueRequest struct {
	App         string                 `json:"app"`
	ReporterID  string                 `json:"reporterId"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CreateIssueResponse is the 201 response.
type CreateIssueResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// List returns federated issues for the calling organisation.
func (s *IssuesService) List(params ListIssuesParams) ([]IssueRecord, error) {
	q := map[string]string{}
	if params.App != "" {
		q["app"] = params.App
	}
	if params.Status != "" {
		q["status"] = params.Status
	}
	var result listIssuesResponse
	err := s.client.doRequestTyped("GET", "/api/v1/issues", nil, q, &result)
	if err != nil {
		return nil, err
	}
	return result.Issues, nil
}

// Create files a new federated issue. A plan-change request (see
// CreateIssueRequest) needs Metadata["requestedPlan"] (400 VALIDATION_ERROR
// without one) and is refused with 409 CONFLICT while another is open. PRO
// and ENTERPRISE come only with a license: such a request is sent to the
// licensing team (an issue, never a plan change) and answered 409
// ErrCodeContactLicensing with Details {"email", "id", "existing"}; while the
// organisation has one open nothing is filed, Details["existing"] is true and
// Details["id"] names the open request (v1.15.0). Each comes back as the client's API error.
func (s *IssuesService) Create(req CreateIssueRequest) (*CreateIssueResponse, error) {
	var result CreateIssueResponse
	err := s.client.doRequestTyped("POST", "/api/v1/issues", req, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
