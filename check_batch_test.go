// =============================================================================
// (c) 2026 Caldera Technologies Ltd.
// Proprietary and confidential.
// Unauthorized copying or distribution is prohibited.
// =============================================================================
//
// CHANGELOG (2026-11-08, v1.11.0 pending):
// - Created. The check-batch route answers a bare array; CheckBatch decoded
//   it as {"results": [...]}, so Results was always empty. Pins the array
//   (and the older object) decoding into Results, with source and roleId.
// =============================================================================

package platformxe

import "testing"

func TestCheckBatchDecodesTheBareArray(t *testing.T) {
	c := newShapeClient(t, "POST", "/api/v1/permissions/check-batch",
		`[{"adminId":"a","path":"bookings","action":"LIST","allowed":true,"source":"role","roleId":"r_1"},`+
			`{"adminId":"a","path":"bookings","action":"DELETE","allowed":false,"source":"default"}]`)
	got, err := c.Permissions.CheckBatch([]map[string]interface{}{
		{"adminId": "a", "path": "bookings", "action": "LIST"},
		{"adminId": "a", "path": "bookings", "action": "DELETE"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Results) != 2 {
		t.Fatalf("want 2 results, got %d", len(got.Results))
	}
	if !got.Results[0].Allowed || got.Results[0].Source != "role" || got.Results[0].RoleID != "r_1" {
		t.Errorf("first result decoded wrongly: %+v", got.Results[0])
	}
	if got.Results[1].Allowed || got.Results[1].Action != "DELETE" {
		t.Errorf("second result decoded wrongly: %+v", got.Results[1])
	}
}

func TestCheckBatchStillReadsAnObjectWithResults(t *testing.T) {
	c := newShapeClient(t, "POST", "/api/v1/permissions/check-batch",
		`{"results":[{"adminId":"a","path":"bookings","action":"LIST","allowed":true}]}`)
	got, err := c.Permissions.CheckBatch([]map[string]interface{}{{"adminId": "a", "path": "bookings", "action": "LIST"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Results) != 1 || !got.Results[0].Allowed {
		t.Fatalf("object answer decoded wrongly: %+v", got.Results)
	}
}
