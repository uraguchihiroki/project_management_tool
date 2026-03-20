# ワークフロー詳細

## パス

`/admin/workflows/[id]`

## 概要

ワークフローの名前・説明の編集と、承認ステップの一覧・追加・並び替え。ステップをクリックで詳細編集へ。sts_start/sts_goal のステップは表示しない。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

- 遷移元: /admin/workflows
- 遷移先: ステップクリック → /admin/workflows/[id]/steps/[stepId]、戻る → /admin/workflows

## UI 要素

- パンくず: ワークフロー一覧 ← ワークフロー名
- ワークフロー名の編集（インライン）
- 承認ステップ一覧（ユーザーステップのみ表示。sts_start/sts_goal は非表示）
- ステップのドラッグ並び替え（即時 API 呼び出しはしない）
- 並び替え用の保存ボタン・キャンセルボタン（保存押下で承認後ステータスが確定）
- ステップ追加フォーム（ステータス、閾値、説明。承認後ステータス欄はなし。ステータス選択はユーザー作成ステータスのみ）
- 追加ボタン、キャンセルボタン
- 最後のユーザーステップ削除時: ダイアログでワークフロー削除を確認し、肯定ならワークフローごと削除

## 備考

- `GET /workflows/:id` でワークフロー・ステップ取得
- `PUT /workflows/:id` でワークフロー更新
- `POST /workflows/:id/steps` でステップ追加（初回は sts_start + user + sts_goal を自動作成）
- `PUT /workflows/:id/steps/reorder` でステップ並び替え・承認後ステータス確定
- `PUT /workflows/:id/steps/:stepId` でステップ更新
- `DELETE /workflows/:id/steps/:stepId` でステップ削除
