// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

import "fmt"

// WorkflowsService handles workflow automation management.
type WorkflowsService struct {
	client *Client
}

// List returns all workflows.
func (s *WorkflowsService) List() (*WorkflowListResponse, error) {
	var result WorkflowListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/workflows", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new workflow.
func (s *WorkflowsService) Create(input map[string]interface{}) (*WorkflowTrigger, error) {
	var result WorkflowTrigger
	err := s.client.doRequestTyped("POST", "/api/v1/workflows", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a workflow by ID.
func (s *WorkflowsService) Get(id string) (*WorkflowTrigger, error) {
	var result WorkflowTrigger
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/workflows/%s", id), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates a workflow.
func (s *WorkflowsService) Update(id string, input map[string]interface{}) (*WorkflowTrigger, error) {
	var result WorkflowTrigger
	err := s.client.doRequestTyped("PATCH", fmt.Sprintf("/api/v1/workflows/%s", id), input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a workflow.
func (s *WorkflowsService) Delete(id string) (*DeleteResult, error) {
	var result DeleteResult
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/workflows/%s", id), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Evaluate evaluates workflows against an event to determine matches.
func (s *WorkflowsService) Evaluate(eventType string, payload map[string]interface{}) (*WorkflowEvalResult, error) {
	var result WorkflowEvalResult
	err := s.client.doRequestTyped("POST", "/api/v1/workflows/evaluate", map[string]interface{}{
		"eventType": eventType, "payload": payload,
	}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DryRun performs a dry-run of a trigger against a context (no side effects).
func (s *WorkflowsService) DryRun(triggerID string, context map[string]interface{}) (*DryRunResult, error) {
	var result DryRunResult
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/workflows/%s/dry-run", triggerID), context, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
