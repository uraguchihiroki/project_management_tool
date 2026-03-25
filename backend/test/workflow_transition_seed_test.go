package test

import (
	"fmt"
	"net/http"
	"testing"
)

// ensureLinearIssueWorkflowTransitions はテスト専用。プロジェクトのデフォルト Issue ワークフローについて、
// statuses の display_order 順に隣接ステータス間の許可遷移だけを API で追加する（本番の自動シードは行わない前提のテスト向け）。
func ensureLinearIssueWorkflowTransitions(t *testing.T, ts *testServer, projectID string) {
	t.Helper()
	st, resp := ts.req(t, "GET", "/api/v1/projects/"+projectID, nil)
	assertStatus(t, st, http.StatusOK, "GET project for linear transitions")
	data := resp["data"].(map[string]interface{})
	rawWf := data["default_workflow_id"]
	if rawWf == nil {
		t.Fatal("project has no default_workflow_id")
	}
	wfID := fmt.Sprintf("%.0f", rawWf.(float64))
	statuses := data["statuses"].([]interface{})
	if len(statuses) < 2 {
		return
	}
	for i := 0; i < len(statuses)-1; i++ {
		from := statuses[i].(map[string]interface{})["id"].(string)
		to := statuses[i+1].(map[string]interface{})["id"].(string)
		st2, _ := ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/transitions", map[string]interface{}{
			"from_status_id": from,
			"to_status_id":   to,
		})
		assertStatus(t, st2, http.StatusCreated, fmt.Sprintf("POST transition %d->%d", i, i+1))
	}
}
