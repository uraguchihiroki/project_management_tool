# ユーザー管理

## パス

`/admin/users`

## 概要

組織内のユーザー作成・更新・削除・グループの管理。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

## UI 要素

- 見出し「ユーザー管理」、説明（組織名のユーザーを管理）
- ユーザー登録フォーム（名前、メール）
- ユーザー一覧テーブル（ユーザー、グループ、管理者、操作）
- グループの編集（インライン）

## 備考

- `GET /admin/users?org_id=xxx` でユーザー一覧
- `POST /admin/users` でユーザー作成
- `PUT /admin/users/:id` で更新
- `DELETE /admin/users/:id` で組織から削除
