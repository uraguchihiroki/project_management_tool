package test

import (
	"fmt"
	"net/http"
	"testing"
)

// createIssueWithWorkflow はワークフロー付きIssueを作成しIssue IDを返します
func createIssueWithWorkflow(t *testing.T, ts *testServer, projectID, statusID, reporterID, wfID string) string {
	t.Helper()
	wfIDUint := uint(0)
	fmt.Sscanf(wfID, "%d", &wfIDUint)
	status, resp := ts.req(t, "POST", "/api/v1/projects/"+projectID+"/issues", map[string]interface{}{
		"title":       "承認テストIssue",
		"status_id":   statusID,
		"reporter_id": reporterID,
		"priority":    "medium",
		"workflow_id": wfIDUint,
	})
	assertStatus(t, status, http.StatusCreated, "createIssueWithWorkflow")
	return mustGetString(t, resp, "data", "id")
}

// setupApprovalFixture は承認テスト用の共通フィクスチャを作成します
// ownerID=課長(level5), directorID=部長(level7)
func setupApprovalFixture(t *testing.T, ts *testServer) (projectID, statusID, ownerID, directorID, wfID, issueID string) {
	t.Helper()
	ownerID = createTestUser(t, ts, "承認者", "approver@example.com")
	projectID = createTestProject(t, ts, "AP", "承認テスト", ownerID)
	statusIDs := getStatusIDs(t, ts, projectID)
	statusID = statusIDs[0]
	if len(statusIDs) < 3 {
		t.Fatal("project needs at least 3 statuses")
	}

	// 役職を作成してownerIDに割り当て（level 5 = 課長級）
	_, kachoResp := ts.req(t, "POST", "/api/v1/roles", map[string]interface{}{
		"name": "課長", "level": 5, "organization_id": testOrgID,
	})
	_, buchoResp := ts.req(t, "POST", "/api/v1/roles", map[string]interface{}{
		"name": "部長", "level": 7, "organization_id": testOrgID,
	})
	kachoID := uint(mustGetFloat(t, kachoResp, "data", "id"))
	buchoID := uint(mustGetFloat(t, buchoResp, "data", "id"))
	ts.req(t, "PUT", "/api/v1/users/"+ownerID+"/roles", map[string]interface{}{
		"role_ids": []float64{float64(kachoID)},
	})

	wfID = createTestWorkflow(t, ts, "テスト承認フロー")
	// Step 1: 課長が承認 → 進行中へ
	ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
		"status_id":     statusIDs[0],
		"next_status_id": statusIDs[1],
		"threshold":     10,
		"approval_objects": []map[string]interface{}{
			{"type": "role", "role_id": kachoID, "role_operator": "gte", "points": 10},
		},
	})
	// Step 2: 部長が承認 → 完了へ
	ts.req(t, "POST", "/api/v1/workflows/"+wfID+"/steps", map[string]interface{}{
		"status_id":     statusIDs[1],
		"next_status_id": statusIDs[2],
		"threshold":     10,
		"approval_objects": []map[string]interface{}{
			{"type": "role", "role_id": buchoID, "role_operator": "gte", "points": 10},
		},
	})

	directorID = createTestUser(t, ts, "部長", "director@example.com")
	ts.req(t, "PUT", "/api/v1/users/"+directorID+"/roles", map[string]interface{}{
		"role_ids": []float64{float64(buchoID)},
	})

	issueID = createIssueWithWorkflow(t, ts, projectID, statusID, ownerID, wfID)
	return
}

