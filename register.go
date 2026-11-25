// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// RegisterRequest is the body for POST /api/v1/register.
type RegisterRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// RegisteredTenant is the compact tenant summary returned on success.
type RegisteredTenant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Plan string `json:"plan"` // always "FREE" for self-registered tenants
}

// RegisterResponse is the 201 response from /api/v1/register.
//
// Save APIKey immediately — the platform does not retain a way to
// retrieve it.
type RegisterResponse struct {
	Message string           `json:"message"`
	Tenant  RegisteredTenant `json:"tenant"`
	APIKey  string           `json:"apiKey"`
	Note    string           `json:"note"`
}

// RegisterOptions overrides the platform base URL and timeout.
type RegisterOptions struct {
	BaseURL string        // default: "https://platformxe.com"
	Timeout time.Duration // default: 10s
}

// Register self-registers a developer tenant and returns an API key.
//
// This is the unauthenticated bootstrap endpoint for new developers — it
// can't sit on a Client because constructing a Client requires an API key.
//
// Usage:
//
//	r, err := platformxe.Register(platformxe.RegisterRequest{
//	    Name:  "Acme",
//	    Email: "dev@acme.test",
//	}, nil)
//	if err != nil { log.Fatal(err) }
//	client := platformxe.NewClient(platformxe.ClientConfig{APIKey: r.APIKey})
func Register(req RegisterRequest, opts *RegisterOptions) (*RegisterResponse, error) {
	baseURL := "https://platformxe.com"
	timeout := 10 * time.Second
	if opts != nil {
		if opts.BaseURL != "" {
			baseURL = opts.BaseURL
		}
		if opts.Timeout > 0 {
			timeout = opts.Timeout
		}
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal register request: %w", err)
	}

	httpClient := &http.Client{Timeout: timeout}
	httpReq, err := http.NewRequest("POST", baseURL+"/api/v1/register", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build register request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute register request: %w", err)
	}
	defer resp.Body.Close()

	var envelope struct {
		Success bool             `json:"success"`
		Data    RegisterResponse `json:"data"`
		Error   *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decode register response: %w", err)
	}

	if !envelope.Success {
		if envelope.Error != nil {
			return nil, fmt.Errorf("register failed [%s]: %s", envelope.Error.Code, envelope.Error.Message)
		}
		return nil, fmt.Errorf("register failed (status %d)", resp.StatusCode)
	}

	return &envelope.Data, nil
}
