# ステータス管理

## パス

`/admin/statuses`

## 概要

組織のステータス（Issue用・プロジェクト用）を管理。カンバンの列として使用。**許可される遷移（形）**・**遷移アラート**は [transition-permissions.md](../transition-permissions.md) で決定（TBD。フォーム項目は合意後に追記）。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

## UI 要素

- 見出し「ステータス管理」、説明
- ステータス追加ボタン
- ステータス追加/編集フォーム（ステータス名、色、タイプ）
- ステータス一覧テーブル（sts_start, sts_goal は表示しない）

## 備考

- `GET /organizations/:orgId/statuses` でステータス一覧。`?exclude_system=1` で sts_start/sts_goal を除外可能
- `POST /organizations/:orgId/statuses`、`PUT /statuses/:id`、`DELETE /statuses/:id`
- システムステータス（sts_start, sts_goal）は編集・削除不可。画面には表示しない
