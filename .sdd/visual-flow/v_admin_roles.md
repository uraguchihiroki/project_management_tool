# 役職管理

## パス

`/admin/roles`

## 概要

役職とヒエラルキーレベルを管理。ステータス遷移の権限表現に使うかは [transition-permissions.md](../transition-permissions.md) で決定（TBD）。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

## UI 要素

- 見出し「役職管理」、説明
- 役職追加ボタン
- 役職追加/編集フォーム（役職名、レベル、説明）
- 役職一覧テーブル（ドラッグで並び替え可能）

## 備考

- `GET /roles?org_id=xxx` で役職一覧
- `POST /roles`、`PUT /roles/:id`、`DELETE /roles/:id`
- `PUT /roles/bulk/reorder` で並び替え