func TestApproval_AutoInitialize(t *testing.T) {
	ts := newTestServer(t)
	projectID, statusID, ownerID, directorID, wfID, issueID := setupApprovalFixture(t, ts)
	_ = projectID
	_ = statusID
	_ = ownerID
	_ = directorID
	_ = wfID

	t.Run("ワークフロー付きIssue作成で承認レコードが自動生成される", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/issues/"+issueID+"/approvals", nil)
		assertStatus(t, status, http.StatusOK, "get approvals")
		approvals := mustGetArray(t, resp, "data")
		if len(approvals) != 2 {
			t.Fatalf("expected 2 approval records, got %d", len(approvals))
		}
	})

	t.Run("全承認レコードの初期状態はpending", func(t *testing.T) {
		_, resp := ts.req(t, "GET", "/api/v1/issues/"+issueID+"/approvals", nil)
		approvals := mustGetArray(t, resp, "data")
		for _, a := range approvals {
			approval := a.(map[string]interface{})
			if approval["status"] != "pending" {
				t.Errorf("expected status=pending, got %v", approval["status"])
			}
		}
	})
}

func TestApproval_Approve(t *testing.T) {
	ts := newTestServer(t)
	projectID, statusID, ownerID, directorID, wfID, issueID := setupApprovalFixture(t, ts)
	_ = projectID
	_ = statusID
	_ = directorID
	_ = wfID

	_, approvalsResp := ts.req(t, "GET", "/api/v1/issues/"+issueID+"/approvals", nil)
	approvals := mustGetArray(t, approvalsResp, "data")

	// Step1（order=1）を取得
	var step1ID string
	for _, a := range approvals {
		approval := a.(map[string]interface{})
		step := approval["workflow_step"].(map[string]interface{})
		if step["order"].(float64) == 1 {
			step1ID = approval["id"].(string)
		}
	}

	t.Run("十分なLevelのユーザーはStep1を承認できる", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/approvals/"+step1ID+"/approve", map[string]interface{}{
			"approver_id": ownerID,
			"comment":     "問題なし",
		})
		assertStatus(t, status, http.StatusOK, "approve step1")
		assertField(t, mustGetString(t, resp, "data", "status"), "approved", "status")
	})

	t.Run("承認後にapprover情報が記録される", func(t *testing.T) {
		_, resp := ts.req(t, "GET", "/api/v1/issues/"+issueID+"/approvals", nil)
		approvals := mustGetArray(t, resp, "data")
		for _, a := range approvals {
			approval := a.(map[string]interface{})
			step := approval["workflow_step"].(map[string]interface{})
			if step["order"].(float64) == 1 {
				if approval["approver_id"] == nil {
					t.Errorf("approver_id should not be nil after approval")
				}
				if approval["acted_at"] == nil {
					t.Errorf("acted_at should not be nil after approval")
				}
			}
		}
	})
}

func TestApproval_LevelCheck(t *testing.T) {
	ts := newTestServer(t)
	projectID, statusID, _, _, wfID, issueID := setupApprovalFixture(t, ts)
	_ = projectID
	_ = statusID
	_ = wfID

	// Levelが低いユーザーを作成（level 3 = 主任）
	lowUserID := createTestUser(t, ts, "主任", "junior@example.com")
	_, lowRoleResp := ts.req(t, "POST", "/api/v1/roles", map[string]interface{}{
		"name": "主任", "level": 3, "organization_id": testOrgID,
	})
	lowRoleID := mustGetFloat(t, lowRoleResp, "data", "id")
	ts.req(t, "PUT", "/api/v1/users/"+lowUserID+"/roles", map[string]interface{}{
		"role_ids": []float64{lowRoleID},
	})

	_, approvalsResp := ts.req(t, "GET", "/api/v1/issues/"+issueID+"/approvals", nil)
	approvals := mustGetArray(t, approvalsResp, "data")
	var step1ID string
	for _, a := range approvals {
		approval := a.(map[string]interface{})
		step := approval["workflow_step"].(map[string]interface{})
		if step["order"].(float64) == 1 {
			step1ID = approval["id"].(string)
		}
	}

	t.Run("Levelが不足しているユーザーは承認できない", func(t *testing.T) {
		status, _ := ts.req(t, "POST", "/api/v1/approvals/"+step1ID+"/approve", map[string]interface{}{
			"approver_id": lowUserID,
			"comment":     "",
		})
		assertStatus(t, status, http.StatusBadRequest, "low level cannot approve")
	})
}

