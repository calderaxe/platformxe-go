// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-12, pending: Task 4 fix round 4): Evaluate sends its map as
// entityData (it sent "payload", which the route ignores, so no trigger saw the
// event data); EvaluateWithOptions takes every request field.
// CHANGELOG (2026-11-20, v1.13.0 pending, fidelity register): DryRun posted to
// /api/v1/workflows/{id}/dry-run, which does not exist (every call was a 404).
// It now posts {entityData} to /api/v1/workflows/{id}, the trigger's dry-run
// route, and returns the route's dryRun result (DryRunResult gained the
// fields the route answers). DryRunEntityData sends entity data directly.
// CHANGELOG (2026-11-20, v1.13.0 pending, fix round): List, Create, Get,
// Update and Delete decode the shapes the routes answer: List {triggers,
// pagination} (it read "workflows": always empty), the others the trigger
// under "trigger" (they decoded the top level: an empty trigger, and Delete
// said Deleted false on success). Delete answers Deleted true when the trigger
// was switched off (the route disables it). The trigger id is path-escaped and
// an empty id is refused before any request. DryRun reads the context as
// before only for the old {eventType, payload} shape or a lone entityData,
// else sends the whole map (it unwrapped any payload key, dropping the rest);
// DryRunResult.TriggerID, Matched and Actions are filled.
// CHANGELOG (2026-11-20, v1.13.0 pending, re-review R2): a context is
// unwrapped when every key is a trigger-context key (triggerContextKeys) and
// it carries an entityData (else payload) object, as the Python SDK; the
// {eventType, entityType, entityId, entityData} shape was nested before.
// CHANGELOG (2026-11-20, v1.13.0 pending, re-review R5): List takes an optional
// *WorkflowListOptions (page, limit, isActive, triggerType, eventType: the GET
// route's query); List() still lists the first page. It could not reach page 2.
// CHANGELOG (2026-11-21, v1.13.0 pending, review N1, L8): List with more than
// one option is an error (only the first was read); an id of "." or ".." is
// refused by the client before any request (ErrDotSegmentPath).
// CHANGELOG (2026-11-22, v1.14.1 pending): docs only. Evaluate and
// EvaluateWithOptions: eventType is a value on the Workflow Event Types page;
// entityData is your own parameters for the event (the custom-parameters
// object), optional: a nil or empty map is not sent (omitempty), which the
// route reads as none, as it reads {} or null.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// workflowByIDPath is a trigger's route; fillPath path-escapes the id.
const workflowByIDPath = "/api/v1/workflows/{id}"

// ErrWorkflowIDRequired is returned before any request when a trigger id is empty.
var ErrWorkflowIDRequired = errors.New("platformxe: workflow trigger id is required")

// workflowTriggerEnvelope is the {trigger} body of the single-trigger routes.
type workflowTriggerEnvelope struct {
	Trigger WorkflowTrigger `json:"trigger"`
}

// WorkflowsService handles workflow automation management.
type WorkflowsService struct {
	client *Client
}

// WorkflowListOptions pages and filters List (the GET /api/v1/workflows
// query). A zero value is left out: the route then answers page 1, 50 per
// page, every trigger. The route answers 400 VALIDATION_ERROR for an unknown
// TriggerType or EventType, and caps Limit at 100.
type WorkflowListOptions struct {
	// Page is the page number, from 1.
	Page int
	// Limit is the page size, 1 to 100 (a larger value is 100).
	Limit int
	// IsActive lists only active (true) or only switched-off (false) triggers.
	IsActive *bool
	// TriggerType is EVENT or SCHEDULE.
	TriggerType string
	// EventType is a workflow event type (for example BOOKING_CREATED).
	EventType string
}

// params is the route's query for the options; nil when none is set.
func (o *WorkflowListOptions) params() map[string]string {
	if o == nil {
		return nil
	}
	params := map[string]string{}
	if o.Page > 0 {
		params["page"] = strconv.Itoa(o.Page)
	}
	if o.Limit > 0 {
		params["limit"] = strconv.Itoa(o.Limit)
	}
	if o.IsActive != nil {
		params["isActive"] = strconv.FormatBool(*o.IsActive)
	}
	if o.TriggerType != "" {
		params["triggerType"] = o.TriggerType
	}
	if o.EventType != "" {
		params["eventType"] = o.EventType
	}
	if len(params) == 0 {
		return nil
	}
	return params
}

// List returns a page of the caller's triggers (GET /api/v1/workflows,
// {triggers, pagination}). opts is optional: List() lists the first page of
// every trigger; List(opts) takes one *WorkflowListOptions (nil is none).
// More than one is an error, before any request: options are never merged
// or silently dropped. Pagination.Total is the number of triggers that
// match, on every page.
func (s *WorkflowsService) List(opts ...*WorkflowListOptions) (*WorkflowListResponse, error) {
	if len(opts) > 1 {
		return nil, fmt.Errorf("platformxe: Workflows.List takes at most one *WorkflowListOptions, got %d", len(opts))
	}
	var params map[string]string
	if len(opts) == 1 {
		params = opts[0].params()
	}
	var result WorkflowListResponse
	err := s.client.doRequestTyped("GET", "/api/v1/workflows", nil, params, &result)
	if err != nil {
		return nil, err
	}
	result.Workflows = result.Triggers
	return &result, nil
}

// Create creates a trigger (POST /api/v1/workflows, answered as {trigger}).
func (s *WorkflowsService) Create(input map[string]interface{}) (*WorkflowTrigger, error) {
	return s.triggerRequest("POST", "/api/v1/workflows", input)
}

