# ワークフロー詳細

## パス

`/admin/workflows/[id]`

## 概要

ワークフローの名前・説明の編集と、承認ステップの一覧・追加・並び替え。ステップをクリックで詳細編集へ。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

- 遷移元: /admin/workflows
- 遷移先: ステップクリック → /admin/workflows/[id]/steps/[stepId]、戻る → /admin/workflows

## UI 要素

- パンくず: ワークフロー一覧 ← ワークフロー名
- ワークフロー名の編集（インライン）
- 承認ステップ一覧（ステータス、閾値、承認オブジェクト数、承認後ステータス）
- ステップ追加フォーム（ステータス、承認後ステータス、閾値、説明）
- 追加ボタン、キャンセルボタン

## 備考

- `GET /workflows/:id` でワークフロー・ステップ取得
- `PUT /workflows/:id` でワークフロー更新
- `POST /workflows/:id/steps` でステップ追加
- `PUT /workflow-steps/:stepId` でステップ更新
- `DELETE /workflows/:id/steps/:stepId` でステップ削除
