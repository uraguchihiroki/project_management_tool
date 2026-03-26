package service

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/pkg/keygen"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
	"gorm.io/gorm"
)

var editorColorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// WorkflowEditorStatusInput は PUT /workflows/:id/editor の statuses 要素。
type WorkflowEditorStatusInput struct {
	ID         *string `json:"id"`
	ClientID   *string `json:"client_id"`
	Name       string  `json:"name"`
	Color      string  `json:"color"`
	IsEntry    bool    `json:"is_entry"`
	IsTerminal bool    `json:"is_terminal"`
}

// WorkflowEditorTransitionInput は許可遷移 1 件。from_ref / to_ref は既存 status の UUID または新規行の client_id。
type WorkflowEditorTransitionInput struct {
	FromRef string `json:"from_ref"`
	ToRef   string `json:"to_ref"`
}

// WorkflowEditorSaveInput は PUT /workflows/:id/editor の本文。
type WorkflowEditorSaveInput struct {
	Name        string                          `json:"name"`
	Description string                          `json:"description"`
	Statuses    []WorkflowEditorStatusInput     `json:"statuses"`
	Transitions []WorkflowEditorTransitionInput `json:"transitions"`
}

// WorkflowEditorService はワークフロー編集の一括保存。
type WorkflowEditorService interface {
	Save(workflowID uint, in *WorkflowEditorSaveInput) error
}

type workflowEditorService struct {
	db *gorm.DB
}

func NewWorkflowEditorService(db *gorm.DB) WorkflowEditorService {
	return &workflowEditorService{db: db}
}