// Get returns a trigger by id (GET /api/v1/workflows/{id}, {trigger, recentExecutions}).
func (s *WorkflowsService) Get(id string) (*WorkflowTrigger, error) {
	if id == "" {
		return nil, ErrWorkflowIDRequired
	}
	return s.triggerRequest("GET", fillPath(workflowByIDPath, id), nil)
}

// Update updates a trigger (PATCH /api/v1/workflows/{id}, answered as {trigger}).
func (s *WorkflowsService) Update(id string, input map[string]interface{}) (*WorkflowTrigger, error) {
	if id == "" {
		return nil, ErrWorkflowIDRequired
	}
	return s.triggerRequest("PATCH", fillPath(workflowByIDPath, id), input)
}

// Delete switches a trigger off (DELETE /api/v1/workflows/{id} disables it and
// answers {trigger}). Deleted is true when the route answered the trigger.
func (s *WorkflowsService) Delete(id string) (*DeleteResult, error) {
	if id == "" {
		return nil, ErrWorkflowIDRequired
	}
	trigger, err := s.triggerRequest("DELETE", fillPath(workflowByIDPath, id), nil)
	if err != nil {
		return nil, err
	}
	return &DeleteResult{Deleted: trigger.ID != ""}, nil
}

// triggerRequest sends a single-trigger request and decodes its {trigger} body.
func (s *WorkflowsService) triggerRequest(method, path string, body interface{}) (*WorkflowTrigger, error) {
	var envelope workflowTriggerEnvelope
	if err := s.client.doRequestTyped(method, path, body, nil, &envelope); err != nil {
		return nil, err
	}
	return &envelope.Trigger, nil
}

// Evaluate evaluates workflows against an event. eventType is one of the values on
// https://docs.platformxe.com/services/workflows/event-types (any other is 400
// VALIDATION_ERROR). payload is your own parameters for the event and is sent as
// entityData, what the triggers' conditions and templates read; for a business event
// type PlatformXe defines no fields. A nil or empty map is not sent (none). The name is
// kept for compatibility; it was sent as "payload", which the route never read.
// EvaluateWithOptions sends the route's other fields.
func (s *WorkflowsService) Evaluate(eventType string, payload map[string]interface{}) (*WorkflowEvalResult, error) {
	return s.EvaluateWithOptions(WorkflowEvaluateInput{EventType: eventType, EntityData: payload})
}

// EvaluateWithOptions evaluates workflows against an event with every request field
// (entity type and id, source app, metadata, dry run). EntityData is your own
// parameters for the event; a nil or empty map is not sent (the route reads none).
func (s *WorkflowsService) EvaluateWithOptions(input WorkflowEvaluateInput) (*WorkflowEvalResult, error) {
	var result WorkflowEvalResult
	err := s.client.doRequestTyped("POST", "/api/v1/workflows/evaluate", input, nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// DryRun dry-runs one trigger (no actions run): POST /api/v1/workflows/{id}
// with {entityData}. A context whose every key is a trigger-context key
// (triggerContextKeys) and that carries an entityData (else payload) object
// sends that object; any other map is the entity data, sent whole. The route
// does not read eventType. Prefer DryRunEntityData.
func (s *WorkflowsService) DryRun(triggerID string, context map[string]interface{}) (*DryRunResult, error) {
	return s.DryRunEntityData(triggerID, dryRunEntityData(context))
}

// DryRunEntityData dry-runs one trigger against entityData (no actions run).
func (s *WorkflowsService) DryRunEntityData(triggerID string, entityData map[string]interface{}) (*DryRunResult, error) {
	if triggerID == "" {
		return nil, ErrWorkflowIDRequired
	}
	if entityData == nil {
		entityData = map[string]interface{}{}
	}
	var envelope struct {
		DryRun DryRunResult `json:"dryRun"`
	}
	err := s.client.doRequestTyped("POST", fillPath(workflowByIDPath, triggerID), map[string]interface{}{"entityData": entityData}, nil, &envelope)
	if err != nil {
		return nil, err
	}
	result := envelope.DryRun
	// The deprecated fields, for code written to the old result.
	result.TriggerID = triggerID
	result.Matched = result.WouldMatch
	result.Actions = result.ActionsWouldExecute
	return &result, nil
}

// triggerContextKeys are the keys of a trigger context: the route's
// TriggerContext fields plus payload (the old name of entityData), sourceApp
// and dryRun. The Python SDK's TRIGGER_CONTEXT_KEYS is the same set.
var triggerContextKeys = map[string]bool{
	"eventType": true, "entityType": true, "entityId": true, "entityData": true, "metadata": true,
	"organizationId": true, "callerService": true, "payload": true, "sourceApp": true, "dryRun": true,
}

// dryRunEntityData reads the entity data from DryRun's context map, as the
// Python SDK's dry_run_request_body does.
func dryRunEntityData(context map[string]interface{}) map[string]interface{} {
	if context == nil {
		return map[string]interface{}{}
	}
	for key := range context {
		if !triggerContextKeys[key] {
			return context
		}
	}
	for _, key := range []string{"entityData", "payload"} {
		if d, ok := asJSONObject(context[key]); ok {
			return d
		}
	}
	return context
}

// asJSONObject returns v as a JSON object map: a map[string]interface{} as it
// is, any other value only when it encodes to a JSON object (a struct, a
// map[string]string). Python's isinstance(v, dict) is the same test.
func asJSONObject(v interface{}) (map[string]interface{}, bool) {
	if v == nil {
		return nil, false
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m, true
	}
	b, err := json.Marshal(v)
	if err != nil || len(b) == 0 || b[0] != '{' {
		return nil, false
	}
	var m map[string]interface{}
	if json.Unmarshal(b, &m) != nil {
		return nil, false
	}
	return m, true
}
