// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

import "fmt"

// TemplatesService handles email/message template management.
type TemplatesService struct {
	client *Client
}

// List returns all templates.
func (s *TemplatesService) List() (*TemplateListResponse, error) {
	var result TemplateListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/templates", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new template.
func (s *TemplatesService) Create(input map[string]interface{}) (*Template, error) {
	var result Template
	err := s.client.doRequestTyped("POST", "/api/v1/templates", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns a template by ID.
func (s *TemplatesService) Get(id string) (*Template, error) {
	var result Template
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/templates/%s", id), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates a template.
func (s *TemplatesService) Update(id string, input map[string]interface{}) (*Template, error) {
	var result Template
	err := s.client.doRequestTyped("PATCH", fmt.Sprintf("/api/v1/templates/%s", id), input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a template.
func (s *TemplatesService) Delete(id string) (*DeleteResult, error) {
	var result DeleteResult
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/templates/%s", id), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Render renders a template with the given variables without sending.
func (s *TemplatesService) Render(id string, variables map[string]interface{}) (*RenderResult, error) {
	var result RenderResult
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/templates/%s/render", id), map[string]interface{}{"variables": variables}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Send renders and sends a template.
func (s *TemplatesService) Send(id string, input map[string]interface{}) (*TemplateSendResult, error) {
	var result TemplateSendResult
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/templates/%s/send", id), input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
