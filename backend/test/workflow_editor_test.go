package test

import (
	"net/http"
	"testing"
)

func TestWorkflowEditor_Put_CreatesStatusesAndTransitions(t *testing.T) {
	ts := newTestServer(t)
	wfID := createTestWorkflow(t, ts, "エディタ一括")

	c1 := "11111111-1111-1111-1111-111111111111"
	c2 := "22222222-2222-2222-2222-222222222222"

	body := map[string]interface{}{
		"name":        "エディタ一括",
		"description": "D1",
		"statuses": []interface{}{
			map[string]interface{}{"client_id": c1, "name": "Todo", "color": "#111111", "is_entry": true, "is_terminal": false},
			map[string]interface{}{"client_id": c2, "name": "Done", "color": "#222222", "is_entry": false, "is_terminal": true},
		},
		"transitions": []interface{}{
			map[string]interface{}{"from_ref": c1, "to_ref": c2},
		},
	}
	st, _ := ts.req(t, "PUT", "/api/v1/workflows/"+wfID+"/editor", body)
	assertStatus(t, st, http.StatusNoContent, "PUT editor")

	st2, resp := ts.req(t, "GET", "/api/v1/workflows/"+wfID+"/statuses", nil)
	assertStatus(t, st2, http.StatusOK, "GET statuses")
	arr := mustGetArray(t, resp, "data")
	if len(arr) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(arr))
	}

	st3, tresp := ts.req(t, "GET", "/api/v1/workflows/"+wfID+"/transitions", nil)
	assertStatus(t, st3, http.StatusOK, "GET transitions")
	tarr := mustGetArray(t, tresp, "data")
	if len(tarr) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(tarr))
	}

	st4, wresp := ts.req(t, "GET", "/api/v1/workflows/"+wfID, nil)
	assertStatus(t, st4, http.StatusOK, "GET workflow")
	assertField(t, mustGetString(t, wresp, "data", "description"), "D1", "description")
}

func TestWorkflowEditor_Put_NoTerminalIs400(t *testing.T) {
	ts := newTestServer(t)
	wfID := createTestWorkflow(t, ts, "終了なし")
	c1 := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	c2 := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	body := map[string]interface{}{
		"name":        "終了なし",
		"description": "",
		"statuses": []interface{}{
			map[string]interface{}{"client_id": c1, "name": "A", "color": "#111111", "is_entry": true, "is_terminal": false},
			map[string]interface{}{"client_id": c2, "name": "B", "color": "#222222", "is_entry": false, "is_terminal": false},
		},
		"transitions": []interface{}{},
	}
	st, _ := ts.req(t, "PUT", "/api/v1/workflows/"+wfID+"/editor", body)
	assertStatus(t, st, http.StatusBadRequest, "no is_terminal")
}

func TestWorkflowEditor_Put_EntryBothTrueIs400(t *testing.T) {
	ts := newTestServer(t)
	wfID := createTestWorkflow(t, ts, "開始二重")
	c1 := "cccccccc-cccc-cccc-cccc-cccccccccccc"
	c2 := "dddddddd-dddd-dddd-dddd-dddddddddddd"
	body := map[string]interface{}{
		"name":        "開始二重",
		"description": "",
		"statuses": []interface{}{
			map[string]interface{}{"client_id": c1, "name": "A", "color": "#111111", "is_entry": true, "is_terminal": false},
			map[string]interface{}{"client_id": c2, "name": "B", "color": "#222222", "is_entry": true, "is_terminal": true},
		},
		"transitions": []interface{}{},
	}
	st, _ := ts.req(t, "PUT", "/api/v1/workflows/"+wfID+"/editor", body)
	assertStatus(t, st, http.StatusBadRequest, "two entries / illegal row")
}

func TestWorkflowEditor_Put_UpdatesExistingByID(t *testing.T) {
	ts := newTestServer(t)
	wfID := createTestWorkflow(t, ts, "既存更新")

	c1 := "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	c2 := "ffffffff-ffff-ffff-ffff-ffffffffffff"
	put1 := map[string]interface{}{
		"name": "既存更新",
		"statuses": []interface{}{
			map[string]interface{}{"client_id": c1, "name": "S1", "color": "#111111", "is_entry": true, "is_terminal": false},
			map[string]interface{}{"client_id": c2, "name": "S2", "color": "#222222", "is_entry": false, "is_terminal": true},
		},
		"transitions": []interface{}{
			map[string]interface{}{"from_ref": c1, "to_ref": c2},
		},
	}
	st0, _ := ts.req(t, "PUT", "/api/v1/workflows/"+wfID+"/editor", put1)
	assertStatus(t, st0, http.StatusNoContent, "first put")

	st, resp := ts.req(t, "GET", "/api/v1/workflows/"+wfID+"/statuses", nil)
	assertStatus(t, st, http.StatusOK, "GET statuses")
	arr := mustGetArray(t, resp, "data")
	if len(arr) != 2 {
		t.Fatalf("want 2 statuses")
	}
	var id1, id2 string
	for _, row := range arr {
		m := row.(map[string]interface{})
		name := m["name"].(string)
		id := m["id"].(string)
		if name == "S1" {
			id1 = id
		}
		if name == "S2" {
			id2 = id
		}
	}
	if id1 == "" || id2 == "" {
		t.Fatalf("missing ids: %q %q", id1, id2)
	}

	put2 := map[string]interface{}{
		"name":        "改名WF",
		"description": "v2",
		"statuses": []interface{}{
			map[string]interface{}{"id": id1, "name": "S1b", "color": "#111111", "is_entry": true, "is_terminal": false},
			map[string]interface{}{"id": id2, "name": "S2b", "color": "#222222", "is_entry": false, "is_terminal": true},
		},
		"transitions": []interface{}{
			map[string]interface{}{"from_ref": id1, "to_ref": id2},
		},
	}
	st1, _ := ts.req(t, "PUT", "/api/v1/workflows/"+wfID+"/editor", put2)
	assertStatus(t, st1, http.StatusNoContent, "second put")

	st2, resp2 := ts.req(t, "GET", "/api/v1/workflows/"+wfID+"/statuses", nil)
	assertStatus(t, st2, http.StatusOK, "GET statuses 2")
	arr2 := mustGetArray(t, resp2, "data")
	found := 0
	for _, row := range arr2 {
		m := row.(map[string]interface{})
		if m["name"].(string) == "S1b" || m["name"].(string) == "S2b" {
			found++
		}
	}
	if found != 2 {
		t.Fatalf("expected renamed statuses, found %d", found)
	}
}
