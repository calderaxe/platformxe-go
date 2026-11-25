// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-18, v1.12.0 pending):
// - Created. The category registry and settings methods call
//   /api/v1/storage/fixed/categories and /settings.
// =============================================================================

package platformxe

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

const categoryJSON = `{"id":"cat_1","callerService":"lettings","key":"LEASE","label":"Lease","description":null,
"sortOrder":0,"color":null,"isActive":true,"retentionType":"LEGAL","retentionDays":2555,"deletableBy":null,
"requiresOverride":false,"requiresWitness":true,"createdAt":"2026-11-18T00:00:00.000Z","updatedAt":"2026-11-18T00:00:00.000Z"}`

func TestListCategories(t *testing.T) {
	var gotQuery string
	srv := fixtureServer(t, json.RawMessage(`{"categories":[`+categoryJSON+`]}`), nil, func(r *http.Request, _ []byte) {
		if r.Method != "GET" || r.URL.Path != "/api/v1/storage/fixed/categories" {
			t.Errorf("sent %s %s", r.Method, r.URL.Path)
		}
		gotQuery = r.URL.Query().Get("includeInactive")
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	res, err := c.Documents.ListCategories(true)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	if gotQuery != "true" {
		t.Errorf("includeInactive = %q", gotQuery)
	}
	if len(res.Categories) != 1 {
		t.Fatalf("categories = %v", res.Categories)
	}
	cat := res.Categories[0]
	if cat.Key != "LEASE" || cat.RetentionType == nil || *cat.RetentionType != RetentionTypeLegal ||
		cat.RetentionDays == nil || *cat.RetentionDays != 2555 || cat.DeletableBy != nil || !cat.RequiresWitness {
		t.Errorf("category = %+v", cat)
	}
}

func TestCreateCategoryOmitsNilFields(t *testing.T) {
	var gotBody map[string]interface{}
	srv := fixtureServer(t, json.RawMessage(`{"category":`+categoryJSON+`}`), nil, func(r *http.Request, body []byte) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/storage/fixed/categories" {
			t.Errorf("sent %s %s", r.Method, r.URL.Path)
		}
		_ = json.Unmarshal(body, &gotBody)
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	rt, days, witness := RetentionTypeLegal, 2555, true
	cat, err := c.Documents.CreateCategory(CreateDocumentCategoryInput{
		Key: "lease", Label: "Lease", RetentionType: &rt, RetentionDays: &days, RequiresWitness: &witness,
	})
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	want := map[string]interface{}{
		"key": "lease", "label": "Lease", "retentionType": "LEGAL", "retentionDays": float64(2555), "requiresWitness": true,
	}
	if !reflect.DeepEqual(gotBody, want) {
		t.Errorf("body = %v", gotBody)
	}
	if cat.ID != "cat_1" {
		t.Errorf("category = %+v", cat)
	}
}

func TestUpdateCategorySendsNullToClear(t *testing.T) {
	var gotBody map[string]interface{}
	var gotKey string
	srv := fixtureServer(t, json.RawMessage(`{"category":`+categoryJSON+`}`), nil, func(r *http.Request, body []byte) {
		if r.Method != "PATCH" || r.URL.Path != "/api/v1/storage/fixed/categories" {
			t.Errorf("sent %s %s", r.Method, r.URL.Path)
		}
		gotKey = r.URL.Query().Get("key")
		_ = json.Unmarshal(body, &gotBody)
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Documents.UpdateCategory("LEASE", map[string]interface{}{"label": "Tenancy", "retentionDays": nil}); err != nil {
		t.Fatalf("UpdateCategory: %v", err)
	}
	if gotKey != "LEASE" {
		t.Errorf("key = %q", gotKey)
	}
	if v, ok := gotBody["retentionDays"]; !ok || v != nil || gotBody["label"] != "Tenancy" {
		t.Errorf("body = %v", gotBody)
	}
}

func TestDeactivateCategory(t *testing.T) {
	srv := fixtureServer(t, json.RawMessage(`{"category":`+categoryJSON+`}`), nil, func(r *http.Request, _ []byte) {
		if r.Method != "DELETE" || r.URL.Path != "/api/v1/storage/fixed/categories" || r.URL.Query().Get("key") != "LEASE" {
			t.Errorf("sent %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	if _, err := c.Documents.DeactivateCategory("LEASE"); err != nil {
		t.Fatalf("DeactivateCategory: %v", err)
	}
}

func TestDocumentSettings(t *testing.T) {
	var gotMethod string
	var gotBody map[string]interface{}
	srv := fixtureServer(t, json.RawMessage(`{"settings":{"enforceCategories":true}}`), nil, func(r *http.Request, body []byte) {
		if r.URL.Path != "/api/v1/storage/fixed/settings" {
			t.Errorf("path = %s", r.URL.Path)
		}
		gotMethod = r.Method
		gotBody = nil
		_ = json.Unmarshal(body, &gotBody)
	})
	c := NewClient(ClientConfig{APIKey: "test", BaseURL: srv.URL})
	got, err := c.Documents.GetSettings()
	if err != nil || !got.EnforceCategories || gotMethod != "GET" {
		t.Fatalf("GetSettings = %+v, %v (%s)", got, err, gotMethod)
	}
	if _, err := c.Documents.UpdateSettings(true); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	if gotMethod != "PATCH" || gotBody["enforceCategories"] != true {
		t.Errorf("sent %s %v", gotMethod, gotBody)
	}
}
