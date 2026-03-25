package test

import (
	"fmt"
	"net/http"
	"testing"
)

// getFirstOrgIssueWorkflowStatusID は組織の「組織Issue」ワークフローの先頭ステータス ID を返す（組織横断 statuses 一覧の先頭は別 WF になりうるため）。
func getFirstOrgIssueWorkflowStatusID(t *testing.T, ts *testServer, orgID string) string {
	t.Helper()
	st, wfListResp := ts.req(t, "GET", "/api/v1/workflows?org_id="+orgID, nil)
	assertStatus(t, st, http.StatusOK, "list workflows for org issue")
	workflows := mustGetArray(t, wfListResp, "data")
	var issueWfID string
	for _, w := range workflows {
		m := w.(map[string]interface{})
		if m["name"].(string) == "組織Issue" {
			issueWfID = fmt.Sprintf("%.0f", m["id"].(float64))
			break
		}
	}
	if issueWfID == "" {
		t.Fatal("組織Issue workflow not found")
	}
	st2, statResp := ts.req(t, "GET", "/api/v1/workflows/"+issueWfID+"/statuses", nil)
	assertStatus(t, st2, http.StatusOK, "GET statuses for 組織Issue")
	arr := mustGetArray(t, statResp, "data")
	if len(arr) == 0 {
		t.Fatal("組織Issue workflow has no statuses")
	}
	return arr[0].(map[string]interface{})["id"].(string)
}

// createTestIssue はテスト用Issueを作成し、そのIssue番号を返します
func createTestIssue(t *testing.T, ts *testServer, projectID, statusID, reporterID, title string) float64 {
	t.Helper()
	status, resp := ts.req(t, "POST", fmt.Sprintf("/api/v1/projects/%s/issues", projectID), map[string]interface{}{
		"title":       title,
		"status_id":   statusID,
		"reporter_id": reporterID,
	})
	assertStatus(t, status, http.StatusCreated, fmt.Sprintf("createIssue(%s)", title))
	return mustGetFloat(t, resp, "data", "number")
}

func TestIssue_Create(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "io@example.com")
	projectID := createTestProject(t, ts, "ISS", "Issue作成テスト", ownerID)
	statusID := getFirstStatusID(t, ts, projectID)

	t.Run("正常系: Issueを作成できる", func(t *testing.T) {
		status, resp := ts.req(t, "POST",
			fmt.Sprintf("/api/v1/projects/%s/issues", projectID),
			map[string]interface{}{
				"title":       "テストIssue",
				"description": "詳細説明",
				"status_id":   statusID,
				"reporter_id": ownerID,
			})
		assertStatus(t, status, http.StatusCreated, "POST /projects/:id/issues")
		assertNotEmpty(t, mustGetString(t, resp, "data", "id"), "id")
		assertField(t, mustGetString(t, resp, "data", "title"), "テストIssue", "title")
		if mustGetFloat(t, resp, "data", "number") <= 0 {
			t.Error("issue number should be positive")
		}
	})

	t.Run("正常系: Issue番号は連番で採番される", func(t *testing.T) {
		num1 := createTestIssue(t, ts, projectID, statusID, ownerID, "Issue #1")
		num2 := createTestIssue(t, ts, projectID, statusID, ownerID, "Issue #2")
		if num2 != num1+1 {
			t.Errorf("issue numbers should be sequential: got %v then %v", num1, num2)
		}
	})

	t.Run("正常系: レスポンスにstatus情報が含まれる", func(t *testing.T) {
		_, resp := ts.req(t, "POST",
			fmt.Sprintf("/api/v1/projects/%s/issues", projectID),
			map[string]interface{}{
				"title":       "ステータス確認",
				"status_id":   statusID,
				"reporter_id": ownerID,
			})
		statusData, ok := resp["data"].(map[string]interface{})["status"].(map[string]interface{})
		if !ok || statusData["id"] == nil {
			t.Error("response should include status object")
		}
	})

	t.Run("正常系: レスポンスにreporter情報が含まれる", func(t *testing.T) {
		_, resp := ts.req(t, "POST",
			fmt.Sprintf("/api/v1/projects/%s/issues", projectID),
			map[string]interface{}{
				"title":       "レポーター確認",
				"status_id":   statusID,
				"reporter_id": ownerID,
			})
		reporter, ok := resp["data"].(map[string]interface{})["reporter"].(map[string]interface{})
		if !ok || reporter["id"] == nil {
			t.Error("response should include reporter object")
		}
	})
}

