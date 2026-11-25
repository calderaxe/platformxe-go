// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================

package platformxe

// CHANGELOG (2026-11-19, v1.13.0 pending, X12):
// - Replay sends fromTimestamp / toTimestamp, the fields the route reads (it
//   sent from / to: every replay was a 400 VALIDATION_ERROR). ReplayTypes
//   also sends eventTypes. EventReplayResult.Replayed is the count the
//   route answers.
// - The subscription paths are constants (eventSubscriptionByIDPath, ...;
//   the id is path-escaped).
// - Docs: 400 RESTRICTED_EVENT_TYPE (ErrCodeRestrictedEventType) on create
//   and update, RestrictedEventTypes on every subscription read, the event
//   log leaves out fraud and compliance events.

// CHANGELOG (2026-11-20, v1.13.0 pending, fidelity W2):
// - Ingest sends a deprecated "payload" key as entityData (the route reads
//   entityData only, so the event data never reached a trigger) and never
//   sends "payload". EmitEventResult gains EntityType and EntityID, which the
//   route answers.
// - Review round 1 (L1): a nil payload or entityData (an untyped nil, or a
//   map, slice or pointer holding nil) is never sent: encoding/json would
//   write null, and the route refuses entityData: null with 400.

// CHANGELOG (2026-11-22, v1.14.1 pending): docs only. eventType is a value on
// the Workflow Event Types page; entityData is your own parameters for the
// event (an empty map is sent as {}, a nil one is not sent). The route now
// accepts entityData: null as none too; the SDK still omits it.

import "reflect"

const (
	eventsIngestPath            = "/api/v1/events"
	eventLogPath                = "/api/v1/event-log"
	eventSubscriptionsPath      = "/api/v1/event-subscriptions"
	eventSubscriptionByIDPath   = "/api/v1/event-subscriptions/{id}"
	eventSubscriptionReplayPath = "/api/v1/event-subscriptions/{id}/replay"
)

// EventsService handles event ingestion, subscriptions, and log queries.
type EventsService struct {
	client *Client
	// Custom is the typed client for tenant-defined custom events.
	// Available at Client.Events.Custom.
	Custom *CustomEventsService
}

