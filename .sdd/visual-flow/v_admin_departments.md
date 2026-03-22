# グループ管理（画面表示名）

## パス

`/admin/departments`（API・データモデルは引き続き `departments`）

## 概要

グループ（開発部、営業部、委員会など）を管理。**部署スコープ付きの遷移権限**（例: 営業部の課長のみ Close）と組み合わせる場合は [transition-permissions.md](../transition-permissions.md) を参照。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

## UI 要素

- 見出し「グループ管理」、説明
- グループ追加ボタン
- グループ追加/編集フォーム（グループ名）
- グループ一覧テーブル（ドラッグで並び替え可能）

## 備考

- `GET /organizations/:orgId/departments` で一覧（リソース名は departments）
- `POST /organizations/:orgId/departments`、`PUT`、`DELETE`
- `PUT /organizations/:orgId/departments/reorder` で並び替え