func normRef(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

func (s *workflowEditorService) Save(workflowID uint, in *WorkflowEditorSaveInput) error {
	if in == nil {
		return fmt.Errorf("リクエスト本文が空です")
	}
	if strings.TrimSpace(in.Name) == "" {
		return fmt.Errorf("名前は必須です")
	}
	if len(in.Name) > 200 {
		return fmt.Errorf("名前が長すぎます")
	}

	if len(in.Statuses) < 2 {
		return fmt.Errorf("ステータスは2つ以上必要です")
	}

	entryCount := 0
	terminalCount := 0
	seenClient := make(map[string]struct{})
	seenExistingID := make(map[uuid.UUID]struct{})
	var existingIDs []uuid.UUID

	for i, st := range in.Statuses {
		hasID := st.ID != nil && strings.TrimSpace(*st.ID) != ""
		hasClient := st.ClientID != nil && strings.TrimSpace(*st.ClientID) != ""
		if hasID == hasClient {
			return fmt.Errorf("statuses[%d]: id と client_id のどちらか一方だけ指定してください", i)
		}
		if hasID {
			u, err := uuid.Parse(strings.TrimSpace(*st.ID))
			if err != nil {
				return fmt.Errorf("statuses[%d]: id が UUID として不正です", i)
			}
			u = uuid.MustParse(normRef(u.String()))
			if _, dup := seenExistingID[u]; dup {
				return fmt.Errorf("statuses[%d]: id が重複しています", i)
			}
			seenExistingID[u] = struct{}{}
			existingIDs = append(existingIDs, u)
		} else {
			raw := normRef(*st.ClientID)
			if _, err := uuid.Parse(raw); err != nil {
				return fmt.Errorf("statuses[%d]: client_id は UUID 形式の文字列にしてください", i)
			}
			if _, dup := seenClient[raw]; dup {
				return fmt.Errorf("statuses[%d]: client_id が重複しています", i)
			}
			seenClient[raw] = struct{}{}
		}

		if strings.TrimSpace(st.Name) == "" {
			return fmt.Errorf("statuses[%d]: ステータス名は必須です", i)
		}
		if len(st.Name) > 50 {
			return fmt.Errorf("statuses[%d]: ステータス名は50文字以内で指定してください", i)
		}
		col := strings.TrimSpace(st.Color)
		if col == "" {
			col = "#6B7280"
		}
		if !editorColorRegex.MatchString(col) {
			return fmt.Errorf("statuses[%d]: 色は#RRGGBB形式で指定してください", i)
		}
		if st.IsEntry {
			entryCount++
		}
		if st.IsTerminal {
			terminalCount++
		}
		if st.IsEntry && st.IsTerminal {
			return fmt.Errorf("statuses[%d]: 開始と終了を同時に付けられません", i)
		}
	}

	if entryCount != 1 {
		return fmt.Errorf("開始ステータスはちょうど1件必要です")
	}
	if terminalCount < 1 {
		return fmt.Errorf("終了ステータスは1件以上必要です")
	}

	// 遷移の検証（参照解決前にペア検証）
	pairSeen := make(map[string]struct{})
	for i, tr := range in.Transitions {
		fr := normRef(tr.FromRef)
		to := normRef(tr.ToRef)
		if fr == "" || to == "" {
			return fmt.Errorf("transitions[%d]: from_ref と to_ref は必須です", i)
		}
		if fr == to {
			return fmt.Errorf("transitions[%d]: 出発と到着を同じステータスにはできません", i)
		}
		key := fr + "\x00" + to
		if _, dup := pairSeen[key]; dup {
			return fmt.Errorf("transitions[%d]: 同じ from→to の遷移が重複しています", i)
		}
		pairSeen[key] = struct{}{}
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		wfRepo := repository.NewWorkflowRepository(tx)
		statusRepo := repository.NewStatusRepository(tx)
		transRepo := repository.NewWorkflowTransitionRepository(tx)

		wf, err := wfRepo.FindByID(workflowID)
		if err != nil {
			return fmt.Errorf("ワークフローが見つかりません")
		}

		currentStatuses, err := statusRepo.FindByWorkflowID(workflowID)
		if err != nil {
			return err
		}
		currentByID := make(map[uuid.UUID]model.Status)
		for _, st := range currentStatuses {
			currentByID[st.ID] = st
		}

		for _, id := range existingIDs {
			if _, ok := currentByID[id]; !ok {
				return fmt.Errorf("指定された既存ステータス id がこのワークフローに存在しません")
			}
		}

		for ck := range seenClient {
			parsed := uuid.MustParse(ck)
			if _, ok := currentByID[parsed]; ok {
				return fmt.Errorf("client_id が既存ステータス id と重複しています")
			}
		}

		toDelete := make([]uuid.UUID, 0)
		for id := range currentByID {
			if _, keep := seenExistingID[id]; !keep {
				toDelete = append(toDelete, id)
			}
		}

		// 先に遷移を削除（参照解除）
		if err := tx.Where("workflow_id = ?", workflowID).Delete(&model.WorkflowTransition{}).Error; err != nil {
			return err
		}

		for _, delID := range toDelete {
			var n int64
			if err := tx.Model(&model.Issue{}).Where("status_id = ?", delID).Count(&n).Error; err != nil {
				return err
			}
			if n > 0 {
				return fmt.Errorf("このステータスは使用中のため削除できません")
			}
		}

		afterCount := len(in.Statuses)
		if afterCount < 2 {
			return fmt.Errorf("ステータスはワークフロー内で最低2つ必要です")
		}

		for _, delID := range toDelete {
			if err := tx.Delete(&model.Status{}, "id = ?", delID).Error; err != nil {
				return err
			}
		}

		clientToNewID := make(map[string]uuid.UUID)
		for _, st := range in.Statuses {
			if st.ClientID != nil {
				ck := normRef(*st.ClientID)
				id := uuid.New()
				clientToNewID[ck] = id
			}
		}

		resolve := func(ref string) (uuid.UUID, error) {
			r := normRef(ref)
			if r == "" {
				return uuid.Nil, fmt.Errorf("参照が空です")
			}
			if u, err := uuid.Parse(r); err == nil {
				if _, ok := seenExistingID[u]; ok {
					return u, nil
				}
				if mapped, ok := clientToNewID[r]; ok {
					return mapped, nil
				}
				return uuid.Nil, fmt.Errorf("遷移の参照が解決できません: %s", ref)
			}
			return uuid.Nil, fmt.Errorf("遷移の参照が UUID として不正です: %s", ref)
		}

		for i, tr := range in.Transitions {
			_, err := resolve(tr.FromRef)
			if err != nil {
				return fmt.Errorf("transitions[%d]: %w", i, err)
			}
			_, err = resolve(tr.ToRef)
			if err != nil {
				return fmt.Errorf("transitions[%d]: %w", i, err)
			}
		}

		// 開始から全ステータスへ到達可能か（許可遷移を順方向に辿る）
		idToName := make(map[uuid.UUID]string)
		var entryNode uuid.UUID
		var entryFound bool
		for _, st := range in.Statuses {
			var id uuid.UUID
			if st.ID != nil && strings.TrimSpace(*st.ID) != "" {
				id = uuid.MustParse(normRef(*st.ID))
			} else {
				id = clientToNewID[normRef(*st.ClientID)]
			}
			idToName[id] = strings.TrimSpace(st.Name)
			if st.IsEntry {
				entryNode = id
				entryFound = true
			}
		}
		if !entryFound {
			return fmt.Errorf("開始ステータスが特定できません")
		}
		adj := make(map[uuid.UUID][]uuid.UUID)
		for _, tr := range in.Transitions {
			from, err := resolve(tr.FromRef)
			if err != nil {
				return err
			}
			to, err := resolve(tr.ToRef)
			if err != nil {
				return err
			}
			adj[from] = append(adj[from], to)
		}
		reachable := make(map[uuid.UUID]bool)
		queue := []uuid.UUID{entryNode}
		reachable[entryNode] = true
		for head := 0; head < len(queue); head++ {
			cur := queue[head]
			for _, nx := range adj[cur] {
				if !reachable[nx] {
					reachable[nx] = true
					queue = append(queue, nx)
				}
			}
		}
		var unreachableNames []string
		for id, name := range idToName {
			if !reachable[id] {
				unreachableNames = append(unreachableNames, name)
			}
		}
		if len(unreachableNames) > 0 {
			sort.Strings(unreachableNames)
			var b strings.Builder
			for i, n := range unreachableNames {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString("「")
				b.WriteString(n)
				b.WriteString("」")
			}
			return fmt.Errorf("開始ステータスから許可遷移を辿ると到達できないステータスがあります: %s", b.String())
		}

		// ステータス作成・更新
		for i, st := range in.Statuses {
			order := i + 1
			col := strings.TrimSpace(st.Color)
			if col == "" {
				col = "#6B7280"
			}
			if st.ID != nil && strings.TrimSpace(*st.ID) != "" {
				uid := uuid.MustParse(normRef(*st.ID))
				row := model.Status{
					ID:           uid,
					Key:          currentByID[uid].Key,
					WorkflowID:   workflowID,
					Name:         strings.TrimSpace(st.Name),
					Color:        col,
					DisplayOrder: order,
					StatusKey:    currentByID[uid].StatusKey,
					IsEntry:      st.IsEntry,
					IsTerminal:   st.IsTerminal,
				}
				if row.Key == "" {
					row.Key = "sts-" + uid.String()
				}
				if err := tx.Save(&row).Error; err != nil {
					return err
				}
			} else {
				ck := normRef(*st.ClientID)
				newID := clientToNewID[ck]
				row := &model.Status{
					ID:           newID,
					Key:          "sts-" + newID.String(),
					WorkflowID:   workflowID,
					Name:         strings.TrimSpace(st.Name),
					Color:        col,
					DisplayOrder: order,
					IsEntry:      st.IsEntry,
					IsTerminal:   st.IsTerminal,
				}
				if err := tx.Create(row).Error; err != nil {
					return err
				}
			}
		}

		// is_entry 排他
		if err := tx.Model(&model.Status{}).
			Where("workflow_id = ? AND deleted_at IS NULL", workflowID).
			Update("is_entry", false).Error; err != nil {
			return err
		}
		var entryID uuid.UUID
		for i, st := range in.Statuses {
			if st.IsEntry {
				if st.ID != nil && strings.TrimSpace(*st.ID) != "" {
					entryID = uuid.MustParse(normRef(*st.ID))
				} else {
					entryID = clientToNewID[normRef(*st.ClientID)]
				}
				_ = i
				break
			}
		}
		if err := tx.Model(&model.Status{}).Where("id = ?", entryID).Update("is_entry", true).Error; err != nil {
			return err
		}

		now := time.Now()
		for i, tr := range in.Transitions {
			fromU, err := resolve(tr.FromRef)
			if err != nil {
				return err
			}
			toU, err := resolve(tr.ToRef)
			if err != nil {
				return err
			}
			row := &model.WorkflowTransition{
				Key:          keygen.UUIDKey(uuid.New()),
				WorkflowID:   workflowID,
				FromStatusID: fromU,
				ToStatusID:   toU,
				DisplayOrder: i + 1,
				CreatedAt:    now,
			}
			if err := transRepo.Create(row); err != nil {
				return err
			}
		}

		wf.Name = strings.TrimSpace(in.Name)
		wf.Description = in.Description
		if err := wfRepo.Update(wf); err != nil {
			return err
		}

		// (name, display_order) 最終チェック（同順位・同名）
		finalRows, err := statusRepo.FindByWorkflowID(workflowID)
		if err != nil {
			return err
		}
		seenPair := make(map[string]struct{})
		for _, r := range finalRows {
			k := fmt.Sprintf("%d\x00%s", r.DisplayOrder, r.Name)
			if _, ok := seenPair[k]; ok {
				return fmt.Errorf("同一ワークフローに同じ表示順・名前のステータスが既にあります")
			}
			seenPair[k] = struct{}{}
		}

		return nil
	})
}
