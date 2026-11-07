// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

// SearchService — federated search index + query.
type SearchService struct {
	client *Client
}

// SearchIndexRequest is the body for POST /api/v1/search/index.
type SearchIndexRequest struct {
	App        string                 `json:"app"`
	EntityType string                 `json:"entityType"`
	EntityID   string                 `json:"entityId"`
	Title      string                 `json:"title"`
	Subtitle   string                 `json:"subtitle,omitempty"`
	Keywords   string                 `json:"keywords,omitempty"`
	Rank       int                    `json:"rank,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// SearchIndexAcceptedResponse is the 202 response.
type SearchIndexAcceptedResponse struct {
	Message string `json:"message"`
}

// SearchResult is a single index row returned by GET /api/v1/search/query.
type SearchResult struct {
	ID             string                 `json:"id"`
	OrganizationID string                 `json:"organizationId"`
	App            string                 `json:"app"`
	EntityType     string                 `json:"entityType"`
	EntityID       string                 `json:"entityId"`
	Title          string                 `json:"title"`
	Subtitle       string                 `json:"subtitle"`
	Keywords       string                 `json:"keywords"`
	Rank           *int                   `json:"rank"`
	Metadata       map[string]interface{} `json:"metadata"`
	IndexedAt      string                 `json:"indexedAt"`
}

type searchQueryResponse struct {
	Results []SearchResult `json:"results"`
}

// Index upserts a resource into the federated search index.
//
// Accepted asynchronously (HTTP 202) — the upsert is performed by the
// inngest worker.
func (s *SearchService) Index(req SearchIndexRequest) (*SearchIndexAcceptedResponse, error) {
	var result SearchIndexAcceptedResponse
	err := s.client.doRequestTyped("POST", "/api/v1/search/index", req, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Query performs an ILIKE search across title + keywords.
//
// Returns at most 20 results. Queries shorter than 2 characters return
// an empty slice.
func (s *SearchService) Query(q string, app string) ([]SearchResult, error) {
	params := map[string]string{"q": q}
	if app != "" {
		params["app"] = app
	}
	var result searchQueryResponse
	err := s.client.doRequestTyped("GET", "/api/v1/search/query", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return result.Results, nil
}
