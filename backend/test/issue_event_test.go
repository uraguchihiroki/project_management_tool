package test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/uraguchihiroki/project_management_tool/internal/model"
)

func TestIssueEvents_ImprintOnStatusChange(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "iev@example.com")
	projectID := createTestProject(t, ts, "IEV", "インプリントテスト", ownerID)
	ensureLinearIssueWorkflowTransitions(t, ts, projectID)
	statusID := getFirstStatusID(t, ts, projectID)
	statusIDs := getStatusIDs(t, ts, projectID)
	if len(statusIDs) < 2 {
		t.Fatal("need at least 2 statuses")
	}

	number := createTestIssue(t, ts, projectID, statusID, ownerID, "イベント用Issue")
	_, getResp := ts.req(t, "GET",
		fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(number)), nil)
	issueID := mustGetString(t, getResp, "data", "id")

	secondStatusID := statusIDs[1]
	ts.req(t, "PUT",
		fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(number)),
		map[string]string{"status_id": secondStatusID})

	st, evResp := ts.req(t, "GET", "/api/v1/issues/"+issueID+"/events", nil)
	assertStatus(t, st, http.StatusOK, "GET issue events")
	arr := mustGetArray(t, evResp, "data")
	if len(arr) != 1 {
		t.Fatalf("want 1 imprint, got %d", len(arr))
	}
	ev := arr[0].(map[string]interface{})
	if ev["event_type"] != model.EventIssueStatusChanged {
		t.Fatalf("event_type: got %v", ev["event_type"])
	}
	if ev["from_status_id"] != statusID {
		t.Fatalf("from_status_id: got %v want %s", ev["from_status_id"], statusID)
	}
	if ev["to_status_id"] != secondStatusID {
		t.Fatalf("to_status_id: got %v want %s", ev["to_status_id"], secondStatusID)
	}
}

func TestIssueEvents_ImprintOnAssigneeChange(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "iea@example.com")
	assigneeID := createTestUser(t, ts, "担当", "iea-assign@example.com")
	projectID := createTestProject(t, ts, "IEA", "担当インプリント", ownerID)
	statusID := getFirstStatusID(t, ts, projectID)
	number := createTestIssue(t, ts, projectID, statusID, ownerID, "担当変更")
	_, getResp := ts.req(t, "GET",
		fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(number)), nil)
	issueID := mustGetString(t, getResp, "data", "id")

	ts.req(t, "PUT",
		fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(number)),
		map[string]string{"assignee_id": assigneeID})

	st, evResp := ts.req(t, "GET", "/api/v1/issues/"+issueID+"/events", nil)
	assertStatus(t, st, http.StatusOK, "GET events after assignee")
	arr := mustGetArray(t, evResp, "data")
	if len(arr) != 1 {
		t.Fatalf("want 1 imprint, got %d", len(arr))
	}
	ev := arr[0].(map[string]interface{})
	if ev["event_type"] != model.EventIssueAssigneeChanged {
		t.Fatalf("event_type: got %v", ev["event_type"])
	}
}

func TestIssueEvents_ListByOrganization(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "ieo@example.com")
	projectID := createTestProject(t, ts, "IEO", "組織イベント", ownerID)
	ensureLinearIssueWorkflowTransitions(t, ts, projectID)
	statusID := getFirstStatusID(t, ts, projectID)
	statusIDs := getStatusIDs(t, ts, projectID)
	number := createTestIssue(t, ts, projectID, statusID, ownerID, "組織一覧用")
	ts.req(t, "PUT",
		fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(number)),
		map[string]string{"status_id": statusIDs[1]})

	path := fmt.Sprintf("/api/v1/organizations/%s/issue-events?event_type=%s", testOrgID, model.EventIssueStatusChanged)
	st, resp := ts.req(t, "GET", path, nil)
	assertStatus(t, st, http.StatusOK, "GET org issue-events")
	arr := mustGetArray(t, resp, "data")
	if len(arr) < 1 {
		t.Fatalf("expected at least 1 event, got %d", len(arr))
	}
}

func TestIssueEvents_TitleOnlyUpdate_NoImprint(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createTestUser(t, ts, "オーナー", "ient@example.com")
	projectID := createTestProject(t, ts, "IENT", "タイトルのみ", ownerID)
	statusID := getFirstStatusID(t, ts, projectID)
	number := createTestIssue(t, ts, projectID, statusID, ownerID, "タイトル")
	_, getResp := ts.req(t, "GET",
		fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(number)), nil)
	issueID := mustGetString(t, getResp, "data", "id")

	ts.req(t, "PUT",
		fmt.Sprintf("/api/v1/projects/%s/issues/%d", projectID, int(number)),
		map[string]string{"title": "だけ変える"})

	st, evResp := ts.req(t, "GET", "/api/v1/issues/"+issueID+"/events", nil)
	assertStatus(t, st, http.StatusOK, "GET events")
	arr := mustGetArray(t, evResp, "data")
	if len(arr) != 0 {
		t.Fatalf("want 0 imprints for title-only update, got %d", len(arr))
	}
}
