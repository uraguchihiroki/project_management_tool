# 部署管理

## パス

`/admin/departments`

## 概要

部署（開発部、営業部、委員会など）を管理。**部署スコープ付きの遷移権限**（例: 営業部の課長のみ Close）と組み合わせる場合は [transition-permissions.md](../transition-permissions.md) を参照。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

## UI 要素

- 見出し「部署管理」、説明
- 部署追加ボタン
- 部署追加/編集フォーム（部署名）
- 部署一覧テーブル（ドラッグで並び替え可能）

## 備考

- `GET /organizations/:orgId/departments` で部署一覧
- `POST /organizations/:orgId/departments`、`PUT`、`DELETE`
- `PUT /organizations/:orgId/departments/reorder` で並び替え
