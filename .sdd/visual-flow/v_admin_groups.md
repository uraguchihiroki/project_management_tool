# グループ管理

## パス

`/admin/groups`（API・データモデルも `groups`）

## 概要

グループ（開発部、営業部、委員会など）を管理。**グループスコープ付きの想定アクター表現**（例: 営業部の課長のみ Close）と組み合わせる場合は [transition-permissions.md](../transition-permissions.md) を参照。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

## UI 要素

- 見出し「グループ管理」、説明
- グループ追加ボタン
- グループ追加/編集フォーム（グループ名）
- グループ一覧テーブル（ドラッグで並び替え可能）

## 備考

- `GET /organizations/:orgId/groups` で一覧
- `POST /organizations/:orgId/groups`、`PUT`、`DELETE`
- `PUT /organizations/:orgId/groups/reorder` で並び替え

