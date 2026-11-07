// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-10-08, v1.7.0 pending — calderasuite/caldera-xadmin#29):
// - Created. Contract tests for KnowledgeService and tool calling on
//   AiService.Complete against testdata/ai_vision.json (the shared fixture
//   set, byte-identical to packages/sdk-python/tests/fixtures/). Each test
//   runs a real httptest server and pins method, path, query and the exact
//   JSON body against the fixture request, and decodes the fixture response
//   into the typed result.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"net/http"
	"os"
	"reflect"
	"testing"
)

type knowledgeToolFixtures struct {
	CompleteWithToolsRequest json.RawMessage `json:"completeWithToolsRequest"`
	CompleteWithToolsResult  json.RawMessage `json:"completeWithToolsResult"`
	KnowledgeUpsertRequest   json.RawMessage `json:"knowledgeUpsertRequest"`
	KnowledgeUpsertResult    json.RawMessage `json:"knowledgeUpsertResult"`
	KnowledgeDocumentList    json.RawMessage `json:"knowledgeDocumentList"`
	KnowledgeDeleteResult    json.RawMessage `json:"knowledgeDeleteResult"`
	KnowledgeSearchRequest   json.RawMessage `json:"knowledgeSearchRequest"`
	KnowledgeSearchResult    json.RawMessage `json:"knowledgeSearchResult"`
}

func loadKnowledgeToolFixtures(t *testing.T) knowledgeToolFixtures {
	t.Helper()
	raw, err := os.ReadFile("testdata/ai_vision.json")
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	var f knowledgeToolFixtures
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("parse fixtures: %v", err)
	}
	return f
}

func TestAiCompleteWithToolsMatchesFixture(t *testing.T) {
	f := loadKnowledgeToolFixtures(t)
	srv := fixtureServer(t, f.CompleteWithToolsResult, nil, func(r *http.Request, body []byte) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/ai/complete" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		assertSameJSON(t, "complete-with-tools body", body, f.CompleteWithToolsRequest)
	})
	var input AiCompleteInput
	decodeInto(t, f.CompleteWithToolsRequest, &input)

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Ai.Complete(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FinishReason != "tool_calls" || len(result.ToolCalls) != 1 || result.ToolCalls[0].Name != "hr.leave_balance" {
		t.Errorf("tool calls not decoded: %#v", result)
	}
	if result.ToolCalls[0].Arguments != `{"employeeRef":"emp_43"}` {
		t.Errorf("arguments not kept verbatim: %q", result.ToolCalls[0].Arguments)
	}
}