// Ingest sends an event to the workflow engine: POST /api/v1/events with
// eventType (one of the values on
// https://docs.platformxe.com/services/workflows/event-types), and optionally
// entityType, entityId, entityData (your own parameters for the event: what
// the triggers' conditions and templates read; for a business event type
// PlatformXe defines no fields; an empty map is sent as {}, a nil one is not
// sent), sourceApp (the key's caller service when absent) and metadata. The
// key's organisation and caller service decide which triggers run. Any other
// event type, or an unknown entity type, is 400 VALIDATION_ERROR (the
// message never lists the values).
//
// A "payload" key is deprecated: the route does not read it, so before
// v1.13.0 the event data never reached a trigger. It is now sent as
// entityData when the map has no entityData (as the TypeScript SDK's ingest
// does), and dropped otherwise.
func (s *EventsService) Ingest(input map[string]interface{}) (*EmitEventResult, error) {
	var result EmitEventResult
	err := s.client.doRequestTyped("POST", eventsIngestPath, ingestBody(input), nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ingestBody maps Ingest's deprecated "payload" key to entityData (when the
// input has no non-nil entityData) and never sends "payload", nor a nil
// entityData (servers before 2026-11-22 refused entityData: null). The
// caller's map is not modified.
func ingestBody(input map[string]interface{}) map[string]interface{} {
	payload, hasPayload := input["payload"]
	entityData, hasEntityData := input["entityData"]
	if !hasPayload && (!hasEntityData || !isNilValue(entityData)) {
		return input
	}
	body := make(map[string]interface{}, len(input))
	for k, v := range input {
		if k != "payload" && k != "entityData" {
			body[k] = v
		}
	}
	switch {
	case hasEntityData && !isNilValue(entityData):
		body["entityData"] = entityData
	case hasPayload && !isNilValue(payload):
		body["entityData"] = payload
	}
	return body
}

// isNilValue is true for an untyped nil and for a map, slice, pointer,
// interface, func or channel holding nil (a typed nil is not == nil inside an
// interface{}, and encoding/json writes it as null).
func isNilValue(v interface{}) bool {
	if v == nil {
		return true
	}
	switch rv := reflect.ValueOf(v); rv.Kind() {
	case reflect.Map, reflect.Slice, reflect.Ptr, reflect.Interface, reflect.Func, reflect.Chan:
		return rv.IsNil()
	}
	return false
}

// Log returns the event log with optional filters (eventTypes comma
// separated, from, to, page, limit). Fraud and compliance events are left
// out, except the operational ones on the deliverable list, which carry
// only their listed payload fields.
func (s *EventsService) Log(params map[string]string) (*EventLogResponse, error) {
	var result EventLogResponse
	err := s.client.doRequestTyped("GET", eventLogPath, nil, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListSubscriptions returns all event subscriptions.
func (s *EventsService) ListSubscriptions() (*EventSubscriptionListResponse, error) {
	var result EventSubscriptionListResponse
	err := s.client.doRequestTyped("GET", eventSubscriptionsPath, nil, nil, &result)
	if err != nil {
		return nil, err
	}
	result.Subscriptions = result.Items
	return &result, nil
}

// CreateSubscription creates a new event subscription. input needs name
// (3+ characters), eventTypes and webhookUrl; a new subscription is ACTIVE.
// The API answers { subscription, secret }: the returned subscription carries
// the signing secret in Secret, which is never returned again. An entry
// naming a fraud or compliance event that is never delivered is 400
// RESTRICTED_EVENT_TYPE (ErrCodeRestrictedEventType).
func (s *EventsService) CreateSubscription(input map[string]interface{}) (*EventSubscription, error) {
	var result struct {
		Subscription EventSubscription `json:"subscription"`
		Secret       string            `json:"secret"`
	}
	err := s.client.doRequestTyped("POST", eventSubscriptionsPath, input, nil, &result)
	if err != nil {
		return nil, err
	}
	result.Subscription.Secret = result.Secret
	return &result.Subscription, nil
}

// GetSubscription returns an event subscription by ID, with
// RestrictedEventTypes (its stored entries that are never delivered).
func (s *EventsService) GetSubscription(id string) (*EventSubscription, error) {
	var result EventSubscription
	err := s.client.doRequestTyped("GET", fillPath(eventSubscriptionByIDPath, id), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateSubscription updates an event subscription with create's checks: a
// name that is not a string or is under 3 characters, a malformed
// eventTypes, an unknown status or an unsafe webhookUrl is 400
// VALIDATION_ERROR; a restricted entry is 400 RESTRICTED_EVENT_TYPE.
// status ACTIVE also resets the failure count.
func (s *EventsService) UpdateSubscription(id string, input map[string]interface{}) (*EventSubscription, error) {
	var result EventSubscription
	err := s.client.doRequestTyped("PATCH", fillPath(eventSubscriptionByIDPath, id), input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteSubscription deletes an event subscription.
func (s *EventsService) DeleteSubscription(id string) (*DeleteResult, error) {
	var result DeleteResult
	err := s.client.doRequestTyped("DELETE", fillPath(eventSubscriptionByIDPath, id), nil, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Replay replays events for a subscription within a time range (to may be
// empty: up to now). Restricted types are never replayed (not counted). 409
// CONFLICT while another replay of the subscription runs.
func (s *EventsService) Replay(id, from, to string) (*EventReplayResult, error) {
	return s.ReplayTypes(id, from, to, nil)
}

// ReplayTypes is Replay limited to eventTypes (nil: every type). A non-nil
// eventTypes must be non-empty with no blank entry (400 VALIDATION_ERROR).
func (s *EventsService) ReplayTypes(id, from, to string, eventTypes []string) (*EventReplayResult, error) {
	body := map[string]interface{}{"fromTimestamp": from}
	if to != "" {
		body["toTimestamp"] = to
	}
	if eventTypes != nil {
		body["eventTypes"] = eventTypes
	}
	var result EventReplayResult
	err := s.client.doRequestTyped("POST", fillPath(eventSubscriptionReplayPath, id), body, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