func TestIssue_List(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "il@example.com")
	projectID := createTestProject(t, ts, "ILS", "Issue一覧テスト", ownerID)
	statusID := getFirstStatusID(t, ts, projectID)

	t.Run("正常系: 空のリスト", func(t *testing.T) {
		status, resp := ts.req(t, "GET", fmt.Sprintf("/api/v1/projects/%s/issues", projectID), nil)
		assertStatus(t, status, http.StatusOK, "GET /projects/:id/issues (empty)")
		arr := mustGetArray(t, resp, "data")
		if len(arr) != 0 {
			t.Errorf("expected 0 issues, got %d", len(arr))
		}
	})

	t.Run("正常系: 作成後に一覧に含まれる", func(t *testing.T) {
		createTestIssue(t, ts, projectID, statusID, ownerID, "Issue A")
		createTestIssue(t, ts, projectID, statusID, ownerID, "Issue B")
		status, resp := ts.req(t, "GET", fmt.Sprintf("/api/v1/projects/%s/issues", projectID), nil)
		assertStatus(t, status, http.StatusOK, "GET /projects/:id/issues")
		arr := mustGetArray(t, resp, "data")
		if len(arr) != 2 {
			t.Errorf("expected 2 issues, got %d", len(arr))
		}
	})
}

func TestIssue_Get(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "ig@example.com")
	projectID := createTestProject(t, ts, "IGT", "Issue取得テスト", ownerID)
	statusID := getFirstStatusID(t, ts, projectID)

	t.Run("正常系: 番号で取得", func(t *testing.T) {
		number := createTestIssue(t, ts, projectID, statusID, ownerID, "取得テストIssue")
		status, resp := ts.req(t, "GET",
			fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(number)), nil)
		assertStatus(t, status, http.StatusOK, "GET /projects/:id/issues/:number")
		assertField(t, mustGetString(t, resp, "data", "title"), "取得テストIssue", "title")
	})

	t.Run("異常系: 存在しない番号は404", func(t *testing.T) {
		status, _ := ts.req(t, "GET",
			fmt.Sprintf("/api/v1/projects/%s/issues/9999", projectID), nil)
		assertStatus(t, status, http.StatusNotFound, "GET issue not found")
	})
}

func TestIssue_Update(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "iu@example.com")
	projectID := createTestProject(t, ts, "IU", "Issue更新テスト", ownerID)
	statusID := getFirstStatusID(t, ts, projectID)

	t.Run("正常系: タイトルを更新", func(t *testing.T) {
		number := createTestIssue(t, ts, projectID, statusID, ownerID, "更新前タイトル")
		status, resp := ts.req(t, "PUT",
			fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(number)),
			map[string]string{"title": "更新後タイトル"})
		assertStatus(t, status, http.StatusOK, "PUT /projects/:id/issues/:number")
		assertField(t, mustGetString(t, resp, "data", "title"), "更新後タイトル", "title")
	})

	t.Run("正常系: ステータスを変更して永続化される", func(t *testing.T) {
		number := createTestIssue(t, ts, projectID, statusID, ownerID, "ステータス変更テスト")

		// 全ステータス取得
		_, projResp := ts.req(t, "GET", "/api/v1/projects/"+projectID, nil)
		statuses := projResp["data"].(map[string]interface{})["statuses"].([]interface{})

		var secondStatusID string
		if len(statuses) >= 2 {
			secondStatusID = statuses[1].(map[string]interface{})["id"].(string)
		} else {
			secondStatusID = statusID
		}

		ts.req(t, "PUT",
			fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(number)),
			map[string]interface{}{"status_id": secondStatusID})

		// GETで確認
		_, getResp := ts.req(t, "GET",
			fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(number)), nil)
		gotStatusID := mustGetString(t, getResp, "data", "status", "id")
		if gotStatusID != secondStatusID {
			t.Errorf("status_id not persisted: got %q, want %q", gotStatusID, secondStatusID)
		}
	})

	t.Run("正常系: 更新後にGETで反映されている", func(t *testing.T) {
		number := createTestIssue(t, ts, projectID, statusID, ownerID, "反映確認前")
		ts.req(t, "PUT",
			fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(number)),
			map[string]string{"title": "反映確認後"})
		_, resp := ts.req(t, "GET",
			fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(number)), nil)
		assertField(t, mustGetString(t, resp, "data", "title"), "反映確認後", "title")
	})
}

