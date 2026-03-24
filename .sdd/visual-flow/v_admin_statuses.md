# ステータス管理

## パス

`/admin/statuses`

## 概要

組織の「組織Issue」ワークフローに紐づく **Issue 用**ステータスを管理（カンバンの列）。プロジェクト進行は別途 **`GET /projects/:id/project-statuses`**。Issue の**許可遷移**・**遷移アラート**は [transition-permissions.md](../transition-permissions.md) を参照。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

## UI 要素

- 見出し「ステータス管理」、説明
- ステータス追加ボタン
- ステータス追加/編集フォーム（ステータス名、色、並び順）
- ステータス一覧テーブル（sts_start, sts_goal は表示しない）

## 備考

- `GET /organizations/:orgId/statuses` でステータス一覧。`?exclude_system=1` で sts_start/sts_goal を除外可能
- `POST /organizations/:orgId/statuses`、`PUT /statuses/:id`、`DELETE /statuses/:id`
- システムステータス（sts_start, sts_goal）は編集・削除不可。画面には表示しない
