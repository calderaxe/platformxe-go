// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

import "fmt"

// EventsService handles event ingestion, subscriptions, and log queries.
type EventsService struct {
	client *Client
	// Custom is the typed client for tenant-defined custom events.
	// Available at Client.Events.Custom.
	Custom *CustomEventsService
}

// Ingest emits a custom event.
func (s *EventsService) Ingest(input map[string]interface{}) (*EmitEventResult, error) {
	var result EmitEventResult
	err := s.client.doRequestTyped("POST", "/api/v1/events", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Log returns the event log with optional filters.
func (s *EventsService) Log(params map[string]string) (*EventLogResponse, error) {
	var result EventLogResponse
	err := s.client.doRequestTyped("GET", "/api/v1/event-log", nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListSubscriptions returns all event subscriptions.
func (s *EventsService) ListSubscriptions() (*EventSubscriptionListResponse, error) {
	var result EventSubscriptionListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/event-subscriptions", nil, nil, &result)
	if err != nil {
		return nil, err
	}
	result.Subscriptions = result.Items
	return &result, nil
}

// CreateSubscription creates a new event subscription. input needs name
// (3+ characters), eventTypes and webhookUrl; a new subscription is ACTIVE.
// The API answers { subscription, secret }: the returned subscription carries
// the signing secret in Secret, which is never returned again.
func (s *EventsService) CreateSubscription(input map[string]interface{}) (*EventSubscription, error) {
	var result struct {
		Subscription EventSubscription `json:"subscription"`
		Secret       string            `json:"secret"`
	}
	err := s.client.doRequestTyped("POST", "/api/v1/event-subscriptions", input, nil, &result)
	if err != nil {
		return nil, err
	}
	result.Subscription.Secret = result.Secret
	return &result.Subscription, nil
}

// GetSubscription returns an event subscription by ID.
func (s *EventsService) GetSubscription(id string) (*EventSubscription, error) {
	var result EventSubscription
	err := s.client.doRequestTyped("GET", fmt.Sprintf("/api/v1/event-subscriptions/%s", id), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateSubscription updates an event subscription.
func (s *EventsService) UpdateSubscription(id string, input map[string]interface{}) (*EventSubscription, error) {
	var result EventSubscription
	err := s.client.doRequestTyped("PATCH", fmt.Sprintf("/api/v1/event-subscriptions/%s", id), input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteSubscription deletes an event subscription.
func (s *EventsService) DeleteSubscription(id string) (*DeleteResult, error) {
	var result DeleteResult
	err := s.client.doRequestTyped("DELETE", fmt.Sprintf("/api/v1/event-subscriptions/%s", id), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Replay replays events for a subscription within a time range.
func (s *EventsService) Replay(id, from, to string) (*EventReplayResult, error) {
	var result EventReplayResult
	err := s.client.doRequestTyped("POST", fmt.Sprintf("/api/v1/event-subscriptions/%s/replay", id), map[string]interface{}{
		"from": from, "to": to,
	}, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
