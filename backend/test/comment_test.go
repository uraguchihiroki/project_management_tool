package test

import (
	"fmt"
	"net/http"
	"testing"
)

// createTestComment はテスト用コメントを作成しそのIDを返します
func createTestComment(t *testing.T, ts *testServer, issueID, authorID, body string) string {
	t.Helper()
	status, resp := ts.req(t, "POST",
		fmt.Sprintf("/api/v1/issues/%s/comments", issueID),
		map[string]string{"body": body, "author_id": authorID})
	assertStatus(t, status, http.StatusCreated, fmt.Sprintf("createComment(%s)", body))
	return mustGetString(t, resp, "data", "id")
}

// getIssueID はプロジェクト内のIssueのIDを取得します
func getIssueID(t *testing.T, ts *testServer, projectID string, issueNumber int) string {
	t.Helper()
	_, resp := ts.req(t, "GET",
		fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, issueNumber), nil)
	return mustGetString(t, resp, "data", "id")
}

func TestComment_Create(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "co@example.com")
	projectID := createTestProject(t, ts, "COM", "コメントテスト", ownerID)
	statusID := getFirstStatusID(t, ts, projectID)
	issueNumber := createTestIssue(t, ts, projectID, statusID, ownerID, "コメントテスト用Issue")
	issueID := getIssueID(t, ts, projectID, int(issueNumber))

	t.Run("正常系: コメントを作成できる", func(t *testing.T) {
		status, resp := ts.req(t, "POST",
			fmt.Sprintf("/api/v1/issues/%s/comments", issueID),
			map[string]string{
				"body":      "テストコメントです",
				"author_id": ownerID,
			})
		assertStatus(t, status, http.StatusCreated, "POST /issues/:id/comments")
		assertNotEmpty(t, mustGetString(t, resp, "data", "id"), "id")
		assertField(t, mustGetString(t, resp, "data", "body"), "テストコメントです", "body")
	})

	t.Run("正常系: レスポンスにauthor情報が含まれる", func(t *testing.T) {
		_, resp := ts.req(t, "POST",
			fmt.Sprintf("/api/v1/issues/%s/comments", issueID),
			map[string]string{"body": "著者確認", "author_id": ownerID})
		author, ok := resp["data"].(map[string]interface{})["author"].(map[string]interface{})
		if !ok || author["id"] == nil {
			t.Error("response should include author object with id")
		}
		if author["id"] != ownerID {
			t.Errorf("author.id = %v, want %v", author["id"], ownerID)
		}
	})
}

func TestComment_List(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "cl@example.com")
	projectID := createTestProject(t, ts, "CLS", "コメント一覧テスト", ownerID)
	statusID := getFirstStatusID(t, ts, projectID)
	issueNumber := createTestIssue(t, ts, projectID, statusID, ownerID, "一覧テスト用Issue")
	issueID := getIssueID(t, ts, projectID, int(issueNumber))

	t.Run("正常系: 空のリスト", func(t *testing.T) {
		status, resp := ts.req(t, "GET", fmt.Sprintf("/api/v1/issues/%s/comments", issueID), nil)
		assertStatus(t, status, http.StatusOK, "GET /issues/:id/comments (empty)")
		arr := mustGetArray(t, resp, "data")
		if len(arr) != 0 {
			t.Errorf("expected 0 comments, got %d", len(arr))
		}
	})

	t.Run("正常系: 作成後に一覧に含まれる", func(t *testing.T) {
		createTestComment(t, ts, issueID, ownerID, "コメント1")
		createTestComment(t, ts, issueID, ownerID, "コメント2")
		status, resp := ts.req(t, "GET", fmt.Sprintf("/api/v1/issues/%s/comments", issueID), nil)
		assertStatus(t, status, http.StatusOK, "GET /issues/:id/comments")
		arr := mustGetArray(t, resp, "data")
		if len(arr) != 2 {
			t.Errorf("expected 2 comments, got %d", len(arr))
		}
	})
}

func TestComment_Update(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "cu@example.com")
	projectID := createTestProject(t, ts, "CU", "コメント更新テスト", ownerID)
	statusID := getFirstStatusID(t, ts, projectID)
	issueNumber := createTestIssue(t, ts, projectID, statusID, ownerID, "更新テスト用Issue")
	issueID := getIssueID(t, ts, projectID, int(issueNumber))

	t.Run("正常系: コメントを更新できる", func(t *testing.T) {
		commentID := createTestComment(t, ts, issueID, ownerID, "変更前コメント")
		status, resp := ts.req(t, "PUT",
			fmt.Sprintf("/api/v1/issues/%s/comments/%s", issueID, commentID),
			map[string]string{"body": "変更後コメント"})
		assertStatus(t, status, http.StatusOK, "PUT /issues/:id/comments/:id")
		assertField(t, mustGetString(t, resp, "data", "body"), "変更後コメント", "body")
	})

	t.Run("正常系: 更新後にGETで反映されている", func(t *testing.T) {
		commentID := createTestComment(t, ts, issueID, ownerID, "反映確認前コメント")
		ts.req(t, "PUT",
			fmt.Sprintf("/api/v1/issues/%s/comments/%s", issueID, commentID),
			map[string]string{"body": "反映確認後コメント"})

		_, listResp := ts.req(t, "GET", fmt.Sprintf("/api/v1/issues/%s/comments", issueID), nil)
		arr := mustGetArray(t, listResp, "data")
		found := false
		for _, item := range arr {
			c := item.(map[string]interface{})
			if c["id"] == commentID {
				if c["body"] != "反映確認後コメント" {
					t.Errorf("updated body = %v, want '反映確認後コメント'", c["body"])
				}
				found = true
			}
		}
		if !found {
			t.Error("updated comment not found in list")
		}
	})

	t.Run("異常系: 存在しないコメントIDは404", func(t *testing.T) {
		status, _ := ts.req(t, "PUT",
			fmt.Sprintf("/api/v1/issues/%s/comments/00000000-0000-0000-0000-000000000000", issueID),
			map[string]string{"body": "不正リクエスト"})
		assertStatus(t, status, http.StatusNotFound, "PUT comment not found")
	})
}

func TestComment_Delete(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "cd@example.com")
	projectID := createTestProject(t, ts, "CDL", "コメント削除テスト", ownerID)
	statusID := getFirstStatusID(t, ts, projectID)
	issueNumber := createTestIssue(t, ts, projectID, statusID, ownerID, "削除テスト用Issue")
	issueID := getIssueID(t, ts, projectID, int(issueNumber))

	t.Run("正常系: コメントを削除できる", func(t *testing.T) {
		commentID := createTestComment(t, ts, issueID, ownerID, "削除対象コメント")
		status, _ := ts.req(t, "DELETE",
			fmt.Sprintf("/api/v1/issues/%s/comments/%s", issueID, commentID), nil)
		assertStatus(t, status, http.StatusOK, "DELETE /issues/:id/comments/:id")

		_, listResp := ts.req(t, "GET", fmt.Sprintf("/api/v1/issues/%s/comments", issueID), nil)
		arr := mustGetArray(t, listResp, "data")
		for _, item := range arr {
			c := item.(map[string]interface{})
			if c["id"] == commentID {
				t.Error("deleted comment still exists in list")
			}
		}
	})
}
