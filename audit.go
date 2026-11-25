// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

// AuditService dispatches federated audit events into the trace pipeline.
//
// Events are accepted asynchronously (HTTP 202). The trace pipeline durably
// records them via the inngest worker.
type AuditService struct {
	client *Client
}

// AuditLogRequest is the body for POST /api/v1/audit/log.
type AuditLogRequest struct {
	App        string                 `json:"app"`
	ActorID    string                 `json:"actorId"`
	EntityType string                 `json:"entityType"`
	EntityID   string                 `json:"entityId"`
	Action     string                 `json:"action"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	OccurredAt string                 `json:"occurredAt,omitempty"`
}

// AuditLogAcceptedResponse is the 202 response.
type AuditLogAcceptedResponse struct {
	Message string `json:"message"`
}

// Log dispatches an audit event into the federated trace pipeline.
func (s *AuditService) Log(req AuditLogRequest) (*AuditLogAcceptedResponse, error) {
	var result AuditLogAcceptedResponse
	err := s.client.doRequestTyped("POST", "/api/v1/audit/log", req, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