func TestAiCompleteForcedToolChoiceIsAnObject(t *testing.T) {
	srv := fixtureServer(t, json.RawMessage(`{"content":"","provider":"gemini","model":"m","finishReason":"tool_calls"}`), nil, func(_ *http.Request, body []byte) {
		var m map[string]interface{}
		_ = json.Unmarshal(body, &m)
		choice, ok := m["toolChoice"].(map[string]interface{})
		if !ok || choice["name"] != "a.b" {
			t.Errorf("toolChoice should be {name}, got %v", m["toolChoice"])
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Ai.Complete(AiCompleteInput{
		FeatureKey: "ai-canvas", Provider: "gemini",
		Messages:   []AiMessage{{Role: "user", Content: "hi"}},
		Tools:      []AiToolDefinition{{Name: "a.b"}},
		ToolChoice: AiToolChoiceName{Name: "a.b"},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAiCompleteWithoutToolsSendsNoToolFields(t *testing.T) {
	srv := fixtureServer(t, json.RawMessage(`{"content":"hi","provider":"gemini","model":"m"}`), nil, func(_ *http.Request, body []byte) {
		var m map[string]interface{}
		_ = json.Unmarshal(body, &m)
		for _, k := range []string{"tools", "toolChoice"} {
			if _, ok := m[k]; ok {
				t.Errorf("%s should be omitted", k)
			}
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Ai.Complete(AiCompleteInput{FeatureKey: "ai-canvas", Provider: "gemini", Messages: []AiMessage{{Role: "user", Content: "hi"}}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKnowledgeUpsertMatchesFixture(t *testing.T) {
	f := loadKnowledgeToolFixtures(t)
	srv := fixtureServer(t, f.KnowledgeUpsertResult, nil, func(r *http.Request, body []byte) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/ai/knowledge/documents" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		assertSameJSON(t, "upsert body", body, f.KnowledgeUpsertRequest)
	})
	var input KnowledgeUpsertDocumentInput
	decodeInto(t, f.KnowledgeUpsertRequest, &input)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Knowledge.UpsertDocument(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var want KnowledgeUpsertDocumentResult
	decodeInto(t, f.KnowledgeUpsertResult, &want)
	if !reflect.DeepEqual(*result, want) {
		t.Errorf("result mismatch: %#v", result)
	}
}

func TestKnowledgeDeleteSendsQuery(t *testing.T) {
	f := loadKnowledgeToolFixtures(t)
	srv := fixtureServer(t, f.KnowledgeDeleteResult, nil, func(r *http.Request, _ []byte) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/ai/knowledge/documents" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("collection") != "hr-policies" || q.Get("sourceRef") != "policy/annual-leave" {
			t.Errorf("unexpected query %v", q)
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Knowledge.DeleteDocument("hr-policies", "policy/annual-leave")
	if err != nil || !result.Deleted {
		t.Fatalf("unexpected result %#v / %v", result, err)
	}
}

func TestKnowledgeListSendsQueryAndDecodes(t *testing.T) {
	f := loadKnowledgeToolFixtures(t)
	srv := fixtureServer(t, f.KnowledgeDocumentList, nil, func(r *http.Request, _ []byte) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/ai/knowledge/documents" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("collection") != "hr-policies" || q.Get("cursor") != "abc" || q.Get("limit") != "50" {
			t.Errorf("unexpected query %v", q)
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Knowledge.ListDocuments(KnowledgeListQuery{Collection: "hr-policies", Cursor: "abc", Limit: 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 || result.NextCursor == nil || result.Items[0].ChunkCount != 1 {
		t.Errorf("list not decoded: %#v", result)
	}
}

func TestKnowledgeListOmitsEmptyCursorAndLimit(t *testing.T) {
	srv := fixtureServer(t, json.RawMessage(`{"items":[],"nextCursor":null}`), nil, func(r *http.Request, _ []byte) {
		q := r.URL.Query()
		if q.Has("cursor") || q.Has("limit") {
			t.Errorf("cursor / limit should be omitted: %v", q)
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Knowledge.ListDocuments(KnowledgeListQuery{Collection: "hr"})
	if err != nil || result.NextCursor != nil {
		t.Fatalf("unexpected %#v / %v", result, err)
	}
}

func TestKnowledgeSearchMatchesFixture(t *testing.T) {
	f := loadKnowledgeToolFixtures(t)
	srv := fixtureServer(t, f.KnowledgeSearchResult, nil, func(r *http.Request, body []byte) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/ai/knowledge/search" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		assertSameJSON(t, "search body", body, f.KnowledgeSearchRequest)
	})
	var input KnowledgeSearchInput
	decodeInto(t, f.KnowledgeSearchRequest, &input)
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Knowledge.Search(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Mode != "hybrid" || len(result.Hits) != 1 || result.Hits[0].VectorRank == nil || *result.Hits[0].VectorRank != 1 {
		t.Errorf("search not decoded: %#v", result)
	}
}

func TestKnowledgeSearchAlwaysSendsAclTags(t *testing.T) {
	srv := fixtureServer(t, json.RawMessage(`{"mode":"lexical","hits":[]}`), nil, func(_ *http.Request, body []byte) {
		var m map[string]interface{}
		_ = json.Unmarshal(body, &m)
		tags, ok := m["aclTags"].([]interface{})
		if !ok || len(tags) != 0 {
			t.Errorf("aclTags should be sent as [], got %v", m["aclTags"])
		}
		if _, ok := m["k"]; ok {
			t.Errorf("k should be omitted when 0")
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Knowledge.Search(KnowledgeSearchInput{FeatureKey: "ai-canvas", Collections: []string{"hr"}, Query: "leave"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
