// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
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

// Create files a new federated issue.
func (s *IssuesService) Create(req CreateIssueRequest) (*CreateIssueResponse, error) {
	var result CreateIssueResponse
	err := s.client.doRequestTyped("POST", "/api/v1/issues", req, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
