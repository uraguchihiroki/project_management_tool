# ステップを編集

## パス

`/admin/workflows/[id]/steps/[stepId]`

## 概要

ワークフローの承認ステップの編集。ステータス、承認後ステータス、閾値、承認オブジェクトを設定。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

- 遷移元: /admin/workflows/[id]（ステップクリック）
- 遷移先: 戻る → /admin/workflows/[id]

## UI 要素

- パンくず: ワークフローに戻る ← ステップを編集: [ステータス名]
- ステータス選択（ドロップダウン、システムステータスは変更不可）
- 承認後ステータス選択
- 閾値入力
- ステップの説明
- 承認オブジェクトセクション（追加ボタン、一覧）
  - 承認オブジェクト: 種類（役職/ユーザー）、対象、点数、exclude_reporter、exclude_assignee
- 保存ボタン、キャンセルボタン

## 備考

- `GET /workflows/:id/steps/:stepId` でステップ取得
- `PUT /workflow-steps/:stepId` でステップ更新
- 承認オブジェクトは type: role/user、role_operator: eq/gte、points 等を設定
