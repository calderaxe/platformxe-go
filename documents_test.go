// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-09-15): Initial. Pins the ACTUAL wire shape the folder and
// deletion-override methods send/receive, against a real net/http/httptest
// server — not just "the service namespace exists" like the rest of this
// package's tests. Written because CreateFolder/ListFolders previously sent
// a body/params the real endpoint always rejected (see documents.go's
// CHANGELOG), so a shallow "method exists" test would never have caught it.
//
// CHANGELOG (2026-09-15): Backward-compatibility cases for the v1 module
// surface — CreateFolder(map) still compiles and decodes folderId,
// CreateFolderForOwner sends ownerType/ownerId, and the deprecated
// ListFolders returns a clear error instead of being removed (removing an
// exported method breaks compilation for v1 consumers).
// =============================================================================

package platformxe

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestServer starts an httptest server whose handler decodes the request
// (method, path, query, JSON body) via check, then replies with the given
// envelope-wrapped data.
func newTestServer(t *testing.T, check func(r *http.Request), data interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			check(r)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    data,
		})
	}))
}

func decodeJSONBody(t *testing.T, r *http.Request) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode request body: %v", err)
	}
	return body
}

func folderCreateServer(t *testing.T) *httptest.Server {
	t.Helper()
	return newTestServer(t, func(r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/storage/fixed/folders" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body := decodeJSONBody(t, r)
		if body["ownerType"] != "AGENT" || body["ownerId"] != "agent_1" {
			t.Errorf("unexpected body: %#v", body)
		}
	}, map[string]interface{}{"folderId": "fld_1"})
}

// CreateFolderForOwner is the typed form.
func TestCreateFolderForOwnerSendsOwnerTypeAndOwnerID(t *testing.T) {
	srv := folderCreateServer(t)
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Documents.CreateFolderForOwner("AGENT", "agent_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FolderID != "fld_1" {
		t.Errorf("expected folderId fld_1, got %q", result.FolderID)
	}
}

// CreateFolder keeps its v1 map signature: a caller passing ownerType/ownerId
// always created the folder; only the result type was wrong.
func TestCreateFolderMapFormStillCompilesAndDecodesFolderID(t *testing.T) {
	srv := folderCreateServer(t)
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	var create func(map[string]interface{}) (*CreateFolderResult, error) = c.Documents.CreateFolder
	result, err := create(map[string]interface{}{"ownerType": "AGENT", "ownerId": "agent_1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FolderID != "fld_1" {
		t.Errorf("expected folderId fld_1, got %q", result.FolderID)
	}
}

// ListFolders stays as a deprecated method (removing it would break v1
// consumers' compilation) but never calls the API — there is no list route.
func TestListFoldersIsDeprecatedAndReturnsAClearError(t *testing.T) {
	calls := 0
	srv := newTestServer(t, func(r *http.Request) { calls++ }, map[string]interface{}{})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	var list func() (*FolderListResponse, error) = c.Documents.ListFolders
	result, err := list()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %#v", result)
	}
	want := "platformxe: ListFolders is not supported by the API — use GetFolderByOwner"
	if err.Error() != want {
		t.Errorf("unexpected error text: %q", err.Error())
	}
	if calls != 0 {
		t.Errorf("expected no HTTP call, got %d", calls)
	}
}

func TestGetFolderByOwnerSendsQueryParams(t *testing.T) {
	srv := newTestServer(t, func(r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		q := r.URL.Query()
		if q.Get("ownerType") != "AGENT" || q.Get("ownerId") != "agent_1" {
			t.Errorf("unexpected query: %v", q)
		}
	}, map[string]interface{}{
		"id": "fld_1", "ownerType": "AGENT", "ownerId": "agent_1",
		"storageLimitBytes": 100, "storageUsedBytes": 10, "maxFileCount": 50,
		"maxFileSizeBytes": 10, "documentCount": 1, "usagePercentage": 10,
		"remainingBytes": 90, "canUpload": true, "allowedCategories": []string{"CONTRACT"},
	})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	folder, err := c.Documents.GetFolderByOwner("AGENT", "agent_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if folder.ID != "fld_1" || !folder.CanUpload {
		t.Errorf("unexpected folder: %#v", folder)
	}
	if folder.DocumentCount != 1 {
		t.Errorf("expected documentCount 1, got %d", folder.DocumentCount)
	}
}

func TestRequestOverrideSendsDocumentIDAndOverrideReason(t *testing.T) {
	reason := make([]byte, 500)
	for i := range reason {
		reason[i] = 'x'
	}
	srv := newTestServer(t, func(r *http.Request) {
		body := decodeJSONBody(t, r)
		if body["documentId"] != "doc_1" {
			t.Errorf("unexpected documentId: %v", body["documentId"])
		}
		if body["overrideReason"] != string(reason) {
			t.Errorf("overrideReason did not round-trip")
		}
	}, map[string]interface{}{
		"id": "ovr_1", "documentId": "doc_1", "documentTitle": "x",
		"overrideReason": string(reason), "requestedById": "service:x", "status": "PENDING",
	})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	override, err := c.Documents.RequestOverride(map[string]interface{}{
		"documentId":     "doc_1",
		"overrideReason": string(reason),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if override.Status != "PENDING" {
		t.Errorf("expected status PENDING, got %q", override.Status)
	}
	if override.RequestedByID != "service:x" {
		t.Errorf("expected RequestedByID to be populated from requestedById, got %q", override.RequestedByID)
	}
}

func TestProcessOverrideApproveSendsAction(t *testing.T) {
	srv := newTestServer(t, func(r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/storage/fixed/overrides/ovr_1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body := decodeJSONBody(t, r)
		if body["action"] != "approve" {
			t.Errorf("expected action=approve, got %v", body["action"])
		}
	}, map[string]interface{}{"id": "ovr_1", "status": "APPROVED"})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	override, err := c.Documents.ProcessOverride("ovr_1", map[string]interface{}{"action": "approve"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if override.Status != "APPROVED" {
		t.Errorf("expected status APPROVED, got %q", override.Status)
	}
}

func TestExecuteOverrideReturnsDistinctShape(t *testing.T) {
	srv := newTestServer(t, func(r *http.Request) {
		body := decodeJSONBody(t, r)
		if body["action"] != "execute" {
			t.Errorf("expected action=execute, got %v", body["action"])
		}
		if _, hasOperator := body["operatorId"]; hasOperator {
			t.Errorf("operatorId should be omitted when empty, got %v", body["operatorId"])
		}
	}, map[string]interface{}{"success": true, "documentId": "doc_1"})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	result, err := c.Documents.ExecuteOverride("ovr_1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success || result.DocumentID != "doc_1" {
		t.Errorf("unexpected result: %#v", result)
	}
}

func TestExecuteOverrideIncludesOperatorIDWhenGiven(t *testing.T) {
	srv := newTestServer(t, func(r *http.Request) {
		body := decodeJSONBody(t, r)
		if body["operatorId"] != "steven" {
			t.Errorf("expected operatorId=steven, got %v", body["operatorId"])
		}
	}, map[string]interface{}{"success": true, "documentId": "doc_1"})
	defer srv.Close()

	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Documents.ExecuteOverride("ovr_1", "steven"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