func TestIssue_Delete(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "id@example.com")
	projectID := createTestProject(t, ts, "IDL", "Issue削除テスト", ownerID)
	statusID := getFirstStatusID(t, ts, projectID)

	t.Run("正常系: Issueを削除できる", func(t *testing.T) {
		number := createTestIssue(t, ts, projectID, statusID, ownerID, "削除対象Issue")
		status, _ := ts.req(t, "DELETE",
			fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(number)), nil)
		assertStatus(t, status, http.StatusOK, "DELETE /projects/:id/issues/:number")

		getStatus, _ := ts.req(t, "GET",
			fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(number)), nil)
		assertStatus(t, getStatus, http.StatusNotFound, "GET after DELETE issue")
	})
}

func TestIssue_OrgScoped(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "org@example.com")
	createTestProject(t, ts, "ORG", "組織Issueテスト", ownerID)
	orgStatusID := getFirstOrgIssueWorkflowStatusID(t, ts, testOrgID)

	t.Run("組織別にIssueを作成できる（project_idなし）", func(t *testing.T) {
		status, resp := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/issues", map[string]interface{}{
			"title":       "組織直下Issue",
			"status_id":   orgStatusID,
			"reporter_id": ownerID,
		})
		assertStatus(t, status, http.StatusCreated, "create org-scoped issue")
		assertField(t, mustGetString(t, resp, "data", "title"), "組織直下Issue", "title")
		if mustGetFloat(t, resp, "data", "number") <= 0 {
			t.Error("issue number should be positive")
		}
	})

	t.Run("組織別Issue一覧を取得できる", func(t *testing.T) {
		ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/issues", map[string]interface{}{
			"title": "一覧用Issue", "status_id": orgStatusID, "reporter_id": ownerID,
		})
		status, resp := ts.req(t, "GET", "/api/v1/organizations/"+testOrgID+"/issues", nil)
		assertStatus(t, status, http.StatusOK, "list org issues")
		arr := mustGetArray(t, resp, "data")
		if len(arr) < 2 {
			t.Errorf("expected at least 2 org issues, got %d", len(arr))
		}
	})

	t.Run("組織別Issueは番号で取得・更新・削除できる", func(t *testing.T) {
		_, createResp := ts.req(t, "POST", "/api/v1/organizations/"+testOrgID+"/issues", map[string]interface{}{
			"title": "CRUDテスト", "status_id": orgStatusID, "reporter_id": ownerID,
		})
		num := int(mustGetFloat(t, createResp, "data", "number"))

		// Get
		status, getResp := ts.req(t, "GET", fmt.Sprintf("/api/v1/organizations/%s/issues/%d", testOrgID, num), nil)
		assertStatus(t, status, http.StatusOK, "get org issue")
		assertField(t, mustGetString(t, getResp, "data", "title"), "CRUDテスト", "title")

		// Update
		ts.req(t, "PUT", fmt.Sprintf("/api/v1/organizations/%s/issues/%d", testOrgID, num),
			map[string]string{"title": "更新後"})
		_, upResp := ts.req(t, "GET", fmt.Sprintf("/api/v1/organizations/%s/issues/%d", testOrgID, num), nil)
		assertField(t, mustGetString(t, upResp, "data", "title"), "更新後", "title")

		// Delete
		delStatus, _ := ts.req(t, "DELETE", fmt.Sprintf("/api/v1/organizations/%s/issues/%d", testOrgID, num), nil)
		assertStatus(t, delStatus, http.StatusOK, "delete org issue")
		afterStatus, _ := ts.req(t, "GET", fmt.Sprintf("/api/v1/organizations/%s/issues/%d", testOrgID, num), nil)
		assertStatus(t, afterStatus, http.StatusNotFound, "get after delete")
	})
}
