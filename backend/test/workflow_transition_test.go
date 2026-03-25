package test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestWorkflowTransitions_PUTAndValidation(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "wf-tr-owner@example.com")
	createTestProject(t, ts, "WFT", "遷移APIテスト", ownerID)
	wfID := createTestWorkflow(t, ts, "遷移検証WF")

	postStatus := func(name string) string {
		t.Helper()
		st, resp := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/statuses", map[string]interface{}{
			"name": name, "color": "#6B7280", "display_order": 0,
		})
		assertStatus(t, st, http.StatusCreated, "POST status "+name)
		return mustGetString(t, resp, "data", "id")
	}
	sA := postStatus("Alpha")
	sB := postStatus("Beta")
	sC := postStatus("Gamma")

	t.Run("POST from==to は400", func(t *testing.T) {
		st, _ := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/transitions", map[string]interface{}{
			"from_status_id": sA,
			"to_status_id":   sA,
		})
		assertStatus(t, st, http.StatusBadRequest, "POST same from/to")
	})

	t.Run("POST 重複ペアは400", func(t *testing.T) {
		st1, _ := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/transitions", map[string]interface{}{
			"from_status_id": sA,
			"to_status_id":   sB,
		})
		assertStatus(t, st1, http.StatusCreated, "POST first A->B")
		st2, _ := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/transitions", map[string]interface{}{
			"from_status_id": sA,
			"to_status_id":   sB,
		})
		assertStatus(t, st2, http.StatusBadRequest, "POST duplicate A->B")
	})

	t.Run("PUT 同一内容は200", func(t *testing.T) {
		stL, respL := ts.req(t, "GET", "/api/v1/workflows/"+wfID+"/transitions", nil)
		assertStatus(t, stL, http.StatusOK, "list transitions")
		arr := mustGetArray(t, respL, "data")
		if len(arr) == 0 {
			t.Fatal("expected at least one transition")
		}
		tid := fmt.Sprintf("%.0f", arr[0].(map[string]interface{})["id"].(float64))

		st, resp := ts.req(t, "PUT", "/api/v1/workflows/"+wfID+"/transitions/"+tid, map[string]interface{}{
			"from_status_id": sA,
			"to_status_id":   sB,
		})
		assertStatus(t, st, http.StatusOK, "PUT keep A->B")
		gotTo := mustGetString(t, resp, "data", "to_status_id")
		if gotTo != sB {
			t.Fatalf("to_status_id = %q, want %q", gotTo, sB)
		}
	})

	tRunCreateSecond := func() string {
		st, resp := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/transitions", map[string]interface{}{
			"from_status_id": sB,
			"to_status_id":   sC,
		})
		assertStatus(t, st, http.StatusCreated, "POST B->C")
		return fmt.Sprintf("%.0f", mustGetFloat(t, resp, "data", "id"))
	}
	tid2 := tRunCreateSecond()

	t.Run("PUT で他行と重複すると400", func(t *testing.T) {
		st, _ := ts.req(t, "PUT", "/api/v1/workflows/"+wfID+"/transitions/"+tid2, map[string]interface{}{
			"from_status_id": sA,
			"to_status_id":   sB,
		})
		assertStatus(t, st, http.StatusBadRequest, "PUT duplicate of existing A->B")
	})

	t.Run("PUT from==to は400", func(t *testing.T) {
		st, _ := ts.req(t, "PUT", "/api/v1/workflows/"+wfID+"/transitions/"+tid2, map[string]interface{}{
			"from_status_id": sC,
			"to_status_id":   sC,
		})
		assertStatus(t, st, http.StatusBadRequest, "PUT same from/to")
	})
}

func TestWorkflowStatus_DeleteBlockedByTransition(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "wf-st-tr@example.com")
	createTestProject(t, ts, "WFTR", "遷移参照削除テスト", ownerID)
	wfID := createTestWorkflow(t, ts, "遷移参照WF")

	postStatus := func(name string) string {
		t.Helper()
		st, resp := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/statuses", map[string]interface{}{
			"name": name, "color": "#6B7280", "display_order": 0,
		})
		assertStatus(t, st, http.StatusCreated, "POST status "+name)
		return mustGetString(t, resp, "data", "id")
	}
	sA := postStatus("T1")
	sB := postStatus("T2")
	postStatus("T3")

	ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/transitions", map[string]interface{}{
		"from_status_id": sA,
		"to_status_id":   sB,
	})

	t.Run("許可遷移で参照中のステータスは削除は400", func(t *testing.T) {
		st, _ := ts.req(t, "DELETE", "/api/v1/statuses/"+sA, nil)
		assertStatus(t, st, http.StatusBadRequest, "DELETE status referenced in transition")
	})
}

func TestWorkflowStatus_DeleteFloorTwo(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "wf-st-del@example.com")
	createTestProject(t, ts, "WFDL", "削除下限テスト", ownerID)
	wfID := createTestWorkflow(t, ts, "削除下限WF")

	postStatus := func(name string) string {
		t.Helper()
		st, resp := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/statuses", map[string]interface{}{
			"name": name, "color": "#112233", "display_order": 0,
		})
		assertStatus(t, st, http.StatusCreated, "POST status "+name)
		return mustGetString(t, resp, "data", "id")
	}
	s1 := postStatus("S1")
	postStatus("S2")

	t.Run("ステータスが2件のワークフローでは削除は400", func(t *testing.T) {
		st, _ := ts.req(t, "DELETE", "/api/v1/statuses/"+s1, nil)
		assertStatus(t, st, http.StatusBadRequest, "DELETE when workflow has only 2 statuses")
	})
}