func TestApproval_OrderCheck(t *testing.T) {
	ts := newTestServer(t)
	projectID, statusID, ownerID, directorID, wfID, issueID := setupApprovalFixture(t, ts)
	_ = projectID
	_ = statusID
	_ = wfID

	_, approvalsResp := ts.req(t, "GET", "/api/v1/issues/"+issueID+"/approvals", nil)
	approvals := mustGetArray(t, approvalsResp, "data")
	var step1ID, step2ID string
	for _, a := range approvals {
		approval := a.(map[string]interface{})
		step := approval["workflow_step"].(map[string]interface{})
		if step["order"].(float64) == 1 {
			step1ID = approval["id"].(string)
		} else {
			step2ID = approval["id"].(string)
		}
	}

	t.Run("Step1が未承認のままStep2は承認できない", func(t *testing.T) {
		status, _ := ts.req(t, "POST", "/api/v1/approvals/"+step2ID+"/approve", map[string]interface{}{
			"approver_id": directorID,
			"comment":     "",
		})
		assertStatus(t, status, http.StatusBadRequest, "cannot skip step")
	})

	t.Run("Step1承認後はStep2を承認できる", func(t *testing.T) {
		// まずStep1を承認
		ts.req(t, "POST", "/api/v1/approvals/"+step1ID+"/approve", map[string]interface{}{
			"approver_id": ownerID, "comment": "",
		})
		// Step2を承認
		status, resp := ts.req(t, "POST", "/api/v1/approvals/"+step2ID+"/approve", map[string]interface{}{
			"approver_id": directorID, "comment": "承認します",
		})
		assertStatus(t, status, http.StatusOK, "approve step2 after step1")
		assertField(t, mustGetString(t, resp, "data", "status"), "approved", "status")
	})
}

func TestApproval_Reject(t *testing.T) {
	ts := newTestServer(t)
	projectID, statusID, ownerID, directorID, wfID, issueID := setupApprovalFixture(t, ts)
	_ = projectID
	_ = statusID
	_ = directorID
	_ = wfID

	_, approvalsResp := ts.req(t, "GET", "/api/v1/issues/"+issueID+"/approvals", nil)
	approvals := mustGetArray(t, approvalsResp, "data")
	var step1ID string
	for _, a := range approvals {
		approval := a.(map[string]interface{})
		step := approval["workflow_step"].(map[string]interface{})
		if step["order"].(float64) == 1 {
			step1ID = approval["id"].(string)
		}
	}

	t.Run("承認を却下できる", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/approvals/"+step1ID+"/reject", map[string]interface{}{
			"approver_id": ownerID,
			"comment":     "内容を修正してください",
		})
		assertStatus(t, status, http.StatusOK, "reject step1")
		assertField(t, mustGetString(t, resp, "data", "status"), "rejected", "status")
	})

	t.Run("却下済みステップは再承認できない", func(t *testing.T) {
		status, _ := ts.req(t, "POST", "/api/v1/approvals/"+step1ID+"/approve", map[string]interface{}{
			"approver_id": ownerID,
		})
		assertStatus(t, status, http.StatusBadRequest, "cannot approve rejected step")
	})
}

func TestApproval_NoWorkflow(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "owner@example.com")
	projectID := createTestProject(t, ts, "NW", "ワークフローなし", ownerID)
	statusID := getFirstStatusID(t, ts, projectID)

	// ワークフローなしでIssue作成
	_, issueResp := ts.req(t, "POST", "/api/v1/projects/"+projectID+"/issues", map[string]interface{}{
		"title":       "通常Issue",
		"status_id":   statusID,
		"reporter_id": ownerID,
	})
	issueID := mustGetString(t, issueResp, "data", "id")

	t.Run("ワークフローなしIssueは承認レコードが生成されない", func(t *testing.T) {
		status, resp := ts.req(t, "GET", "/api/v1/issues/"+issueID+"/approvals", nil)
		assertStatus(t, status, http.StatusOK, "get approvals no workflow")
		approvals := mustGetArray(t, resp, "data")
		if len(approvals) != 0 {
			t.Errorf("expected 0 approvals, got %d", len(approvals))
		}
	})
}
